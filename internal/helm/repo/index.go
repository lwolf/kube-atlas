package repo

import (
    "fmt"
    "io"
    "net/http"
    "sort"
    "time"

    "github.com/Masterminds/semver/v3"
    "gopkg.in/yaml.v3"
)

// Index represents the Helm repo index.yaml structure.
type Index struct {
    APIVersion string                     `yaml:"apiVersion"`
    Entries    map[string][]ChartVersion  `yaml:"entries"`
    Generated  time.Time                  `yaml:"generated"`
}

type ChartVersion struct {
    Name        string   `yaml:"name"`
    Version     string   `yaml:"version"`
    Description string   `yaml:"description,omitempty"`
    URLs        []string `yaml:"urls"`
    Digest      string   `yaml:"digest,omitempty"`
    Created     time.Time `yaml:"created,omitempty"`
}

// LoadIndex fetches and parses the index.yaml from a given URL.
// In a real app we would cache this file.
func LoadIndex(url string) (*Index, error) {
    // Basic HTTP get
    // We expect the URL to point to the repo root, so we append index.yaml
    // unless the URL already ends with it (unlikely for proper repo URL).
    // Helm convention: repo URL + "/index.yaml"
    
    // Normalize URL
    if url[len(url)-1] != '/' {
        url += "/"
    }
    indexURL := url + "index.yaml"
    
    resp, err := http.Get(indexURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch index from %s: %w", indexURL, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch index: status %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    var idx Index
    if err := yaml.Unmarshal(body, &idx); err != nil {
        return nil, fmt.Errorf("failed to parse index yaml: %w", err)
    }

    return &idx, nil
}

// GetVersion returns the specific ChartVersion for a chart and version constraint.
// If constraint is specific version (e.g. "1.2.3"), returns that.
// If constraint is semver range (e.g. "^1.0.0"), returns highest compatible.
func (i *Index) GetVersion(chartName, constraint string) (*ChartVersion, error) {
    versions, ok := i.Entries[chartName]
    if !ok {
        return nil, fmt.Errorf("chart %q not found in index", chartName)
    }

    // Sort versions desceding (newest first)
    // We parse them as semver to sort correctly
    // However, some versions might be invalid semver, we skip them or handle gracefully?
    // Helm SDK helps here, but we try to keep deps minimal. 
    // We imported Masterminds/semver/v3 which is standard.

    c, err := semver.NewConstraint(constraint)
    if err != nil {
        return nil, fmt.Errorf("invalid version constraint %q: %w", constraint, err)
    }

    // Filter compatible versions
    var compatible []*semver.Version
    versionMap := make(map[string]*ChartVersion)

    for idx, v := range versions {
        sv, err := semver.NewVersion(v.Version)
        if err != nil {
            continue // Skip invalid versions
        }
        if c.Check(sv) {
            compatible = append(compatible, sv)
            versionMap[sv.Original()] = &versions[idx]
        }
    }

    if len(compatible) == 0 {
        return nil, fmt.Errorf("no version found for chart %q matching %q", chartName, constraint)
    }

    // Sort compatible versions, generic sort is ascending, so we reverse
    sort.Sort(sort.Reverse(semver.Collection(compatible)))

    // Pick highest
    best := compatible[0]
    return versionMap[best.Original()], nil
}
