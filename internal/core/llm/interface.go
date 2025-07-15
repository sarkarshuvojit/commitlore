package llm

import (
	"context"
)

// LLMProvider defines the interface for all LLM implementations
type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}