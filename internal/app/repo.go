package app

import (
	"fmt"
	"path/filepath"

	"github.com/lwolf/kube-atlas/internal/config"
)

type AddRepositoryOptions struct {
	RepoRoot string
	Name     string
	URL      string
}

type EditRepositoryOptions struct {
	Name string // Identifies which repository to edit
	URL  string
}

func (a *App) AddRepository(opts AddRepositoryOptions) error {
	cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")

	// Load existing or create new if not exists?
	// Usually Load fails if not exists.
	// Let's assume user calls `atlas init` or manually touches it,
	// OR we be nice and create it if adding repo?
	// For now, assume config exists (as per Load contract).
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config (init first?): %w", err)
	}

	// Check duplicate
	for _, r := range cfg.Repositories {
		if r.Name == opts.Name {
			return fmt.Errorf("repository %q already exists", opts.Name)
		}
	}

	cfg.Repositories = append(cfg.Repositories, config.Repository{
		Name: opts.Name,
		URL:  opts.URL,
	})

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added repository %q (%s)\n", opts.Name, opts.URL)
	return nil
}

// EditRepository updates an existing repository's configuration.
func (a *App) EditRepository(opts EditRepositoryOptions) error {
	cfgPath := "atlas.yaml" // Assume we're in repo root
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	found := false
	for i, repo := range cfg.Repositories {
		if repo.Name == opts.Name {
			cfg.Repositories[i].URL = opts.URL
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("repository %q not found", opts.Name)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Updated repository %q in atlas.yaml\n", opts.Name)
	return nil
}
