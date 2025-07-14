package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// ExtractTopics analyzes changesets and extracts relevant topics for content creation
func ExtractTopics(provider LLMProvider, changesets []Changeset) ([]string, error) {
	if len(changesets) == 0 {
		return []string{}, nil
	}
	
	// Build changeset string from the provided changesets
	changesetString := buildChangesetString(changesets)
	
	systemPrompt := `You are an expert at analyzing git commit changes and extracting meaningful topics for content creation. Your task is to analyze the provided changesets and extract 3-5 key topics that would be interesting for technical blog posts, social media content, or developer stories.

Guidelines:
- Focus on technical achievements, patterns, and insights
- Consider the broader impact and learnings from the changes
- Prioritize topics that would resonate with other developers
- Make topics specific enough to be actionable but broad enough to be interesting
- Return only the topic titles, one per line
- No numbering, bullets, or additional formatting`

	userPrompt := fmt.Sprintf("Analyze the following git changesets and extract 3-5 key topics for content creation:\n\n%s", changesetString)
	
	ctx := context.Background()
	response, err := provider.GenerateContentWithSystemPrompt(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract topics from LLM: %w", err)
	}
	
	// Parse the response to extract individual topics
	topics := parseTopicsFromResponse(response)
	
	return topics, nil
}

// buildChangesetString converts changesets into a formatted string for LLM analysis
func buildChangesetString(changesets []Changeset) string {
	var buffer bytes.Buffer
	
	for i, changeset := range changesets {
		buffer.WriteString(fmt.Sprintf("=== Commit %d ===\n", i+1))
		buffer.WriteString(fmt.Sprintf("Hash: %s\n", changeset.CommitHash))
		buffer.WriteString(fmt.Sprintf("Author: %s\n", changeset.Author))
		buffer.WriteString(fmt.Sprintf("Date: %s\n", changeset.Date.Format("2006-01-02 15:04:05")))
		buffer.WriteString(fmt.Sprintf("Subject: %s\n", changeset.Subject))
		
		if changeset.Body != "" {
			buffer.WriteString(fmt.Sprintf("Body: %s\n", changeset.Body))
		}
		
		buffer.WriteString(fmt.Sprintf("Files: %v\n", changeset.Files))
		
		if changeset.Diff != "" {
			// Truncate diff if too long to keep within token limits
			diff := changeset.Diff
			if len(diff) > 2000 {
				diff = diff[:2000] + "\n... (truncated)"
			}
			buffer.WriteString(fmt.Sprintf("Diff:\n%s\n", diff))
		}
		
		buffer.WriteString("\n")
	}
	
	return buffer.String()
}

// parseTopicsFromResponse extracts individual topics from the LLM response
func parseTopicsFromResponse(response string) []string {
	rawLines := strings.Split(response, "\n")
	
	var topics []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Remove common prefixes like numbers, bullets, dashes
		line = strings.TrimLeft(line, "0123456789.-• ")
		line = strings.TrimSpace(line)
		
		if line != "" && len(line) > 10 { // Filter out very short lines
			topics = append(topics, line)
		}
	}
	
	return topics
}