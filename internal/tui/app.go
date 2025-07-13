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
	currentView  ViewState
	commits      []core.Commit
	currentPage  int
	perPage      int
	totalCommits int
	repoPath     string
	errorMsg     string
	cursor       int
	viewport     int
	maxViewport  int
}

func initialModel() model {
	cwd, _ := os.Getwd()
	gitRoot, isGit, _ := core.GetGitDirectory(cwd)
	
	m := model{
		currentView: ListingView,
		currentPage: 1,
		perPage:     20,
		repoPath:    gitRoot,
		cursor:      0,
		viewport:    0,
		maxViewport: 8,
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
		
		row := m.renderCommitRow(commit, isSelected)
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

func (m model) renderCommitRow(commit core.Commit, isSelected bool) string {
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
	
	// Cursor indicator for selected row
	cursor := "  "
	if isSelected {
		cursor = "▶ "
	}
	
	if isSelected {
		// Selected row with full width background highlight
		hashText := selectedHashStyle.Render(hash)
		subjectText := selectedSubjectStyle.Render(subject)
		authorText := selectedAuthorStyle.Render(author)
		dateText := selectedDateStyle.Render(date)
		
		firstLine := fmt.Sprintf("%s%s %s", cursor, hashText, subjectText)
		secondLine := fmt.Sprintf("  %s • %s", authorText, dateText)
		
		rowContent := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
		
		// Apply background with full width
		return selectedCommitRowStyle.
			Width(76).
			Align(lipgloss.Left).
			Render(rowContent)
	} else {
		// Regular row with subtle styling
		hashText := hashStyle.Render(hash)
		subjectText := subjectStyle.Render(subject)
		authorText := authorStyle.Render(author)
		dateText := dateStyle.Render(date)
		
		firstLine := fmt.Sprintf("%s%s %s", cursor, hashText, subjectText)
		secondLine := fmt.Sprintf("  %s • %s", authorText, dateText)
		
		rowContent := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
		return commitRowStyle.Render(rowContent)
	}
}

func (m model) renderStatusBar() string {
	// Create help sections
	navHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("↑↓/jk"), helpDescStyle.Render("navigate"))
	jumpHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("g/G"), helpDescStyle.Render("top/bottom"))
	quitHelp := fmt.Sprintf("%s %s", helpKeyStyle.Render("q"), helpDescStyle.Render("quit"))
	
	// Position indicator
	position := positionStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits)))
	
	// Combine help text
	helpText := lipgloss.JoinHorizontal(lipgloss.Left, navHelp, " • ", jumpHelp, " • ", quitHelp)
	
	// Create status bar with help on left and position on right
	statusContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", 20), // Spacer
		position,
	)
	
	return statusBarStyle.Render(statusContent)
}

func RunApp() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}