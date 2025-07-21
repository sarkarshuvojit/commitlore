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
		return m.renderErrorState()
	}

	if m.loading {
		return m.renderLoadingState()
	}

	if len(m.providers) == 0 {
		return m.renderEmptyState()
	}

	return m.renderMainView()
}

// New beautiful rendering methods

func (m *ProviderModel) renderErrorState() string {
	// Sophisticated error display with gradient border
	errorIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ef4444")).
		SetString("󰀪")

	errorTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ef4444")).
		Bold(true).
		SetString("Connection Error")

	errorMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748b")).
		SetString(m.errorMsg)

	errorCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ef4444")).
		Padding(2, 4).
		Width(60).
		Align(lipgloss.Center)

	errorContent := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Left, errorIcon.Render(), " ", errorTitle.Render()),
		"",
		errorMsg.Render(),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b")).Render("Press 'esc' to go back"))

	return lipgloss.Place(100, 30, lipgloss.Center, lipgloss.Center, errorCard.Render(errorContent))
}

func (m *ProviderModel) renderLoadingState() string {
	// Elegant loading animation with spinner
	spinner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6366f1")).
		Bold(true).
		SetString("◐")

	loadingText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8")).
		SetString("Discovering AI providers...")

	loadingCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Padding(2, 4).
		Width(40).
		Align(lipgloss.Center)

	loadingContent := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Left, spinner.Render(), " ", loadingText.Render()))

	return lipgloss.Place(100, 30, lipgloss.Center, lipgloss.Center, loadingCard.Render(loadingContent))
}

func (m *ProviderModel) renderEmptyState() string {
	// Beautiful empty state with illustration
	emptyIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748b")).
		SetString("󰋘")

	emptyTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8")).
		Bold(true).
		SetString("No AI Providers Available")

	emptyMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748b")).
		SetString("Configure your preferred AI provider to get started")

	emptyCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Padding(3, 6).
		Width(50).
		Align(lipgloss.Center)

	emptyContent := lipgloss.JoinVertical(lipgloss.Center,
		emptyIcon.Render(),
		"",
		emptyTitle.Render(),
		"",
		emptyMsg.Render(),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b")).Render("Press 'esc' to go back"))

	return lipgloss.Place(100, 30, lipgloss.Center, lipgloss.Center, emptyCard.Render(emptyContent))
}

func (m *ProviderModel) renderMainView() string {
	// Modern card-based layout with sophisticated spacing
	header := m.renderModernHeader()
	providerGrid := m.renderProviderGrid()
	footer := m.renderModernFooter()

	mainContainer := lipgloss.NewStyle().
		Padding(2, 4).
		Width(96)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		providerGrid,
		"",
		footer)

	return mainContainer.Render(content)
}

func (m *ProviderModel) renderModernHeader() string {
	// Elegant header with gradient effect
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8fafc")).
		Bold(true).
		SetString("AI Provider Selection")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8")).
		SetString("Choose your preferred AI assistant")

	// Active provider indicator
	activeProviderText := ""
	if m.providerConfig != nil {
		for _, provider := range m.providers {
			if provider.ID == m.providerConfig.ActiveProviderID {
				activeIndicator := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#10b981")).
					SetString("●")
				
				activeName := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#10b981")).
					Bold(true).
					SetString(provider.Name)

				activeProviderText = lipgloss.JoinHorizontal(lipgloss.Left,
					activeIndicator.Render(), " Currently using: ", activeName.Render())
				break
			}
		}
	}

	headerContent := lipgloss.JoinVertical(lipgloss.Left,
		title.Render(),
		subtitle.Render())
	
	if activeProviderText != "" {
		headerContent = lipgloss.JoinVertical(lipgloss.Left,
			headerContent,
			"",
			activeProviderText)
	}

	return headerContent
}

func (m *ProviderModel) renderProviderGrid() string {
	// Modern card-based provider grid
	var cards []string

	for i, provider := range m.providers {
		card := m.renderProviderCard(provider, i == m.cursor)
		cards = append(cards, card)
	}

	// Arrange cards in a clean vertical layout with proper spacing
	return lipgloss.JoinVertical(lipgloss.Left, cards...)
}

