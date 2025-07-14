package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewState int

const (
	ListingView ViewState = iota
	TopicSelectionView
	FormatSelectionView
	ContentCreationView
)

type flashTimerMsg struct{}

// mockLLMProvider provides mock responses when no API key is available
type mockLLMProvider struct{}

func (m *mockLLMProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return m.GenerateContentWithSystemPrompt(ctx, "", prompt)
}

func (m *mockLLMProvider) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Return mock topics based on the prompt content
	mockTopics := []string{
		"Implementing modern Go patterns and best practices",
		"Building terminal user interfaces with Bubble Tea",
		"Git repository analysis and commit processing",
		"Error handling and robust software design",
		"API integration and external service communication",
	}
	
	result := ""
	for _, topic := range mockTopics {
		result += topic + "\n"
	}
	
	return result, nil
}

type model struct {
	currentView     ViewState
	commits         []core.Commit
	currentPage     int
	perPage         int
	totalCommits    int
	repoPath        string
	errorMsg        string
	cursor          int
	viewport        int
	maxViewport     int
	selectedCommits map[int]bool // tracks selected commit indices
	selectionMode   bool         // true when in visual selection mode
	rangeStart      int          // start index for range selection
	flashLimit      bool         // true when showing red flash for limit reached
	topics          []string     // extracted topics for topic selection view
	llmProvider     core.LLMProvider // LLM provider for content generation
	llmProviderType string           // type of LLM provider being used
	topicCursor     int          // cursor for topic selection
	selectedTopic   string       // currently selected topic
	formatCursor    int          // cursor for format selection
	formats         []string     // available content formats
	selectedFormat  string       // currently selected format
	promptText      string       // user's prompt input
	generatedContent string      // generated content from LLM
	isEditingPrompt bool         // whether user is editing the prompt
}

