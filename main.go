package main

import (
	"fmt"
	"go-test-llama/database"
	"go-test-llama/instruction"
	"go-test-llama/ollama"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var ollamaApi = ollama.OllamaApi{}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Erro ao fazer upgrade para WebSocket:", err)
		return
	}
	defer conn.Close()
	db, err := database.ConnectDB("bem_estar.db")
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
		message := " O usuário se chama: " + msg.User + " e enviou: " + msg.Message
		logrus.Info("Sending message to OLLAMA:", message)
		response := ""
		err = ollamaApi.SendOllamaChatStream(message, "llama3.2", func(chunk string) {
			logrus.Info("Response OLLAMA:", chunk)
			err = conn.WriteJSON(map[string]string{"response": chunk})
			if err != nil {
				logrus.Error("error when send websocket:", err)
			}
			response += chunk
		})
		instructions := instruction.HandleCallback(response)
		for _, ins := range instructions {
			switch ins.Name {
			case "health":
				logrus.Info("health instruction received with: " + ins.Args[0])
				err = database.InsertHealthData(db, ins.Args[0], ins.Args[1])
				if err != nil {
					logrus.Error("Error when insert health data:", err)
				}
				break
			case "sensorial":
				logrus.Info("sensorial instruction received with: " + ins.Args[0] + " " + ins.Args[1] + " " + ins.Args[2])
				err = database.InsertSensorialData(db, ins.Args[0], ins.Args[1], ins.Args[2], ins.Args[3])
				if err != nil {
					logrus.Error("Error when insert sensorial data:", err)
				}
				break
			default:
				logrus.Error("Instruction invalid:", ins.Name)
				continue
			}
		}
		if err != nil {
			logrus.Error("Erro ao enviar mensagem para OLLAMA:", err)
			err = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			if err != nil {
				logrus.Error("error when send websocket:", err)
			}
			break
		}
	}
}

func main() {

	if os.Getenv("OLLAMA_URL") == "" {
		logrus.Info("OLLAMA_URL não está definido. Usando o padrão http://127.0.0.1:11434")
		os.Setenv("OLLAMA_URL", "http://127.0.0.1:11434")
	}

	db, err := database.ConnectDB("bem_estar.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = database.CreateTable(db)
	if err != nil {
		panic(err)
	}

	ollamaApi = *ollama.New()
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	fmt.Println("Servidor WebSocket iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
