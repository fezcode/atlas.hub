package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"atlas.hub/internal/install"
	"atlas.hub/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const spinnerInterval = 80 * time.Millisecond

const (
	StateList int = iota
	StateInstalling
	StateDone
)

// displayRow is a single row in the scrollable list: either a category header or a tool.
type displayRow struct {
	IsCategory bool
	Category   string
	ToolIndex  int
}

type Model struct {
	Manager          *install.Manager
	Tools            []model.Tool
	InstallPath      string
	Cursor           int
	Top              int
	Width            int
	Height           int
	State            int
	Quitting         bool
	ShowDescriptions bool
	SearchMode       bool
	SearchQuery      string
	SpinnerFrame     int

	displayRows []displayRow
}

var categoryOrder = []string{
	"Core", "Development", "System", "Productivity", "Security",
	"CLI", "Utility", "Media", "Entertainment", "Fun", "Lifestyle",
}

func categoryRank(cat string) int {
	for i, c := range categoryOrder {
		if c == cat {
			return i
		}
	}
	return len(categoryOrder)
}

func NewModel(manager *install.Manager, tools []model.Tool, installPath string) Model {
	for i := range tools {
		if tools[i].IsHub {
			tools[i].Selected = true
		}
	}
	m := Model{
		Manager:     manager,
		Tools:       tools,
		InstallPath: installPath,
		State:       StateList,
	}
	m.rebuildRows()
	return m
}

func (m *Model) rebuildRows() {
	m.displayRows = nil

	groups := map[string][]int{}
	for i, t := range m.Tools {
		if m.SearchQuery != "" {
			q := strings.ToLower(m.SearchQuery)
			if !strings.Contains(strings.ToLower(t.Name), q) &&
				!strings.Contains(strings.ToLower(t.Description), q) &&
				!strings.Contains(strings.ToLower(t.Category), q) {
				continue
			}
		}
		groups[t.Category] = append(groups[t.Category], i)
	}

	cats := make([]string, 0, len(groups))
	for c := range groups {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		return categoryRank(cats[i]) < categoryRank(cats[j])
	})

	for _, cat := range cats {
		m.displayRows = append(m.displayRows, displayRow{IsCategory: true, Category: cat})
		for _, idx := range groups[cat] {
			m.displayRows = append(m.displayRows, displayRow{ToolIndex: idx})
		}
	}

	if m.Cursor >= len(m.displayRows) {
		m.Cursor = len(m.displayRows) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	m.skipCategoryRow(1)
}