func initialModel() model {
	cwd, _ := os.Getwd()
	gitRoot, isGit, _ := core.GetGitDirectory(cwd)
	
	// Initialize LLM provider - prefer Claude CLI, fallback to API, then mock
	var llmProvider core.LLMProvider
	var llmProviderType string
	if core.IsClaudeCLIAvailable() {
		if cliClient, err := core.NewClaudeCLIClient(); err == nil {
			llmProvider = cliClient
			llmProviderType = "Claude CLI"
		} else {
			// Fallback to API if CLI initialization fails
			if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
				llmProvider = core.NewClaudeClient(apiKey)
				llmProviderType = "Claude API"
			} else {
				llmProvider = &mockLLMProvider{}
				llmProviderType = "Mock"
			}
		}
	} else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		llmProvider = core.NewClaudeClient(apiKey)
		llmProviderType = "Claude API"
	} else {
		// Create a mock provider for when no Claude CLI or API key is available
		llmProvider = &mockLLMProvider{}
		llmProviderType = "Mock"
	}
	
	m := model{
		currentView:     ListingView,
		currentPage:     1,
		perPage:         20,
		repoPath:        gitRoot,
		cursor:          0,
		viewport:        0,
		llmProvider:     llmProvider,
		llmProviderType: llmProviderType,
		maxViewport:     8,
		selectedCommits: make(map[int]bool),
		selectionMode:   false,
		rangeStart:      -1,
		flashLimit:      false,
		formats:         []string{"Blog Article", "Twitter Thread"},
		formatCursor:    0,
		promptText:      "",
		generatedContent: "",
		isEditingPrompt: true,
	}
	
	if !isGit {
		m.errorMsg = "Not in a git repository"
		return m
	}
	
	m.loadCommits()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flashTimerMsg:
		m.flashLimit = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.currentView == ListingView {
				if m.cursor > 0 {
					m.cursor--
					// Ensure cursor stays visible in viewport
					if m.cursor < m.viewport {
						m.viewport = m.cursor
					}
				}
			} else if m.currentView == TopicSelectionView {
				if m.topicCursor > 0 {
					m.topicCursor--
				}
			} else if m.currentView == FormatSelectionView {
				if m.formatCursor > 0 {
					m.formatCursor--
				}
			}
		case "down", "j":
			if m.currentView == ListingView {
				if m.cursor < len(m.commits)-1 {
					m.cursor++
					// Ensure cursor stays visible in viewport
					if m.cursor >= m.viewport+m.maxViewport {
						m.viewport = m.cursor - m.maxViewport + 1
					}
				}
			} else if m.currentView == TopicSelectionView {
				if m.topicCursor < len(m.topics)-1 {
					m.topicCursor++
				}
			} else if m.currentView == FormatSelectionView {
				if m.formatCursor < len(m.formats)-1 {
					m.formatCursor++
				}
			}
		case "home", "g":
			if m.currentView == ListingView {
				m.cursor = 0
				m.viewport = 0
			} else if m.currentView == TopicSelectionView {
				m.topicCursor = 0
			} else if m.currentView == FormatSelectionView {
				m.formatCursor = 0
			}
		case "end", "G":
			if m.currentView == ListingView {
				if len(m.commits) > 0 {
					m.cursor = len(m.commits) - 1
					// Adjust viewport to show the last item
					if len(m.commits) > m.maxViewport {
						m.viewport = len(m.commits) - m.maxViewport
					} else {
						m.viewport = 0
					}
				}
			} else if m.currentView == TopicSelectionView {
				if len(m.topics) > 0 {
					m.topicCursor = len(m.topics) - 1
				}
			} else if m.currentView == FormatSelectionView {
				if len(m.formats) > 0 {
					m.formatCursor = len(m.formats) - 1
				}
			}
		case "v":
			// Toggle selection for current commit (only in ListingView)
			if m.currentView == ListingView {
				if len(m.selectedCommits) < 5 || m.selectedCommits[m.cursor] {
					if m.selectedCommits[m.cursor] {
						delete(m.selectedCommits, m.cursor)
					} else {
						m.selectedCommits[m.cursor] = true
					}
				} else {
					// Flash the limit indicator
					m.flashLimit = true
					return m, tea.Tick(time.Millisecond*300, func(t time.Time) tea.Msg {
						return flashTimerMsg{}
					})
				}
			}
		case "V":
			// Range selection mode (only in ListingView)
			if m.currentView == ListingView {
				if !m.selectionMode {
					// Start range selection
					m.selectionMode = true
					m.rangeStart = m.cursor
					m.selectedCommits[m.cursor] = true
				} else {
					// Complete range selection
					start := m.rangeStart
					end := m.cursor
					if start > end {
						start, end = end, start
					}
					
					// Check if we can select the range without exceeding limit
					rangeSize := end - start + 1
					if len(m.selectedCommits)+rangeSize <= 5 {
						for i := start; i <= end; i++ {
							m.selectedCommits[i] = true
						}
					} else {
						// Flash the limit indicator
						m.flashLimit = true
						m.selectionMode = false
						m.rangeStart = -1
						return m, tea.Tick(time.Millisecond*300, func(t time.Time) tea.Msg {
							return flashTimerMsg{}
						})
					}
					m.selectionMode = false
					m.rangeStart = -1
				}
			}
		case "escape":
			if m.currentView == ContentCreationView {
				// Go back to format selection view
				m.currentView = FormatSelectionView
				m.errorMsg = ""
			} else if m.currentView == FormatSelectionView {
				// Go back to topic selection view
				m.currentView = TopicSelectionView
				m.errorMsg = ""
			} else if m.currentView == TopicSelectionView {
				// Go back to listing view
				m.currentView = ListingView
				m.errorMsg = ""
			} else if m.currentView == ListingView {
				// Clear selection mode and all selections
				m.selectionMode = false
				m.rangeStart = -1
				m.selectedCommits = make(map[int]bool)
			}
		case "enter":
			if m.currentView == TopicSelectionView && len(m.topics) > 0 {
				// Select topic and move to format selection
				m.selectedTopic = m.topics[m.topicCursor]
				m.currentView = FormatSelectionView
				m.formatCursor = 0
			} else if m.currentView == FormatSelectionView && len(m.formats) > 0 {
				// Select format and move to content creation view
				m.selectedFormat = m.formats[m.formatCursor]
				m.currentView = ContentCreationView
				m.isEditingPrompt = true
				m.promptText = fmt.Sprintf("Create a %s about %s", strings.ToLower(m.selectedFormat), m.selectedTopic)
			}
		case "d":
			// Deselect current commit (only in ListingView)
			if m.currentView == ListingView && m.selectedCommits[m.cursor] {
				delete(m.selectedCommits, m.cursor)
			}
		case "n", "N":
			// Move to next step if at least one commit is selected
			if len(m.selectedCommits) > 0 && m.currentView == ListingView {
				return m.transitionToTopicSelection()
			}
		case "backspace":
			// Handle backspace in content creation view
			if m.currentView == ContentCreationView && m.isEditingPrompt && len(m.promptText) > 0 {
				m.promptText = m.promptText[:len(m.promptText)-1]
			}
		case "ctrl+enter":
			// Generate content with LLM
			if m.currentView == ContentCreationView && m.promptText != "" {
				return m.generateContent()
			}
		default:
			// Handle text input in content creation view
			if m.currentView == ContentCreationView && m.isEditingPrompt {
				if len(msg.String()) == 1 {
					m.promptText += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case ListingView:
		return m.renderListingView()
	case TopicSelectionView:
		return m.renderTopicSelectionView()
	case FormatSelectionView:
		return m.renderFormatSelectionView()
	case ContentCreationView:
		return m.renderContentCreationView()
	default:
		return "Unknown view"
	}
}

