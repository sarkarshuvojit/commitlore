package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette - modern, accessible colors
	primaryColor   = lipgloss.Color("#6366f1")   // Indigo
	secondaryColor = lipgloss.Color("#8b5cf6")   // Purple
	accentColor    = lipgloss.Color("#06b6d4")   // Cyan
	successColor   = lipgloss.Color("#10b981")   // Emerald
	warningColor   = lipgloss.Color("#f59e0b")   // Amber
	errorColor     = lipgloss.Color("#ef4444")   // Red
	
	// Neutral colors
	textPrimary   = lipgloss.Color("#f8fafc")    // Slate 50
	textSecondary = lipgloss.Color("#94a3b8")    // Slate 400
	textMuted     = lipgloss.Color("#64748b")    // Slate 500
	
	// Background colors
	bgPrimary     = lipgloss.Color("#0f172a")    // Slate 900
	bgSecondary   = lipgloss.Color("#1e293b")    // Slate 800
	bgAccent      = lipgloss.Color("#334155")    // Slate 700
	bgSelected    = lipgloss.Color("#1e40af")    // Blue 800
	
	// Border colors
	borderPrimary   = lipgloss.Color("#475569")  // Slate 600
	borderSecondary = lipgloss.Color("#334155")  // Slate 700
	borderAccent    = lipgloss.Color("#6366f1")  // Indigo 500
)

// Header styles
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Padding(1, 2).
			MarginBottom(2)
	
	titleStyle = lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Italic(true)
	
	subtitleStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			Italic(true)
	
	dimStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Italic(true)
)

// Commit row styles
var (
	// Base commit row
	commitRowStyle = lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1)
	
	// Selected commit row
	selectedCommitRowStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(0, 2).
				MarginBottom(1)

	// Multi-selected commit row
	multiSelectedCommitRowStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(0, 2).
				MarginBottom(1)

	// Range selection mode indicator
	rangeSelectionRowStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(0, 2).
				MarginBottom(1)
	
	// Hash style
	hashStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
	
	selectedHashStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Bold(true)
	
	// Subject style
	subjectStyle = lipgloss.NewStyle().
			Foreground(textPrimary)
	
	selectedSubjectStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Bold(true)
	
	// Author style
	authorStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)
	
	selectedAuthorStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Bold(true)
	
	// Date style
	dateStyle = lipgloss.NewStyle().
			Foreground(textMuted)
	
	selectedDateStyle = lipgloss.NewStyle().
				Foreground(textSecondary)
	
	// Cursor indicator
	cursorStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)

// Status bar styles
var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			Padding(0, 2).
			MarginTop(2)
	
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
	
	helpDescStyle = lipgloss.NewStyle().
			Foreground(textSecondary)
	
	positionStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
	
	flashStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)
)

// Container styles
var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)
	
	contentStyle = lipgloss.NewStyle().
			Width(100)
	
	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Right)
)

// Error and empty state styles
var (
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(1, 2).
			MarginTop(2).
			MarginBottom(2)
	
	emptyStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Italic(true).
			Align(lipgloss.Center).
			MarginTop(4).
			MarginBottom(4)
)