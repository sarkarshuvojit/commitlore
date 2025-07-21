package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sarkarshuvojit/commitlore/internal/core"
)

// Compile-time interface compliance check
var _ LLMProvider = (*ClaudeClient)(nil)

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(apiKey string) *ClaudeClient {
	logger := core.GetLogger()
	logger.Info("Creating new Claude API client", "provider", "claude-api", "model", "claude-3-5-sonnet-20241022")
	
	return &ClaudeClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.anthropic.com/v1",
		model:   "claude-3-5-sonnet-20241022",
	}
}

// GenerateContent generates content using Claude API with a simple prompt
func (c *ClaudeClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with Claude API", "provider", "claude-api", "prompt_length", len(prompt))
	
	return c.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

// GenerateContentWithSystemPrompt generates content using Claude API with system and user prompts
func (c *ClaudeClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with system prompt", 
		"provider", "claude-api",
		"system_prompt_length", len(systemPrompt),
		"user_prompt_length", len(userPrompt),
		"model", c.model)
	
	start := time.Now()
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
		logger.Error("Failed to marshal Claude API request", "provider", "claude-api", "error", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	logger.Debug("Marshaled request", "request_size", len(reqBody))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Failed to create HTTP request", "provider", "claude-api", "error", err, "url", c.baseURL+"/messages")
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	logger.Debug("Created HTTP request", "url", c.baseURL+"/messages", "method", "POST")

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	logger.Debug("Making HTTP request to Claude API")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Failed to make HTTP request to Claude API", "provider", "claude-api", "error", err, "duration", time.Since(start))
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	logger.Debug("Received HTTP response", "status_code", resp.StatusCode, "duration", time.Since(start))

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "provider", "claude-api", "error", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	logger.Debug("Read response body", "response_size", len(respBody))

	if resp.StatusCode != http.StatusOK {
		logger.Error("Claude API request failed", 
			"provider", "claude-api",
			"status_code", resp.StatusCode, 
			"response_body", string(respBody),
			"duration", time.Since(start))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		logger.Error("Failed to unmarshal Claude API response", "provider", "claude-api", "error", err, "response_body", string(respBody))
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	logger.Debug("Unmarshaled response", "content_blocks", len(claudeResp.Content))

	if len(claudeResp.Content) == 0 {
		logger.Error("No content in Claude API response", "provider", "claude-api", "response_id", claudeResp.ID)
		return "", fmt.Errorf("no content in response")
	}

	responseText := claudeResp.Content[0].Text
	logger.Info("Successfully generated content with Claude API", 
		"provider", "claude-api",
		"response_length", len(responseText),
		"duration", time.Since(start),
		"response_id", claudeResp.ID,
		"input_tokens", claudeResp.Usage.InputTokens,
		"output_tokens", claudeResp.Usage.OutputTokens)
	
	return responseText, nil
}