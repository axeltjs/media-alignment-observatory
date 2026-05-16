package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

type ClaudeClient struct {
	APIKey     string
	Model      string
	httpClient *http.Client
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	return &ClaudeClient{
		APIKey:     apiKey,
		Model:      model,
		httpClient: &http.Client{},
	}
}

func (c *ClaudeClient) IsConfigured() bool {
	return c.APIKey != ""
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *ClaudeClient) Complete(ctx context.Context, system, userPrompt string, maxTokens int) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("claude client not configured")
	}

	payload := claudeRequest{
		Model:     c.Model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []claudeMessage{{Role: "user", Content: userPrompt}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result claudeResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic api: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}
	return result.Content[0].Text, nil
}
