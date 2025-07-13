package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sarkarshuvojit/commitlore/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewState int

const (
	ListingView ViewState = iota
)

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
}

func initialModel() model {
	cwd, _ := os.Getwd()
	gitRoot, isGit, _ := core.GetGitDirectory(cwd)
	
	m := model{
		currentView:     ListingView,
		currentPage:     1,
		perPage:         20,
		repoPath:        gitRoot,
		cursor:          0,
		viewport:        0,
		maxViewport:     8,
		selectedCommits: make(map[int]bool),
		selectionMode:   false,
		rangeStart:      -1,
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Ensure cursor stays visible in viewport
				if m.cursor < m.viewport {
					m.viewport = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.commits)-1 {
				m.cursor++
				// Ensure cursor stays visible in viewport
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
				// Adjust viewport to show the last item
				if len(m.commits) > m.maxViewport {
					m.viewport = len(m.commits) - m.maxViewport
				} else {
					m.viewport = 0
				}
			}
		case "v":
			// Toggle selection for current commit
			if len(m.selectedCommits) < 5 || m.selectedCommits[m.cursor] {
				if m.selectedCommits[m.cursor] {
					delete(m.selectedCommits, m.cursor)
				} else {
					m.selectedCommits[m.cursor] = true
				}
			}
		case "V":
			// Range selection mode
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
				}
				m.selectionMode = false
				m.rangeStart = -1
			}
		case "escape":
			// Clear selection mode and all selections
			m.selectionMode = false
			m.rangeStart = -1
			m.selectedCommits = make(map[int]bool)
		case "d":
			// Deselect current commit
			if m.selectedCommits[m.cursor] {
				delete(m.selectedCommits, m.cursor)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case ListingView:
		return m.renderListingView()
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
		Width(80).
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
	if len(subject) > 50 {
		subject = subject[:47] + "..."
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
			Width(76).
			Align(lipgloss.Left).
			Render(rowContent)
	}
	
	return style.Render(rowContent)
}

func (m model) renderStatusBar() string {
	// Create help sections
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	selectHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("v"), helpDescStyle.Render("select"))
	rangeHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("V"), helpDescStyle.Render("range"))
	clearHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("esc"), helpDescStyle.Render("clear"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	// Selection counter
	selectionCount := len(m.selectedCommits)
	selectionText := ""
	if selectionCount > 0 {
		selectionText = fmt.Sprintf(" • %s", positionStyle.Render(fmt.Sprintf("%d/5 selected", selectionCount)))
	}
	
	// Selection mode indicator
	modeText := ""
	if m.selectionMode {
		modeText = fmt.Sprintf(" • %s", helpKeyStyle.Render("RANGE MODE"))
	}
	
	// Position indicator
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits)))
	
	// Combine help text
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", selectHelp, " • ", rangeHelp, " • ", clearHelp, " • ", quitHelp)
	
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

func RunApp() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}