func (m *Model) skipCategoryRow(dir int) {
	if len(m.displayRows) == 0 {
		return
	}
	for m.displayRows[m.Cursor].IsCategory {
		m.Cursor += dir
		if m.Cursor < 0 {
			m.Cursor = 0
			return
		}
		if m.Cursor >= len(m.displayRows) {
			m.Cursor = len(m.displayRows) - 1
			return
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type InstallMsg struct {
	Index int
	Err   error
}

type SpinnerTickMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateViewport()

	case SpinnerTickMsg:
		if m.State == StateInstalling {
			m.SpinnerFrame++
			return m, tea.Tick(spinnerInterval, func(_ time.Time) tea.Msg {
				return SpinnerTickMsg{}
			})
		}

	case tea.KeyMsg:
		if m.SearchMode {
			switch msg.String() {
			case "esc":
				m.SearchMode = false
				m.SearchQuery = ""
				m.rebuildRows()
				m.updateViewport()
			case "enter":
				m.SearchMode = false
			case "backspace":
				if len(m.SearchQuery) > 0 {
					m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
					m.rebuildRows()
					m.updateViewport()
				}
			case "ctrl+c":
				m.Quitting = true
				return m, tea.Quit
			default:
				if len(msg.String()) == 1 {
					m.SearchQuery += msg.String()
					m.rebuildRows()
					m.updateViewport()
				}
			}
			return m, nil
		}

		if m.State == StateInstalling {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}

		if m.State == StateDone {
			if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "enter" {
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
				m.skipCategoryRow(-1)
				m.updateViewport()
			}
		case "down", "j":
			if m.Cursor < len(m.displayRows)-1 {
				m.Cursor++
				m.skipCategoryRow(1)
				m.updateViewport()
			}
		case "pgup":
			m.Cursor -= m.listHeight()
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.skipCategoryRow(1)
			m.updateViewport()
		case "pgdown":
			m.Cursor += m.listHeight()
			if m.Cursor >= len(m.displayRows) {
				m.Cursor = len(m.displayRows) - 1
			}
			m.skipCategoryRow(-1)
			m.updateViewport()
		case "home", "g":
			m.Cursor = 0
			m.skipCategoryRow(1)
			m.Top = 0
			m.updateViewport()
		case "end", "G":
			m.Cursor = len(m.displayRows) - 1
			m.skipCategoryRow(-1)
			m.updateViewport()
		case "h":
			m.ShowDescriptions = !m.ShowDescriptions
		case "/":
			m.SearchMode = true
			m.SearchQuery = ""
		case " ":
			if m.Cursor >= 0 && m.Cursor < len(m.displayRows) && !m.displayRows[m.Cursor].IsCategory {
				idx := m.displayRows[m.Cursor].ToolIndex
				if !m.Tools[idx].IsHub {
					m.Tools[idx].Selected = !m.Tools[idx].Selected
				}
			}
		case "a":
			allSelected := true
			for _, row := range m.displayRows {
				if !row.IsCategory && !m.Tools[row.ToolIndex].IsHub && !m.Tools[row.ToolIndex].Selected {
					allSelected = false
					break
				}
			}
			for _, row := range m.displayRows {
				if !row.IsCategory && !m.Tools[row.ToolIndex].IsHub {
					m.Tools[row.ToolIndex].Selected = !allSelected
				}
			}
		case "enter":
			if m.State == StateList {
				m.State = StateInstalling
				m.SpinnerFrame = 0
				mdl, cmd := m.installNext()
				return mdl, tea.Batch(cmd, tea.Tick(spinnerInterval, func(_ time.Time) tea.Msg {
					return SpinnerTickMsg{}
				}))
			}
		}

	case InstallMsg:
		if msg.Err != nil {
			m.Tools[msg.Index].Status = "error"
			m.Tools[msg.Index].Error = msg.Err
		} else {
			m.Tools[msg.Index].Status = "done"
			m.Manager.CheckInstalledVersion(&m.Tools[msg.Index])
		}
		return m.installNext()
	}

	return m, nil
}

// listHeight returns how many rows are available for the scrollable item list.
func (m Model) listHeight() int {
	// We use exactly m.Height lines total. Layout:
	//   line 1: title bar
	//   line 2: subtitle / info
	//   line 3: search bar or blank
	//   line 4..H-2: scrollable list items
	//   line H-1: status bar
	//   line H:   help bar
	// So list area = Height - 5
	h := m.Height - 5
	if h < 1 {
		h = 1
	}
	return h
}

func (m *Model) updateViewport() {
	if m.Height == 0 {
		return
	}
	h := m.listHeight()
	if m.Cursor < m.Top {
		m.Top = m.Cursor
	} else if m.Cursor >= m.Top+h {
		m.Top = m.Cursor - h + 1
	}
	if m.Top < 0 {
		m.Top = 0
	}
}

func (m Model) installNext() (tea.Model, tea.Cmd) {
	for i := range m.Tools {
		if m.Tools[i].Selected && m.Tools[i].Status == "" {
			m.Tools[i].Status = "installing"
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

func (m Model) selectedCount() (selected, total int) {
	for _, t := range m.Tools {
		total++
		if t.Selected {
			selected++
		}
	}
	return
}

func (m Model) toolRowCount() int {
	n := 0
	for _, r := range m.displayRows {
		if !r.IsCategory {
			n++
		}
	}
	return n
}

func (m Model) cursorToolIndex() int {
	idx := 0
	for i, r := range m.displayRows {
		if r.IsCategory {
			continue
		}
		if i == m.Cursor {
			return idx
		}
		idx++
	}
	return idx
}

// View renders the entire UI into exactly m.Height lines.
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	if m.Height == 0 || m.Width == 0 {
		return "Loading..."
	}

	// Build lines array — exactly m.Height lines
	lines := make([]string, m.Height)
	w := m.Width

	// Line 0: Title bar
	title := titleStyle.Render(" ATLAS.HUB ") + " " + subtitleStyle.Render("Installer")
	lines[0] = title

	if m.State == StateList {
		m.renderList(lines, w)
	} else {
		m.renderProgress(lines, w)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderList(lines []string, w int) {
	// Line 1: subtitle
	sel, total := m.selectedCount()
	lines[1] = dimTextStyle.Render(fmt.Sprintf("Select tools to install (%d/%d selected)", sel, total))

	// Line 2: search or link
	if m.SearchMode {
		lines[2] = searchStyle.Render(" / ") + searchInputStyle.Render(m.SearchQuery) + searchStyle.Render("_")
	} else if m.SearchQuery != "" {
		lines[2] = searchStyle.Render(" Filter: ") + searchInputStyle.Render(m.SearchQuery) +
			dimTextStyle.Render(fmt.Sprintf(" (%d results)", m.toolRowCount()))
	} else {
		lines[2] = faintTextStyle.Render("https://github.com/stars/fezcode/lists/atlas")
	}

	// Lines 3..H-3: scrollable list
	listH := m.listHeight()
	end := m.Top + listH
	if end > len(m.displayRows) {
		end = len(m.displayRows)
	}

	lineIdx := 3
	for i := m.Top; i < end && lineIdx < m.Height-2; i++ {
		row := m.displayRows[i]

		if row.IsCategory {
			icon := categoryIcon(row.Category)
			lines[lineIdx] = " " + categoryStyle.Render(icon+" "+strings.ToUpper(row.Category))
			lineIdx++
			continue
		}

		tool := m.Tools[row.ToolIndex]

		// Cursor
		cursor := "   "
		if m.Cursor == i {
			cursor = " " + cursorStyle.Render("❯") + " "
		}

		// Checkbox
		check := checkboxStyle.Render("○")
		if tool.Selected {
			check = checkedStyle.Render("●")
		}

		// Name
		name := tool.Name
		if m.Cursor == i {
			name = highlightStyle.Render(name)
		}

		// Badge
		badge := ""
		if tool.InstalledVersion != "" {
			if tool.InstalledVersion != tool.LatestVersion && tool.LatestVersion != "" {
				badge = " " + updateBadge.Render(tool.InstalledVersion+" -> "+tool.LatestVersion)
			} else {
				badge = " " + installedBadge.Render("✓") + " " + versionDimStyle.Render("v"+tool.InstalledVersion)
			}
		} else if tool.LatestVersion != "" {
			badge = " " + versionDimStyle.Render("v"+tool.LatestVersion)
		}

		line := cursor + check + " " + name + badge

		// Description
		if m.ShowDescriptions && tool.Description != "" {
			used := lipgloss.Width(line)
			avail := w - used - 4
			if avail > 10 {
				desc := tool.Description
				if len(desc) > avail {
					desc = desc[:avail-3] + "..."
				}
				line += " " + descriptionStyle.Render("- "+desc)
			}
		}

		lines[lineIdx] = line
		lineIdx++
	}

	// Scroll indicator on right side of list area
	if len(m.displayRows) > listH {
		// Show a scroll position hint at the last list line
		pos := ""
		if m.Top > 0 && end < len(m.displayRows) {
			pos = scrollStyle.Render(fmt.Sprintf("  [%d-%d of %d]", m.Top+1, end, len(m.displayRows)))
		} else if m.Top > 0 {
			pos = scrollStyle.Render("  [end]")
		} else {
			pos = scrollStyle.Render(fmt.Sprintf("  [1-%d of %d]", end, len(m.displayRows)))
		}
		// Place on the last empty list line, or append to last used line
		if lineIdx < m.Height-2 {
			lines[lineIdx] = pos
		} else {
			lines[m.Height-3] = lines[m.Height-3] + pos
		}
	}

	// Line H-2: status bar
	statusLine := m.Height - 2
	toolIdx := m.cursorToolIndex() + 1
	toolTotal := m.toolRowCount()
	left := fmt.Sprintf(" %d/%d selected", sel, total)
	right := fmt.Sprintf(" %d/%d ", toolIdx, toolTotal)
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	barContent := left + strings.Repeat(" ", gap) + right
	lines[statusLine] = lipgloss.NewStyle().
		Foreground(purple).
		Background(statusBarBg).
		Bold(true).
		Width(w).
		Render(barContent)

	// Line H-1: help bar
	helpLine := m.Height - 1
	lines[helpLine] = helpKeyStyle.Render(" j/k") + helpStyle.Render(" nav ") +
		helpKeyStyle.Render("space") + helpStyle.Render(" sel ") +
		helpKeyStyle.Render("a") + helpStyle.Render(" all ") +
		helpKeyStyle.Render("h") + helpStyle.Render(" desc ") +
		helpKeyStyle.Render("/") + helpStyle.Render(" find ") +
		helpKeyStyle.Render("pgup/dn") + helpStyle.Render(" page ") +
		helpKeyStyle.Render("enter") + helpStyle.Render(" install ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
}

func (m Model) renderProgress(lines []string, w int) {
	// Gather selected tools
	type selTool struct {
		tool  model.Tool
		index int
	}
	var selected []selTool
	for i, t := range m.Tools {
		if t.Selected {
			selected = append(selected, selTool{t, i})
		}
	}

	// Line 1: blank
	lines[1] = ""

	if len(selected) == 0 {
		lines[2] = dimTextStyle.Render("  No tools selected.")
		return
	}

	// Counts
	done, errors := 0, 0
	for _, st := range selected {
		switch st.tool.Status {
		case "done":
			done++
		case "error":
			errors++
		}
	}
	total := len(selected)

	// Line 2: progress bar
	barWidth := 30
	if w > 80 {
		barWidth = 40
	}
	filled := 0
	if total > 0 {
		filled = (done + errors) * barWidth / total
	}
	bar := progressFilledStyle.Render(strings.Repeat("█", filled)) +
		progressEmptyStyle.Render(strings.Repeat("░", barWidth-filled))
	lines[2] = fmt.Sprintf("  %s %s", bar,
		dimTextStyle.Render(fmt.Sprintf("%d/%d", done+errors, total)))

	// Lines 3+: tool statuses
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinnerFrames[m.SpinnerFrame%len(spinnerFrames)]

	lineIdx := 4
	for _, st := range selected {
		if lineIdx >= m.Height-2 {
			break
		}

		var icon, status string
		var style lipgloss.Style

		switch st.tool.Status {
		case "installing":
			icon = statusInstallingStyle.Render(spinner)
			status = "installing..."
			style = statusInstallingStyle
		case "done":
			icon = statusDoneStyle.Render("✓")
			status = "done"
			style = statusDoneStyle
		case "error":
			icon = statusErrorStyle.Render("✗")
			status = "error"
			style = statusErrorStyle
		default:
			icon = statusPendingStyle.Render("○")
			status = "waiting"
			style = statusPendingStyle
		}

		lines[lineIdx] = fmt.Sprintf("  %s %s %s", icon, st.tool.Name, style.Render(status))
		lineIdx++

		if st.tool.Status == "error" && st.tool.Error != nil && lineIdx < m.Height-2 {
			errMsg := st.tool.Error.Error()
			if len(errMsg) > 60 {
				errMsg = errMsg[:57] + "..."
			}
			lines[lineIdx] = "    " + statusErrorStyle.Render(errMsg)
			lineIdx++
		}
	}

	// Footer
	if m.State == StateDone {
		helpIdx := m.Height - 2
		if errors > 0 {
			lines[helpIdx] = "  " + statusDoneStyle.Render(fmt.Sprintf("✓ %d installed", done)) +
				"  " + statusErrorStyle.Render(fmt.Sprintf("✗ %d failed", errors))
		} else {
			lines[helpIdx] = "  " + statusDoneStyle.Render(fmt.Sprintf("✓ All %d tools installed!", done))
		}
		lines[m.Height-1] = "  " + dimTextStyle.Render("Path: "+m.InstallPath) +
			"  " + helpKeyStyle.Render("q") + helpStyle.Render("/") +
			helpKeyStyle.Render("enter") + helpStyle.Render(" quit")
	} else {
		lines[m.Height-1] = "  " + dimTextStyle.Render("Installing... ") +
			helpKeyStyle.Render("ctrl+c") + helpStyle.Render(" cancel")
	}
}

func categoryIcon(cat string) string {
	switch cat {
	case "Core":
		return "◆"
	case "Development":
		return "⚙"
	case "System":
		return "⊞"
	case "Productivity":
		return "▶"
	case "Security":
		return "◈"
	case "CLI":
		return "⌘"
	case "Utility":
		return "◇"
	case "Media":
		return "♫"
	case "Entertainment":
		return "★"
	case "Fun":
		return "✦"
	case "Lifestyle":
		return "♥"
	default:
		return "·"
	}
}
