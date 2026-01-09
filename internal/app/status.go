package app

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/fs"
    "github.com/lwolf/kube-atlas/internal/helm/render"
    "github.com/lwolf/kube-atlas/internal/kust"
    "github.com/lwolf/kube-atlas/internal/lock"
    "github.com/lwolf/kube-atlas/internal/workspace"
    "github.com/sergi/go-diff/diffmatchpatch"
)

// Check Status
type StatusOptions struct {
    RepoRoot string
}

func (a *App) Status(opts StatusOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return err
    }

    lockPath := filepath.Join(opts.RepoRoot, "atlas", "lock.yaml")
    lck, err := lock.Load(lockPath)
    if err != nil {
        lck = &lock.Lockfile{}
    }

    fmt.Printf("%-20s %-15s %-15s\n", "RELEASE", "CHART STATUS", "RENDER STATUS")
    fmt.Println("---------------------------------------------------------")

    for _, r := range cfg.Releases {
        layout := workspace.Resolve(opts.RepoRoot, r)
        
        // Check Chart Status
        chartStatus := "OK"
        if _, err := os.Stat(layout.ChartDir); os.IsNotExist(err) {
            chartStatus = "MISSING"
        } else {
             // Check Hash
             _, ok := lck.Get(r.ID)
             if !ok {
                 chartStatus = "UNKNOWN"
             } else {
                 if dirty, err := lck.IsDirty(r.ID, layout.ChartDir); err != nil {
                     chartStatus = "ERR"
                 } else if dirty {
                     chartStatus = "DIRTY"
                 } else {
                     // Also check if lock matches installed version? 
                     // IsDirty checks content vs Digest.
                     // But if repo updated version but didn't fetch?
                     // That's config vs lock mismatch, separate check.
                 }
             }
        }
        
        // Check Render Status
        renderStatus := "OK"
        outFile := layout.RenderedFile
        if _, err := os.Stat(outFile); os.IsNotExist(err) {
            renderStatus = "MISSING"
        } else {
            // We could compare with in-memory render? content-wise?
            // "atlas status" usually implicitly checks if drift?
            // Let's just check existence for basic status.
            // Or hash check if we stored render hash.
        }

        fmt.Printf("%-20s %-15s %-15s\n", r.ID, chartStatus, renderStatus)
    }

    return nil
}

// Diff
type DiffOptions struct {
    RepoRoot string
    ReleaseID string
}

func (a *App) Diff(opts DiffOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return err
    }

    renderer := render.New()
    patcher := kust.New()
    fsys := fs.RealFS{}

    for _, r := range cfg.Releases {
        if opts.ReleaseID != "" && r.ID != opts.ReleaseID {
            continue
        }

        fmt.Printf("Diffing %s...\n", r.ID)

        layout := workspace.Resolve(opts.RepoRoot, r)

        // 1. Render in-memory
        newContent, err := a.renderRelease(r, opts.RepoRoot, fsys, renderer, patcher)
        if err != nil {
            return fmt.Errorf("render failed: %w", err)
        }

        // 2. Read existing
        outFile := layout.RenderedFile
        oldContentBytes, err := os.ReadFile(outFile)
        oldContent := string(oldContentBytes)
        if err != nil && !os.IsNotExist(err) {
            return fmt.Errorf("failed to read existing render: %w", err)
        }

        // 3. Diff
        dmp := diffmatchpatch.New()
        diffs := dmp.DiffMain(oldContent, newContent, true)
        
        // Pretty print
        // Only print if changes
        if len(diffs) == 0 || (len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual) {
             fmt.Println("No changes.")
        } else {
             fmt.Println(dmp.DiffPrettyText(diffs))
        }
    }
    return nil
}
