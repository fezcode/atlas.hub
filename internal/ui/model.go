package ui

import (
	"fmt"

	"atlas.hub/internal/install"
	"atlas.hub/internal/model"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	StateList int = iota
	StateInstalling
	StateDone
)

type Model struct {
	Manager       *install.Manager
	Tools         []model.Tool
	InstallPath   string
	Cursor        int
	State         int
	Quitting      bool
	
	// Progress
	Progress      progress.Model
	CurrentAction string
	CurrentIndex  int // Currently installing tool index
	InstallCh     chan interface{} 
}

func NewModel(manager *install.Manager, tools []model.Tool, installPath string) Model {
	for i := range tools {
		if tools[i].IsHub {
			tools[i].Selected = true
		}
	}
	
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	return Model{
		Manager:     manager,
		Tools:       tools,
		InstallPath: installPath,
		State:       StateList,
		Progress:    p,
		CurrentIndex: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Messages
type installProgressMsg string
type installDoneMsg error

func waitForInstall(ch chan interface{}) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil 
		}
		switch v := msg.(type) {
		case string:
			return installProgressMsg(v)
		case error:
			return installDoneMsg(v)
		case nil: // nil error means success
			return installDoneMsg(nil)
		default:
			return nil
		}
	}
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
				// Start installation
				m.State = StateInstalling
				return m.installNext()
			}
		}

	case installProgressMsg:
		m.CurrentAction = string(msg)
		return m, waitForInstall(m.InstallCh)

	case installDoneMsg:
		// Mark current as done/error
		if msg != nil {
			m.Tools[m.CurrentIndex].Status = "error"
			m.Tools[m.CurrentIndex].Error = msg
		} else {
			m.Tools[m.CurrentIndex].Status = "done"
		}
		m.CurrentAction = ""
		// Install next
		return m.installNext()
		
	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m Model) installNext() (tea.Model, tea.Cmd) {
	// Find the next tool that is selected but not started
	for i := range m.Tools {
		if m.Tools[i].Selected && m.Tools[i].Status == "" {
			m.Tools[i].Status = "installing"
			m.CurrentIndex = i
			m.CurrentAction = "Starting..."
			
			// Count total selected and completed for progress bar
			total := 0
			done := 0
			for _, t := range m.Tools {
				if t.Selected {
					total++
					if t.Status == "done" || t.Status == "error" {
						done++
					}
				}
			}
			// Current one is starting, so done is previous count.
			// But we want smooth progress?
			// Let's set progress to done/total at start of this step
			pct := float64(done) / float64(total)
			
			// Note: We can't easily animate the bar smoothly WITHIN a tool install without more granular events.
			// But we can update it step-by-step.
			
			m.InstallCh = make(chan interface{})
			
			tool := m.Tools[i]
			go func() {
				err := m.Manager.Install(&tool, func(status string) {
					m.InstallCh <- status
				})
				m.InstallCh <- err
				close(m.InstallCh)
			}()
			
			return m, tea.Batch(
				waitForInstall(m.InstallCh),
				m.Progress.SetPercent(pct),
			)
		}
	}
	
	// All done
	m.State = StateDone
	m.CurrentIndex = -1
	return m, m.Progress.SetPercent(1.0)
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

			line := fmt.Sprintf("%s %s %s", cursor, checked, tool.Name)
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
		
		// Count selected tools
		count := 0
		for _, tool := range m.Tools {
			if tool.Selected {
				count++
			}
		}

		if count == 0 {
			s += "  " + itemStyle.Render("No tools selected.") + "\n"
		} else {
			// Limit list height if too many?
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
		
		s += "\n"
		if m.State == StateInstalling {
			s += m.Progress.View() + "\n"
			s += helpStyle.Render(m.CurrentAction) + "\n"
		} else {
			s += m.Progress.View() + "\n"
			s += "\n" + headerStyle.Render("Installation complete!") + "\n"
			s += itemStyle.Render(fmt.Sprintf("Make sure %s is in your PATH.", m.InstallPath)) + "\n"
			s += helpStyle.Render("Press 'q' to quit.")
		}
	}

	return appStyle.Render(s)
}
