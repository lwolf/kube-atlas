package app

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/workspace"
)

type DriftOptions struct {
    RepoRoot  string
    ReleaseID string
}

func (a *App) Drift(ctx context.Context, opts DriftOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    for _, r := range cfg.Releases {
        if opts.ReleaseID != "" && r.ID != opts.ReleaseID {
            continue
        }

        layout := workspace.Resolve(opts.RepoRoot, r)

        // Ensure rendered file exists
        if _, err := os.Stat(layout.RenderedFile); os.IsNotExist(err) {
            return fmt.Errorf("release %s is not rendered (run 'atlas render' first)", r.ID)
        }

        fmt.Printf("Checking drift for %s...\n", r.ID)
        
        // constructed command: kubectl diff -f <rendered-file>
        cmd := exec.CommandContext(ctx, "kubectl", "diff", "-f", layout.RenderedFile)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        
        err := cmd.Run()
        if err != nil {
            if exitError, ok := err.(*exec.ExitError); ok {
                if exitError.ExitCode() == 1 {
                    // Exit code 1 means drift detected
                    fmt.Printf("Drift detected for %s\n", r.ID)
                    continue // Continue checking other releases
                }
            }
            return fmt.Errorf("kubectl diff failed for %s: %w", r.ID, err)
        }
        
        fmt.Printf("No drift for %s\n", r.ID)
    }

    return nil
}