func (m *ProviderModel) renderProviderCard(provider config.Provider, isSelected bool) string {
	// Sophisticated card design with status indicators
	isActive := provider.ID == m.providerConfig.ActiveProviderID

	// Define card styling based on state
	var borderColor, bgColor lipgloss.Color
	var borderStyle lipgloss.Border = lipgloss.RoundedBorder()

	if isSelected {
		if isActive {
			borderColor = lipgloss.Color("#10b981") // Green for active selection
		} else if !provider.Enabled {
			borderColor = lipgloss.Color("#64748b") // Gray for disabled selection
		} else if !provider.Available {
			borderColor = lipgloss.Color("#f59e0b") // Amber for unavailable selection
		} else {
			borderColor = lipgloss.Color("#6366f1") // Primary for available selection
		}
		bgColor = lipgloss.Color("#1e293b") // Darker background for selected
	} else {
		borderColor = lipgloss.Color("#334155") // Subtle border for unselected
		bgColor = lipgloss.Color("#0f172a")     // Dark background for unselected
	}

	// Selection indicator
	cursor := ""
	if isSelected {
		cursor = lipgloss.NewStyle().
			Foreground(borderColor).
			Bold(true).
			SetString("▶ ").Render()
	} else {
		cursor = "  "
	}

	// Provider type icon with modern styling
	typeIcon := m.getModernTypeIcon(provider.Type)
	typeIconStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8b5cf6")).
		Bold(true).
		SetString(typeIcon)

	// Provider name with proper hierarchy
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8fafc")).
		Bold(true)
	if !provider.Enabled {
		nameStyle = nameStyle.Foreground(lipgloss.Color("#64748b"))
	}
	providerName := nameStyle.SetString(provider.Name)

	// Status badges with modern design
	statusBadge := m.renderStatusBadge(provider, isActive)

	// Provider description with subtle styling
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8"))
	if !provider.Enabled {
		descStyle = descStyle.Foreground(lipgloss.Color("#64748b")).Italic(true)
	}

	description := provider.Description
	if !provider.Enabled {
		description = "Under development"
	}

	// Availability hint for unavailable providers
	var availabilityHint string
	if provider.Enabled && !provider.Available {
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")).
			Italic(true)
		availabilityHint = hintStyle.SetString("⚡ " + m.getAvailabilityHint(provider)).Render()
	}

	// Card header with icon, name, and status
	cardHeader := lipgloss.JoinHorizontal(lipgloss.Left,
		cursor,
		typeIconStyled.Render(), " ",
		providerName.Render(), " ",
		statusBadge)

	// Card content assembly
	var cardContent string
	if availabilityHint != "" {
		cardContent = lipgloss.JoinVertical(lipgloss.Left,
			cardHeader,
			"",
			descStyle.Render(description),
			"",
			availabilityHint)
	} else {
		cardContent = lipgloss.JoinVertical(lipgloss.Left,
			cardHeader,
			"",
			descStyle.Render(description))
	}

	// Final card styling
	cardStyle := lipgloss.NewStyle().
		Border(borderStyle).
		BorderForeground(borderColor).
		Background(bgColor).
		Padding(1, 2).
		Margin(0, 0, 1, 0).
		Width(84)

	return cardStyle.Render(cardContent)
}

func (m *ProviderModel) renderStatusBadge(provider config.Provider, isActive bool) string {
	if isActive {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#10b981")).
			Padding(0, 1).
			Bold(true).
			SetString("ACTIVE").Render()
	}

	if !provider.Enabled {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#64748b")).
			Padding(0, 1).
			SetString("BETA").Render()
	}

	if !provider.Available {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#f59e0b")).
			Padding(0, 1).
			SetString("SETUP REQUIRED").Render()
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#6366f1")).
		Padding(0, 1).
		SetString("READY").Render()
}

func (m *ProviderModel) renderModernFooter() string {
	// Elegant footer with helpful keyboard shortcuts
	var shortcuts []string

	if len(m.providers) > 0 {
		shortcuts = append(shortcuts,
			m.renderShortcut("↑↓", "navigate"),
			m.renderShortcut("enter", "select"),
			m.renderShortcut("r", "refresh"))
	}

	shortcuts = append(shortcuts,
		m.renderShortcut("esc", "back"),
		m.renderShortcut("q", "quit"))

	shortcutText := lipgloss.JoinHorizontal(lipgloss.Left, shortcuts...)

	// Position indicator
	position := ""
	if len(m.providers) > 0 {
		posStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6366f1")).
			Bold(true)
		position = posStyle.SetString(fmt.Sprintf("%d/%d", m.cursor+1, len(m.providers))).Render()
	}

	// Create footer layout
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748b")).
		Width(84)

	if position != "" {
		footer := lipgloss.JoinHorizontal(lipgloss.Left,
			shortcutText,
			strings.Repeat(" ", max(0, 84-lipgloss.Width(shortcutText)-lipgloss.Width(position))),
			position)
		return footerStyle.Render(footer)
	}

	return footerStyle.Render(shortcutText)
}

func (m *ProviderModel) renderShortcut(key, desc string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8b5cf6")).
		Bold(true)
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748b"))

	return lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render(key), " ", descStyle.Render(desc), "  ")
}

func (m *ProviderModel) getModernTypeIcon(providerType config.ProviderType) string {
	switch providerType {
	case config.APIProviderType:
		return "󰖟"  // Cloud icon
	case config.CLIProviderType:
		return "󰆍"  // Terminal icon
	case config.LocalProviderType:
		return "󰟀"  // Computer icon
	default:
		return "󰋘"  // Generic icon
	}
}

// Helper function for max calculation
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

