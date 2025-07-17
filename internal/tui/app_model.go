package tui

import (
	"context"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/config"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// mockLLMProvider provides mock responses when no API key is available
type mockLLMProvider struct{}

func (m *mockLLMProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return m.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

func (m *mockLLMProvider) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	mockTopics := []string{
		"Implementing modern Go patterns and best practices",
		"Building terminal user interfaces with Bubble Tea",
		"Git repository analysis and commit processing",
		"Error handling and robust software design",
		"API integration and external service communication",
	}
	
	result := ""
	for _, topic := range mockTopics {
		result += topic + "\n"
	}
	
	return result, nil
}

// NewAppModel creates a new app model with all sub-models
func NewAppModel() *AppModel {
	logger := core.GetLogger()
	cwd, _ := os.Getwd()
	gitRoot, isGit, _ := core.GetGitDirectory(cwd)
	
	// Load provider configuration
	providerConfig, err := config.LoadProviderConfig()
	if err != nil {
		logger.Error("Failed to load provider config, using defaults", "error", err)
		providerConfig = config.DefaultProviderConfig()
	}
	
	// Update provider availability
	config.UpdateProviderAvailability(providerConfig)
	
	// Create provider factory
	factory := config.NewProviderFactory(providerConfig)
	
	// Initialize LLM provider using factory
	var llmProvider llm.LLMProvider
	var llmProviderType string
	
	provider, providerName, err := factory.CreateActiveProvider()
	if err != nil {
		logger.Warn("Failed to create active provider, falling back to mock", "error", err)
		llmProvider = &mockLLMProvider{}
		llmProviderType = "Mock (No providers available)"
	} else {
		llmProvider = provider
		llmProviderType = providerName
	}
	
	baseModel := BaseModel{
		repoPath:        gitRoot,
		llmProvider:     llmProvider,
		llmProviderType: llmProviderType,
	}
	
	if !isGit {
		baseModel.errorMsg = "Not in a git repository"
	}
	
	app := &AppModel{
		BaseModel:       baseModel,
		currentView:     SplashView,
		selectedCommits: make(map[int]bool),
	}
	
	// Initialize sub-models
	app.splashModel = NewSplashModel(baseModel)
	app.listingModel = NewListingModel(baseModel)
	app.topicModel = NewTopicModel(baseModel)
	app.formatModel = NewFormatModel(baseModel)
	app.contentModel = NewContentModel(baseModel)
	app.providerModel = NewProviderModel(baseModel)
	
	return app
}

func (m *AppModel) Init() tea.Cmd {
	return m.getCurrentModel().Init()
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case NextMsg:
		return m.handleNext()
	case BackMsg:
		return m.handleBack()
	case ProviderMsg:
		if m.currentView != ProviderView {
			m.currentView = ProviderView
			return m, m.providerModel.Init()
		}
		return m, nil
	case ErrorMsg:
		m.errorMsg = msg.Error
		return m, nil
	case providerChangedMsg:
		// Provider was changed, reload the base model
		return m.reloadProvider()
	}
	
	// Delegate to current view model
	currentModel := m.getCurrentModel()
	updatedModel, cmd := currentModel.Update(msg)
	m.setCurrentModel(updatedModel)
	
	return m, cmd
}

func (m *AppModel) View() string {
	return m.getCurrentModel().View()
}

func (m *AppModel) getCurrentModel() ViewInterface {
	switch m.currentView {
	case SplashView:
		return m.splashModel
	case ListingView:
		return m.listingModel
	case TopicSelectionView:
		return m.topicModel
	case FormatSelectionView:
		return m.formatModel
	case ContentCreationView:
		return m.contentModel
	case ProviderView:
		return m.providerModel
	default:
		return m.splashModel
	}
}

