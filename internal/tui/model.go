package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	profile     string
	currentView int
	views       []string
	lastRefresh time.Time
	width       int
	height      int
}

func NewModel(profile string) Model {
	return Model{
		profile: profile,
		views: []string{
			"Dashboard",
			"Quotes",
			"Orders",
			"Portfolio",
		},
		lastRefresh: time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentView = (m.currentView + 1) % len(m.views)
			return m, nil
		case "r":
			m.lastRefresh = time.Now()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("zerodha interactive mode\n")
	b.WriteString(fmt.Sprintf("profile: %s\n", m.profile))
	b.WriteString(fmt.Sprintf("view: %s\n", m.views[m.currentView]))
	b.WriteString(fmt.Sprintf("last refresh: %s\n", m.lastRefresh.Format("2006-01-02 15:04:05")))
	b.WriteString("\nKeys: tab=next view, r=refresh timestamp, q=quit\n")
	b.WriteString("\nThis TUI scaffold is ready for account-aware widgets.\n")
	return b.String()
}
