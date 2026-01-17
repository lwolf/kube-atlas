package chartsearch

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lwolf/kube-atlas/internal/config"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
	repos       []config.Repository
	searchInput textinput.Model
	results     []string
	selected    int
	loading     bool
	query       string
}

type ChartSelectedMsg struct {
	Chart   string
	Version string
}

type BackMsg struct{}

func New(repos []config.Repository) Model {
	ti := textinput.New()
	ti.Placeholder = "Search for charts..."
	ti.Focus()

	return Model{
		repos:       repos,
		searchInput: ti,
		results:     []string{},
		selected:    0,
		loading:     false,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.loading {
				return m, nil
			}
			if m.selected >= 0 && m.selected < len(m.results) {
				// For now, just return the chart name with a default version
				// In a full implementation, we'd get the latest version
				return m, func() tea.Msg {
					return ChartSelectedMsg{
						Chart:   m.results[m.selected],
						Version: "latest",
					}
				}
			}
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.results)-1 {
				m.selected++
			}
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}

	case SearchResultsMsg:
		m.results = msg.Results
		m.loading = false
		if len(m.results) > 0 && m.selected >= len(m.results) {
			m.selected = len(m.results) - 1
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Check if search query changed
	newQuery := m.searchInput.Value()
	if newQuery != m.query && newQuery != "" {
		m.query = newQuery
		m.loading = true
		m.results = []string{}
		m.selected = 0
		return m, tea.Cmd(func() tea.Msg {
			// Simulate async search - in real implementation, this would search across repos
			results := []string{}
			for _, repo := range m.repos {
				// This is a placeholder - real implementation would fetch and search repo indices
				if strings.Contains(strings.ToLower(repo.Name), strings.ToLower(newQuery)) {
					results = append(results, fmt.Sprintf("%s/%s", repo.Name, "example-chart"))
				}
			}
			return SearchResultsMsg{Results: results}
		})
	}

	return m, cmd
}

type SearchResultsMsg struct {
	Results []string
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Search Charts") + "\n\n")

	// Search input
	b.WriteString("Search: " + m.searchInput.View() + "\n\n")

	if m.loading {
		b.WriteString("Searching...\n\n")
	} else if len(m.results) == 0 && m.query != "" {
		b.WriteString("No charts found.\n\n")
	} else if len(m.results) > 0 {
		b.WriteString("Results:\n")
		for i, result := range m.results {
			if i == m.selected {
				b.WriteString(selectedItemStyle.Render(fmt.Sprintf("▶ %s", result)) + "\n")
			} else {
				b.WriteString(itemStyle.Render(fmt.Sprintf("  %s", result)) + "\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.StatusStyle.Render("Enter: Select • ↑/↓ or j/k: Navigate • Esc: Back"))

	return b.String()
}

var (
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)
