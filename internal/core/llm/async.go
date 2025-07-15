package llm

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// LLMResponse represents the response from an LLM call
type LLMResponse struct {
	Content string
	Error   error
}

// LLMResponseMsg is a Bubble Tea message for LLM responses
type LLMResponseMsg struct {
	Content string
	Error   string
}

// AsyncLLMWrapper wraps LLM calls to run them asynchronously with channels
type AsyncLLMWrapper struct {
	provider LLMProvider
	timeout  time.Duration
}

// NewAsyncLLMWrapper creates a new async LLM wrapper
func NewAsyncLLMWrapper(provider LLMProvider, timeout time.Duration) *AsyncLLMWrapper {
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}
	return &AsyncLLMWrapper{
		provider: provider,
		timeout:  timeout,
	}
}

// GenerateContentAsync runs GenerateContent in a goroutine and sends response to channel
func (a *AsyncLLMWrapper) GenerateContentAsync(ctx context.Context, prompt string, responseChan chan<- LLMResponse) {
	go func() {
		defer close(responseChan)
		
		// Create context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, a.timeout)
		defer cancel()
		
		content, err := a.provider.GenerateContent(timeoutCtx, prompt)
		
		select {
		case responseChan <- LLMResponse{Content: content, Error: err}:
		case <-timeoutCtx.Done():
			// Context cancelled or timed out
			if timeoutCtx.Err() == context.DeadlineExceeded {
				responseChan <- LLMResponse{Content: "", Error: context.DeadlineExceeded}
			}
		}
	}()
}

// GenerateContentWithSystemPromptAsync runs GenerateContentWithSystemPrompt in a goroutine
func (a *AsyncLLMWrapper) GenerateContentWithSystemPromptAsync(ctx context.Context, systemPrompt, userPrompt string, responseChan chan<- LLMResponse) {
	go func() {
		defer close(responseChan)
		
		// Create context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, a.timeout)
		defer cancel()
		
		content, err := a.provider.GenerateContentWithSystemPrompt(timeoutCtx, systemPrompt, userPrompt)
		
		select {
		case responseChan <- LLMResponse{Content: content, Error: err}:
		case <-timeoutCtx.Done():
			// Context cancelled or timed out
			if timeoutCtx.Err() == context.DeadlineExceeded {
				responseChan <- LLMResponse{Content: "", Error: context.DeadlineExceeded}
			}
		}
	}()
}

// WaitForLLMResponse creates a tea.Cmd that waits for LLM response on a channel
func WaitForLLMResponse(responseChan <-chan LLMResponse) tea.Cmd {
	return func() tea.Msg {
		response := <-responseChan
		
		errorMsg := ""
		if response.Error != nil {
			errorMsg = response.Error.Error()
		}
		
		return LLMResponseMsg{
			Content: response.Content,
			Error:   errorMsg,
		}
	}
}

// CreateLLMResponseChannel creates a buffered channel for LLM responses
func CreateLLMResponseChannel() chan LLMResponse {
	return make(chan LLMResponse, 1)
}