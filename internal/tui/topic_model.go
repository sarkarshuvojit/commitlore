package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
)

// TopicModel handles the topic selection view
type TopicModel struct {
	BaseModel
	topics        []string
	cursor        int
	selectedTopic string
	asyncWrapper  *llm.AsyncLLMWrapper
	isExtracting  bool
	extractionStartTime time.Time
	hourglassFrame int
}

// NewTopicModel creates a new topic model
func NewTopicModel(base BaseModel) *TopicModel {
	// Create async wrapper with 60 second timeout
	var asyncWrapper *llm.AsyncLLMWrapper
	if base.llmProvider != nil {
		asyncWrapper = llm.NewAsyncLLMWrapper(base.llmProvider, 120*time.Second)
	}

	return &TopicModel{
		BaseModel:    base,
		topics:       []string{},
		cursor:       0,
		asyncWrapper: asyncWrapper,
		isExtracting: false,
	}
}

func (m *TopicModel) Init() tea.Cmd {
	return nil
}

func (m *TopicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if m.isExtracting {
			m.hourglassFrame = (m.hourglassFrame + 1) % 4
			return m, doTick()
		}
		return m, nil
	case llm.LLMResponseMsg:
		m.isExtracting = false
		if msg.Error != "" {
			m.errorMsg = msg.Error
			m.topics = []string{}
		} else {
			m.errorMsg = ""
			// Parse topics from response (assuming comma-separated)
			topics := strings.Split(msg.Content, ",")
			for i, topic := range topics {
				topics[i] = strings.TrimSpace(topic)
			}
			m.SetTopics(topics)
		}
		return m, nil
	case tea.KeyMsg:
		// Don't allow input while extracting topics
		if m.isExtracting {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.topics)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			if len(m.topics) > 0 {
				m.cursor = len(m.topics) - 1
			}
		case "enter":
			if len(m.topics) > 0 {
				m.selectedTopic = m.topics[m.cursor]
				return m, func() tea.Msg { return NextMsg{} }
			}
		case "escape":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m *TopicModel) View() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}

	if m.isExtracting {
		header := titleStyle.Render("📝 Extracting Topics")
		hourglass := m.getHourglassFrame()
		elapsedTime := m.getElapsedTime()
		subtitle := subtitleStyle.Render(fmt.Sprintf("🤖 Analyzing commits with AI... %s (%s)", hourglass, elapsedTime))
		headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
		headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)

		generatingHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render(hourglass), helpDescStyle.Render("extracting topics..."))
		quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
		helpText := lipgloss.JoinHorizontal(lipgloss.Left, generatingHelp, " • ", quitHelp)
		statusBar := statusBarStyle.Render(helpText)

		main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, statusBar)
		return appStyle.Render(main)
	}

	header := titleStyle.Render("📝 Select Topic for Content Creation")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Choose from %d extracted topics", len(m.topics)))

	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)

	var topicRows []string
	for i, topic := range m.topics {
		isSelected := i == m.cursor

		cursor := "  "
		if isSelected {
			cursor = "▶ "
		}

		var topicText string
		if isSelected {
			topicText = selectedSubjectStyle.Render(topic)
		} else {
			topicText = subjectStyle.Render(topic)
		}

		row := fmt.Sprintf("%s%s", cursor, topicText)

		if isSelected {
			row = selectedCommitRowStyle.Width(96).Align(lipgloss.Left).Render(row)
		} else {
			row = commitRowStyle.Render(row)
		}

		topicRows = append(topicRows, row)
	}

	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, topicRows...))

	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("select"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))

	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.topics)))
	providerInfo := positionStyle.Render(fmt.Sprintf("Provider: %s", m.llmProviderType))

	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", backHelp, " • ", quitHelp)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 5),
		providerInfo,
		strings.Repeat(" ", 5),
		position,
	)
	statusBar := statusBarStyle.Render(statusContent)

	main := lipgloss.JoinVertical(lipgloss.Left, headerWithBg, content, statusBar)
	return appStyle.Render(main)
}

// SetTopics sets the topics for the model
func (m *TopicModel) SetTopics(topics []string) {
	m.topics = topics
	m.cursor = 0
}

// GetSelectedTopic returns the selected topic
func (m *TopicModel) GetSelectedTopic() string {
	return m.selectedTopic
}

// ExtractTopics extracts topics from selected commits using async LLM calls
func (m *TopicModel) ExtractTopics(commits []core.Commit, selectedCommits map[int]bool) tea.Cmd {
	logger := core.GetLogger()
	logger.Info("Starting topic extraction", "selected_commits", len(selectedCommits), "provider", m.llmProviderType)

	if m.asyncWrapper == nil {
		m.errorMsg = "LLM provider not configured"
		logger.Error("LLM provider not configured for topic extraction", "provider", m.llmProviderType)
		return nil
	}

	m.isExtracting = true
	m.errorMsg = ""
	m.topics = []string{}
	m.extractionStartTime = time.Now()
	m.hourglassFrame = 0

	// Create channel for async response
	responseChan := llm.CreateLLMResponseChannel()

	// Get selected commit data
	var selectedCommitHashes []string
	for index := range selectedCommits {
		if index < len(commits) {
			selectedCommitHashes = append(selectedCommitHashes, commits[index].Hash)
		}
	}

	// Build comprehensive changelist data for topic extraction
	var commitDetails []string
	for index := range selectedCommits {
		if index < len(commits) {
			commit := commits[index]
			
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

	systemPrompt := `You are a developer story assistant. Your task is to analyze commit changesets including diffs and extract meaningful topics that could be used for creating developer content like blog posts, social media posts, or technical articles.

Analyze the provided commits with their full changesets and extract 3-5 relevant topics that would be interesting for developer content creation. Focus on:
- Technical concepts and implementations revealed in the code changes
- Problem-solving approaches shown in the diffs
- Development practices and patterns demonstrated
- Technology stack and tools used
- Performance improvements or optimizations
- Architectural decisions and refactoring patterns
- Bug fixes and their underlying issues

Return only the topics as a comma-separated list, with no additional text or explanations.`

	userPrompt := fmt.Sprintf(`Analyze these commits with their full changesets and extract meaningful topics for content creation:

%s

Provide 3-5 topics as a comma-separated list.`, strings.Join(commitDetails, "\n"))

	// Start async LLM call
	ctx := context.Background()
	m.asyncWrapper.GenerateContentWithSystemPromptAsync(ctx, systemPrompt, userPrompt, responseChan)

	logger.Info("Started async LLM call for topic extraction", "provider", m.llmProviderType)

	// Return command to wait for response
	return tea.Batch(llm.WaitForLLMResponse(responseChan), doTick())
}

// getHourglassFrame returns the current frame of the hourglass animation
func (m *TopicModel) getHourglassFrame() string {
	frames := []string{"⧖", "⧗", "⧑", "⧒"}
	return frames[m.hourglassFrame]
}

// getElapsedTime returns human-readable elapsed time
func (m *TopicModel) getElapsedTime() string {
	if m.extractionStartTime.IsZero() {
		return ""
	}
	elapsed := time.Since(m.extractionStartTime)
	
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
