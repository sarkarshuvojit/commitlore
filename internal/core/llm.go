package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeResponse struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Role    string          `json:"role"`
	Content []ClaudeContent `json:"content"`
	Model   string          `json:"model"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	model      string
}

// Compile-time interface compliance check
var _ LLMProvider = (*ClaudeClient)(nil)

func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.anthropic.com/v1",
		model:   "claude-3-5-sonnet-20241022",
	}
}

func (c *ClaudeClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return c.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

func (c *ClaudeClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 4000,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	if systemPrompt != "" {
		req.System = systemPrompt
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return claudeResp.Content[0].Text, nil
}