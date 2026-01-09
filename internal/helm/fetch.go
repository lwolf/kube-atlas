package helm

import (
    "fmt"
    "io"
    "io/fs"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "archive/tar"
    "compress/gzip"

    "github.com/lwolf/kube-atlas/internal/config"
    "github.com/lwolf/kube-atlas/internal/helm/oci"
    "github.com/lwolf/kube-atlas/internal/helm/repo"
    "github.com/lwolf/kube-atlas/internal/lock"
)

// Fetcher handles downloading and extracting charts.
type Fetcher struct {
    repoManager *repo.Manager
    ociClient   *oci.Client
}

func NewFetcher(cfg *config.Config) *Fetcher {
    return &Fetcher{
        repoManager: repo.NewManager(cfg.Repositories),
        ociClient:   oci.NewClient(),
    }
}

// Fetch downloads the chart for the given release and extracts it to destDir.
// It returns the resolved lock info.
func (f *Fetcher) Fetch(release config.Release, destDir string) (*lock.ResolvedLock, error) {
    // 1. Resolve
    resolved, err := f.repoManager.ResolveChart(release.Chart, release.Version)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve chart: %w", err)
    }
    
    // 2. Download and Verify
    // Create temp dir for download
    tmpDir, err := os.MkdirTemp("", "atlas-fetch")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)

    if resolved.Ref != "" {
        // OCI Pull
        if err := f.ociClient.Pull(resolved.Ref, resolved.Version, tmpDir); err != nil {
            return nil, err
        }
    } else {
        downloadURL := resolved.URL
        if downloadURL == "" {
             return nil, fmt.Errorf("chart %s has no download URL", release.Chart)
        }

        resp, err := http.Get(downloadURL)
        if err != nil {
             return nil, fmt.Errorf("failed to download chart: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            return nil, fmt.Errorf("failed to download chart from %s: %s", downloadURL, resp.Status)
        }
        
        if err := extractTarGz(resp.Body, tmpDir); err != nil {
            return nil, fmt.Errorf("failed to extract chart: %w", err)
        }
    }
    
    // 3. Move/Swap to destDir
    // Find chart root
    var chartRoot string
    // Walk to find Chart.yaml
    err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if !d.IsDir() && d.Name() == "Chart.yaml" {
            chartRoot = filepath.Dir(path)
            // Found it
            // We can return special error to stop or just check if chartRoot set
            // Stopping walk via SkipAll (Go 1.16+)
             return fs.SkipAll
        }
        return nil
    })
    
    if chartRoot == "" {
        return nil, fmt.Errorf("Chart.yaml not found in downloaded package")
    }

    // Prepare destDir
    // Remove existing content to ensure clean state
    if err := os.RemoveAll(destDir); err != nil {
        return nil, fmt.Errorf("failed to clear dest dir: %w", err)
    }
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create dest dir: %w", err)
    }

    // Copy from chartRoot to destDir
    if err := copyDir(chartRoot, destDir); err != nil {
         return nil, fmt.Errorf("failed to copy chart files: %w", err)
    }
    
    // 4. Return Lock
    return &lock.ResolvedLock{
        Chart:   release.Chart,
        Version: resolved.Version,
        Digest:  resolved.Digest,
    }, nil
}


func extractTarGz(r io.Reader, dest string) error {
    gzr, err := gzip.NewReader(r)
    if err != nil {
        return err
    }
    defer gzr.Close()

    tr := tar.NewReader(gzr)

    for {
        header, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        path := filepath.Join(dest, header.Name)
        
        // ZipSlip protection
        if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
             return fmt.Errorf("illegal file path: %s", path)
        }

        switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(path, 0755); err != nil {
                return err
            }
        case tar.TypeReg:
             if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
                return err
            }
            f, err := os.Create(path)
            if err != nil {
                return err
            }
            if _, err := io.Copy(f, tr); err != nil {
                f.Close()
                return err
            }
            f.Close()
        }
    }
    return nil
}

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
