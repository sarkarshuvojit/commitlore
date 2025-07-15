package tui

import (
	"context"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
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
	cwd, _ := os.Getwd()
	gitRoot, isGit, _ := core.GetGitDirectory(cwd)
	
	// Initialize LLM provider
	var llmProvider llm.LLMProvider
	var llmProviderType string
	if llm.IsClaudeCLIAvailable() {
		if cliClient, err := llm.NewClaudeCLIClient(); err == nil {
			llmProvider = cliClient
			llmProviderType = "Claude CLI"
		} else {
			if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
				llmProvider = llm.NewClaudeClient(apiKey)
				llmProviderType = "Claude API"
			} else {
				llmProvider = &mockLLMProvider{}
				llmProviderType = "Mock"
			}
		}
	} else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		llmProvider = llm.NewClaudeClient(apiKey)
		llmProviderType = "Claude API"
	} else {
		llmProvider = &mockLLMProvider{}
		llmProviderType = "Mock"
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
	case ErrorMsg:
		m.errorMsg = msg.Error
		return m, nil
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
		m.contentModel.SetContext(m.selectedTopic, m.selectedFormat)
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