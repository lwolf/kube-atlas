package dashboard

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
    table table.Model
    releases []config.Release
}

func New(releases []config.Release) Model {
    columns := []table.Column{
        {Title: "ID", Width: 20},
        {Title: "Namespace", Width: 15},
        {Title: "Chart", Width: 30},
        {Title: "Version", Width: 10},
    }

    rows := make([]table.Row, len(releases))
    for i, r := range releases {
        rows[i] = table.Row{r.ID, r.Namespace, r.Chart, r.Version}
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    s := table.DefaultStyles()
    s.Header = styles.TableHeaderStyle
    s.Selected = styles.SelectedRowStyle
    t.SetStyles(s)

    return Model{
        table:    t,
        releases: releases,
    }
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    base := lipgloss.NewStyle().
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240"))
    return base.Render(m.table.View()) + "\n"
}

func (m Model) SelectedRelease() *config.Release {
    idx := m.table.Cursor()
    if idx >= 0 && idx < len(m.releases) {
        return &m.releases[idx]
    }
    return nil
}
