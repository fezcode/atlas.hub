package ui

import "github.com/charmbracelet/lipgloss"

var (
	gold   = lipgloss.Color("#FFD700")
	silver = lipgloss.Color("#CCCCCC")
	grey   = lipgloss.Color("#555555")
	red    = lipgloss.Color("#FF5F5F")
	green  = lipgloss.Color("#5FFF5F")
	onyx   = lipgloss.Color("#121212")

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(onyx).
			Background(gold).
			Padding(0, 1).
			Bold(true)

	headerStyle = lipgloss.NewStyle().Foreground(gold).Bold(true)
	
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(gold).
				Bold(true).
				PaddingLeft(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(silver).
			PaddingLeft(2)

	checkedStyle = lipgloss.NewStyle().Foreground(green).Bold(true)
	uncheckedStyle = lipgloss.NewStyle().Foreground(grey)

	statusPendingStyle = lipgloss.NewStyle().Foreground(grey)
	statusInstallingStyle = lipgloss.NewStyle().Foreground(gold).Bold(true)
	statusDoneStyle = lipgloss.NewStyle().Foreground(green).Bold(true)
	statusErrorStyle = lipgloss.NewStyle().Foreground(red).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(grey).
			MarginTop(1)
)
