package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	purple      = lipgloss.Color("#6B50FF")
	brightGreen = lipgloss.Color("#00D787")
	cyan        = lipgloss.Color("#00D7FF")
	pink        = lipgloss.Color("#FF5F87")
	dimGray     = lipgloss.Color("#555555")
	midGray     = lipgloss.Color("#626262")
	lightGray   = lipgloss.Color("#888888")
	faintGray   = lipgloss.Color("#444444")
	white       = lipgloss.Color("#FFFDF5")
	yellow      = lipgloss.Color("#FFD700")

	titleStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Padding(0, 1).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B3B3FF")).
			Bold(true)

	dimTextStyle = lipgloss.NewStyle().
			Foreground(midGray)

	faintTextStyle = lipgloss.NewStyle().
			Foreground(faintGray)

	categoryStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	checkboxStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	checkedStyle = lipgloss.NewStyle().
			Foreground(brightGreen)

	installedBadge = lipgloss.NewStyle().
			Foreground(brightGreen).
			Bold(true)

	updateBadge = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	versionDimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	statusBarBg = lipgloss.Color("#282828")

	searchStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(white).
				Bold(true)

	statusPendingStyle    = lipgloss.NewStyle().Foreground(midGray)
	statusInstallingStyle = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	statusDoneStyle       = lipgloss.NewStyle().Foreground(brightGreen).Bold(true)
	statusErrorStyle      = lipgloss.NewStyle().Foreground(pink).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(midGray)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lightGray).
				Italic(true)

	scrollStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	progressFilledStyle = lipgloss.NewStyle().
				Foreground(brightGreen)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(faintGray)
)
