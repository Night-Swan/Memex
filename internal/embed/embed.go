package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func GenerateEmbedding(text string) ([]float32, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		return nil, fmt.Errorf("OLLAMA_HOST environment variable not set")
	}

	reqBody := embedRequest{
		Model: "nomic-embed-text",
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	resp, err := http.Post(host+"/api/embed", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not reach ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0], nil
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
}

func GenerateText(prompt string) (string, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		return "", fmt.Errorf("OLLAMA_HOST environment variable not set")
	}

	reqBody := generateRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("could not marshal request: %w", err)
	}

	resp, err := http.Post(host+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not reach ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("could not decode response: %w", err)
	}

	return result.Response, nil
}