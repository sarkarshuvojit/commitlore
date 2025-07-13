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
