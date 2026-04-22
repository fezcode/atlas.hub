package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	accent      = lipgloss.Color("#0EA5E9")
	accentLight = lipgloss.Color("#7DD3FC")
	mint        = lipgloss.Color("#34D399")
	sky         = lipgloss.Color("#38BDF8")
	rose        = lipgloss.Color("#FB7185")
	dimGray     = lipgloss.Color("#4B5563")
	midGray     = lipgloss.Color("#6B7280")
	lightGray   = lipgloss.Color("#9CA3AF")
	faintGray   = lipgloss.Color("#374151")
	white       = lipgloss.Color("#F9FAFB")
	amber       = lipgloss.Color("#FBBF24")
	darkBg      = lipgloss.Color("#0F172A")

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(accent).
			Padding(0, 1).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(accentLight).
			Bold(true)

	// Tabs
	activeTabStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(accent).
			Bold(true).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(midGray).
				Padding(0, 1)

	tabGapStyle = lipgloss.NewStyle().
			Foreground(faintGray)

	// Separators
	separatorStyle = lipgloss.NewStyle().
			Foreground(faintGray)

	// List items
	cursorStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	checkboxStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	checkedStyle = lipgloss.NewStyle().
			Foreground(mint)

	hubStyle = lipgloss.NewStyle().
			Foreground(rose)

	installedBadge = lipgloss.NewStyle().
			Foreground(mint)

	updateBadge = lipgloss.NewStyle().
			Foreground(sky).
			Bold(true)

	versionDimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Italic(true)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Background(darkBg)

	// Confirmation bar
	confirmBarStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(lipgloss.Color("#7C3AED"))

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDD6FE"))

	confirmYesStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(rose).
			Bold(true).
			Padding(0, 1)

	confirmNoStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(dimGray).
			Bold(true).
			Padding(0, 1)

	// Search
	searchStyle = lipgloss.NewStyle().
			Foreground(amber).
			Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(white).
				Bold(true)

	// Text
	dimTextStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true)

	// Progress
	statusPendingStyle    = lipgloss.NewStyle().Foreground(midGray)
	statusInstallingStyle = lipgloss.NewStyle().Foreground(sky).Bold(true)
	statusDoneStyle       = lipgloss.NewStyle().Foreground(mint).Bold(true)
	statusErrorStyle      = lipgloss.NewStyle().Foreground(rose).Bold(true)

	progressFilledStyle = lipgloss.NewStyle().
				Foreground(mint)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(faintGray)
)
