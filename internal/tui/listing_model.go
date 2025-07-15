package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
)

// ListingModel handles the commit listing view
type ListingModel struct {
	BaseModel
	commits         []core.Commit
	currentPage     int
	perPage         int
	totalCommits    int
	cursor          int
	viewport        int
	maxViewport     int
	selectedCommits map[int]bool
	selectionMode   bool
	rangeStart      int
	flashLimit      bool
}

// NewListingModel creates a new listing model
func NewListingModel(base BaseModel) *ListingModel {
	m := &ListingModel{
		BaseModel:       base,
		currentPage:     1,
		perPage:         100,
		cursor:          0,
		viewport:        0,
		maxViewport:     8,
		selectedCommits: make(map[int]bool),
		selectionMode:   false,
		rangeStart:      -1,
		flashLimit:      false,
	}

	m.loadCommits()
	return m
}

func (m *ListingModel) Init() tea.Cmd {
	return nil
}

func (m *ListingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flashTimerMsg:
		m.flashLimit = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewport {
					m.viewport = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.commits)-1 {
				m.cursor++
				if m.cursor >= m.viewport+m.maxViewport {
					m.viewport = m.cursor - m.maxViewport + 1
				}
			}
		case "home", "g":
			m.cursor = 0
			m.viewport = 0
		case "end", "G":
			if len(m.commits) > 0 {
				m.cursor = len(m.commits) - 1
				if len(m.commits) > m.maxViewport {
					m.viewport = len(m.commits) - m.maxViewport
				} else {
					m.viewport = 0
				}
			}
		case "v":
			if len(m.selectedCommits) < 5 || m.selectedCommits[m.cursor] {
				if m.selectedCommits[m.cursor] {
					delete(m.selectedCommits, m.cursor)
				} else {
					m.selectedCommits[m.cursor] = true
				}
			} else {
				m.flashLimit = true
				return m, tea.Tick(time.Millisecond*300, func(t time.Time) tea.Msg {
					return flashTimerMsg{}
				})
			}
		case "V":
			if !m.selectionMode {
				m.selectionMode = true
				m.rangeStart = m.cursor
				m.selectedCommits[m.cursor] = true
			} else {
				start := m.rangeStart
				end := m.cursor
				if start > end {
					start, end = end, start
				}

				rangeSize := end - start + 1
				if len(m.selectedCommits)+rangeSize <= 5 {
					for i := start; i <= end; i++ {
						m.selectedCommits[i] = true
					}
				} else {
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
		case "d":
			if m.selectedCommits[m.cursor] {
				delete(m.selectedCommits, m.cursor)
			}
		case "escape":
			m.selectionMode = false
			m.rangeStart = -1
			m.selectedCommits = make(map[int]bool)
		case "n", "N":
			if len(m.selectedCommits) > 0 {
				return m, func() tea.Msg { return NextMsg{} }
			}
		}
	}
	return m, nil
}

func (m *ListingModel) View() string {
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

	header := m.renderHeader()
	content := m.renderCommitList()
	statusBar := m.renderStatusBar()

	main := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
	return appStyle.Render(main)
}

func (m *ListingModel) loadCommits() {
	page, err := core.GetCommitLogs(m.repoPath, m.perPage, m.currentPage)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error loading commits: %v", err)
		return
	}

	m.commits = page.Commits
	m.totalCommits = page.Total
	m.errorMsg = ""
}

func (m *ListingModel) renderHeader() string {
	title := titleStyle.Render("✨ CommitLore")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Page %d • %d commits total", m.currentPage, m.totalCommits))

	headerContent := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
	headerWithBg := headerStyle.Width(100).Align(lipgloss.Left).Render(headerContent)

	return headerWithBg
}

func (m *ListingModel) renderCommitList() string {
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

func (m *ListingModel) renderCommitRow(commit core.Commit, isSelected bool, isMultiSelected bool, isInRange bool) string {
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

	var style lipgloss.Style
	var hashText, subjectText, authorText, dateText string
	var needsFullWidth bool

	if isSelected {
		style = selectedCommitRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else if isInRange {
		style = rangeSelectionRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else if isMultiSelected {
		style = multiSelectedCommitRowStyle
		needsFullWidth = true
		hashText = selectedHashStyle.Render(hash)
		subjectText = selectedSubjectStyle.Render(subject)
		authorText = selectedAuthorStyle.Render(author)
		dateText = selectedDateStyle.Render(date)
	} else {
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

	if needsFullWidth {
		return style.Width(96).Align(lipgloss.Left).Render(rowContent)
	}

	return style.Render(rowContent)
}

func (m *ListingModel) calculateTokensForSelection() int {
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

func (m *ListingModel) renderStatusBar() string {
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("v"), helpDescStyle.Render("select"))
	rangeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("V"), helpDescStyle.Render("range"))
	nextHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("N"), helpDescStyle.Render("next"))
	clearHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("clear"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))

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

	modeText := ""
	if m.selectionMode {
		modeText = fmt.Sprintf(" • %s", helpKeyStyle.Render("RANGE MODE"))
	}

	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits)))

	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", rangeHelp, " • ", nextHelp, " • ", clearHelp, " • ", quitHelp)

	rightSide := fmt.Sprintf("%s%s%s", position, selectionText, modeText)
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 10),
		rightSide,
	)

	return statusBarStyle.Render(statusContent)
}

// GetSelectedCommits returns the selected commits for sharing with other models
func (m *ListingModel) GetSelectedCommits() ([]core.Commit, map[int]bool) {
	return m.commits, m.selectedCommits
}
