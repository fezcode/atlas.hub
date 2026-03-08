package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6B50FF")).
			Padding(0, 1).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B3B3FF")).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("212")).
				Bold(true)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585858"))

	checkedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D787"))

	statusPendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	statusInstallingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Bold(true)
	statusDoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D787")).Bold(true)
	statusErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#777777")).
				Italic(true)
)