package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingClient calls a configurable embedding API endpoint.
// Supports Ollama (/api/embeddings) and OpenAI-compatible APIs.
type EmbeddingClient struct {
	BaseURL    string
	Model      string
	APIKey     string
	httpClient *http.Client
}

func NewEmbeddingClient(baseURL, model, apiKey string) *EmbeddingClient {
	return &EmbeddingClient{
		BaseURL:    baseURL,
		Model:      model,
		APIKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *EmbeddingClient) IsConfigured() bool {
	return c.BaseURL != "" && c.Model != ""
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float64, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("embedding client not configured")
	}

	payload := ollamaEmbedRequest{Model: c.Model, Prompt: text}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API returned %d: %s", resp.StatusCode, string(raw))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}