func (m *model) loadCommits() {
	page, err := core.GetCommitLogs(m.repoPath, m.perPage, m.currentPage)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error loading commits: %v", err)
		return
	}
	
	m.commits = page.Commits
	m.totalCommits = page.Total
	m.errorMsg = ""
}

func (m model) renderListingView() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	if len(m.commits) == 0 {
		emptyContent := emptyStyle.Render("📭 No commits found in this repository")
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Center, emptyContent, helpText))
	}
	
	// Create header
	header := m.renderHeader()
	
	// Create content
	content := m.renderCommitList()
	
	// Create status bar
	statusBar := m.renderStatusBar()
	
	// Combine all sections
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)
	
	return appStyle.Render(main)
}

func (m model) renderHeader() string {
	title := titleStyle.Render("✨ CommitLore")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Page %d • %d commits total", m.currentPage, m.totalCommits))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
	
	// Create header with full width background
	headerWithBg := headerStyle.
		Width(100).
		Align(lipgloss.Left).
		Render(headerContent)
	
	return headerWithBg
}

func (m model) renderCommitList() string {
	// Calculate visible range
	start := m.viewport
	end := start + m.maxViewport
	if end > len(m.commits) {
		end = len(m.commits)
	}
	if start < 0 {
		start = 0
	}
	
	var rows []string
	
	for i := start; i < end; i++ {
		commit := m.commits[i]
		isSelected := i == m.cursor
		isMultiSelected := m.selectedCommits[i]
		isInRange := m.selectionMode && ((m.rangeStart <= i && i <= m.cursor) || (m.cursor <= i && i <= m.rangeStart))
		
		row := m.renderCommitRow(commit, isSelected, isMultiSelected, isInRange)
		rows = append(rows, row)
	}
	
	// Add scroll indicators if needed
	var scrollIndicators []string
	if m.viewport > 0 {
		scrollIndicators = append(scrollIndicators, scrollIndicatorStyle.Render("↑ More above"))
	}
	if end < len(m.commits) {
		scrollIndicators = append(scrollIndicators, scrollIndicatorStyle.Render("↓ More below"))
	}
	
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if len(scrollIndicators) > 0 {
		indicators := lipgloss.JoinVertical(lipgloss.Right, scrollIndicators...)
		content = lipgloss.JoinVertical(lipgloss.Left, content, indicators)
	}
	
	return contentStyle.Render(content)
}

