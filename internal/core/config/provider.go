package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sarkarshuvojit/commitlore/internal/core"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	APIProviderType   ProviderType = "api"
	CLIProviderType   ProviderType = "cli"
	LocalProviderType ProviderType = "local"
)

// Provider represents an LLM provider configuration
type Provider struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        ProviderType      `json:"type"`
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
	Available   bool              `json:"available"` // Runtime availability check
	Config      map[string]string `json:"config"`    // Provider-specific config
}

// ProviderConfig manages the configuration of all LLM providers
type ProviderConfig struct {
	Providers        []Provider `json:"providers"`
	ActiveProviderID string     `json:"active_provider_id"`
}

// DefaultProviderConfig creates a default provider configuration
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Providers: []Provider{
			{
				ID:          "claude-api",
				Name:        "Claude API",
				Type:        APIProviderType,
				Description: "Anthropic Claude via API (requires ANTHROPIC_API_KEY)",
				Enabled:     true,
				Available:   false, // Will be checked at runtime
				Config: map[string]string{
					"model":   "claude-3-5-sonnet-20241022",
					"api_key": "ANTHROPIC_API_KEY", // Environment variable name
				},
			},
			{
				ID:          "claude-cli",
				Name:        "Claude CLI",
				Type:        CLIProviderType,
				Description: "Anthropic Claude via CLI tool",
				Enabled:     true,
				Available:   false, // Will be checked at runtime
				Config:      map[string]string{},
			},
			{
				ID:          "openai-api",
				Name:        "OpenAI API",
				Type:        APIProviderType,
				Description: "OpenAI GPT models via API (requires OPENAI_API_KEY)",
				Enabled:     false, // Disabled until implemented
				Available:   false,
				Config: map[string]string{
					"model":   "gpt-4",
					"api_key": "OPENAI_API_KEY",
				},
			},
			{
				ID:          "gemini-api",
				Name:        "Gemini API",
				Type:        APIProviderType,
				Description: "Google Gemini via API (requires GEMINI_API_KEY)",
				Enabled:     false, // Disabled until implemented
				Available:   false,
				Config: map[string]string{
					"model":   "gemini-pro",
					"api_key": "GEMINI_API_KEY",
				},
			},
			{
				ID:          "ollama",
				Name:        "Ollama",
				Type:        LocalProviderType,
				Description: "Local models via Ollama",
				Enabled:     false, // Disabled until implemented
				Available:   false,
				Config: map[string]string{
					"endpoint": "http://localhost:11434",
					"model":    "llama2",
				},
			},
		},
		ActiveProviderID: "claude-cli", // Default to Claude CLI
	}
}

// LoadProviderConfig loads provider configuration from user's home directory
func LoadProviderConfig() (*ProviderConfig, error) {
	logger := core.GetLogger()
	configPath, err := getConfigPath()
	if err != nil {
		logger.Error("Failed to get config path", "error", err)
		return nil, err
	}

	logger.Debug("Loading provider config", "path", configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("Provider config file does not exist, creating default", "path", configPath)
		config := DefaultProviderConfig()
		if saveErr := SaveProviderConfig(config); saveErr != nil {
			logger.Error("Failed to save default config", "error", saveErr)
			return nil, saveErr
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("Failed to read provider config file", "error", err, "path", configPath)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProviderConfig
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Error("Failed to unmarshal provider config", "error", err)
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	logger.Info("Successfully loaded provider config", "providers_count", len(config.Providers))
	return &config, nil
}

// SaveProviderConfig saves provider configuration to user's home directory
func SaveProviderConfig(config *ProviderConfig) error {
	logger := core.GetLogger()
	configPath, err := getConfigPath()
	if err != nil {
		logger.Error("Failed to get config path", "error", err)
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create config directory", "error", err, "dir", configDir)
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal provider config", "error", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		logger.Error("Failed to write provider config file", "error", err, "path", configPath)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Info("Successfully saved provider config", "path", configPath)
	return nil
}

// CheckProviderAvailability checks if a provider is available at runtime
func CheckProviderAvailability(provider *Provider) bool {
	logger := core.GetLogger()
	logger.Debug("Checking provider availability", "provider_id", provider.ID, "type", provider.Type)

	switch provider.Type {
	case APIProviderType:
		// Check if API key environment variable is set
		if envVar, exists := provider.Config["api_key"]; exists {
			apiKey := os.Getenv(envVar)
			available := apiKey != ""
			logger.Debug("API provider availability check", 
				"provider_id", provider.ID, 
				"env_var", envVar, 
				"available", available)
			return available
		}
		return false

	case CLIProviderType:
		// Check if CLI tool is available in PATH
		switch provider.ID {
		case "claude-cli":
			_, err := exec.LookPath("claude")
			available := err == nil
			logger.Debug("CLI provider availability check", 
				"provider_id", provider.ID, 
				"available", available)
			return available
		}
		return false

	case LocalProviderType:
		// Check if local service is running (e.g., Ollama)
		switch provider.ID {
		case "ollama":
			// TODO: Implement Ollama availability check via HTTP ping
			logger.Debug("Local provider availability check", 
				"provider_id", provider.ID, 
				"available", false)
			return false
		}
		return false

	default:
		logger.Warn("Unknown provider type", "provider_id", provider.ID, "type", provider.Type)
		return false
	}
}

// UpdateProviderAvailability updates the availability status of all providers
func UpdateProviderAvailability(config *ProviderConfig) {
	logger := core.GetLogger()
	logger.Debug("Updating provider availability for all providers")

	for i := range config.Providers {
		config.Providers[i].Available = CheckProviderAvailability(&config.Providers[i])
		logger.Debug("Provider availability updated", 
			"provider_id", config.Providers[i].ID, 
			"available", config.Providers[i].Available)
	}
}

// GetAvailableProviders returns only enabled and available providers
func GetAvailableProviders(config *ProviderConfig) []Provider {
	var available []Provider
	for _, provider := range config.Providers {
		if provider.Enabled && provider.Available {
			available = append(available, provider)
		}
	}
	return available
}

// GetProviderByID returns a provider by its ID
func GetProviderByID(config *ProviderConfig, id string) *Provider {
	for i := range config.Providers {
		if config.Providers[i].ID == id {
			return &config.Providers[i]
		}
	}
	return nil
}

// getConfigPath returns the path to the provider configuration file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".commitlore", "providers.json"), nil
}