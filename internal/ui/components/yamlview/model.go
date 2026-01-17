package yamlview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lwolf/kube-atlas/internal/config"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
	release config.Release
	yaml    string
}

type BackMsg struct{}

func New(release config.Release, yaml string) Model {
	return Model{
		release: release,
		yaml:    yaml,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Rendered YAML: "+m.release.ID) + "\n\n")

	// Display YAML content in a bordered box
	yamlBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Render(m.yaml)

	b.WriteString(yamlBox + "\n\n")
	b.WriteString(styles.StatusStyle.Render("Esc or q: Back"))

	return b.String()
}
