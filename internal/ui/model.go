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
	SearchMode       bool
	SearchQuery      string
	SpinnerFrame     int

	// Tab navigation
	ActiveTab  int
	Categories []string

	// Filtered view
	filtered []int // indices into Tools
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
	m.buildCategories()
	m.rebuildFiltered()
	return m
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

func (m *Model) buildCategories() {
	seen := map[string]bool{}
	for _, t := range m.Tools {
		seen[t.Category] = true
	}
	m.Categories = []string{"All"}
	sorted := make([]string, 0, len(seen))
	for c := range seen {
		sorted = append(sorted, c)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return categoryRank(sorted[i]) < categoryRank(sorted[j])
	})
	m.Categories = append(m.Categories, sorted...)
}

func (m *Model) rebuildFiltered() {
	m.filtered = nil
	activeCategory := ""
	if m.ActiveTab > 0 && m.ActiveTab < len(m.Categories) {
		activeCategory = m.Categories[m.ActiveTab]
	}

	for i, t := range m.Tools {
		// Category filter
		if activeCategory != "" && t.Category != activeCategory {
			continue
		}
		// Search filter
		if m.SearchQuery != "" {
			q := strings.ToLower(m.SearchQuery)
			if !strings.Contains(strings.ToLower(t.Name), q) &&
				!strings.Contains(strings.ToLower(t.Description), q) {
				continue
			}
		}
		m.filtered = append(m.filtered, i)
	}

	if m.Cursor >= len(m.filtered) {
		m.Cursor = len(m.filtered) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	m.Top = 0
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
				m.rebuildFiltered()
				m.updateViewport()
			case "enter":
				m.SearchMode = false
			case "backspace":
				if len(m.SearchQuery) > 0 {
					m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
					m.rebuildFiltered()
					m.updateViewport()
				}
			case "ctrl+c":
				m.Quitting = true
				return m, tea.Quit
			default:
				if len(msg.String()) == 1 {
					m.SearchQuery += msg.String()
					m.rebuildFiltered()
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
				m.updateViewport()
			}
		case "down", "j":
			if m.Cursor < len(m.filtered)-1 {
				m.Cursor++
				m.updateViewport()
			}
		case "pgup":
			m.Cursor -= m.listHeight()
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.updateViewport()
		case "pgdown":
			m.Cursor += m.listHeight()
			if m.Cursor >= len(m.filtered) {
				m.Cursor = len(m.filtered) - 1
			}
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.updateViewport()
		case "home", "g":
			m.Cursor = 0
			m.Top = 0
		case "end", "G":
			m.Cursor = len(m.filtered) - 1
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.updateViewport()
		case "tab":
			m.ActiveTab++
			if m.ActiveTab >= len(m.Categories) {
				m.ActiveTab = 0
			}
			m.rebuildFiltered()
			m.updateViewport()
		case "shift+tab":
			m.ActiveTab--
			if m.ActiveTab < 0 {
				m.ActiveTab = len(m.Categories) - 1
			}
			m.rebuildFiltered()
			m.updateViewport()
		case "/":
			m.SearchMode = true
			m.SearchQuery = ""
		case " ":
			if m.Cursor >= 0 && m.Cursor < len(m.filtered) {
				idx := m.filtered[m.Cursor]
				if !m.Tools[idx].IsHub {
					m.Tools[idx].Selected = !m.Tools[idx].Selected
				}
			}
		case "a":
			allSelected := true
			for _, fi := range m.filtered {
				if !m.Tools[fi].IsHub && !m.Tools[fi].Selected {
					allSelected = false
					break
				}
			}
			for _, fi := range m.filtered {
				if !m.Tools[fi].IsHub {
					m.Tools[fi].Selected = !allSelected
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

func (m Model) listHeight() int {
	// Layout: title(1) + tabs(1) + separator(1) + list + separator(1) + status(1) + help(1)
	h := m.Height - 6
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

// View renders the entire UI into exactly m.Height lines.
func (m Model) View() string {
	if m.Quitting {
		return ""
	}
	if m.Height == 0 || m.Width == 0 {
		return "Loading..."
	}

	lines := make([]string, m.Height)
	w := m.Width

	// Line 0: Title
	lines[0] = titleStyle.Render(" ATLAS.HUB ") + " " + subtitleStyle.Render("Suite Installer")

	if m.State == StateList {
		m.renderList(lines, w)
	} else {
		m.renderProgress(lines, w)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderTabs(w int) string {
	var tabs []string
	for i, cat := range m.Categories {
		if i == m.ActiveTab {
			tabs = append(tabs, activeTabStyle.Render(" "+cat+" "))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(" "+cat+" "))
		}
	}

	row := strings.Join(tabs, tabGapStyle.Render(" "))

	// If search is active, show it on the right
	if m.SearchMode {
		search := searchStyle.Render("/") + searchInputStyle.Render(m.SearchQuery) + searchStyle.Render("_")
		gap := w - lipgloss.Width(row) - lipgloss.Width(search) - 2
		if gap > 0 {
			row = row + strings.Repeat(" ", gap) + search
		}
	} else if m.SearchQuery != "" {
		search := searchStyle.Render("filter: ") + searchInputStyle.Render(m.SearchQuery)
		gap := w - lipgloss.Width(row) - lipgloss.Width(search) - 2
		if gap > 0 {
			row = row + strings.Repeat(" ", gap) + search
		}
	}

	return row
}

func (m Model) renderList(lines []string, w int) {
	// Line 1: Tab bar
	lines[1] = " " + m.renderTabs(w)

	// Line 2: Separator
	sepW := w
	if sepW < 0 {
		sepW = 0
	}
	lines[2] = separatorStyle.Render(strings.Repeat("─", sepW))

	// Lines 3..H-3: Items
	listH := m.listHeight()
	end := m.Top + listH
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	// Compute column widths
	nameCol := 20
	for _, fi := range m.filtered {
		if l := len(m.Tools[fi].Name); l > nameCol {
			nameCol = l
		}
	}
	nameCol += 1 // padding

	verCol := 10
	for _, fi := range m.filtered {
		t := m.Tools[fi]
		if l := len(t.InstalledVersion); l > verCol {
			verCol = l
		}
		if l := len(t.LatestVersion); l > verCol {
			verCol = l
		}
	}
	verCol += 1 // padding

	statusCol := 12

	lineIdx := 3
	for i := m.Top; i < end && lineIdx < m.Height-3; i++ {
		tool := m.Tools[m.filtered[i]]
		isCurrent := i == m.Cursor

		// Col 1: cursor + checkbox (4 chars)
		cursor := "  "
		if isCurrent {
			cursor = cursorStyle.Render("› ")
		}

		check := checkboxStyle.Render("○")
		if tool.Selected {
			check = checkedStyle.Render("●")
		}

		// Col 2: name (fixed width)
		nameTxt := tool.Name
		namePadding := nameCol - len(nameTxt)
		if namePadding < 0 {
			namePadding = 0
		}
		padded := nameTxt + strings.Repeat(" ", namePadding)
		if isCurrent {
			padded = highlightStyle.Render(nameTxt) + strings.Repeat(" ", namePadding)
		} else {
			padded = nameStyle.Render(nameTxt) + strings.Repeat(" ", namePadding)
		}

		// Col 3: version (fixed width)
		ver := ""
		if tool.InstalledVersion != "" {
			ver = tool.InstalledVersion
		} else if tool.LatestVersion != "" {
			ver = tool.LatestVersion
		}
		verPad := ver
		verPadding := verCol - len(ver)
		if verPadding > 0 {
			verPad = ver + strings.Repeat(" ", verPadding)
		} else if len(ver) > verCol {
			verPad = ver[:verCol]
		}
		verRendered := versionDimStyle.Render(verPad)

		// Col 4: status badge (fixed width)
		status := ""
		if tool.InstalledVersion != "" {
			if tool.InstalledVersion != tool.LatestVersion && tool.LatestVersion != "" {
				s := "update"
				statusPadding := statusCol - len(s)
				if statusPadding < 0 {
					statusPadding = 0
				}
				status = updateBadge.Render(s) + strings.Repeat(" ", statusPadding)
			} else {
				s := "installed"
				statusPadding := statusCol - len(s)
				if statusPadding < 0 {
					statusPadding = 0
				}
				status = installedBadge.Render(s) + strings.Repeat(" ", statusPadding)
			}
		} else {
			statusPadding := statusCol
			if statusPadding < 0 {
				statusPadding = 0
			}
			status = strings.Repeat(" ", statusPadding)
		}

		// Col 5: description (fill remaining)
		desc := ""
		used := 4 + nameCol + verCol + statusCol + 1
		avail := w - used
		if avail > 10 && tool.Description != "" {
			d := tool.Description
			if len(d) > avail {
				d = d[:avail-3] + "..."
			}
			desc = descriptionStyle.Render(d)
		}

		lines[lineIdx] = cursor + check + " " + padded + verRendered + status + desc
		lineIdx++
	}

	// Line H-3: Separator
	sepLine := m.Height - 3
	sepW2 := w
	if sepW2 < 0 {
		sepW2 = 0
	}
	lines[sepLine] = separatorStyle.Render(strings.Repeat("─", sepW2))

	// Line H-2: Status bar
	sel, total := m.selectedCount()
	statusLine := m.Height - 2

	left := fmt.Sprintf(" %d/%d selected", sel, total)

	// Pagination info
	page, totalPages := m.pageInfo()
	right := fmt.Sprintf(" %d of %d   page %d/%d ", m.Cursor+1, len(m.filtered), page, totalPages)

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	barContent := left + strings.Repeat(" ", gap) + right
	lines[statusLine] = statusBarStyle.Width(w).Render(barContent)

	// Line H-1: Help
	helpLine := m.Height - 1
	lines[helpLine] = " " +
		helpKeyStyle.Render("j/k") + helpStyle.Render(" nav  ") +
		helpKeyStyle.Render("space") + helpStyle.Render(" select  ") +
		helpKeyStyle.Render("a") + helpStyle.Render(" all  ") +
		helpKeyStyle.Render("tab") + helpStyle.Render(" category  ") +
		helpKeyStyle.Render("/") + helpStyle.Render(" search  ") +
		helpKeyStyle.Render("enter") + helpStyle.Render(" install  ") +
		helpKeyStyle.Render("q") + helpStyle.Render(" quit")
}

func (m Model) pageInfo() (page, totalPages int) {
	h := m.listHeight()
	if h <= 0 {
		return 1, 1
	}
	totalPages = (len(m.filtered) + h - 1) / h
	if totalPages < 1 {
		totalPages = 1
	}
	page = (m.Top / h) + 1
	if page > totalPages {
		page = totalPages
	}
	return
}

func (m Model) renderProgress(lines []string, w int) {
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

	lines[1] = ""

	if len(selected) == 0 {
		lines[2] = dimTextStyle.Render("  No tools selected.")
		return
	}

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

	// Progress bar
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
	pct := 0
	if total > 0 {
		pct = (done + errors) * 100 / total
	}
	lines[2] = fmt.Sprintf("  %s %s %s",
		bar,
		dimTextStyle.Render(fmt.Sprintf("%d/%d", done+errors, total)),
		dimTextStyle.Render(fmt.Sprintf("(%d%%)", pct)))

	lines[3] = ""

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
			icon = statusDoneStyle.Render("OK")
			status = "done"
			style = statusDoneStyle
		case "error":
			icon = statusErrorStyle.Render("!!")
			status = "error"
			style = statusErrorStyle
		default:
			icon = statusPendingStyle.Render("..")
			status = "waiting"
			style = statusPendingStyle
		}

		lines[lineIdx] = fmt.Sprintf("  %s  %-25s %s", icon, st.tool.Name, style.Render(status))
		lineIdx++

		if st.tool.Status == "error" && st.tool.Error != nil && lineIdx < m.Height-2 {
			errMsg := st.tool.Error.Error()
			if len(errMsg) > 60 {
				errMsg = errMsg[:57] + "..."
			}
			lines[lineIdx] = "      " + statusErrorStyle.Render(errMsg)
			lineIdx++
		}
	}

	if m.State == StateDone {
		helpIdx := m.Height - 2
		if errors > 0 {
			lines[helpIdx] = "  " + statusDoneStyle.Render(fmt.Sprintf("%d installed", done)) +
				"  " + statusErrorStyle.Render(fmt.Sprintf("%d failed", errors))
		} else {
			lines[helpIdx] = "  " + statusDoneStyle.Render(fmt.Sprintf("All %d tools installed successfully.", done))
		}
		lines[m.Height-1] = "  " + dimTextStyle.Render("Path: "+m.InstallPath) +
			"  " + helpKeyStyle.Render("q") + helpStyle.Render("/") +
			helpKeyStyle.Render("enter") + helpStyle.Render(" quit")
	} else {
		lines[m.Height-1] = "  " + dimTextStyle.Render("Installing... ") +
			helpKeyStyle.Render("ctrl+c") + helpStyle.Render(" cancel")
	}
}
