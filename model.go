package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab int

const (
	tabOverview tab = iota
	tabDetails
	tabRaw
	tabCount
)

const pollInterval = 2 * time.Second

type pollMsg struct{}

type model struct {
	supplies    []PowerSupply
	selected    int
	activeTab   tab
	width       int
	height      int
	loading     bool
	lastUpdated time.Time
	spinner     spinner.Model
	ready       bool
	viewport    viewport.Model
	content     string
}

func initialModel() model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
	s.Spinner = spinner.Dot

	return model{
		supplies:  []PowerSupply{},
		selected:  0,
		activeTab: tabOverview,
		loading:   true,
		spinner:   s,
		viewport:  viewport.New(80, 20),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg { return pollMsg{} },
	)
}

func (m model) regenerateContent() string {
	switch m.activeTab {
	case tabOverview:
		return m.renderOverview()
	case tabDetails:
		return m.renderDetails()
	case tabRaw:
		return m.renderRaw()
	}
	return ""
}

func (m model) updateViewport() model {
	m.content = m.regenerateContent()
	m.viewport.SetContent(m.content)
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		m = m.updateViewport()
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "l":
			m.activeTab = (m.activeTab + 1) % tabCount
			m = m.updateViewport()
			m.viewport.GotoTop()
			return m, nil
		case "h":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m = m.updateViewport()
			m.viewport.GotoTop()
			return m, nil
		case "1":
			m.activeTab = tabOverview
			m = m.updateViewport()
			m.viewport.GotoTop()
			return m, nil
		case "2":
			m.activeTab = tabDetails
			m = m.updateViewport()
			m.viewport.GotoTop()
			return m, nil
		case "3":
			m.activeTab = tabRaw
			m = m.updateViewport()
			m.viewport.GotoTop()
			return m, nil
		case "j", "down":
			if m.activeTab == tabOverview {
				m.selected = min(m.selected+1, max(0, len(m.supplies)-1))
			}
			return m, nil
		case "k", "up":
			if m.activeTab == tabOverview {
				m.selected = max(m.selected-1, 0)
			}
			return m, nil
		case "r":
			m.loading = true
			return m, func() tea.Msg { return pollMsg{} }
		case "g":
			if m.activeTab == tabOverview {
				m.selected = 0
			}
			return m, nil
		case "G":
			if m.activeTab == tabOverview {
				m.selected = max(0, len(m.supplies)-1)
			}
			return m, nil
		}

	case pollMsg:
		supplies := ScanPowerSupplies()
		m.supplies = supplies
		m.selected = clamp(m.selected, 0, max(0, len(supplies)-1))
		m.loading = false
		m.lastUpdated = time.Now()
		m = m.updateViewport()
		return m, tea.Tick(pollInterval, func(t time.Time) tea.Msg {
			return pollMsg{}
		})

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.renderHeader()
	tabs := m.renderTabs()
	footer := m.renderFooter()

	contentStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	content := contentStyle.Render(m.viewport.View())

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		tabs,
		content,
		footer,
	)
}

func (m model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7aa2f7")).
		Render("battui")

	var updated string
	if !m.lastUpdated.IsZero() {
		updated = dimStyle.Render(m.lastUpdated.Format("15:04:05"))
	}

	var status string
	if m.loading {
		status = m.spinner.View() + " scanning..."
	} else {
		count := 0
		for _, s := range m.supplies {
			if s.IsBattery() {
				count++
			}
		}
		if count > 0 {
			status = accentStyle.Render(fmt.Sprintf("%d battery", count))
		} else {
			status = criticalStyle.Render("no battery")
		}
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		lipgloss.NewStyle().Width(1).Render(""),
		status,
	)

	right := lipgloss.JoinHorizontal(lipgloss.Center, updated)

	fill := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right)-4)
	padding := dimStyle.Render(strings.Repeat("─", fill))

	bar := lipgloss.JoinHorizontal(lipgloss.Center, left, padding, right)
	return headerBarStyle.Width(m.width).Render(bar)
}

func (m model) renderTabs() string {
	names := []string{"Overview", "Details", "All Props"}
	var tabs []string
	for i, name := range names {
		if tab(i) == m.activeTab {
			tabs = append(tabs, tabActiveStyle.Render(name))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(name))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, tabs...)
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#24283b")).
		Width(m.width).
		Padding(0, 1).
		Render(tabBar)
}

func (m model) renderFooter() string {
	hint := func(k, d string) string {
		return lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7aa2f7")).Render(k),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89")).Render(" "+d),
		)
	}

	bindings := lipgloss.JoinHorizontal(lipgloss.Center,
		hint("q", "quit"),
		dimStyle.Render(" · "),
		hint("hl/Tab", "tabs"),
		dimStyle.Render(" · "),
		hint("↑↓/jk", "select"),
		dimStyle.Render(" · "),
		hint("r", "refresh"),
		dimStyle.Render(" · "),
		hint("1-3", "tabs"),
	)

	right := ""
	if len(m.supplies) > 0 {
		right = dimStyle.Render(fmt.Sprintf("%d devices", len(m.supplies)))
	}

	fill := max(0, m.width-lipgloss.Width(bindings)-lipgloss.Width(right)-4)
	padding := strings.Repeat(" ", fill)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, bindings, padding, right)
	return footerStyle.Width(m.width).Render(bar)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}


