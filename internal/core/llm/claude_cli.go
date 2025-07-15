package llm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Compile-time interface compliance check
var _ LLMProvider = (*ClaudeCLIClient)(nil)

// IsClaudeCLIAvailable checks if Claude CLI is installed and available
func IsClaudeCLIAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// NewClaudeCLIClient creates a new Claude CLI client
func NewClaudeCLIClient() (*ClaudeCLIClient, error) {
	execPath, err := exec.LookPath("claude")
	if err != nil {
		return nil, fmt.Errorf("claude CLI not found in PATH: %w", err)
	}
	
	return &ClaudeCLIClient{
		execPath: execPath,
	}, nil
}

// GenerateContent generates content using Claude CLI with a simple prompt
func (c *ClaudeCLIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return c.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

// GenerateContentWithSystemPrompt generates content using Claude CLI with system and user prompts
func (c *ClaudeCLIClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	var cmd *exec.Cmd
	
	if systemPrompt != "" {
		// Combine system prompt and user prompt
		fullPrompt := fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
		cmd = exec.CommandContext(ctx, c.execPath, "--print", "--output-format", "text", fullPrompt)
	} else {
		cmd = exec.CommandContext(ctx, c.execPath, "--print", "--output-format", "text", userPrompt)
	}
	
	// Set environment variables to ensure proper execution
	cmd.Env = os.Environ()
	
	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("claude CLI execution failed: %w (stderr: %s)", err, stderr.String())
	}
	
	response := strings.TrimSpace(stdout.String())
	if response == "" {
		return "", fmt.Errorf("claude CLI returned empty response (stderr: %s)", stderr.String())
	}
	
	return response, nil
}