func (m model) renderCommitRow(commit core.Commit, isSelected bool, isMultiSelected bool, isInRange bool) string {
	// Truncate text to fit nicely
	subject := commit.Subject
	if len(subject) > 70 {
		subject = subject[:67] + "..."
	}
	
	author := commit.Author
	if len(author) > 20 {
		author = author[:17] + "..."
	}
	
	hash := commit.Hash[:7]
	date := commit.Date.Format("Jan 02, 15:04")
	
	// Cursor and selection indicators
	cursor := "  "
	selectionIndicator := ""
	
	if isSelected {
		cursor = "▶ "
	}
	
	if isMultiSelected {
		selectionIndicator = "✓ "
	} else if isInRange {
		selectionIndicator = "~ "
	}
	
	// Determine styling based on state priority
	var style lipgloss.Style
	var hashText, subjectText, authorText, dateText string
	var needsFullWidth bool
	
	if isSelected {
		// Current cursor position (highest priority)
		style = selectedCommitRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else if isInRange {
		// Range selection preview
		style = rangeSelectionRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else if isMultiSelected {
		// Multi-selected items
		style = multiSelectedCommitRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else {
		// Regular row
		style = commitRowStyle
		needsFullWidth = false
		hashText = hashStyle.Render(hash)
		subjectText = subjectStyle.Render(subject)
		authorText = authorStyle.Render(author)
		dateText = dateStyle.Render(date)
	}
	
	firstLine := fmt.Sprintf("%s%s%s %s", cursor, selectionIndicator, hashText, subjectText)
	secondLine := fmt.Sprintf("  %s • %s", authorText, dateText)
	
	rowContent := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
	
	// Apply full width background if needed
	if needsFullWidth {
		return style.
			Width(96).
			Align(lipgloss.Left).
			Render(rowContent)
	}
	
	return style.Render(rowContent)
}

func (m model) calculateTokensForSelection() int {
	if len(m.selectedCommits) == 0 {
		return 0
	}
	
	totalTokens := 0
	for index := range m.selectedCommits {
		if index < len(m.commits) {
			commit := m.commits[index]
			diff, err := core.GetCommitDiff(m.repoPath, commit.Hash)
			if err == nil {
				tokens := core.EstimateTokenCount(string(diff))
				totalTokens += tokens
			}
		}
	}
	
	return totalTokens
}

func (m model) renderStatusBar() string {
	// Create help sections
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("v"), helpDescStyle.Render("select"))
	rangeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("V"), helpDescStyle.Render("range"))
	nextHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("N"), helpDescStyle.Render("next"))
	clearHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("clear"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	// Selection counter and token count
	selectionCount := len(m.selectedCommits)
	selectionText := ""
	if selectionCount > 0 {
		style := positionStyle
		if m.flashLimit {
			style = flashStyle
		}
		
		tokenCount := m.calculateTokensForSelection()
		tokenText := core.FormatTokenCount(tokenCount)
		
		selectionText = fmt.Sprintf(" • %s • %s • %s", 
			style.Render(fmt.Sprintf("%d/5 selected", selectionCount)),
			positionStyle.Render(fmt.Sprintf("Tokens: 🪙 %s", tokenText)),
			positionStyle.Render(fmt.Sprintf("Provider: %s", m.llmProviderType)))
	}
	
	// Selection mode indicator
	modeText := ""
	if m.selectionMode {
		modeText = fmt.Sprintf(" • %s", helpKeyStyle.Render("RANGE MODE"))
	}
	
	// Position indicator
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits)))
	
	// Combine help text
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", rangeHelp, " • ", nextHelp, " • ", clearHelp, " • ", quitHelp)
	
	// Create status bar with help on left and position/selection on right
	rightSide := fmt.Sprintf("%s%s%s", position, selectionText, modeText)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 10), // Spacer
		rightSide,
	)
	
	return statusBarStyle.Render(statusContent)
}

func (m model) transitionToTopicSelection() (tea.Model, tea.Cmd) {
	// Extract topics from selected commits
	topics, err := m.extractTopicsFromSelection()
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error extracting topics: %v", err)
		return m, nil
	}
	
	m.currentView = TopicSelectionView
	m.topics = topics
	m.topicCursor = 0
	return m, nil
}

func (m model) extractTopicsFromSelection() ([]string, error) {
	// Get the selected commits and their diffs
	var selectedCommitHashes []string
	for index := range m.selectedCommits {
		if index < len(m.commits) {
			selectedCommitHashes = append(selectedCommitHashes, m.commits[index].Hash)
		}
	}
	
	// Get changesets for the selected commits
	changesets := make([]core.Changeset, 0, len(selectedCommitHashes))
	for _, hash := range selectedCommitHashes {
		changeset, err := core.GetChangesForCommit(m.repoPath, hash)
		if err != nil {
			return nil, err
		}
		changesets = append(changesets, changeset)
	}
	
	// Extract topics using LLM
	topics, err := core.ExtractTopics(m.llmProvider, changesets)
	if err != nil {
		return nil, err
	}
	
	return topics, nil
}

