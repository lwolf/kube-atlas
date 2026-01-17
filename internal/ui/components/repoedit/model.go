package repoedit

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lwolf/kube-atlas/internal/app"
	"github.com/lwolf/kube-atlas/internal/config"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
	repo       config.Repository
	focusIndex int
	inputs     []textinput.Model
}

type ResultMsg struct {
	Options app.EditRepositoryOptions
}

type CancelMsg struct{}

func New(repo config.Repository) Model {
	m := Model{
		repo:   repo,
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = styles.TitleStyle
		t.CharLimit = 200

		switch i {
		case 0: // Name - immutable
			t.Placeholder = "Name (immutable)"
			t.SetValue(repo.Name)
			t.Blur() // Keep blurred since it's immutable
		case 1: // URL
			t.Placeholder = "URL (e.g. https://charts.bitnami.com/bitnami)"
			t.SetValue(repo.URL)
		}

		m.inputs[i] = t
	}

	// Focus the first editable field (URL)
	m.focusIndex = 1
	m.inputs[1].Focus()

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return CancelMsg{} }

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, m.submit()
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 1 // Skip name field
			} else if m.focusIndex < 1 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i > 0 { // Skip name field (index 0)
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
	}
	return tea.Batch(cmds...)
}

func (m Model) submit() tea.Cmd {
	return func() tea.Msg {
		return ResultMsg{
			Options: app.EditRepositoryOptions{
				Name: m.repo.Name, // Keep original name
				URL:  m.inputs[1].Value(),
			},
		}
	}
}

func (m Model) View() string {
	var b strings.Builder

	inputs := []string{
		"Name (immutable)",
		"URL",
	}

	for i := range m.inputs {
		b.WriteString(inputs[i] + "\n")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	button := styles.StatusStyle.Render("[ Update Repository ]")
	if m.focusIndex == len(m.inputs) {
		button = styles.TitleStyle.Render("[ Update Repository ]")
	}

	b.WriteString(button)

	return b.String()
}
