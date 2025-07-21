package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/config"
)

// ProviderModel handles the provider management view
type ProviderModel struct {
	BaseModel
	cursor         int
	providers      []config.Provider
	providerConfig *config.ProviderConfig
	loading        bool
}

// NewProviderModel creates a new provider model
func NewProviderModel(base BaseModel) *ProviderModel {
	return &ProviderModel{
		BaseModel:      base,
		cursor:         0,
		providers:      []config.Provider{},
		providerConfig: nil,
		loading:        true,
	}
}

func (m *ProviderModel) Init() tea.Cmd {
	return m.loadProviders
}

func (m *ProviderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.providers)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			if len(m.providers) > 0 {
				m.cursor = len(m.providers) - 1
			}
		case "enter":
			if len(m.providers) > 0 && m.cursor < len(m.providers) {
				selectedProvider := m.providers[m.cursor]
				if selectedProvider.Enabled && selectedProvider.Available {
					// Select this provider and go back
					m.providerConfig.ActiveProviderID = selectedProvider.ID
					return m, tea.Batch(
						m.saveProviderConfig,
						func() tea.Msg { return providerChangedMsg{} },
						func() tea.Msg { return BackMsg{} },
					)
				}
			}
		case "r":
			// Refresh provider availability
			return m, m.loadProviders
		case "escape":
			return m, func() tea.Msg { return BackMsg{} }
		}
	case providerLoadedMsg:
		m.loading = false
		m.providerConfig = msg.config
		m.providers = msg.config.Providers
		// Update availability
		config.UpdateProviderAvailability(m.providerConfig)
		return m, nil
	case ErrorMsg:
		m.loading = false
		m.errorMsg = msg.Error
		return m, nil
	}
	return m, nil
}

func (m *ProviderModel) View() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'escape' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}

	if m.loading {
		loadingContent := emptyStyle.Render("⏳ Loading providers...")
		return appStyle.Render(loadingContent)
	}

	if len(m.providers) == 0 {
		emptyContent := emptyStyle.Render("📭 No providers configured")
		helpText := helpDescStyle.Render("Press 'escape' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Center, emptyContent, helpText))
	}

	header := m.renderHeader()
	content := m.renderProviderList()
	statusBar := m.renderStatusBar()

	main := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
	return appStyle.Render(main)
}

func (m *ProviderModel) renderHeader() string {
	title := titleStyle.Render("🔧 Provider Management")
	subtitle := subtitleStyle.Render(fmt.Sprintf("%d providers configured", len(m.providers)))

	headerContent := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)

	return headerWithBg
}

func (m *ProviderModel) renderProviderList() string {
	var rows []string

	for i, provider := range m.providers {
		isSelected := i == m.cursor
		isActive := provider.ID == m.providerConfig.ActiveProviderID

		row := m.renderProviderRow(provider, isSelected, isActive)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return contentStyle.Render(content)
}

func (m *ProviderModel) renderProviderRow(provider config.Provider, isSelected bool, isActive bool) string {
	cursor := "  "
	if isSelected {
		cursor = "▶ "
	}

	// Provider name with type indicator
	name := provider.Name
	typeIndicator := m.getTypeIndicator(provider.Type)
	
	// Status indicators
	statusIndicators := []string{}
	
	if isActive {
		statusIndicators = append(statusIndicators, "🟢 ACTIVE")
	}
	
	if !provider.Enabled {
		statusIndicators = append(statusIndicators, "🔒 DISABLED")
	} else if !provider.Available {
		statusIndicators = append(statusIndicators, "❌ UNAVAILABLE")
	} else {
		statusIndicators = append(statusIndicators, "✅ AVAILABLE")
	}

	// Build the row content
	firstLine := fmt.Sprintf("%s%s %s %s", cursor, typeIndicator, name, strings.Join(statusIndicators, " "))
	
	// Use "Under development" for disabled providers, otherwise use the actual description
	var description string
	if !provider.Enabled {
		description = "Under development"
	} else {
		description = provider.Description
	}
	secondLine := fmt.Sprintf("  %s", description)

	// Add availability details for unavailable providers
	thirdLine := ""
	if provider.Enabled && !provider.Available {
		thirdLine = fmt.Sprintf("  💡 %s", m.getAvailabilityHint(provider))
	}

	var rowContent string
	if thirdLine != "" {
		rowContent = lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine, thirdLine)
	} else {
		rowContent = lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
	}

	// Apply styling based on state with consistent width and alignment
	var style lipgloss.Style
	if isSelected {
		if !provider.Enabled {
			style = disabledSelectedRowStyle
		} else if !provider.Available {
			style = unavailableSelectedRowStyle
		} else {
			style = selectedCommitRowStyle
		}
	} else {
		if !provider.Enabled {
			style = disabledRowStyle
		} else if !provider.Available {
			style = unavailableRowStyle
		} else {
			style = commitRowStyle
		}
	}
	
	// Apply consistent width and alignment to all rows
	return style.Width(96).Align(lipgloss.Left).Render(rowContent)
}

