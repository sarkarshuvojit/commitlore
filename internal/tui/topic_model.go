package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	"github.com/sarkarshuvojit/commitlore/internal/core/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// TopicModel handles the topic selection view
type TopicModel struct {
	BaseModel
	topics      []string
	cursor      int
	selectedTopic string
}

// NewTopicModel creates a new topic model
func NewTopicModel(base BaseModel) *TopicModel {
	return &TopicModel{
		BaseModel: base,
		topics:    []string{},
		cursor:    0,
	}
}

func (m *TopicModel) Init() tea.Cmd {
	return nil
}

func (m *TopicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

// ExtractTopics extracts topics from selected commits
func (m *TopicModel) ExtractTopics(commits []core.Commit, selectedCommits map[int]bool) error {
	var selectedCommitHashes []string
	for index := range selectedCommits {
		if index < len(commits) {
			selectedCommitHashes = append(selectedCommitHashes, commits[index].Hash)
		}
	}
	
	changesets := make([]core.Changeset, 0, len(selectedCommitHashes))
	for _, hash := range selectedCommitHashes {
		changeset, err := core.GetChangesForCommit(m.repoPath, hash)
		if err != nil {
			return err
		}
		changesets = append(changesets, changeset)
	}
	
	// Convert core.Changeset to llm.Changeset for topic extraction
	llmChangesets := make([]llm.Changeset, 0, len(changesets))
	for _, cs := range changesets {
		llmChangeset := llm.Changeset{
			CommitHash: cs.CommitHash,
			Author:     cs.Author,
			Date:       cs.Date,
			Subject:    cs.Subject,
			Body:       cs.Body,
			Files:      cs.Files,
			Diff:       cs.Diff,
		}
		llmChangesets = append(llmChangesets, llmChangeset)
	}
	
	topics, err := llm.ExtractTopics(m.llmProvider, llmChangesets)
	if err != nil {
		return err
	}
	
	m.SetTopics(topics)
	return nil
}