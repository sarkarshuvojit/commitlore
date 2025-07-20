package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// FormatModel handles the format selection view
type FormatModel struct {
	BaseModel
	formats        []string
	cursor         int
	selectedFormat string
	selectedTopic  string
}

// NewFormatModel creates a new format model
func NewFormatModel(base BaseModel) *FormatModel {
	return &FormatModel{
		BaseModel: base,
		formats:   []string{ContentFormatBlogArticle, ContentFormatTwitterThread, ContentFormatLinkedInPost},
		cursor:    0,
	}
}

func (m *FormatModel) Init() tea.Cmd {
	return nil
}

func (m *FormatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.formats)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			if len(m.formats) > 0 {
				m.cursor = len(m.formats) - 1
			}
		case "enter":
			if len(m.formats) > 0 {
				m.selectedFormat = m.formats[m.cursor]
				return m, func() tea.Msg { return NextMsg{} }
			}
		case "escape":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m *FormatModel) View() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	header := titleStyle.Render("📄 Select Content Format")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Topic: %s", m.selectedTopic))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)
	
	var formatRows []string
	for i, format := range m.formats {
		isSelected := i == m.cursor
		
		cursor := "  "
		if isSelected {
			cursor = "▶ "
		}
		
		var formatText string
		if isSelected {
			formatText = selectedSubjectStyle.Render(format)
		} else {
			formatText = subjectStyle.Render(format)
		}
		
		var description string
		switch format {
		case ContentFormatBlogArticle:
			description = ContentFormatBlogArticleDesc
		case ContentFormatTwitterThread:
			description = ContentFormatTwitterThreadDesc
		case ContentFormatLinkedInPost:
			description = ContentFormatLinkedInPostDesc
		}
		
		firstLine := fmt.Sprintf("%s%s", cursor, formatText)
		secondLine := fmt.Sprintf("  %s", authorStyle.Render(description))
		
		rowContent := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
		
		if isSelected {
			row := selectedCommitRowStyle.Width(96).Align(lipgloss.Left).Render(rowContent)
			formatRows = append(formatRows, row)
		} else {
			row := commitRowStyle.Render(rowContent)
			formatRows = append(formatRows, row)
		}
	}
	
	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, formatRows...))
	
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("select"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.formats)))
	
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", backHelp, " • ", quitHelp)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 10),
		position,
	)
	statusBar := statusBarStyle.Render(statusContent)
	
	main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, content, statusBar)
	return appStyle.Render(main)
}

// SetSelectedTopic sets the selected topic
func (m *FormatModel) SetSelectedTopic(topic string) {
	m.selectedTopic = topic
}

// GetSelectedFormat returns the selected format
func (m *FormatModel) GetSelectedFormat() string {
	return m.selectedFormat
}