package app

import (
    "fmt"
    "path/filepath"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/fs"
    "github.com/lwolf/kube-atlas/internal/workspace"
)

type AddReleaseOptions struct {
    ID        string
    Namespace string
    Chart     string
    Version   string
    RepoRoot  string
    NoFetch   bool // Placeholder for future use
    NoRender  bool // Placeholder for future use
}

// AddRelease adds a new release to the configuration and scaffolds directories.
func (a *App) AddRelease(opts AddReleaseOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Check for existing release ID
    for _, r := range cfg.Releases {
        if r.ID == opts.ID {
            return fmt.Errorf("release %q already exists", opts.ID)
        }
    }

    // Add new release
    newRelease := config.Release{
        ID:        opts.ID,
        Namespace: opts.Namespace,
        Chart:     opts.Chart,
        Version:   opts.Version,
    }

    // Validate the modified config (which checks ID constraints, uniqueness, etc.)
    // We append temporarily to validate
    tempCfg := *cfg
    tempCfg.Releases = append(tempCfg.Releases, newRelease)
    if err := tempCfg.Validate(); err != nil {
        return fmt.Errorf("invalid release configuration: %w", err)
    }

    // Scaffold directories
    if err := a.scaffoldRelease(opts.RepoRoot, newRelease); err != nil {
        return fmt.Errorf("failed to scaffold directories: %w", err)
    }

    // Save config
    cfg.Releases = append(cfg.Releases, newRelease)
    if err := cfg.Save(cfgPath); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    fmt.Printf("Added release %q to atlas.yaml\n", opts.ID)
    return nil
}

func (a *App) scaffoldRelease(repoRoot string, release config.Release) error {
    layout := workspace.Resolve(repoRoot, release)
    fsys := fs.RealFS{}

    // Create directories
    dirs := []string{
        layout.OverlayDir,
        layout.ValuesDir,
        layout.PatchesDir,
        layout.ResourcesDir,
        filepath.Join(layout.PatchesDir, "patches"), // Standard structure
    }

    for _, dir := range dirs {
        if err := fsys.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("failed to create directory %q: %w", dir, err)
        }
    }

    // Create 00-base.yaml in values
    baseValues := []byte("# 00-base.yaml - Base values for this release.\n# Files in this directory are merged in lexical order.\n")
    if err := fs.AtomicWrite(fsys, filepath.Join(layout.ValuesDir, "00-base.yaml"), baseValues, 0644); err != nil {
         return fmt.Errorf("failed to create base values: %w", err)
    }

    // Create kustomization.yaml in patches
    kust := []byte("apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources: []\npatches: []\n")
    if err := fs.AtomicWrite(fsys, filepath.Join(layout.PatchesDir, "kustomization.yaml"), kust, 0644); err != nil {
        return fmt.Errorf("failed to create kustomization.yaml: %w", err)
    }

    return nil
}
