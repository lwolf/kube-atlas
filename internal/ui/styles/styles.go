package styles

import (
    "github.com/charmbracelet/lipgloss"
)

var (
    AppStyle = lipgloss.NewStyle().Margin(1, 2)

    TitleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFDF5")).
        Background(lipgloss.Color("#25A065")).
        Padding(0, 1)

    StatusStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#A4A4A4"))

    ErrorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF3333"))

    // Table styles
    TableHeaderStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Bold(true)
    
    SelectedRowStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("57")).
        Foreground(lipgloss.Color("229"))
)
