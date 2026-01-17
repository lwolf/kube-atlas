package releasedetails

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lwolf/kube-atlas/internal/config"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
	release  config.Release
	selected int
}

type EditReleaseMsg struct {
	Release config.Release
}

type ViewYAMLMsg struct {
	Release config.Release
}

type BackMsg struct{}

func New(release config.Release) Model {
	return Model{
		release:  release,
		selected: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < 2 {
				m.selected++
			}
		case "enter":
			switch m.selected {
			case 0: // Edit Release
				return m, func() tea.Msg { return EditReleaseMsg{Release: m.release} }
			case 1: // View YAML
				return m, func() tea.Msg { return ViewYAMLMsg{Release: m.release} }
			case 2: // Back
				return m, func() tea.Msg { return BackMsg{} }
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Release info
	b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Release: %s", m.release.ID)) + "\n\n")
	b.WriteString(fmt.Sprintf("ID: %s\n", m.release.ID))
	b.WriteString(fmt.Sprintf("Namespace: %s\n", m.release.Namespace))
	b.WriteString(fmt.Sprintf("Chart: %s\n", m.release.Chart))
	b.WriteString(fmt.Sprintf("Version: %s\n\n", m.release.Version))

	// Action menu
	actions := []string{"Edit Release", "View YAML", "Back"}
	for i, action := range actions {
		if i == m.selected {
			b.WriteString(selectedItemStyle.Render(fmt.Sprintf("▶ %s", action)) + "\n")
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("  %s", action)) + "\n")
		}
	}

	b.WriteString("\n" + styles.StatusStyle.Render("↑/↓ or j/k: navigate • enter: select • esc: back"))

	return b.String()
}

var (
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)
