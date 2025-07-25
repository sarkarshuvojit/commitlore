package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
)

// ContentGeneratedMsg represents a message sent when content generation is complete
type ContentGeneratedMsg struct {
	Content string
	Error   string
}

// TickMsg represents a tick for animation
type TickMsg struct{}

// Tick command for animation
func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// ContentModel handles the content creation view
type ContentModel struct {
	BaseModel
	selectedTopic    string
	selectedFormat   string
	textarea         textarea.Model
	generatedContent string
	isEditingPrompt  bool
	isGenerating     bool
	viewport         viewport.Model
	showFinalOutput  bool
	asyncWrapper     *llm.AsyncLLMWrapper
	commits          []core.Commit
	selectedCommits  map[int]bool
	generationStartTime time.Time
	hourglassFrame   int
}

// NewContentModel creates a new content model
func NewContentModel(base BaseModel) *ContentModel {
	vp := viewport.New(80, 20)

	// Create async wrapper with 2 minute timeout
	var asyncWrapper *llm.AsyncLLMWrapper
	if base.llmProvider != nil {
		asyncWrapper = llm.NewAsyncLLMWrapper(base.llmProvider, 2*time.Minute)
	}

	// Initialize textarea with proper configuration
	ta := textarea.New()
	ta.SetWidth(94)    // Match the width of the prompt box
	ta.SetHeight(8)    // Use most of the available height
	ta.Placeholder = "Enter your instructions for content generation..."
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false

	return &ContentModel{
		BaseModel:        base,
		textarea:         ta,
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
	case TickMsg:
		if m.isGenerating {
			m.hourglassFrame = (m.hourglassFrame + 1) % 4
			return m, doTick()
		}
		return m, nil
	case llm.LLMResponseMsg:
		m.isGenerating = false
		if msg.Error != "" {
			m.errorMsg = msg.Error
			if !m.showFinalOutput {
				m.generatedContent = ""
			}
		} else {
			m.errorMsg = ""
			m.statusMessage = nil
			// If this is a save success message, show it as status
			if m.showFinalOutput && msg.Content != m.generatedContent {
				// This is a save success message, show it briefly
				m.statusMessage = NewSuccessMessage(msg.Content)
			} else {
				// This is generated content
				m.generatedContent = msg.Content
				m.showFinalOutput = true
				// Wrap text to fit viewport width (94 chars to account for padding)
				wrappedContent := wordwrap.String(msg.Content, 94)
				m.viewport.SetContent(wrappedContent)
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
			m.statusMessage = nil
			// If this is a save success message, show it as status
			if m.showFinalOutput && msg.Content != m.generatedContent {
				// This is a save success message, show it briefly
				m.statusMessage = NewSuccessMessage(msg.Content)
			} else {
				// This is generated content
				m.generatedContent = msg.Content
				m.showFinalOutput = true
				// Wrap text to fit viewport width (94 chars to account for padding)
				wrappedContent := wordwrap.String(msg.Content, 94)
				m.viewport.SetContent(wrappedContent)
			}
		}
		return m, nil
	case tea.KeyMsg:
		// Don't allow input while generating content
		if m.isGenerating {
			return m, nil
		}

		// Handle Enter key specifically - check for plain Enter
		if msg.Type == tea.KeyEnter {
			if msg.String() == "enter" {
				// Plain Enter - trigger content generation
				if m.isEditingPrompt && !m.showFinalOutput {
					m.isGenerating = true
					m.errorMsg = ""
					m.generationStartTime = time.Now()
					m.hourglassFrame = 0
					model, cmd := m.generateContent()
					return model, tea.Batch(cmd, doTick())
				}
			} else {
				// Shift+Enter, Ctrl+Enter, Alt+Enter - pass to textarea for new line
				if m.isEditingPrompt {
					var cmd tea.Cmd
					m.textarea, cmd = m.textarea.Update(msg)
					return m, cmd
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "escape":
			if m.showFinalOutput {
				m.showFinalOutput = false
			} else {
				return m, func() tea.Msg { return BackMsg{} }
			}
		default:
			if m.showFinalOutput {
				// Handle save command when viewing final output
				if (msg.String() == "s" || msg.String() == "S") && m.generatedContent != "" {
					return m, m.saveContent()
				}
				// Handle viewport scrolling
				m.viewport, _ = m.viewport.Update(msg)
			} else if m.isEditingPrompt {
				// Handle textarea updates for all other keys
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m *ContentModel) View() string {
	// Handle error messages (legacy support)
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	// Handle status messages (new system)
	if m.statusMessage != nil {
		statusContent := RenderStatusMessage(m.statusMessage)
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, statusContent, helpText))
	}

	header := titleStyle.Render("✍️ Content Creation")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Topic: %s • Format: %s", m.selectedTopic, m.selectedFormat))

	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)

	if m.showFinalOutput {
		return m.renderFinalOutput(headerWithBg)
	}

	promptTitle := subjectStyle.Render("📝 Your Instructions")
	promptBox := commitRowStyle.
		Width(96).
		Height(10).
		Padding(1).
		Render(m.textarea.View())

	content := lipgloss.JoinVertical(lipgloss.Left, promptTitle, promptBox)

	var helpText string
	if m.isGenerating {
		hourglass := m.getHourglassFrame()
		elapsedTime := m.getElapsedTime()
		generatingHelp := fmt.Sprintf("%s %s (%s)", helpKeyStyle.Render(hourglass), helpDescStyle.Render("generating content..."), elapsedTime)
		backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
		quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
		helpText = lipgloss.JoinHorizontal(lipgloss.Left, generatingHelp, " • ", backHelp, " • ", quitHelp)
	} else {
		typeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("type"), helpDescStyle.Render("edit prompt"))
		newlineHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("shift+enter"), helpDescStyle.Render("new line"))
		generateHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("generate"))
		backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
		quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
		helpText = lipgloss.JoinHorizontal(lipgloss.Left, typeHelp, " • ", newlineHelp, " • ", generateHelp, " • ", backHelp, " • ", quitHelp)
	}
	statusBar := statusBarStyle.Render(helpText)

	main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, content, statusBar)
	return appStyle.Render(main)
}

