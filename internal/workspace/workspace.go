package workspace

import (
    "path/filepath"

    "github.com/lwolf/kube-atlas/internal/config"
)

// Layout defines the absolute paths for a specific release.
type Layout struct {
    ReleaseRoot string

    // Source directories
    ChartDir     string // <root>/chart
    OverlayDir   string // <root>/overlay
    ValuesDir    string // <root>/values
    PatchesDir   string // <root>/patches
    ResourcesDir string // <root>/resources

    // Output file
    OutputDir    string // <root>/atlas/rendered/namespaces/<ns>
    RenderedFile string
}

// Resolve returns the filesystem layout for a given release.
// repoRoot should be the absolute path to the repository root (containing atlas.yaml).
func Resolve(repoRoot string, release config.Release) Layout {
    // Default base path: atlas/releases/namespaces/<ns>/<id>
    basePath := filepath.Join(repoRoot, "atlas", "releases", "namespaces", release.Namespace, release.ID)

    l := Layout{
        ReleaseRoot: basePath,
    }

    // Chart dir is always fixed convention: <base>/chart
    l.ChartDir = filepath.Join(basePath, "chart")

    // Overlay dir: convention or override
    if release.OverlayDir != "" {
        if filepath.IsAbs(release.OverlayDir) {
            l.OverlayDir = release.OverlayDir
        } else {
            l.OverlayDir = filepath.Join(repoRoot, release.OverlayDir)
        }
    } else {
        l.OverlayDir = filepath.Join(basePath, "overlay")
    }

    // Values dir
    l.ValuesDir = filepath.Join(basePath, "values")

    // Patches dir: convention or override
    if release.PatchesDir != "" {
         if filepath.IsAbs(release.PatchesDir) {
            l.PatchesDir = release.PatchesDir
        } else {
            l.PatchesDir = filepath.Join(repoRoot, release.PatchesDir)
        }
    } else {
        l.PatchesDir = filepath.Join(basePath, "patches")
    }
    
    // Resources dir: convention or override
    if release.ResourcesDir != "" {
         if filepath.IsAbs(release.ResourcesDir) {
            l.ResourcesDir = release.ResourcesDir
        } else {
            l.ResourcesDir = filepath.Join(repoRoot, release.ResourcesDir)
        }
    } else {
        l.ResourcesDir = filepath.Join(basePath, "resources")
    }

    // Rendered output: atlas/rendered/namespaces/<ns>/<id>.yaml
    l.OutputDir = filepath.Join(repoRoot, "atlas", "rendered", "namespaces", release.Namespace)
    l.RenderedFile = filepath.Join(l.OutputDir, release.ID+".yaml")

    return l
}
