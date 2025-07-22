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
var _ LLMProvider = (*OpenAIClient)(nil)

// NewOpenAIClient creates a new OpenAI API client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	logger := core.GetLogger()
	logger.Info("Creating new OpenAI API client", "provider", "openai-api", "model", "gpt-3.5-turbo")
	
	return &OpenAIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-3.5-turbo",
	}
}

// GenerateContent generates content using OpenAI API with a simple prompt
func (c *OpenAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with OpenAI API", "provider", "openai-api", "prompt_length", len(prompt))
	
	return c.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

// GenerateContentWithSystemPrompt generates content using OpenAI API with system and user prompts
func (c *OpenAIClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with system prompt", 
		"provider", "openai-api",
		"system_prompt_length", len(systemPrompt),
		"user_prompt_length", len(userPrompt),
		"model", c.model)
	
	start := time.Now()
	
	// Build messages array
	messages := []OpenAIMessage{}
	
	if systemPrompt != "" {
		messages = append(messages, OpenAIMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: userPrompt,
	})

	req := OpenAIRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   4000,
		Temperature: 0.7,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		logger.Error("Failed to marshal OpenAI API request", "provider", "openai-api", "error", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	logger.Debug("Marshaled request", "request_size", len(reqBody))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Failed to create HTTP request", "provider", "openai-api", "error", err, "url", c.baseURL+"/chat/completions")
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	logger.Debug("Created HTTP request", "url", c.baseURL+"/chat/completions", "method", "POST")

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	logger.Debug("Making HTTP request to OpenAI API")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Failed to make HTTP request to OpenAI API", "provider", "openai-api", "error", err, "duration", time.Since(start))
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	logger.Debug("Received HTTP response", "status_code", resp.StatusCode, "duration", time.Since(start))

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "provider", "openai-api", "error", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	logger.Debug("Read response body", "response_size", len(respBody))

	if resp.StatusCode != http.StatusOK {
		logger.Error("OpenAI API request failed", 
			"provider", "openai-api",
			"status_code", resp.StatusCode, 
			"response_body", string(respBody),
			"duration", time.Since(start))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		logger.Error("Failed to unmarshal OpenAI API response", "provider", "openai-api", "error", err, "response_body", string(respBody))
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	logger.Debug("Unmarshaled response", "choices", len(openaiResp.Choices))

	if len(openaiResp.Choices) == 0 {
		logger.Error("No choices in OpenAI API response", "provider", "openai-api", "response_id", openaiResp.ID)
		return "", fmt.Errorf("no choices in response")
	}

	responseText := openaiResp.Choices[0].Message.Content
	logger.Info("Successfully generated content with OpenAI API", 
		"provider", "openai-api",
		"response_length", len(responseText),
		"duration", time.Since(start),
		"response_id", openaiResp.ID,
		"prompt_tokens", openaiResp.Usage.PromptTokens,
		"completion_tokens", openaiResp.Usage.CompletionTokens,
		"total_tokens", openaiResp.Usage.TotalTokens)
	
	return responseText, nil
}