// SetContext sets the topic and format for content generation
func (m *ContentModel) SetContext(topic, format string) {
	m.selectedTopic = topic
	m.selectedFormat = format
	m.textarea.SetValue("")
	m.isEditingPrompt = true
	m.showFinalOutput = false
}

// SetContextWithCommits sets the topic, format, and commit data for content generation
func (m *ContentModel) SetContextWithCommits(topic, format string, commits []core.Commit, selectedCommits map[int]bool) {
	m.selectedTopic = topic
	m.selectedFormat = format
	m.textarea.SetValue("")
	m.isEditingPrompt = true
	m.showFinalOutput = false
	m.commits = commits
	m.selectedCommits = selectedCommits
}

func (m *ContentModel) generateContent() (tea.Model, tea.Cmd) {
	logger := core.GetLogger()
	logger.Info("Starting content generation",
		"topic", m.selectedTopic,
		"format", m.selectedFormat,
		"prompt_length", len(m.textarea.Value()),
		"provider", m.llmProviderType)

	if m.asyncWrapper == nil {
		m.errorMsg = "LLM provider not configured"
		logger.Error("LLM provider not configured for content generation", "provider", m.llmProviderType)
		return m, nil
	}

	m.generatedContent = ""

	// Create channel for async response
	responseChan := llm.CreateLLMResponseChannel()

	// Get the appropriate system prompt based on format
	var systemPrompt string
	switch m.selectedFormat {
	case ContentFormatTwitterThread:
		systemPrompt = llm.TwitterThreadPrompt
	case ContentFormatBlogArticle:
		systemPrompt = llm.BlogPostPrompt
	case ContentFormatLinkedInPost:
		systemPrompt = llm.LinkedInPostPrompt
	default:
		systemPrompt = llm.ContentGenerationPrompt
	}

	// Build comprehensive changelist data for content generation
	var changelistData string
	if m.selectedCommits != nil && len(m.selectedCommits) > 0 {
		var commitDetails []string
		for index := range m.selectedCommits {
			if index < len(m.commits) {
				commit := m.commits[index]
				
				// Get changelist data for this commit
				changeset, err := core.GetChangesForCommit(m.repoPath, commit.Hash)
				if err != nil {
					logger.Error("Failed to get changeset for commit", "hash", commit.Hash, "error", err, "provider", m.llmProviderType)
					// Fall back to basic commit info
					detail := fmt.Sprintf("- %s: %s", commit.Hash[:8], commit.Subject)
					commitDetails = append(commitDetails, detail)
					continue
				}

				// Create detailed commit information with changelist
				detail := fmt.Sprintf(`Commit: %s
Author: %s
Date: %s  
Subject: %s
Body: %s
Files Changed: %s
Diff:
%s

---`, 
					commit.Hash[:8], 
					changeset.Author, 
					changeset.Date.Format("2006-01-02 15:04:05"),
					changeset.Subject,
					changeset.Body,
					strings.Join(changeset.Files, ", "),
					changeset.Diff)
				
				commitDetails = append(commitDetails, detail)
			}
		}
		changelistData = strings.Join(commitDetails, "\n")
	}

	// Use the user's prompt text as the user prompt, including changelist data
	userPrompt := fmt.Sprintf(`Create %s content about: %s

Please ensure the content is:
- Technically accurate and up-to-date
- Engaging and valuable to developers
- Properly formatted for the target platform
- Includes relevant code examples where applicable
- Optimized for engagement and sharing
- Instead of being generic, tries to actively target the content based on the actual code changes shown below

Additional user instructions: %s

Based on the following commit changesets from the selected commits:

%s`, m.selectedFormat, m.selectedTopic, m.textarea.Value(), changelistData)

	// Start async LLM call
	ctx := context.Background()
	m.asyncWrapper.GenerateContentWithSystemPromptAsync(ctx, systemPrompt, userPrompt, responseChan)

	logger.Info("Started async LLM call for content generation", "provider", m.llmProviderType)

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

// getHourglassFrame returns the current frame of the hourglass animation
func (m *ContentModel) getHourglassFrame() string {
	frames := []string{"⧖", "⧗", "⧑", "⧒"}
	return frames[m.hourglassFrame]
}

// getElapsedTime returns human-readable elapsed time
func (m *ContentModel) getElapsedTime() string {
	if m.generationStartTime.IsZero() {
		return ""
	}
	elapsed := time.Since(m.generationStartTime)
	
	if elapsed < time.Second {
		return fmt.Sprintf("%.0fms", float64(elapsed.Nanoseconds())/1e6)
	} else if elapsed < time.Minute {
		return fmt.Sprintf("%.0fs", elapsed.Seconds())
	} else {
		minutes := int(elapsed.Minutes())
		seconds := int(elapsed.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
}
