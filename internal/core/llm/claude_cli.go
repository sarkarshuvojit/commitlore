package llm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sarkarshuvojit/commitlore/internal/core"
)

// Compile-time interface compliance check
var _ LLMProvider = (*ClaudeCLIClient)(nil)

// IsClaudeCLIAvailable checks if Claude CLI is installed and available
func IsClaudeCLIAvailable() bool {
	logger := core.GetLogger()
	logger.Debug("Checking if Claude CLI is available")
	
	execPath, err := exec.LookPath("claude")
	available := err == nil
	
	logger.Info("Claude CLI availability check", "available", available, "path", execPath)
	return available
}

// NewClaudeCLIClient creates a new Claude CLI client
func NewClaudeCLIClient() (*ClaudeCLIClient, error) {
	logger := core.GetLogger()
	logger.Info("Creating new Claude CLI client")
	
	execPath, err := exec.LookPath("claude")
	if err != nil {
		logger.Error("Claude CLI not found in PATH", "error", err)
		return nil, fmt.Errorf("claude CLI not found in PATH: %w", err)
	}
	
	logger.Info("Claude CLI client created successfully", "exec_path", execPath)
	return &ClaudeCLIClient{
		execPath: execPath,
	}, nil
}

// GenerateContent generates content using Claude CLI with a simple prompt
func (c *ClaudeCLIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with Claude CLI", "prompt_length", len(prompt))
	
	return c.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

// GenerateContentWithSystemPrompt generates content using Claude CLI with system and user prompts
func (c *ClaudeCLIClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger := core.GetLogger()
	logger.Info("Generating content with Claude CLI and system prompt", 
		"system_prompt_length", len(systemPrompt),
		"user_prompt_length", len(userPrompt),
		"exec_path", c.execPath)
	
	start := time.Now()
	var cmd *exec.Cmd
	
	if systemPrompt != "" {
		// Combine system prompt and user prompt
		fullPrompt := fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
		logger.Debug("Using system prompt with Claude CLI", "full_prompt_length", len(fullPrompt))
		cmd = exec.CommandContext(ctx, c.execPath, "--print", "--output-format", "text", fullPrompt)
	} else {
		logger.Debug("Using user prompt only with Claude CLI")
		cmd = exec.CommandContext(ctx, c.execPath, "--print", "--output-format", "text", userPrompt)
	}
	
	logger.Debug("Prepared Claude CLI command", "args", cmd.Args)
	
	// Set environment variables to ensure proper execution
	cmd.Env = os.Environ()
	logger.Debug("Set environment variables for Claude CLI")
	
	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	logger.Debug("Starting Claude CLI execution")
	
	err := cmd.Run()
	if err != nil {
		logger.Error("Claude CLI execution failed", 
			"error", err,
			"stderr", stderr.String(),
			"duration", time.Since(start),
			"command", cmd.Args)
		return "", fmt.Errorf("claude CLI execution failed: %w (stderr: %s)", err, stderr.String())
	}
	
	logger.Debug("Claude CLI execution completed", 
		"duration", time.Since(start),
		"stdout_length", stdout.Len(),
		"stderr_length", stderr.Len())
	
	response := strings.TrimSpace(stdout.String())
	if response == "" {
		logger.Error("Claude CLI returned empty response", 
			"stderr", stderr.String(),
			"duration", time.Since(start))
		return "", fmt.Errorf("claude CLI returned empty response (stderr: %s)", stderr.String())
	}
	
	logger.Info("Successfully generated content with Claude CLI", 
		"response_length", len(response),
		"duration", time.Since(start))
	
	return response, nil
}