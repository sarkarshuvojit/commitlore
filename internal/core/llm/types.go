package llm

import (
	"net/http"
	"time"
)

// ClaudeMessage represents a message in the Claude API format
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the request payload for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

// ClaudeContent represents the content structure in Claude responses
type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ClaudeResponse represents the response from Claude API
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

// ClaudeClient represents the Claude API client
type ClaudeClient struct {
	apiKey     string
	httpClient interface{ Do(req *http.Request) (*http.Response, error) }
	baseURL    string
	model      string
}

// ClaudeCLIClient represents the Claude CLI client
type ClaudeCLIClient struct {
	execPath string
}

// Changeset represents a git changeset for analysis
type Changeset struct {
	CommitHash string
	Author     string
	Date       time.Time
	Subject    string
	Body       string
	Files      []string
	Diff       string
}