package tui

import (
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

func RunApp() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}