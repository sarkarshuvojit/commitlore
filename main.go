package main

import (
	"fmt"
	"os"

	"github.com/sarkarshuvojit/commitlore/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewState int

const (
	HelloView ViewState = iota
)

type model struct {
	currentView ViewState
}

func initialModel() model {
	return model{
		currentView: HelloView,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case HelloView:
		return m.renderHelloView()
	default:
		return "Unknown view"
	}
}

func (m model) renderHelloView() string {
	return "Hello, World!\n\nPress 'q' or Ctrl+C to quit."
}


func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	
	isGitRepo, err := core.IsGitRepository(cwd)
	if err != nil {
		fmt.Printf("Error checking Git repository: %v\n", err)
		os.Exit(1)
	}
	
	if !isGitRepo {
		fmt.Println("Error: Current directory is not a Git repository")
		os.Exit(1)
	}
	
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}