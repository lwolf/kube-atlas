package app

import (
    "fmt"
    "path/filepath"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/fs"
    "github.com/lwolf/kube-atlas/internal/helm"
    "github.com/lwolf/kube-atlas/internal/lock"
    "github.com/lwolf/kube-atlas/internal/workspace"
)

type FetchOptions struct {
    RepoRoot string
    // Filter by release ID if needed
    ReleaseID string
}

// Fetch downloads charts for releases and updates the lockfile.
func (a *App) Fetch(opts FetchOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    lockPath := filepath.Join(opts.RepoRoot, "atlas", "lock.yaml")
    lck, err := lock.Load(lockPath)
    if err != nil {
        // If not exists, create empty
        lck = &lock.Lockfile{}
    }

    fetcher := helm.NewFetcher(cfg)
    fsys := fs.RealFS{}

    for _, r := range cfg.Releases {
        if opts.ReleaseID != "" && r.ID != opts.ReleaseID {
            continue
        }

        fmt.Printf("Fetching %s (%s)...\n", r.ID, r.Chart)

        layout := workspace.Resolve(opts.RepoRoot, r)
        
        // Ensure chart dir exists
        if err := fsys.MkdirAll(layout.ChartDir, 0755); err != nil {
            return fmt.Errorf("failed to create chart dir: %w", err)
        }

        // Fetch
        resolved, err := fetcher.Fetch(r, layout.ChartDir)
        if err != nil {
            return fmt.Errorf("failed to fetch release %q: %w", r.ID, err)
        }

        // Compute hash of the vendored chart to detect dirty changes later
        h, err := lock.HashTree(layout.ChartDir)
        if err != nil {
            return fmt.Errorf("failed to hash chart dir: %w", err)
        }
        
        relLock := lock.ReleaseLock{
            Resolved:   *resolved,
            VendorHash: h,
        }

        lck.Update(r.ID, relLock)
    }

    if err := lck.Save(lockPath); err != nil {
        return fmt.Errorf("failed to save lockfile: %w", err)
    }

    return nil
}
