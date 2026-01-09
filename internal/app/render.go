package app

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/fs"
    "github.com/lwolf/kube-atlas/internal/helm/render"
    "github.com/lwolf/kube-atlas/internal/helm/values"
    "github.com/lwolf/kube-atlas/internal/kust"
    "github.com/lwolf/kube-atlas/internal/normalize"
    "github.com/lwolf/kube-atlas/internal/policy"
    "github.com/lwolf/kube-atlas/internal/resources"
    "github.com/lwolf/kube-atlas/internal/validate"
    "github.com/lwolf/kube-atlas/internal/workspace"
)

type RenderOptions struct {
    RepoRoot string
    ReleaseID string
}

func (a *App) Render(opts RenderOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    renderer := render.New()
    patcher := kust.New()
    fsys := fs.RealFS{}

    for _, r := range cfg.Releases {
        if opts.ReleaseID != "" && r.ID != opts.ReleaseID {
            continue
        }

        // Render logic extracted
        out, err := a.renderRelease(r, opts.RepoRoot, fsys, renderer, patcher)
        if err != nil {
            return fmt.Errorf("failed to render %s: %w", r.ID, err)
        }

        // 8. Write Output
        layout := workspace.Resolve(opts.RepoRoot, r)
        if err := fsys.MkdirAll(layout.OutputDir, 0755); err != nil {
            return err
        }
        outFile := layout.RenderedFile
        if err := fs.AtomicWrite(fsys, outFile, []byte(out), 0644); err != nil {
            return fmt.Errorf("failed to write output: %w", err)
        }
        
        fmt.Printf("Rendered to %s\n", outFile)
    }
    
    return nil
}

type CheckOptions struct {
    RepoRoot string
    ReleaseID string
}

func (a *App) Check(ctx context.Context, opts CheckOptions) error {
    cfgPath := filepath.Join(opts.RepoRoot, "atlas.yaml")
    cfg, err := config.Load(cfgPath)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    if len(cfg.Policy.Command) == 0 {
        return fmt.Errorf("no policy command configured in atlas.yaml")
    }
    
    checker := policy.New(cfg.Policy.Command)
    
    for _, r := range cfg.Releases {
        if opts.ReleaseID != "" && r.ID != opts.ReleaseID {
            continue
        }
        
        layout := workspace.Resolve(opts.RepoRoot, r)
        
        // Ensure rendered file exists
        if _, err := os.Stat(layout.RenderedFile); os.IsNotExist(err) {
             return fmt.Errorf("release %s is not rendered (run 'atlas render' first)", r.ID)
        }
        
        fmt.Printf("Checking %s (%s)...\n", r.ID, layout.RenderedFile)
        if err := checker.Check(ctx, layout.RenderedFile); err != nil {
             return fmt.Errorf("policy check failed for %s: %w", r.ID, err)
        }
    }
    
    return nil
}

func (a *App) renderRelease(r config.Release, repoRoot string, fsys fs.FS, renderer *render.Engine, patcher *kust.Patcher) (string, error) {
        layout := workspace.Resolve(repoRoot, r)
        
        buildDir, err := os.MkdirTemp("", "atlas-build-"+r.ID)
        if err != nil {
            return "", err
        }
        defer os.RemoveAll(buildDir)
        
        if err := copyDir(layout.ChartDir, buildDir); err != nil {
            return "", fmt.Errorf("failed to copy chart to build dir: %w", err)
        }
        
        if _, err := os.Stat(layout.OverlayDir); err == nil {
             if err := copyDir(layout.OverlayDir, buildDir); err != nil {
                return "", fmt.Errorf("failed to apply overlay: %w", err)
            }
        }

        vals, err := values.Load(layout.ValuesDir, nil)
        if err != nil {
            return "", fmt.Errorf("failed to load values: %w", err)
        }

        docs, err := renderer.Render(r, buildDir, vals)
        if err != nil {
            return "", fmt.Errorf("render failed: %w", err)
        }

        extraDocs, err := resources.Load(layout.ResourcesDir, r.Namespace)
        if err != nil {
            return "", fmt.Errorf("failed to load resources: %w", err)
        }
        docs = append(docs, extraDocs...)

        kustPath := filepath.Join(layout.PatchesDir, "kustomization.yaml")
        if _, err := os.Stat(kustPath); err == nil {
            patched, err := patcher.Patch(docs, layout.PatchesDir)
            if err != nil {
                return "", fmt.Errorf("patch failed: %w", err)
            }
            docs = splitYaml(patched)
        }

        out, err := normalize.Normalize(docs, []string{r.Namespace}, r.Namespace)
        if err != nil {
             return "", fmt.Errorf("normalization failed: %w", err)
        }
        
        if err := validate.Validate(splitYaml(out)); err != nil {
            return "", fmt.Errorf("validation failed: %w", err)
        }
        
        return out, nil
}


// Helpers

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        destPath := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(destPath, info.Mode())
        }
        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        return os.WriteFile(destPath, data, info.Mode())
    })
}



func splitYaml(in string) []string {
    if in == "" {
        return nil
    }
    // Naive split, same as resources loader
    // TODO: move to util
    parts := strings.Split(in, "\n---")
    var out []string
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            out = append(out, p)
        }
    }
    return out
}
