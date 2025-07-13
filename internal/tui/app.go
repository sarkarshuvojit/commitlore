package tui

import (
	"fmt"
	"os"
	"strings"

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
		maxViewport: 15,
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
		return fmt.Sprintf("\n  Error: %s\n\n  Press 'q' or Ctrl+C to quit.\n", m.errorMsg)
	}
	
	if len(m.commits) == 0 {
		return "\n  No commits found.\n\n  Press 'q' or Ctrl+C to quit.\n"
	}
	
	var sb strings.Builder
	
	// Top margin
	sb.WriteString("\n")
	
	// Header with left margin
	sb.WriteString(fmt.Sprintf("  📋 Git Commits (Page %d) - %d total commits\n", m.currentPage, m.totalCommits))
	sb.WriteString("  " + strings.Repeat("─", 76) + "\n\n")
	
	// Calculate visible range
	start := m.viewport
	end := start + m.maxViewport
	if end > len(m.commits) {
		end = len(m.commits)
	}
	
	// Ensure we don't go below 0
	if start < 0 {
		start = 0
	}
	
	// Render visible commits with left margin
	for i := start; i < end; i++ {
		commit := m.commits[i]
		isSelected := i == m.cursor
		
		// Commit hash and message line
		hashMsg := fmt.Sprintf("%s %s", commit.Hash[:8], commit.Subject)
		if len(hashMsg) > 68 {
			hashMsg = hashMsg[:65] + "..."
		}
		
		// Author and date line
		authorDate := fmt.Sprintf("%s • %s", commit.Author, commit.Date.Format("2006-01-02 15:04"))
		if len(authorDate) > 68 {
			authorDate = authorDate[:65] + "..."
		}
		
		if isSelected {
			// Highlighted row with margin
			sb.WriteString(fmt.Sprintf("  ► ┌%s┐\n", strings.Repeat("─", 70)))
			sb.WriteString(fmt.Sprintf("    │ %-68s │\n", hashMsg))
			sb.WriteString(fmt.Sprintf("    │ %-68s │\n", authorDate))
			sb.WriteString(fmt.Sprintf("    └%s┘\n", strings.Repeat("─", 70)))
		} else {
			// Regular row with margin
			sb.WriteString(fmt.Sprintf("    %s\n", hashMsg))
			sb.WriteString(fmt.Sprintf("    %s\n", authorDate))
			sb.WriteString("\n")
		}
	}
	
	// Status bar with margin
	sb.WriteString("\n  " + strings.Repeat("─", 76) + "\n")
	sb.WriteString(fmt.Sprintf("  Navigate: ↑/↓ or j/k • Jump: g(top)/G(bottom) • Quit: q/Ctrl+C • Item %d/%d\n", 
		m.cursor+1, len(m.commits)))
	
	// Bottom margin
	sb.WriteString("\n")
	
	return sb.String()
}

func RunApp() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}