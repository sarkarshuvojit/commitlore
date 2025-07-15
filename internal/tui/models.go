package tui

import (
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewState represents the different states of the application
type ViewState int

const (
	ListingView ViewState = iota
	TopicSelectionView
	FormatSelectionView
	ContentCreationView
)

// BaseModel contains common data needed by all models
type BaseModel struct {
	repoPath        string
	llmProvider     llm.LLMProvider
	llmProviderType string
	errorMsg        string
}

// AppModel is the main model that manages view state and delegation
type AppModel struct {
	BaseModel
	currentView ViewState
	
	// Individual view models
	listingModel   *ListingModel
	topicModel     *TopicModel
	formatModel    *FormatModel
	contentModel   *ContentModel
	
	// Shared data between views
	selectedCommits map[int]bool
	selectedTopic   string
	selectedFormat  string
}

// Common messages used across views
type (
	BackMsg        struct{}
	NextMsg        struct{}
	ErrorMsg       struct{ Error string }
	SelectionMsg   struct{ Selection interface{} }
	flashTimerMsg  struct{}
)

// ViewInterface defines the common interface for all view models
type ViewInterface interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}