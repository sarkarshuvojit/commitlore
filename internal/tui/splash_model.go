package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SplashModel struct {
	BaseModel
}

func NewSplashModel(base BaseModel) *SplashModel {
	return &SplashModel{
		BaseModel: base,
	}
}

func (m *SplashModel) Init() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return splashTimerMsg{}
	})
}

func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, func() tea.Msg { return NextMsg{} }
		case "p", "P":
			return m, func() tea.Msg { return ProviderMsg{} }
		}
	case splashTimerMsg:
		return m, func() tea.Msg { return NextMsg{} }
	}
	return m, nil
}

func (m *SplashModel) View() string {
	if m.errorMsg != "" {
		return errorStyle.Render("Error: " + m.errorMsg)
	}

	logo := `
   ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗██╗      ██████╗ ██████╗ ███████╗
  ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝██║     ██╔═══██╗██╔══██╗██╔════╝
  ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║   ██║     ██║   ██║██████╔╝█████╗  
  ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║   ██║     ██║   ██║██╔══██╗██╔══╝  
  ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║   ███████╗╚██████╔╝██║  ██║███████╗
   ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
`

	subtitle := "Transform your Git history into compelling stories"
	
	// Center the logo and subtitle
	lines := strings.Split(logo, "\n")
	var centeredLines []string
	
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			centeredLines = append(centeredLines, titleStyle.Render(line))
		}
	}
	
	centeredSubtitle := subtitleStyle.Render(subtitle)
	
	content := strings.Join(centeredLines, "\n") + "\n\n" + centeredSubtitle
	
	// Add provider information
	providerInfo := dimStyle.Render("Active Provider: " + m.llmProviderType)
	
	// Add keyboard shortcuts
	shortcuts := dimStyle.Render("Press ENTER to continue • Press P for provider settings")
	
	// Add some spacing and content
	content += "\n\n" + providerInfo + "\n\n" + shortcuts
	
	return appStyle.Render(content)
}