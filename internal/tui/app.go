package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func RunApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}