package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
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
	viewport         viewport.Model
	showFinalOutput  bool
	asyncWrapper     *llm.AsyncLLMWrapper
}

// NewContentModel creates a new content model
func NewContentModel(base BaseModel) *ContentModel {
	vp := viewport.New(80, 20)
	
	// Create async wrapper with 60 second timeout
	var asyncWrapper *llm.AsyncLLMWrapper
	if base.llmProvider != nil {
		asyncWrapper = llm.NewAsyncLLMWrapper(base.llmProvider, 60*time.Second)
	}
	
	return &ContentModel{
		BaseModel:        base,
		promptText:       "",
		generatedContent: "",
		isEditingPrompt:  true,
		isGenerating:     false,
		viewport:         vp,
		showFinalOutput:  false,
		asyncWrapper:     asyncWrapper,
	}
}

func (m *ContentModel) Init() tea.Cmd {
	return nil
}

func (m *ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case llm.LLMResponseMsg:
		m.isGenerating = false
		if msg.Error != "" {
			m.errorMsg = msg.Error
			if !m.showFinalOutput {
				m.generatedContent = ""
			}
		} else {
			m.errorMsg = ""
			// If this is a save success message, show it as status
			if m.showFinalOutput && msg.Content != m.generatedContent {
				// This is a save success message, show it briefly
				m.errorMsg = msg.Content
			} else {
				// This is generated content
				m.generatedContent = msg.Content
				m.showFinalOutput = true
				m.viewport.SetContent(msg.Content)
			}
		}
		return m, nil
	case ContentGeneratedMsg:
		m.isGenerating = false
		if msg.Error != "" {
			m.errorMsg = msg.Error
			if !m.showFinalOutput {
				m.generatedContent = ""
			}
		} else {
			m.errorMsg = ""
			// If this is a save success message, show it as status
			if m.showFinalOutput && msg.Content != m.generatedContent {
				// This is a save success message, show it briefly
				m.errorMsg = msg.Content
			} else {
				// This is generated content
				m.generatedContent = msg.Content
				m.showFinalOutput = true
				m.viewport.SetContent(msg.Content)
			}
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
			if m.showFinalOutput {
				m.showFinalOutput = false
			} else {
				return m, func() tea.Msg { return BackMsg{} }
			}
		case "s", "S":
			if m.showFinalOutput && m.generatedContent != "" {
				return m, m.saveContent()
			}
		default:
			if m.showFinalOutput {
				m.viewport, _ = m.viewport.Update(msg)
			} else if m.isEditingPrompt && len(msg.String()) == 1 {
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

	if m.showFinalOutput {
		return m.renderFinalOutput(headerWithBg)
	}

	leftWidth := 48
	rightWidth := 48

	promptTitle := subjectStyle.Render("📝 Your Instructions")
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
	m.promptText = ""
	m.isEditingPrompt = true
	m.showFinalOutput = false
}

func (m *ContentModel) generateContent() (tea.Model, tea.Cmd) {
	logger := core.GetLogger()
	logger.Info("Starting content generation", 
		"topic", m.selectedTopic,
		"format", m.selectedFormat,
		"prompt_length", len(m.promptText))
	
	if m.asyncWrapper == nil {
		m.errorMsg = "LLM provider not configured"
		logger.Error("LLM provider not configured for content generation")
		return m, nil
	}
	
	m.generatedContent = ""
	
	// Create channel for async response
	responseChan := llm.CreateLLMResponseChannel()
	
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

Additional user instructions: %s`, m.selectedFormat, m.selectedTopic, m.promptText)
	
	// Start async LLM call
	ctx := context.Background()
	m.asyncWrapper.GenerateContentWithSystemPromptAsync(ctx, systemPrompt, userPrompt, responseChan)
	
	logger.Info("Started async LLM call for content generation")
	
	// Return command to wait for response
	return m, llm.WaitForLLMResponse(responseChan)
}

// renderFinalOutput renders the final output view with scrollable viewport
func (m *ContentModel) renderFinalOutput(headerWithBg string) string {
	contentTitle := subjectStyle.Render("📄 Generated Content")
	
	// Update viewport dimensions
	m.viewport.Width = 96
	m.viewport.Height = 15
	
	viewportContent := commitRowStyle.
		Width(96).
		Height(15).
		Padding(1).
		Render(m.viewport.View())
	
	content := lipgloss.JoinVertical(lipgloss.Left, contentTitle, viewportContent)
	
	saveHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("S"), helpDescStyle.Render("save to file"))
	scrollHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓"), helpDescStyle.Render("scroll"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, saveHelp, " • ", scrollHelp, " • ", backHelp, " • ", quitHelp)
	
	statusBar := statusBarStyle.Render(helpText)
	
	main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, content, statusBar)
	return appStyle.Render(main)
}

// saveContent saves the generated content to a file
func (m *ContentModel) saveContent() tea.Cmd {
	return func() tea.Msg {
		// Generate filename based on topic and format
		topic := m.sanitizeFilename(m.selectedTopic)
		format := m.sanitizeFilename(m.selectedFormat)
		filename := fmt.Sprintf("%s_%s.txt", topic, format)
		
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return ContentGeneratedMsg{
				Error: fmt.Sprintf("Failed to get current directory: %v", err),
			}
		}
		
		// Create full path
		fullPath := filepath.Join(cwd, filename)
		
		// Write content to file
		err = os.WriteFile(fullPath, []byte(m.generatedContent), 0644)
		if err != nil {
			return ContentGeneratedMsg{
				Error: fmt.Sprintf("Failed to save file: %v", err),
			}
		}
		
		// Return success message (we'll handle this in the Update method)
		return ContentGeneratedMsg{
			Content: fmt.Sprintf("✅ Content saved to: %s", fullPath),
			Error:   "",
		}
	}
}

// sanitizeFilename removes invalid characters from filename
func (m *ContentModel) sanitizeFilename(filename string) string {
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")
	
	// Remove invalid characters
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	filename = reg.ReplaceAllString(filename, "")
	
	// Convert to lowercase
	filename = strings.ToLower(filename)
	
	return filename
}