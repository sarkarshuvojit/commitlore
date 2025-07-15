package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sarkarshuvojit/commitlore/internal/core"
)

func RunApp() error {
	logger := core.GetLogger()
	logger.Info("Initializing TUI application")
	
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	if err != nil {
		logger.Error("TUI program execution failed", "error", err)
	} else {
		logger.Info("TUI application terminated successfully")
	}
	return err
}