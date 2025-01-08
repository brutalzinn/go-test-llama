package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func connectDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createTable(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS bem_estar (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario TEXT NOT NULL,
		sentimento TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sensorial (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario TEXT NOT NULL,
		descricao TEXT NOT NULL,
		gatilho TEXT NOT NULL,
		nivel INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func armazenarDadoSensorial(db *sql.DB, usuario string, descricao string, gatilho string, nivel string) error {
	stmt, err := db.Prepare("INSERT INTO sensorial(usuario, descricao, gatilho, nivel) VALUES(?, ?, ? ,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(usuario, descricao, gatilho, nivel)
	return err
}

func armazenarBemEstar(db *sql.DB, usuario string, sentimento string) error {
	stmt, err := db.Prepare("INSERT INTO bem_estar(usuario, sentimento) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(usuario, sentimento)
	return err
}

var upgrader = websocket.Upgrader{}

var ollamaURL = "http://192.168.0.175:11434/api/generate" ///meu homeserver rodando ollama com o modelo llama3.2 4g
var model = "llama3.2"

var prompt = `Você é o Espectrum, um assistente de IA útil e amigável, projetado para apoiar pessoas com autismo e TPS.

Seus objetivos são:
* ** Você fornece meios de pessoas autistas relatarem sua rotina e ajuda pessoas autistas como uma ferramenta de apoio para que possam manter seu bem estar
* **Fornecer suporte empático e paciente:** Compreender e responder a sinais emocionais, oferecendo respostas calmas e de apoio.
* **Oferecer informações claras e concisas:** Fornecer informações e explicações úteis de forma fácil de entender.
** Você deve sempre focar em captar informações de monitamento para que os usuários tenham seus dados registrados e possam rever depois.
* **Monitorar o bem-estar do usuário:** 
    * Quando o usuário expressar algo sobre como se sente sentimentalmente (ex: "Estou feliz", "Me sinto triste", "Estou ansioso"), identifique o sentimento.
    * Escreva a seguinte mensagem: '#register("nomedousuario","sentimento")' quando identificar um sentimento.
    * Substitua "nomedousuario" pelo nome do usuário que expressou o sentimento e "sentimento" pelo sentimento identificado.
* **Monitorar a regulação sensorial 
	* Quando o usuário expressar que está em desregulação sensorial, pergunte se ele consegue descrever o que está sentindo e pergunte se houve algum estímulo específico(som, luz, toque e vibração) que causou isso.
	* Peça para que esse usuário escolha um número de 1 - 10 para representar o impacto dessa desregulação. Caso não informe um nível, o nível sempre será 0
	* Quando usuário explicar o gatilho e descrever o que aconteceu, escreva a seguinte mensagem '#registersensorial("nomedousuario", "descricao", "gatilho", "nivel")'.
	* Substitua "nomedousuario" pelo nome do usuário, "descrição" pelos detalhes fornecidos sobre a desregulação e "gatilho" sendo o que causou essa desregulação e "nível" sendo um número de 0 a 10 que representa o impacto dessa desregulação
Lembre-se de:

* **Ser respeitoso e compreensivo:** Sempre aborde as conversas com empatia e respeito.
* **Evitar estereótipos:** Trate cada indivíduo como único.
* **Manter a confidencialidade:** Proteger a privacidade de qualquer informação pessoal compartilhada.

Como posso te ajudar hoje?`

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Erro ao fazer upgrade para WebSocket:", err)
		return
	}
	defer conn.Close()
	db, err := connectDB("bem_estar.db")
	if err != nil {
		logrus.Error("Erro ao conectar ao banco de dados:", err)
		return
	}
	defer db.Close()
	for {
		var msg struct {
			User    string `json:"user"`
			Message string `json:"message"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			logrus.Error("Erro ao ler mensagem JSON:", err)
			break
		}
		requestPayload := OllamaRequest{
			Model:  model,
			Prompt: prompt + " O usuário se chama: " + msg.User + " e enviou: " + msg.Message,
			Stream: true,
		}

		jsonPayload, err := json.Marshal(requestPayload)
		if err != nil {
			logrus.Error(err)
		}
		resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			logrus.Error(err)
		}
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}

			var ollamaResponse OllamaResponse
			err = json.Unmarshal(line, &ollamaResponse)
			if err != nil {
				logrus.Error("Erro ao decodificar JSON:", err)
				continue
			}
			fullResponse.WriteString(ollamaResponse.Response)
			handleBemEstar(db, fullResponse.String(), func() {
				fullResponse.Reset()
			})
			handleSensorial(db, fullResponse.String(), func() {
				fullResponse.Reset()
			})
			// ollamaResponse.Response = re.ReplaceAllString(ollamaResponse.Response, "")
			err = conn.WriteMessage(websocket.TextMessage, []byte(ollamaResponse.Response))
			if err != nil {
				logrus.Error("Erro ao escrever mensagem:", err)
				break
			}
		}
	}
}

func handleBemEstar(db *sql.DB, response string, callback func()) {
	re := regexp.MustCompile(`#register\("([^"]*)", "([^"]*)"\)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) == 3 {
		usuario := matches[1]
		sentimento := matches[2]
		err := armazenarBemEstar(db, usuario, sentimento)
		if err != nil {
			logrus.Error("Erro ao armazenar dados de bem-estar:", err)
		}
		callback()
	}
}

func handleSensorial(db *sql.DB, response string, callback func()) {
	re := regexp.MustCompile(`#registersensorial\("([^"]*)", "([^"]*)"\)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) == 5 {
		usuario := matches[1]
		descricao := matches[2]
		gatilho := matches[3]
		nivel := matches[4]
		err := armazenarDadoSensorial(db, usuario, descricao, gatilho, nivel)
		if err != nil {
			logrus.Error("Erro ao armazenar dados de bem-estar:", err)
		}
		// fullResponse.Reset()
		callback()
	}
}

func main() {
	// Conectar ao banco de dados
	db, err := connectDB("bem_estar.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Criar a tabela
	err = createTable(db)
	if err != nil {
		panic(err)
	}

	// Rota para o WebSocket
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	fmt.Println("Servidor WebSocket iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
