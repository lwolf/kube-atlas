package repoform

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lwolf/kube-atlas/internal/app"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
}

type ResultMsg struct {
	Options app.AddRepositoryOptions
}

type CancelMsg struct{}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = styles.TitleStyle
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "Name (e.g. bitnami)"
			t.Focus()
		case 1:
			t.Placeholder = "URL (e.g. https://charts.bitnami.com/bitnami)"
		}

		m.inputs[i] = t
	}

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
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
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
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m Model) submit() tea.Cmd {
	return func() tea.Msg {
		return ResultMsg{
			Options: app.AddRepositoryOptions{
				Name: m.inputs[0].Value(),
				URL:  m.inputs[1].Value(),
			},
		}
	}
}

func (m Model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := styles.StatusStyle.Render("[ Submit ]")
	if m.focusIndex == len(m.inputs) {
		button = styles.TitleStyle.Render("[ Submit ]")
	}

	b.WriteString("\n\n")
	b.WriteString(button)

	return b.String()
}
