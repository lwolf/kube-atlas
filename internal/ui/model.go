package ui

import (
    "context"
    "fmt"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/lwolf/kube-atlas/internal/app"
    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/ui/components/releaseform"
    "github.com/lwolf/kube-atlas/internal/ui/components/repoform"
    "github.com/lwolf/kube-atlas/internal/ui/components/repolist"
    "github.com/lwolf/kube-atlas/internal/ui/dashboard"
    "github.com/lwolf/kube-atlas/internal/ui/styles"
)

type State int

const (
    StateDashboard State = iota
    StateAddRelease
    StateRepos
    StateAddRepo
)

type Model struct {
    ctx      context.Context
    app      *app.App
    repoRoot string
    
    state       State
    dashboard   dashboard.Model
    releaseForm releaseform.Model
    repoList    repolist.Model
    repoForm    repoform.Model
    
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
        ctx:         ctx,
        app:         application,
        repoRoot:    repoRoot,
        state:       StateDashboard,
        dashboard:   dashboard.New(cfg.Releases),
        releaseForm: releaseform.New(),
        repoList:    repolist.New(cfg.Repositories),
        repoForm:    repoform.New(),
        cfg:         cfg,
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
            if m.state == StateDashboard || m.state == StateRepos {
                return m, tea.Quit
            }
        case "r":
            if m.state == StateDashboard {
                m.state = StateRepos
                return m, nil
            }
        case "d":
            if m.state == StateRepos {
                m.state = StateDashboard
                return m, nil
            }
        }
    
    case releaseform.ResultMsg:
        msg.Options.RepoRoot = m.repoRoot
        if err := m.app.AddRelease(msg.Options); err != nil {
             m.err = err
        } else {
             if err := m.reloadConfig(); err != nil {
                 m.err = err
             }
             m.state = StateDashboard
             m.releaseForm = releaseform.New()
        }
        return m, nil

    case releaseform.CancelMsg:
        m.state = StateDashboard
        return m, nil

    case repoform.ResultMsg:
        msg.Options.RepoRoot = m.repoRoot
        // AddRepository logic
        if err := m.app.AddRepository(msg.Options); err != nil {
             m.err = err
        } else {
             if err := m.reloadConfig(); err != nil {
                 m.err = err
             }
             m.state = StateRepos
             m.repoForm = repoform.New()
        }
        return m, nil

    case repoform.CancelMsg:
        m.state = StateRepos
        return m, nil
    }

    // Handle view specific updates
    switch m.state {
    case StateDashboard:
        if key, ok := msg.(tea.KeyMsg); ok && key.String() == "a" {
            m.state = StateAddRelease
            return m, m.releaseForm.Init()
        }
        m.dashboard, cmd = m.dashboard.Update(msg)
        cmds = append(cmds, cmd)

    case StateAddRelease:
        m.releaseForm, cmd = m.releaseForm.Update(msg)
        cmds = append(cmds, cmd)

    case StateRepos:
        if key, ok := msg.(tea.KeyMsg); ok && key.String() == "a" {
            m.state = StateAddRepo
            return m, m.repoForm.Init()
        }
        m.repoList, cmd = m.repoList.Update(msg)
        cmds = append(cmds, cmd)

    case StateAddRepo:
        m.repoForm, cmd = m.repoForm.Update(msg)
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
    m.repoList = repolist.New(cfg.Repositories)
    return nil
}

func (m Model) View() string {
    if m.err != nil {
        return styles.ErrorStyle.Render("Error: " + m.err.Error()) + "\nPress q to quit."
    }

    switch m.state {
    case StateAddRelease:
         return styles.AppStyle.Render(
            styles.TitleStyle.Render("Add New Release") + "\n\n" +
            m.releaseForm.View(),
        )
    case StateRepos:
        return styles.AppStyle.Render(
            styles.TitleStyle.Render("Repositories") + "\n\n" +
            m.repoList.View() + "\n" +
            styles.StatusStyle.Render(fmt.Sprintf("Checking: %s", m.repoRoot)) + "\n" +
            styles.StatusStyle.Render("a: Add Repo | d: Dashboard | q: Quit"),
        )
    case StateAddRepo:
         return styles.AppStyle.Render(
            styles.TitleStyle.Render("Add New Repository") + "\n\n" +
            m.repoForm.View(),
        )
    default:
        return styles.AppStyle.Render(
            styles.TitleStyle.Render("Atlas TUI") + "\n\n" +
            m.dashboard.View() + "\n" +
            styles.StatusStyle.Render(fmt.Sprintf("Checking: %s", m.repoRoot)) + "\n" +
            styles.StatusStyle.Render("a: Add Release | r: Repos | q: Quit"),
        )
    }
}
