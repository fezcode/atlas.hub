package ui

import (
	"fmt"

	"atlas.hub/internal/install"
	"atlas.hub/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateList int = iota
	StateInstalling
	StateDone
)

type Model struct {
	Manager          *install.Manager
	Tools            []model.Tool
	InstallPath      string
	Cursor           int
	Top              int // Top visible item index
	Width            int
	Height           int
	State            int
	Quitting         bool
	ShowDescriptions bool
}

func NewModel(manager *install.Manager, tools []model.Tool, installPath string) Model {
	for i := range tools {
		if tools[i].IsHub {
			tools[i].Selected = true
		}
	}
	return Model{
		Manager:     manager,
		Tools:       tools,
		InstallPath: installPath,
		State:       StateList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type InstallMsg struct {
	Index int
	Err   error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateViewport()

	case tea.KeyMsg:
		if m.State == StateInstalling {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		
		if m.State == StateDone {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				m.updateViewport()
			}
		case "down", "j":
			if m.Cursor < len(m.Tools)-1 {
				m.Cursor++
				m.updateViewport()
			}
		case "h":
			m.ShowDescriptions = !m.ShowDescriptions
			m.updateViewport()
		case " ", "space":
			if !m.Tools[m.Cursor].IsHub {
				m.Tools[m.Cursor].Selected = !m.Tools[m.Cursor].Selected
			}
		case "enter":
			if m.State == StateList {
				m.State = StateInstalling
				return m.installNext()
			}
		}

	case InstallMsg:
		if msg.Err != nil {
			m.Tools[msg.Index].Status = "error"
			m.Tools[msg.Index].Error = msg.Err
		} else {
			m.Tools[msg.Index].Status = "done"
			// Refresh version
			m.Manager.CheckInstalledVersion(&m.Tools[msg.Index])
		}
		return m.installNext()
	}

	return m, nil
}

func (m *Model) updateViewport() {
	if m.Height == 0 {
		return
	}

	headerHeight := 6 // Title + Welcome text + spacing + Repo link
	footerHeight := 2 // Help text + spacing
	
	visibleHeight := m.Height - headerHeight - footerHeight
	if visibleHeight <= 0 {
		visibleHeight = 1
	}

	if m.Cursor < m.Top {
		m.Top = m.Cursor
	} else if m.Cursor >= m.Top+visibleHeight {
		m.Top = m.Cursor - visibleHeight + 1
	}
}

func (m Model) installNext() (tea.Model, tea.Cmd) {
	for i := range m.Tools {
		if m.Tools[i].Selected && m.Tools[i].Status == "" {
			m.Tools[i].Status = "installing"
			
			// Capture variables for closure
			idx := i
			tool := m.Tools[i]
			
			cmd := func() tea.Msg {
				err := m.Manager.Install(&tool)
				return InstallMsg{Index: idx, Err: err}
			}
			return m, cmd
		}
	}
	
	m.State = StateDone
	return m, nil
}

func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	s := titleStyle.Render(" ATLAS.HUB ") + " " + headerStyle.Render("Installer") + "\n\n"

	if m.State == StateList {
		s += "Select tools to install (Space to toggle, Enter to confirm):\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("View all apps here: https://github.com/stars/fezcode/lists/atlas") + "\n\n"
		
		headerHeight := 6
		footerHeight := 2
		visibleHeight := m.Height - headerHeight - footerHeight
		if visibleHeight <= 0 {
			visibleHeight = 10 // Fallback
		}

		end := m.Top + visibleHeight
		if end > len(m.Tools) {
			end = len(m.Tools)
		}

		for i := m.Top; i < end; i++ {
			tool := m.Tools[i]
			cursor := " "
			if m.Cursor == i {
				cursor = cursorStyle.Render("❯")
			}

			checked := checkboxStyle.Render("☐")
			if tool.Selected {
				checked = checkedStyle.Render("☑")
			}

			// Version info
			verInfo := ""
			if tool.InstalledVersion != "" {
				if tool.InstalledVersion != tool.LatestVersion && tool.LatestVersion != "" {
					verInfo = fmt.Sprintf(" [%s -> %s]", tool.InstalledVersion, tool.LatestVersion)
					verInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Render(verInfo)
				} else {
					verInfo = fmt.Sprintf(" [v%s]", tool.InstalledVersion)
					verInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(verInfo)
				}
			}

			content := fmt.Sprintf("%s %s %s%s", cursor, checked, tool.Name, verInfo)
			if m.Cursor == i {
				s += selectedItemStyle.Render(content) + "\n"
			} else {
				s += itemStyle.Render(content) + "\n"
			}

			if m.ShowDescriptions {
				desc := "   " + lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Italic(true).Render(tool.Description)
				s += desc + "\n"
			}
		}

		// Fill remaining space if needed to keep help at bottom
		renderedCount := end - m.Top
		if m.ShowDescriptions {
			// This is a rough estimation, but better than nothing for filling space
			renderedCount *= 2
		}
		if renderedCount < visibleHeight {
			for i := 0; i < visibleHeight-renderedCount; i++ {
				s += "\n"
			}
		}

		s += "\n" + helpStyle.Render("j/k: navigate • space: select • h: description • enter: install • q: quit")
	} else if m.State == StateInstalling || m.State == StateDone {
		s += "Installation Progress:\n\n"
		
		// For progress screen, we might also need scrolling if many tools selected
		// but usually it's fewer. Let's apply basic scrolling here too if needed.
		
		count := 0
		selectedTools := []model.Tool{}
		for _, tool := range m.Tools {
			if tool.Selected {
				count++
				selectedTools = append(selectedTools, tool)
			}
		}

		if count == 0 {
			s += "  " + itemStyle.Render("No tools selected.") + "\n"
		} else {
			// Reuse same height logic
			headerHeight := 5
			footerHeight := 3
			visibleHeight := m.Height - headerHeight - footerHeight
			if visibleHeight <= 0 { visibleHeight = 10 }

			// In progress/done state, we don't have a cursor, 
			// maybe we just show the first N or follow the active installation?
			// For now let's just show as many as fit.
			displayLimit := visibleHeight
			if len(selectedTools) > displayLimit {
				selectedTools = selectedTools[:displayLimit]
				// TODO: Better progress scrolling
			}

			for _, tool := range selectedTools {
				status := "PENDING"
				style := statusPendingStyle
				if tool.Status == "installing" {
					status = "INSTALLING"
					style = statusInstallingStyle
				} else if tool.Status == "done" {
					status = "DONE"
					style = statusDoneStyle
				} else if tool.Status == "error" {
					status = "ERROR"
					style = statusErrorStyle
				}

				s += fmt.Sprintf("  %s: %s\n", tool.Name, style.Render(status))
				if tool.Status == "error" {
					s += fmt.Sprintf("    %s\n", tool.Error)
				}
			}
		}
		
		if m.State == StateDone {
			s += "\n" + headerStyle.Render("Installation complete!") + "\n"
			s += itemStyle.Render(fmt.Sprintf("Make sure %s is in your PATH.", m.InstallPath)) + "\n"
			s += helpStyle.Render("Press 'q' to quit.")
		} else {
			s += "\n" + helpStyle.Render("Installing... Please wait.")
		}
	}

	return appStyle.Render(s)
}
