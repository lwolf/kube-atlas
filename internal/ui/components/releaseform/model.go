package releaseform

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lwolf/kube-atlas/internal/app"
)

type Model struct {
	inputs  []textinput.Model
	focused int
	err     error
}

// ResultMsg is sent when the form is submitted successfully
type ResultMsg struct {
	Options app.AddReleaseOptions
}

// CancelMsg is sent when the form is cancelled
type CancelMsg struct{}

type ChartSearchMsg struct{}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "Release ID (e.g. my-app)"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		case 1:
			t.Placeholder = "Namespace (e.g. default)"
		case 2:
			t.Placeholder = "Chart (e.g. oci://... or repo/chart)"
			t.CharLimit = 200
		case 3:
			t.Placeholder = "Version (e.g. 1.0.0)"
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
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" {
				if m.focused == len(m.inputs) {
					// Search Charts button
					return m, func() tea.Msg { return ChartSearchMsg{} }
				} else if m.focused == len(m.inputs)+1 {
					// Submit button
					return m, m.submit()
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs)+1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) + 1
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs) {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focused {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}

			return m, tea.Batch(cmds...)

		case "esc":
			return m, func() tea.Msg { return CancelMsg{} }
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	inputs := []string{
		"ID",
		"Namespace",
		"Chart",
		"Version",
	}

	for i := range m.inputs {
		b.WriteString(inputs[i] + "\n")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	searchButton := "[ Search Charts ]"
	if m.focused == len(m.inputs) {
		searchButton = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("[ Search Charts ]")
	}

	submitButton := "[ Submit ]"
	if m.focused == len(m.inputs)+1 {
		submitButton = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("[ Submit ]")
	}

	b.WriteString(searchButton + "  " + submitButton + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Esc to cancel"))

	return b.String()
}

func (m Model) submit() tea.Cmd {
	return func() tea.Msg {
		return ResultMsg{
			Options: app.AddReleaseOptions{
				ID:        m.inputs[0].Value(),
				Namespace: m.inputs[1].Value(),
				Chart:     m.inputs[2].Value(),
				Version:   m.inputs[3].Value(),
			},
		}
	}
}
