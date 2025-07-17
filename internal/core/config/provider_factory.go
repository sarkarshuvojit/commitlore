package config

import (
	"fmt"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
)

// ProviderFactory creates LLM provider instances based on configuration
type ProviderFactory struct {
	config *ProviderConfig
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(config *ProviderConfig) *ProviderFactory {
	return &ProviderFactory{
		config: config,
	}
}

// CreateActiveProvider creates an instance of the currently active provider
func (f *ProviderFactory) CreateActiveProvider() (llm.LLMProvider, string, error) {
	logger := core.GetLogger()
	
	activeProvider := GetProviderByID(f.config, f.config.ActiveProviderID)
	if activeProvider == nil {
		logger.Error("Active provider not found", "provider_id", f.config.ActiveProviderID)
		return nil, "", fmt.Errorf("active provider '%s' not found", f.config.ActiveProviderID)
	}

	if !activeProvider.Enabled {
		logger.Error("Active provider is disabled", "provider_id", activeProvider.ID)
		return nil, "", fmt.Errorf("active provider '%s' is disabled", activeProvider.ID)
	}

	if !CheckProviderAvailability(activeProvider) {
		logger.Error("Active provider is not available", "provider_id", activeProvider.ID)
		return nil, "", fmt.Errorf("active provider '%s' is not available", activeProvider.ID)
	}

	provider, err := f.createProvider(activeProvider)
	if err != nil {
		logger.Error("Failed to create active provider", "provider_id", activeProvider.ID, "error", err)
		return nil, "", fmt.Errorf("failed to create provider '%s': %w", activeProvider.ID, err)
	}

	logger.Info("Successfully created active provider", "provider_id", activeProvider.ID, "provider_name", activeProvider.Name)
	return provider, activeProvider.Name, nil
}

// CreateProvider creates an instance of a specific provider by ID
func (f *ProviderFactory) CreateProvider(providerID string) (llm.LLMProvider, string, error) {
	logger := core.GetLogger()
	
	provider := GetProviderByID(f.config, providerID)
	if provider == nil {
		logger.Error("Provider not found", "provider_id", providerID)
		return nil, "", fmt.Errorf("provider '%s' not found", providerID)
	}

	if !provider.Enabled {
		logger.Error("Provider is disabled", "provider_id", provider.ID)
		return nil, "", fmt.Errorf("provider '%s' is disabled", provider.ID)
	}

	if !CheckProviderAvailability(provider) {
		logger.Error("Provider is not available", "provider_id", provider.ID)
		return nil, "", fmt.Errorf("provider '%s' is not available", provider.ID)
	}

	llmProvider, err := f.createProvider(provider)
	if err != nil {
		logger.Error("Failed to create provider", "provider_id", provider.ID, "error", err)
		return nil, "", fmt.Errorf("failed to create provider '%s': %w", provider.ID, err)
	}

	logger.Info("Successfully created provider", "provider_id", provider.ID, "provider_name", provider.Name)
	return llmProvider, provider.Name, nil
}

// createProvider creates the actual provider instance based on its configuration
func (f *ProviderFactory) createProvider(provider *Provider) (llm.LLMProvider, error) {
	logger := core.GetLogger()
	logger.Debug("Creating provider instance", "provider_id", provider.ID, "type", provider.Type)

	switch provider.Type {
	case APIProviderType:
		return f.createAPIProvider(provider)
	case CLIProviderType:
		return f.createCLIProvider(provider)
	case LocalProviderType:
		return f.createLocalProvider(provider)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

// createAPIProvider creates an API-based provider
func (f *ProviderFactory) createAPIProvider(provider *Provider) (llm.LLMProvider, error) {
	logger := core.GetLogger()
	logger.Debug("Creating API provider", "provider_id", provider.ID)

	switch provider.ID {
	case "claude-api":
		envVar, exists := provider.Config["api_key"]
		if !exists {
			return nil, fmt.Errorf("API key environment variable not configured")
		}

		apiKey := os.Getenv(envVar)
		if apiKey == "" {
			return nil, fmt.Errorf("API key not found in environment variable %s", envVar)
		}

		logger.Info("Creating Claude API client", "model", provider.Config["model"])
		return llm.NewClaudeClient(apiKey), nil

	case "openai-api":
		// TODO: Implement OpenAI API provider
		return nil, fmt.Errorf("OpenAI API provider not yet implemented")

	case "gemini-api":
		// TODO: Implement Gemini API provider
		return nil, fmt.Errorf("Gemini API provider not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported API provider: %s", provider.ID)
	}
}

// createCLIProvider creates a CLI-based provider
func (f *ProviderFactory) createCLIProvider(provider *Provider) (llm.LLMProvider, error) {
	logger := core.GetLogger()
	logger.Debug("Creating CLI provider", "provider_id", provider.ID)

	switch provider.ID {
	case "claude-cli":
		logger.Info("Creating Claude CLI client")
		return llm.NewClaudeCLIClient()

	default:
		return nil, fmt.Errorf("unsupported CLI provider: %s", provider.ID)
	}
}

// createLocalProvider creates a local model provider
func (f *ProviderFactory) createLocalProvider(provider *Provider) (llm.LLMProvider, error) {
	logger := core.GetLogger()
	logger.Debug("Creating local provider", "provider_id", provider.ID)

	switch provider.ID {
	case "ollama":
		// TODO: Implement Ollama provider
		return nil, fmt.Errorf("Ollama provider not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported local provider: %s", provider.ID)
	}
}

// GetAvailableProviderNames returns a list of available provider names for display
func (f *ProviderFactory) GetAvailableProviderNames() []string {
	availableProviders := GetAvailableProviders(f.config)
	names := make([]string, len(availableProviders))
	for i, provider := range availableProviders {
		names[i] = provider.Name
	}
	return names
}

// SetActiveProvider sets the active provider and saves the configuration
func (f *ProviderFactory) SetActiveProvider(providerID string) error {
	logger := core.GetLogger()
	logger.Debug("Setting active provider", "provider_id", providerID)

	provider := GetProviderByID(f.config, providerID)
	if provider == nil {
		return fmt.Errorf("provider '%s' not found", providerID)
	}

	if !provider.Enabled {
		return fmt.Errorf("provider '%s' is disabled", providerID)
	}

	if !CheckProviderAvailability(provider) {
		return fmt.Errorf("provider '%s' is not available", providerID)
	}

	f.config.ActiveProviderID = providerID
	
	if err := SaveProviderConfig(f.config); err != nil {
		logger.Error("Failed to save provider config after setting active provider", "error", err)
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Info("Successfully set active provider", "provider_id", providerID, "provider_name", provider.Name)
	return nil
}