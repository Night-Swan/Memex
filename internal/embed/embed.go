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
	reqBody := embedRequest{
		Model: "nomic-embed-text",
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", err)
	}

	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		return nil, fmt.Errorf("OLLAMA_HOST environment variable not set")
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