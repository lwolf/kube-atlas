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
