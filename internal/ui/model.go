package ui

import (
    "context"
    "fmt"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/lwolf/kube-atlas/internal/app"
    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/ui/components/releaseform"
    "github.com/lwolf/kube-atlas/internal/ui/dashboard"
    "github.com/lwolf/kube-atlas/internal/ui/styles"
)

type State int

const (
    StateDashboard State = iota
    StateAddRelease
)

type Model struct {
    ctx      context.Context
    app      *app.App
    repoRoot string
    
    state       State
    dashboard   dashboard.Model
    releaseForm releaseform.Model
    
    cfg       *config.Config
    err       error
}

func NewModel(ctx context.Context, repoRoot string, application *app.App) (Model, error) {
    // Load config
    cfgPath := filepath.Join(repoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return Model{err: err}, nil
    }

    return Model{
        ctx:       ctx,
        app:       application,
        repoRoot:  repoRoot,
        state:     StateDashboard,
        dashboard: dashboard.New(cfg.Releases),
        releaseForm: releaseform.New(),
        cfg:       cfg,
    }, nil
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            if m.state == StateDashboard {
                return m, tea.Quit
            }
        }
    
    case releaseform.ResultMsg:
        msg.Options.RepoRoot = m.repoRoot
        if err := m.app.AddRelease(msg.Options); err != nil {
             m.err = err
        } else {
             // Reload config and switch back
             if err := m.reloadConfig(); err != nil {
                 m.err = err
             }
             m.state = StateDashboard
             // Reset form?
             m.releaseForm = releaseform.New()
        }
        return m, nil

    case releaseform.CancelMsg:
        m.state = StateDashboard
        return m, nil
    }

    // Handle view specific updates
    switch m.state {
    case StateDashboard:
        // Intercept 'a' for add
        if key, ok := msg.(tea.KeyMsg); ok && key.String() == "a" {
            m.state = StateAddRelease
            return m, m.releaseForm.Init() // Focus form
        }
        
        m.dashboard, cmd = m.dashboard.Update(msg)
        cmds = append(cmds, cmd)

    case StateAddRelease:
        m.releaseForm, cmd = m.releaseForm.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}

func (m *Model) reloadConfig() error {
    cfgPath := filepath.Join(m.repoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return err
    }
    m.cfg = cfg
    m.dashboard = dashboard.New(cfg.Releases)
    return nil
}

func (m Model) View() string {
    if m.err != nil {
        // Simple error view
        return styles.ErrorStyle.Render("Error: " + m.err.Error()) + "\nPress q to quit."
    }

    switch m.state {
    case StateAddRelease:
         return styles.AppStyle.Render(
            styles.TitleStyle.Render("Add New Release") + "\n\n" +
            m.releaseForm.View(),
        )
    default:
        return styles.AppStyle.Render(
            styles.TitleStyle.Render("Atlas TUI") + "\n\n" +
            m.dashboard.View() + "\n" +
            styles.StatusStyle.Render(fmt.Sprintf("Checking: %s", m.repoRoot)) + "\n" +
            styles.StatusStyle.Render("a: Add Release | q: Quit"),
        )
    }
}
