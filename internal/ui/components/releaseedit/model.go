package releaseedit

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lwolf/kube-atlas/internal/app"
	"github.com/lwolf/kube-atlas/internal/config"
)

type Model struct {
	release config.Release
	inputs  []textinput.Model
	focused int
	err     error
}

type ResultMsg struct {
	Options app.EditReleaseOptions
}

type CancelMsg struct{}

func New(release config.Release) Model {
	m := Model{
		release: release,
		inputs:  make([]textinput.Model, 3), // ID (disabled), Chart, Version
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		t.CharLimit = 200

		switch i {
		case 0: // ID - immutable
			t.Placeholder = "Release ID (immutable)"
			t.SetValue(release.ID)
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			t.Blur() // Keep blurred since it's immutable
		case 1: // Chart
			t.Placeholder = "Chart (e.g. oci://... or repo/chart)"
			t.SetValue(release.Chart)
			if m.focused == 1 {
				t.Focus()
				t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			}
		case 2: // Version
			t.Placeholder = "Version (e.g. 1.0.0)"
			t.SetValue(release.Version)
		}

		m.inputs[i] = t
	}

	// Focus the first editable field (Chart)
	m.focused = 1
	m.inputs[1].Focus()
	m.inputs[1].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.inputs[1].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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
			if s == "enter" && m.focused == len(m.inputs) {
				return m, func() tea.Msg {
					return ResultMsg{
						Options: app.EditReleaseOptions{
							ID:      m.release.ID, // Keep original ID
							Chart:   m.inputs[1].Value(),
							Version: m.inputs[2].Value(),
						},
					}
				}
			}

			// Cycle indexes (skip ID field at index 0)
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs) {
				m.focused = 1 // Skip ID field
			} else if m.focused < 1 {
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
				if i == 0 {
					// ID field stays dimmed
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				} else {
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				}
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
		if i > 0 { // Skip ID field (index 0)
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	inputs := []string{
		"ID (immutable)",
		"Chart",
		"Version",
	}

	for i := range m.inputs {
		b.WriteString(inputs[i] + "\n")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	button := "[ Update Release ]"
	if m.focused == len(m.inputs) {
		button = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("[ Update Release ]")
	}
	b.WriteString(button + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Esc to cancel"))

	return b.String()
}
