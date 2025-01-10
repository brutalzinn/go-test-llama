package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

type OllamaApi struct {
	URL string
}

func New() *OllamaApi {
	return &OllamaApi{
		URL: os.Getenv("OLLAMA_URL"),
	}
}

func (api OllamaApi) SendOllamaChatStream(msg string, model string, responseCallback func(string)) error {

	requestPayload := OllamaRequest{
		Model:  model,
		Prompt: prompt + msg,
		Stream: true,
	}

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		logrus.Error(err)
		return err
	}
	apiurl := fmt.Sprintf("%s/%s", api.URL, "api/generate")
	resp, err := http.Post(apiurl, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		logrus.Error(err)
		return err
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
	}
	responseCallback(fullResponse.String())
	return nil
}
