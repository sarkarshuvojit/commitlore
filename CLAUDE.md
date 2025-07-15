# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CommitLore is an AI-powered TUI (Terminal User Interface) application built with Go and Bubble Tea that transforms Git commit histories into compelling developer stories and content. The application analyzes commit patterns and generates social media posts, blog articles, and technical narratives.

## Development Commands

**Build and Run:**
```bash
go run main.go
```

**Build Binary:**
```bash
go build -o commitlore main.go
```

**Add Dependencies:**
```bash
go get <package-name>
go mod tidy
```

**Module Management:**
```bash
go mod download    # Download dependencies
go mod verify      # Verify dependencies
```

## Architecture

The application follows the Bubble Tea MVU (Model-View-Update) pattern:

- **State Management**: Uses a `ViewState` enum to manage different application views
- **Model**: The `model` struct contains the application state including `currentView`
- **Views**: Each view state has its own render method (e.g., `renderHelloView()`)
- **Event Handling**: The `Update()` method processes all user input and state transitions

## Key Files

- `main.go`: Contains the complete TUI application with state machine implementation
- `go.mod`: Defines the module and Bubble Tea dependencies
- `readme.md`: Product specification and business model documentation

## TUI Framework

The application uses Charm's Bubble Tea framework (`github.com/charmbracelet/bubbletea`) for building terminal user interfaces. Key concepts:

- **Init()**: Initial commands when the program starts
- **Update()**: Handles messages and updates model state
- **View()**: Renders the current state as a string
- **tea.Cmd**: Commands that can be returned to perform side effects

## Current Implementation

The current codebase implements a basic "Hello World" view with:
- Quit functionality via 'q' or Ctrl+C
- State machine ready for additional views
- Proper error handling in the main function

## Running the Application

The application requires a proper TTY environment. It will not run correctly in environments without TTY support (like some CI systems or IDEs' embedded terminals).

## Code style guides

- maintain an internal pkg inside which all of the logic needs to reside
- create and use subpackage roots as follows
- internal/tui: This should hold display layer code ONLY
- internal/core: This would contain all core code like interfacing with git, interfacing with llm, disk code etc. Make this code highly testable. Don't generate tests unless asked, but make sure the implementations are testable.

## Screen/View Model Standards

### Creating New Screens

Each screen in the TUI should follow these standards for consistency and maintainability:

#### 1. Model Structure
- **Individual Screen Models**: Each screen should have its own model struct (e.g., `ListingModel`, `TopicModel`, `FormatModel`)
- **BaseModel Embedding**: All screen models should embed `BaseModel` to inherit common functionality:
  ```go
  type ScreenModel struct {
      BaseModel
      // screen-specific fields
  }
  ```
- **BaseModel Contents**: Common fields like `repoPath`, `llmProvider`, `llmProviderType`, `errorMsg`

#### 2. Constructor Pattern
- **Constructor Function**: Each screen model should have a `NewScreenModel(base BaseModel)` constructor
- **Field Initialization**: Initialize all screen-specific fields with sensible defaults
- **BaseModel Inheritance**: Pass the BaseModel to embed shared functionality

#### 3. tea.Model Interface Implementation
Every screen model must implement the three required methods:

```go
func (m *ScreenModel) Init() tea.Cmd {
    // Return initial commands or nil
}

func (m *ScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle input events and return updated model + commands
}

func (m *ScreenModel) View() string {
    // Render the screen and return formatted string
}
```

#### 4. ViewInterface Implementation
All screen models should implement the `ViewInterface` for consistency:
```go
type ViewInterface interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

#### 5. Message Handling Patterns
- **Standard Navigation**: Handle common keys like `up/k`, `down/j`, `home/g`, `end/G`
- **Screen Transitions**: Use `NextMsg{}` and `BackMsg{}` for navigation between screens
- **Error Handling**: Use `ErrorMsg{Error: string}` for error propagation
- **Input Validation**: Validate input before processing and show helpful feedback

#### 6. View Rendering Standards
- **Error Display**: Always check for and display `m.errorMsg` first
- **Header Section**: Use consistent header styling with title and subtitle
- **Content Section**: Render main content with proper styling
- **Status Bar**: Include navigation help and status information
- **Styling Consistency**: Use shared styles from `styles.go`

#### 7. State Management
- **Screen-Specific State**: Keep screen-specific state in the model struct
- **Shared State**: Use AppModel for state that needs to be shared between screens
- **Data Passing**: Use methods like `GetSelectedCommits()` to pass data between screens

#### 8. Navigation Integration
- **AppModel Integration**: Register the screen model in AppModel's constructor
- **View State**: Add new view state to `ViewState` enum
- **getCurrentModel()**: Add case in AppModel's getCurrentModel() switch
- **setCurrentModel()**: Add case in AppModel's setCurrentModel() switch

#### 9. File Organization
- **Separate Files**: Each screen model should be in its own file (e.g., `listing_model.go`)
- **Consistent Naming**: Use `*_model.go` naming convention
- **Package Structure**: Keep all TUI models in the `internal/tui` package

#### 10. Common Patterns
- **Cursor Management**: Use `cursor` field for item selection
- **Selection States**: Use boolean maps for multi-selection (e.g., `map[int]bool`)
- **Pagination**: Implement viewport-based pagination for large datasets
- **Loading States**: Show loading indicators for async operations

### Example Screen Model Template

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
)

type NewScreenModel struct {
    BaseModel
    cursor int
    items  []string
}

func NewNewScreenModel(base BaseModel) *NewScreenModel {
    return &NewScreenModel{
        BaseModel: base,
        cursor:    0,
        items:     []string{},
    }
}

func (m *NewScreenModel) Init() tea.Cmd {
    return nil
}

func (m *NewScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.items)-1 {
                m.cursor++
            }
        case "enter":
            // Handle selection
            return m, func() tea.Msg { return NextMsg{} }
        case "escape":
            return m, func() tea.Msg { return BackMsg{} }
        }
    }
    return m, nil
}

func (m *NewScreenModel) View() string {
    if m.errorMsg != "" {
        // Render error state
    }
    
    // Render header, content, and status bar
    return appStyle.Render(content)
}
```

## Development Notes

- you cannot run this inside claude code because you don't have TTY support so stop trying to test it that way