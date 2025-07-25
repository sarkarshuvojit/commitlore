package tui

import (
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewState represents the different states of the application
type ViewState int

const (
	SplashView ViewState = iota
	ListingView
	TopicSelectionView
	FormatSelectionView
	ContentCreationView
	ProviderView
)

// MessageType represents the type of message to display
type MessageType int

const (
	MessageTypeInfo MessageType = iota
	MessageTypeWarning
	MessageTypeError
	MessageTypeSuccess
)

// StatusMessage represents a message with a type
type StatusMessage struct {
	Content string
	Type    MessageType
}

// BaseModel contains common data needed by all models
type BaseModel struct {
	repoPath        string
	llmProvider     llm.LLMProvider
	llmProviderType string
	statusMessage   *StatusMessage
	errorMsg        string // Deprecated: use statusMessage instead
}

// AppModel is the main model that manages view state and delegation
type AppModel struct {
	BaseModel
	currentView ViewState
	
	// Individual view models
	splashModel    *SplashModel
	listingModel   *ListingModel
	topicModel     *TopicModel
	formatModel    *FormatModel
	contentModel   *ContentModel
	providerModel  *ProviderModel
	
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
	ProviderMsg    struct{}
	flashTimerMsg  struct{}
	splashTimerMsg struct{}
)

// ViewInterface defines the common interface for all view models
type ViewInterface interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}