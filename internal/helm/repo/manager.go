package repo

import (
    "fmt"
    
    "github.com/lwolf/kube-atlas/internal/config"
)

// Repository is a generic interface for chart repositories.
type Repository interface {
    // Resolve returns the specific version and download URL/Ref for a chart.
    Resolve(chart string, version string) (*ResolvedChart, error)
}

type ResolvedChart struct {
    Version string
    URL     string // For HTTP
    Ref     string // For OCI
    Digest  string
}

// Manager handles multiple repositories.
type Manager struct {
    repos map[string]config.Repository
    // Cache for loaded indices
    indices map[string]*Index
}

func NewManager(repos []config.Repository) *Manager {
    rMap := make(map[string]config.Repository)
    for _, r := range repos {
        rMap[r.Name] = r
    }
    return &Manager{
        repos:   rMap,
        indices: make(map[string]*Index),
    }
}

// ResolveChart resolves a chart reference (repo/chart or oci://...).
// Returns the resolved details tailored for downloading.
func (m *Manager) ResolveChart(chartRef string, versionConstraint string) (*ResolvedChart, error) {
    // Check if OCI
    // TODO: stricter check, maybe use helm registry utils if available
    if len(chartRef) > 6 && chartRef[:6] == "oci://" {
         return &ResolvedChart{
            Version: versionConstraint, // OCI usually requires exact version or latest tag logic
            Ref:     chartRef,
        }, nil
    }

    // Split repo/chart
    // We assume format "repoName/chartName"
    var repoName, chartName string
    // Simple split
    for i := 0; i < len(chartRef); i++ {
        if chartRef[i] == '/' {
            repoName = chartRef[:i]
            chartName = chartRef[i+1:]
            break
        }
    }
    
    if repoName == "" || chartName == "" {
        return nil, fmt.Errorf("invalid chart reference %q (expected repo/chart)", chartRef)
    }

    repoConfig, ok := m.repos[repoName]
    if !ok {
        return nil, fmt.Errorf("repository %q not found in config", repoName)
    }

    // Load index (lazy)
    // TODO: Persistence invocation
    idx, ok := m.indices[repoName]
    if !ok {
        var err error
        idx, err = LoadIndex(repoConfig.URL)
        if err != nil {
            return nil, fmt.Errorf("failed to load index for repo %q: %w", repoName, err)
        }
        m.indices[repoName] = idx
    }

    v, err := idx.GetVersion(chartName, versionConstraint)
    if err != nil {
        return nil, err
    }

    // Return resolved info
    // We pick the first URL
    if len(v.URLs) == 0 {
        return nil, fmt.Errorf("no URLs found for chart %s-%s", chartName, v.Version)
    }
    
    rawURL := v.URLs[0]
    
    // Handle relative URLs
    var finalURL string
    if len(rawURL) > 4 && rawURL[:4] == "http" {
        finalURL = rawURL
    } else {
        // Simple join
        baseURL := repoConfig.URL
        if baseURL[len(baseURL)-1] != '/' {
            baseURL += "/"
        }
        finalURL = baseURL + rawURL
    }

    return &ResolvedChart{
        Version: v.Version,
        URL:     finalURL,
        Digest:  v.Digest,
    }, nil
}
