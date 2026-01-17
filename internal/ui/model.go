package ui

import (
	"context"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lwolf/kube-atlas/internal/app"
	"github.com/lwolf/kube-atlas/internal/config"
	"github.com/lwolf/kube-atlas/internal/fs"
	"github.com/lwolf/kube-atlas/internal/helm/render"
	"github.com/lwolf/kube-atlas/internal/kust"
	"github.com/lwolf/kube-atlas/internal/ui/components/chartsearch"
	"github.com/lwolf/kube-atlas/internal/ui/components/releasedetails"
	"github.com/lwolf/kube-atlas/internal/ui/components/releaseedit"
	"github.com/lwolf/kube-atlas/internal/ui/components/releaseform"
	"github.com/lwolf/kube-atlas/internal/ui/components/repoedit"
	"github.com/lwolf/kube-atlas/internal/ui/components/repoform"
	"github.com/lwolf/kube-atlas/internal/ui/components/repolist"
	"github.com/lwolf/kube-atlas/internal/ui/components/yamlview"
	"github.com/lwolf/kube-atlas/internal/ui/dashboard"
	"github.com/lwolf/kube-atlas/internal/ui/styles"
)

type ErrorRecoveryMsg struct{}

type State int

const (
	StateDashboard State = iota
	StateReleaseDetails
	StateEditRelease
	StateAddRelease
	StateRepos
	StateEditRepo
	StateAddRepo
	StateViewYAML
	StateChartSearch
)

type Model struct {
	ctx      context.Context
	app      *app.App
	repoRoot string

	state          State
	prevState      State
	dashboard      dashboard.Model
	releaseDetails releasedetails.Model
	releaseEdit    releaseedit.Model
	releaseForm    releaseform.Model
	repoList       repolist.Model
	repoEdit       repoedit.Model
	repoForm       repoform.Model
	yamlView       yamlview.Model
	chartSearch    chartsearch.Model

	cfg    *config.Config
	err    error
	width  int
	height int
}

