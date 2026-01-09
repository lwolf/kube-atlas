package render

import (
    "fmt"
    "sort"
    "strings"


    "helm.sh/helm/v3/pkg/chart/loader"
    "helm.sh/helm/v3/pkg/chartutil"
    "helm.sh/helm/v3/pkg/engine"
    "sigs.k8s.io/yaml"
    
    "github.com/lwolf/kube-atlas/internal/config"
)

type Engine struct {
    // We could wrap helm engine here
}

func New() *Engine {
    return &Engine{}
}

// Render renders the chart with the given values.
// chartDir is the path to the chart root (resolved effective FS root if possible, but Helm SDK works on disk mostly).
// For v1 we assume chartDir is on disk.
func (e *Engine) Render(release config.Release, chartDir string, values map[string]interface{}) ([]string, error) {
    // 1. Load Chart
    c, err := loader.Load(chartDir)
    if err != nil {
        return nil, fmt.Errorf("failed to load chart from %q: %w", chartDir, err)
    }

    // 2. Prepare Values
    // Helm engine requires values to be processed
    valOpts := chartutil.ReleaseOptions{
        Name:      release.ID,
        Namespace: release.Namespace,
        Revision:  1,
        IsInstall: true,
    }
    
    vals, err := chartutil.ToRenderValues(c, values, valOpts, nil)
    if err != nil {
         return nil, fmt.Errorf("failed to process values: %w", err)
    }

    // 3. Render
    m, err := engine.Render(c, vals)
    if err != nil {
        return nil, fmt.Errorf("failed to render chart: %w", err)
    }

    // 4. Output processing
    // Helm returns map[filename]content. We want a list of docs.
    // Also filter hooks.
    
    var docs []string
    
    // Sort keys for deterministic order (lexical by filename)
    filenames := make([]string, 0, len(m))
    for k := range m {
        filenames = append(filenames, k)
    }
    sort.Strings(filenames)

    for _, filename := range filenames {
        content := m[filename]
        
        // Split multi-doc output if any (templates shouldn't usually, but good practice)
        // Helm engine returns string per template file.
        // Ignore NOTES.txt and other non-yaml if they leak (usually excluded by Engine, but let's check extension)
        if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
            continue
        }
        
        // Check for hook annotation
        if isHook(content) {
            continue
        }

        docs = append(docs, content)
    }

    return docs, nil
}

// isHook checks if the manifest contains helm.sh/hook annotation.
// This is best-effort string check or naive parse.
func isHook(content string) bool {
    // Fast check
    if !strings.Contains(content, "helm.sh/hook") {
        return false
    }
    
    // Parse to be sure
    // We only need metadata
    type Meta struct {
        Metadata struct {
            Annotations map[string]string `yaml:"annotations"`
        } `yaml:"metadata"`
    }
    
    var m Meta
    if err := yaml.Unmarshal([]byte(content), &m); err != nil {
        return false // If we can't parse, assume not a hook or invalid yaml
    }
    
    _, ok := m.Metadata.Annotations["helm.sh/hook"]
    return ok
}