func (m *ProviderModel) getTypeIndicator(providerType config.ProviderType) string {
	switch providerType {
	case config.APIProviderType:
		return "🌐"
	case config.CLIProviderType:
		return "⚡"
	case config.LocalProviderType:
		return "💻"
	default:
		return "❓"
	}
}

func (m *ProviderModel) getAvailabilityHint(provider config.Provider) string {
	switch provider.Type {
	case config.APIProviderType:
		if envVar, exists := provider.Config["api_key"]; exists {
			return fmt.Sprintf("Set environment variable: %s", envVar)
		}
		return "API key required"
	case config.CLIProviderType:
		switch provider.ID {
		case "claude-cli":
			return "Install Claude CLI: https://claude.ai/download"
		}
		return "CLI tool not found in PATH"
	case config.LocalProviderType:
		switch provider.ID {
		case "ollama":
			return "Install and start Ollama: https://ollama.ai"
		}
		return "Local service not running"
	default:
		return "Unknown availability issue"
	}
}

func (m *ProviderModel) renderStatusBar() string {
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("select"))
	refreshHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("r"), helpDescStyle.Render("refresh"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))

	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.providers)))

	var activeProvider string
	if m.providerConfig != nil {
		for _, provider := range m.providers {
			if provider.ID == m.providerConfig.ActiveProviderID {
				activeProvider = fmt.Sprintf("Active: %s", provider.Name)
				break
			}
		}
	}

	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", refreshHelp, " • ", backHelp, " • ", quitHelp)

	rightSide := fmt.Sprintf("%s • %s", position, activeProvider)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 5),
		rightSide,
	)

	return statusBarStyle.Render(statusContent)
}

// loadProviders is a command that loads provider configuration
func (m *ProviderModel) loadProviders() tea.Msg {
	logger := core.GetLogger()
	logger.Debug("Loading provider configuration")

	config, err := config.LoadProviderConfig()
	if err != nil {
		logger.Error("Failed to load provider config", "error", err)
		return ErrorMsg{Error: fmt.Sprintf("Failed to load providers: %v", err)}
	}

	logger.Info("Successfully loaded provider configuration", "providers_count", len(config.Providers))
	return providerLoadedMsg{config: config}
}

// saveProviderConfig is a command that saves provider configuration
func (m *ProviderModel) saveProviderConfig() tea.Msg {
	logger := core.GetLogger()
	logger.Debug("Saving provider configuration")

	if err := config.SaveProviderConfig(m.providerConfig); err != nil {
		logger.Error("Failed to save provider config", "error", err)
		return ErrorMsg{Error: fmt.Sprintf("Failed to save providers: %v", err)}
	}

	logger.Info("Successfully saved provider configuration")
	return nil
}

// Custom messages for provider management
type providerLoadedMsg struct {
	config *config.ProviderConfig
}

// Additional styles for provider display
var (
	disabledRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	unavailableRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("203"))

	disabledSelectedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Italic(true)

	unavailableSelectedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("203")).
		Foreground(lipgloss.Color("255")).
		Bold(true)
)