func (m *AppModel) setCurrentModel(model tea.Model) {
	switch m.currentView {
	case SplashView:
		m.splashModel = model.(*SplashModel)
	case ListingView:
		m.listingModel = model.(*ListingModel)
	case TopicSelectionView:
		m.topicModel = model.(*TopicModel)
	case FormatSelectionView:
		m.formatModel = model.(*FormatModel)
	case ContentCreationView:
		m.contentModel = model.(*ContentModel)
	case ProviderView:
		m.providerModel = model.(*ProviderModel)
	}
}

func (m *AppModel) handleNext() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case SplashView:
		m.currentView = ListingView
		return m, m.listingModel.Init()
	case ListingView:
		// Get selected commits and extract topics
		commits, selectedCommits := m.listingModel.GetSelectedCommits()
		m.selectedCommits = selectedCommits
		
		// Start async topic extraction
		cmd := m.topicModel.ExtractTopics(commits, selectedCommits)
		
		m.currentView = TopicSelectionView
		return m, cmd
		
	case TopicSelectionView:
		// Get selected topic and move to format selection
		m.selectedTopic = m.topicModel.GetSelectedTopic()
		m.formatModel.SetSelectedTopic(m.selectedTopic)
		m.currentView = FormatSelectionView
		return m, m.formatModel.Init()
		
	case FormatSelectionView:
		// Get selected format and move to content creation
		m.selectedFormat = m.formatModel.GetSelectedFormat()
		commits, selectedCommits := m.listingModel.GetSelectedCommits()
		m.contentModel.SetContextWithCommits(m.selectedTopic, m.selectedFormat, commits, selectedCommits)
		m.currentView = ContentCreationView
		return m, m.contentModel.Init()
	}
	
	return m, nil
}

func (m *AppModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ListingView:
		m.currentView = SplashView
		return m, m.splashModel.Init()
	case TopicSelectionView:
		m.currentView = ListingView
		return m, m.listingModel.Init()
	case FormatSelectionView:
		m.currentView = TopicSelectionView
		return m, m.topicModel.Init()
	case ContentCreationView:
		m.currentView = FormatSelectionView
		return m, m.formatModel.Init()
	case ProviderView:
		m.currentView = SplashView
		return m, m.splashModel.Init()
	case SplashView:
		// Clear selections
		m.selectedCommits = make(map[int]bool)
		if m.listingModel != nil {
			m.listingModel.selectedCommits = make(map[int]bool)
		}
		return m, nil
	}
	
	return m, nil
}

// providerChangedMsg is sent when the active provider has been changed
type providerChangedMsg struct{}

// reloadProvider reloads the provider after a change
func (m *AppModel) reloadProvider() (tea.Model, tea.Cmd) {
	logger := core.GetLogger()
	logger.Debug("Reloading provider after configuration change")

	// Load updated provider configuration
	providerConfig, err := config.LoadProviderConfig()
	if err != nil {
		logger.Error("Failed to reload provider config", "error", err)
		m.errorMsg = "Failed to reload provider configuration"
		return m, nil
	}

	// Update provider availability
	config.UpdateProviderAvailability(providerConfig)

	// Create provider factory
	factory := config.NewProviderFactory(providerConfig)

	// Create new provider instance
	provider, providerName, err := factory.CreateActiveProvider()
	if err != nil {
		logger.Warn("Failed to create active provider after reload, falling back to mock", "error", err)
		m.llmProvider = &mockLLMProvider{}
		m.llmProviderType = "Mock (No providers available)"
	} else {
		m.llmProvider = provider
		m.llmProviderType = providerName
	}

	// Update all sub-models with new base model
	baseModel := BaseModel{
		repoPath:        m.repoPath,
		llmProvider:     m.llmProvider,
		llmProviderType: m.llmProviderType,
		errorMsg:        m.errorMsg,
	}

	// Update all existing models
	m.listingModel.BaseModel = baseModel
	m.topicModel.BaseModel = baseModel
	m.formatModel.BaseModel = baseModel
	m.contentModel.BaseModel = baseModel
	m.providerModel.BaseModel = baseModel

	logger.Info("Successfully reloaded provider", "provider_name", m.llmProviderType)
	return m, nil
}