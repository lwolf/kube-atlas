package workspace

import (
    "path/filepath"
    "testing"

    "github.com/lwolf/kube-atlas/internal/config"
)

func TestResolve(t *testing.T) {
    repoRoot := "/tmp/repo"
    
    tests := []struct {
        name    string
        release config.Release
        check   func(t *testing.T, l Layout)
    }{
        {
            name: "defaults",
            release: config.Release{
                ID:        "nginx",
                Namespace: "web",
            },
            check: func(t *testing.T, l Layout) {
                expectedBase := "/tmp/repo/atlas/releases/namespaces/web/nginx"
                if l.ReleaseRoot != expectedBase {
                    t.Errorf("got ReleaseRoot %q, want %q", l.ReleaseRoot, expectedBase)
                }
                if l.ChartDir != filepath.Join(expectedBase, "chart") {
                    t.Errorf("got ChartDir %q", l.ChartDir)
                }
                if l.OverlayDir != filepath.Join(expectedBase, "overlay") {
                   t.Errorf("got OverlayDir %q", l.OverlayDir)
                }
                if l.RenderedFile != "/tmp/repo/atlas/rendered/namespaces/web/nginx.yaml" {
                    t.Errorf("got RenderedFile %q", l.RenderedFile)
                }
            },
        },
        {
            name: "overrides",
            release: config.Release{
                ID:           "app",
                Namespace:    "default",
                OverlayDir:   "custom/overlay",
                PatchesDir:   "custom/patches",
                ResourcesDir: "custom/resources",
            },
            check: func(t *testing.T, l Layout) {
                if l.OverlayDir != "/tmp/repo/custom/overlay" {
                    t.Errorf("got OverlayDir %q", l.OverlayDir)
                }
                if l.PatchesDir != "/tmp/repo/custom/patches" {
                    t.Errorf("got PatchesDir %q", l.PatchesDir)
                }
                 if l.ResourcesDir != "/tmp/repo/custom/resources" {
                    t.Errorf("got ResourcesDir %q", l.ResourcesDir)
                }
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            l := Resolve(repoRoot, tt.release)
            tt.check(t, l)
        })
    }
}
