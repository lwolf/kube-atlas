package repolist

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/ui/styles"
)

type Model struct {
    table table.Model
}

func New(repos []config.Repository) Model {
    columns := []table.Column{
        {Title: "Name", Width: 15},
        {Title: "URL", Width: 40},
    }

    var rows []table.Row
    for _, r := range repos {
        rows = append(rows, table.Row{r.Name, r.URL})
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    s := styles.TableStyles()
    t.SetStyles(s)

    return Model{table: t}
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
    return m.table.View()
}