func (m model) renderTopicSelectionView() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	// Create header
	header := titleStyle.Render("📝 Select Topic for Content Creation")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Choose from %d extracted topics", len(m.topics)))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.
		Width(100).
		Align(lipgloss.Left).
		Render(headerContent)
	
	// Create topic list
	var topicRows []string
	for i, topic := range m.topics {
		isSelected := i == m.topicCursor
		
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
			row = selectedCommitRowStyle.
				Width(96).
				Align(lipgloss.Left).
				Render(row)
		} else {
			row = commitRowStyle.Render(row)
		}
		
		topicRows = append(topicRows, row)
	}
	
	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, topicRows...))
	
	// Create status bar
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("select"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.topicCursor+1, len(m.topics)))
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
	
	// Combine all sections
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		headerWithBg,
		content,
		statusBar,
	)
	
	return appStyle.Render(main)
}

func (m model) renderFormatSelectionView() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	// Create header
	header := titleStyle.Render("📄 Select Content Format")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Topic: %s", m.selectedTopic))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.
		Width(100).
		Align(lipgloss.Left).
		Render(headerContent)
	
	// Create format list
	var formatRows []string
	for i, format := range m.formats {
		isSelected := i == m.formatCursor
		
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
		
		// Add format descriptions
		var description string
		switch format {
		case "Blog Article":
			description = "Long-form technical article suitable for dev.to, Medium, or personal blog"
		case "Twitter Thread":
			description = "Engaging tweet series optimized for Twitter's format and audience"
		}
		
		firstLine := fmt.Sprintf("%s%s", cursor, formatText)
		secondLine := fmt.Sprintf("  %s", authorStyle.Render(description))
		
		rowContent := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
		
		if isSelected {
			row := selectedCommitRowStyle.
				Width(96).
				Align(lipgloss.Left).
				Render(rowContent)
			formatRows = append(formatRows, row)
		} else {
			row := commitRowStyle.Render(rowContent)
			formatRows = append(formatRows, row)
		}
	}
	
	content := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, formatRows...))
	
	// Create status bar
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("enter"), helpDescStyle.Render("select"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.formatCursor+1, len(m.formats)))
	
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", backHelp, " • ", quitHelp)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 10),
		position,
	)
	statusBar := statusBarStyle.Render(statusContent)
	
	// Combine all sections
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		headerWithBg,
		content,
		statusBar,
	)
	
	return appStyle.Render(main)
}

func (m model) renderContentCreationView() string {
	if m.errorMsg != "" {
		errorContent := errorStyle.Render(fmt.Sprintf("⚠ Error: %s", m.errorMsg))
		helpText := helpDescStyle.Render("Press 'q' or Ctrl+C to quit • 'esc' to go back")
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, errorContent, helpText))
	}
	
	// Create header
	header := titleStyle.Render("✍️ Content Creation")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Topic: %s • Format: %s", m.selectedTopic, m.selectedFormat))
	
	headerContent := lipgloss.JoinVertical(lipgloss.Left, header, subtitle)
	headerWithBg := headerStyle.
		Width(100).
		Align(lipgloss.Left).
		Render(headerContent)
	
	// Create two-column layout
	leftWidth := 48
	rightWidth := 48
	
	// Left panel - Prompt input
	promptTitle := subjectStyle.Render("📝 Prompt Instructions")
	promptBox := commitRowStyle.
		Width(leftWidth).
		Height(10).
		Padding(1).
		Render(m.promptText + "█") // Add cursor
	
	leftPanel := lipgloss.JoinVertical(lipgloss.Left, promptTitle, promptBox)
	
	// Right panel - Generated content
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
	
	// Combine panels
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	
	// Create status bar
	typeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("type"), helpDescStyle.Render("edit prompt"))
	generateHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("ctrl+enter"), helpDescStyle.Render("generate"))
	backHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("back"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, typeHelp, " • ", generateHelp, " • ", backHelp, " • ", quitHelp)
	statusBar := statusBarStyle.Render(helpText)
	
	// Combine all sections
	main := lipgloss.JoinVertical(
		lipgloss.Left,
		headerWithBg,
		content,
		statusBar,
	)
	
	return appStyle.Render(main)
}

func (m model) generateContent() (tea.Model, tea.Cmd) {
	// For now, generate mock content - will implement LLM integration later
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

func RunApp() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}