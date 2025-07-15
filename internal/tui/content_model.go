package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// ContentModel handles the content creation view
type ContentModel struct {
	BaseModel
	selectedTopic    string
	selectedFormat   string
	promptText       string
	generatedContent string
	isEditingPrompt  bool
}

// NewContentModel creates a new content model
func NewContentModel(base BaseModel) *ContentModel {
	return &ContentModel{
		BaseModel:        base,
		promptText:       "",
		generatedContent: "",
		isEditingPrompt:  true,
	}
}

func (m *ContentModel) Init() tea.Cmd {
	return nil
}

func (m *ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.isEditingPrompt && len(m.promptText) > 0 {
				m.promptText = m.promptText[:len(m.promptText)-1]
			}
		case "ctrl+enter":
			if m.promptText != "" {
				return m.generateContent()
			}
		case "escape":
			return m, func() tea.Msg { return BackMsg{} }
		default:
			if m.isEditingPrompt && len(msg.String()) == 1 {
				m.promptText += msg.String()
			}
		}
	}
	return m, nil
}

func (m *ContentModel) View() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	header := titleStyle.Render("✍️ Content Creation")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Topic: %s • Format: %s", m.selectedTopic, m.selectedFormat))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)
	
	leftWidth := 48
	rightWidth := 48
	
	promptTitle := subjectStyle.Render("📝 Prompt Instructions")
	promptBox := commitRowStyle.
		Width(leftWidth).
		Height(10).
		Padding(1).
		Render(m.promptText + "█")
	
	leftPanel := lipgloss.JoinVertical(lipgloss.Left, promptTitle, promptBox)
	
	contentTitle := subjectStyle.Render("📄 Generated Content")
	contentText := m.generatedContent
	if contentText == "" {
		contentText = "Generated content will appear here after you provide instructions..."
	}
	
	contentBox := commitRowStyle.
		Width(rightWidth).
		Height(10).
		Padding(1).
		Render(contentText)
	
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, contentTitle, contentBox)
	
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	
	typeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("type"), helpDescStyle.Render("edit prompt"))
	generateHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("ctrl+enter"), helpDescStyle.Render("generate"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, typeHelp, " • ", generateHelp, " • ", backHelp, " • ", quitHelp)
	statusBar := statusBarStyle.Render(helpText)
	
	main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, content, statusBar)
	return appStyle.Render(main)
}

// SetContext sets the topic and format for content generation
func (m *ContentModel) SetContext(topic, format string) {
	m.selectedTopic = topic
	m.selectedFormat = format
	m.promptText = llm.GetContentCreationPrompt(format, topic)
	m.isEditingPrompt = true
}

func (m *ContentModel) generateContent() (tea.Model, tea.Cmd) {
	content := ""
	
	switch m.selectedFormat {
	case "Blog Article":
		content = fmt.Sprintf(`# %s

## Introduction

In this article, we'll explore the implementation details and lessons learned from recent development work on %s.

## Key Insights

- Understanding the core concepts and patterns
- Implementation challenges and solutions
- Best practices and recommendations

## Technical Details

The implementation involved several key components:

1. **Architecture Design**: Careful consideration of the overall system structure
2. **Error Handling**: Robust error management strategies
3. **Performance**: Optimization techniques and considerations

## Conclusion

Working on %s provided valuable insights into modern development practices and helped reinforce important software engineering principles.

---

*Generated based on: %s*`, 
			m.selectedTopic, 
			strings.ToLower(m.selectedTopic), 
			strings.ToLower(m.selectedTopic),
			m.promptText)
			
	case "Twitter Thread":
		content = fmt.Sprintf(`🧵 Thread: %s

1/5 Just finished working on %s and wanted to share some key insights! 

2/5 The main challenge was understanding how to properly implement the core patterns while maintaining code quality.

3/5 Key takeaways:
• Clean architecture really matters
• Error handling saves you time later
• Testing early prevents headaches

4/5 The implementation taught me valuable lessons about software design and helped me understand why certain patterns exist.

5/5 Overall, working on %s was a great learning experience that reinforced important development principles. What are your thoughts on this approach?

#SoftwareDevelopment #Coding #TechLearning

---
Generated based on: %s`, 
			m.selectedTopic,
			strings.ToLower(m.selectedTopic),
			strings.ToLower(m.selectedTopic),
			m.promptText)
	}
	
	m.generatedContent = content
	return m, nil
}