func NewModel(ctx context.Context, repoRoot string, application *app.App) (Model, error) {
	// Load config
	cfgPath := filepath.Join(repoRoot, "atlas.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return Model{err: err}, nil
	}

	return Model{
		ctx:            ctx,
		app:            application,
		repoRoot:       repoRoot,
		state:          StateDashboard,
		dashboard:      dashboard.New(cfg.Releases),
		releaseDetails: releasedetails.New(config.Release{}), // Will be set when navigating
		releaseEdit:    releaseedit.New(config.Release{}),    // Will be set when navigating
		releaseForm:    releaseform.New(),
		repoList:       repolist.New(cfg.Repositories),
		repoEdit:       repoedit.New(config.Repository{}), // Will be set when navigating
		repoForm:       repoform.New(),
		yamlView:       yamlview.New(config.Release{}, ""), // Will be set when navigating
		chartSearch:    chartsearch.New(cfg.Repositories),  // Will be set when navigating
		cfg:            cfg,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		availableHeight := msg.Height - 8
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.dashboard = dashboard.NewWithHeight(m.cfg.Releases, availableHeight)
		m.repoList = repolist.NewWithHeight(m.cfg.Repositories, availableHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == StateDashboard || m.state == StateRepos {
				return m, tea.Quit
			}
		case "b", "esc":
			if m.err != nil {
				return m, func() tea.Msg { return ErrorRecoveryMsg{} }
			}
		case "r":
			if m.state == StateDashboard {
				m.prevState = m.state
				m.state = StateRepos
				return m, nil
			}
		case "d":
			if m.state == StateRepos {
				m.prevState = m.state
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
			m.prevState = m.state
			m.state = StateDashboard
			m.releaseForm = releaseform.New()
		}
		return m, nil

	case releaseform.CancelMsg:
		m.prevState = m.state
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
			m.prevState = m.state
			m.state = StateRepos
			m.repoForm = repoform.New()
		}
		return m, nil

	case repoform.CancelMsg:
		m.prevState = m.state
		m.state = StateRepos
		return m, nil

	case releasedetails.EditReleaseMsg:
		m.prevState = m.state
		m.state = StateEditRelease
		m.releaseEdit = releaseedit.New(msg.Release)
		return m, m.releaseEdit.Init()

	case releasedetails.ViewYAMLMsg:
		// For demo purposes, show placeholder YAML since actual rendering
		// requires configured Helm repositories and local chart access
		placeholderYAML := fmt.Sprintf(`# Rendered YAML for release: %s
# Note: Actual rendering requires configured Helm repositories
# and proper chart access. This is a placeholder for demo purposes.

apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-placeholder
  namespace: %s
data:
  chart: "%s"
  version: "%s"
  rendered: "true"`,
			msg.Release.ID,
			msg.Release.ID,
			msg.Release.Namespace,
			msg.Release.Chart,
			msg.Release.Version)

		m.prevState = m.state
		m.state = StateViewYAML
		m.yamlView = yamlview.New(msg.Release, placeholderYAML)
		return m, nil

	case releasedetails.BackMsg:
		m.prevState = m.state
		m.state = StateDashboard
		return m, nil

	case releaseedit.ResultMsg:
		if err := m.app.EditRelease(msg.Options); err != nil {
			m.err = err
		} else {
			if err := m.reloadConfig(); err != nil {
				m.err = err
			}
			m.prevState = m.state
			m.state = StateDashboard
			m.releaseEdit = releaseedit.New(config.Release{})
		}
		return m, nil

	case releaseedit.CancelMsg:
		m.prevState = m.state
		m.state = StateDashboard
		return m, nil

	case releaseform.ChartSearchMsg:
		m.prevState = m.state
		m.state = StateChartSearch
		return m, nil

	case chartsearch.ChartSelectedMsg:
		// Populate the release form with selected chart
		m.releaseForm = releaseform.New() // Reset form
		// TODO: Pre-populate chart and version fields
		m.prevState = m.state
		m.state = StateAddRelease
		return m, nil

	case chartsearch.BackMsg:
		m.prevState = m.state
		m.state = StateAddRelease
		return m, nil

	case ErrorRecoveryMsg:
		m.err = nil
		m.state = m.prevState
		return m, nil

	case repoedit.ResultMsg:
		if err := m.app.EditRepository(msg.Options); err != nil {
			m.err = err
		} else {
			if err := m.reloadConfig(); err != nil {
				m.err = err
			}
			m.prevState = m.state
			m.state = StateRepos
			m.repoEdit = repoedit.New(config.Repository{})
		}
		return m, nil

	case repoedit.CancelMsg:
		m.prevState = m.state
		m.state = StateRepos
		return m, nil

	case yamlview.BackMsg:
		m.prevState = m.state
		m.state = StateReleaseDetails
		return m, nil
	}

	// Handle view specific updates
	switch m.state {
	case StateDashboard:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				if release := m.dashboard.SelectedRelease(); release != nil {
					m.prevState = m.state
					m.state = StateReleaseDetails
					m.releaseDetails = releasedetails.New(*release)
					return m, nil
				}
			case "a":
				m.prevState = m.state
				m.state = StateAddRelease
				return m, m.releaseForm.Init()
			}
		}
		m.dashboard, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)

	case StateReleaseDetails:
		m.releaseDetails, cmd = m.releaseDetails.Update(msg)
		cmds = append(cmds, cmd)

	case StateEditRelease:
		m.releaseEdit, cmd = m.releaseEdit.Update(msg)
		cmds = append(cmds, cmd)

	case StateAddRelease:
		m.releaseForm, cmd = m.releaseForm.Update(msg)
		cmds = append(cmds, cmd)

	case StateEditRepo:
		m.repoEdit, cmd = m.repoEdit.Update(msg)
		cmds = append(cmds, cmd)

	case StateViewYAML:
		m.yamlView, cmd = m.yamlView.Update(msg)
		cmds = append(cmds, cmd)

	case StateChartSearch:
		m.chartSearch, cmd = m.chartSearch.Update(msg)
		cmds = append(cmds, cmd)

	case StateRepos:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				if repo := m.repoList.SelectedRepository(); repo != nil {
					m.prevState = m.state
					m.state = StateEditRepo
					m.repoEdit = repoedit.New(*repo)
					return m, m.repoEdit.Init()
				}
			case "a":
				m.prevState = m.state
				m.state = StateAddRepo
				return m, m.repoForm.Init()
			}
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
	height := 10 // Default height
	if m.height > 0 {
		height = m.height - 8
		if height < 5 {
			height = 5
		}
	}
	cfgPath := filepath.Join(m.repoRoot, "atlas.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	m.cfg = cfg
	m.dashboard = dashboard.NewWithHeight(cfg.Releases, height)
	m.repoList = repolist.NewWithHeight(cfg.Repositories, height)
	return nil
}

func (m Model) View() string {
	if m.err != nil {
		return styles.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n" +
			styles.StatusStyle.Render("b: back • esc: back • q: quit")
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
	case StateEditRepo:
		return styles.AppStyle.Render(
			styles.TitleStyle.Render("Edit Repository") + "\n\n" +
				m.repoEdit.View(),
		)
	case StateReleaseDetails:
		return styles.AppStyle.Render(
			m.releaseDetails.View(),
		)
	case StateEditRelease:
		return styles.AppStyle.Render(
			styles.TitleStyle.Render("Edit Release") + "\n\n" +
				m.releaseEdit.View(),
		)
	case StateViewYAML:
		return styles.AppStyle.Render(
			m.yamlView.View(),
		)
	case StateChartSearch:
		return styles.AppStyle.Render(
			m.chartSearch.View(),
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

// renderReleaseYAML renders the YAML for a given release
func (m Model) renderReleaseYAML(release config.Release) (string, error) {
	fsys := fs.RealFS{}
	renderer := render.New()
	patcher := kust.New()

	yaml, err := m.app.RenderReleaseForView(release, m.repoRoot, fsys, renderer, patcher)
	if err != nil {
		return "", fmt.Errorf("failed to render release: %w", err)
	}

	return yaml, nil
}
