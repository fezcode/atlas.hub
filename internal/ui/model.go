package ui

import (
	"fmt"

	"atlas.hub/internal/install"
	"atlas.hub/internal/model"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateList int = iota
	StateInstalling
	StateDone
)

type Model struct {
	Manager     *install.Manager
	Tools       []model.Tool
	InstallPath string
	Cursor      int
	State       int
	Quitting    bool
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
			}
		case "down", "j":
			if m.Cursor < len(m.Tools)-1 {
				m.Cursor++
			}
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
		s += "Select tools to install (Space to toggle, Enter to confirm):\n\n"
		for i, tool := range m.Tools {
			cursor := " "
			if m.Cursor == i {
				cursor = ">"
			}

			checked := "[ ]"
			if tool.Selected {
				checked = "[x]"
			}

			// Version info
			verInfo := ""
			if tool.InstalledVersion != "" {
				if tool.InstalledVersion != tool.LatestVersion && tool.LatestVersion != "" {
					verInfo = fmt.Sprintf(" [%s -> %s]", tool.InstalledVersion, tool.LatestVersion)
					verInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Render(verInfo) // Cyan for update
				} else {
					verInfo = fmt.Sprintf(" [v%s]", tool.InstalledVersion)
					verInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(verInfo) // Grey for same
				}
			} else if tool.LatestVersion != "" {
				// verInfo = fmt.Sprintf(" (v%s)", tool.LatestVersion)
			}

			line := fmt.Sprintf("%s %s %s%s", cursor, checked, tool.Name, verInfo)
			if m.Cursor == i {
				line = selectedItemStyle.Render(line)
			} else {
				line = itemStyle.Render(line)
			}
			s += line + "\n"
		}
		s += "\n" + helpStyle.Render("j/k: navigate • space: select • enter: install • q: quit")
	} else if m.State == StateInstalling || m.State == StateDone {
		s += "Installation Progress:\n\n"
		
		count := 0
		for _, tool := range m.Tools {
			if tool.Selected {
				count++
			}
		}

		if count == 0 {
			s += "  " + itemStyle.Render("No tools selected.") + "\n"
		} else {
			for _, tool := range m.Tools {
				if !tool.Selected {
					continue
				}

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
