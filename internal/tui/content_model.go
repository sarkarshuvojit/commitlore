package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// ContentGeneratedMsg represents a message sent when content generation is complete
type ContentGeneratedMsg struct {
	Content string
	Error   string
}

// ContentModel handles the content creation view
type ContentModel struct {
	BaseModel
	selectedTopic    string
	selectedFormat   string
	promptText       string
	generatedContent string
	isEditingPrompt  bool
	isGenerating     bool
}

// NewContentModel creates a new content model
func NewContentModel(base BaseModel) *ContentModel {
	return &ContentModel{
		BaseModel:        base,
		promptText:       "",
		generatedContent: "",
		isEditingPrompt:  true,
		isGenerating:     false,
	}
}

func (m *ContentModel) Init() tea.Cmd {
	return nil
}

func (m *ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ContentGeneratedMsg:
		m.isGenerating = false
		if msg.Error != "" {
			m.errorMsg = msg.Error
			m.generatedContent = ""
		} else {
			m.errorMsg = ""
			m.generatedContent = msg.Content
		}
		return m, nil
	case tea.KeyMsg:
		// Don't allow input while generating content
		if m.isGenerating {
			return m, nil
		}
		
		switch msg.String() {
		case "backspace":
			if m.isEditingPrompt && len(m.promptText) > 0 {
				m.promptText = m.promptText[:len(m.promptText)-1]
			}
		case "enter", "shift+enter":
			if m.promptText != "" && !m.isGenerating {
				m.isGenerating = true
				m.errorMsg = ""
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
		if m.isGenerating {
			contentText = "🤖 Generating content with AI... Please wait..."
		} else {
			contentText = "Generated content will appear here after you provide instructions..."
		}
	}
	
	contentBox := commitRowStyle.
		Width(rightWidth).
		Height(10).
		Padding(1).
		Render(contentText)
	
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, contentTitle, contentBox)
	
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	
	var helpText string
	if m.isGenerating {
		generatingHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("⏳"), helpDescStyle.Render("generating content..."))
		backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
		quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
		helpText = lipgloss.JoinHorizontal(lipgloss.Left, generatingHelp, " • ", backHelp, " • ", quitHelp)
	} else {
		typeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("type"), helpDescStyle.Render("edit prompt"))
		generateHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("generate"))
		backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
		quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
		helpText = lipgloss.JoinHorizontal(lipgloss.Left, typeHelp, " • ", generateHelp, " • ", backHelp, " • ", quitHelp)
	}
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
	logger := core.GetLogger()
	logger.Info("Starting content generation", 
		"topic", m.selectedTopic,
		"format", m.selectedFormat,
		"prompt_length", len(m.promptText))
	
	if m.llmProvider == nil {
		m.errorMsg = "LLM provider not configured"
		logger.Error("LLM provider not configured for content generation")
		return m, nil
	}
	
	m.generatedContent = ""
	
	return m, tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		logger.Info("Calling LLM provider for content generation")
		
		// Get the appropriate system prompt based on format
		var systemPrompt string
		switch m.selectedFormat {
		case "Twitter Thread":
			systemPrompt = llm.TwitterThreadPrompt
		case "Blog Article":
			systemPrompt = llm.BlogPostPrompt
		case "LinkedIn Post":
			systemPrompt = llm.LinkedInPostPrompt
		default:
			systemPrompt = llm.ContentGenerationPrompt
		}
		
		// Use the user's prompt text as the user prompt
		userPrompt := fmt.Sprintf(`Create %s content about: %s

Please ensure the content is:
- Technically accurate and up-to-date
- Engaging and valuable to developers
- Properly formatted for the target platform
- Includes relevant code examples where applicable
- Optimized for engagement and sharing

Additional instructions: %s`, m.selectedFormat, m.selectedTopic, m.promptText)
		
		content, err := m.llmProvider.GenerateContentWithSystemPrompt(ctx, systemPrompt, userPrompt)
		if err != nil {
			logger.Error("Failed to generate content", "error", err)
			return ContentGeneratedMsg{
				Content: "",
				Error:   fmt.Sprintf("Failed to generate content: %v", err),
			}
		}
		
		logger.Info("Content generated successfully", 
			"content_length", len(content),
			"topic", m.selectedTopic,
			"format", m.selectedFormat)
		
		return ContentGeneratedMsg{
			Content: content,
			Error:   "",
		}
	})
}