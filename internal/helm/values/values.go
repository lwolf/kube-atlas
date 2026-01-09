package values

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"

    "sigs.k8s.io/yaml"
)

// Loader loads and merges values from various sources.
type Loader struct {
    // We can use Helm's values merger
}

// Load loads all values for a release.
// valuesDir: path to directory containing user values (lexical load)
// specificFiles: explicit list of files (overrides valuesDir if non-empty, or adds to it? Plan said "OR")
// For v1 we implemented: files in values/ (lexical) OR explicit list
func Load(valuesDir string, specificFiles []string) (map[string]interface{}, error) {
    base := make(map[string]interface{})

    // Gather file list
    var files []string
    if len(specificFiles) > 0 {
        // If specific files provided, resolve them relative to valuesDir or repo root
        // For simplicity in this function, we assume they are absolute or resolvable.
        files = specificFiles
    } else {
        // Walk valuesDir
        entries, err := os.ReadDir(valuesDir)
        if err != nil {
            if os.IsNotExist(err) {
                 return base, nil
            }
            return nil, fmt.Errorf("failed to read values dir: %w", err)
        }
        
        for _, e := range entries {
            if e.IsDir() {
                continue
            }
            ext := filepath.Ext(e.Name())
            if ext == ".yaml" || ext == ".yml" {
                files = append(files, filepath.Join(valuesDir, e.Name()))
            }
        }
        sort.Strings(files)
    }

    // Merge in order
    // We use Helm's value merging logic via a simple util function or standard generic merge
    // Helm's 'values.Options' uses getter providers.
    
    // We can use a simplified approach since we have local files
    for _, f := range files {
        data, err := os.ReadFile(f)
        if err != nil {
            return nil, fmt.Errorf("failed to read values file %q: %w", f, err)
        }
        
        var current map[string]interface{}
        if err := yaml.Unmarshal(data, &current); err != nil {
            return nil, fmt.Errorf("failed to parse values file %q: %w", f, err)
        }
        
        base = MergeMaps(base, current)
    }

    return base, nil
}

// MergeMaps merges two maps. override takes precedence.
// Arrays are replaced.
func MergeMaps(base, override map[string]interface{}) map[string]interface{} {
    out := make(map[string]interface{}, len(base))
    for k, v := range base {
        out[k] = v
    }
    
    for k, v := range override {
        // If both keys exist and are maps, recurse
        if v, ok := v.(map[string]interface{}); ok {
            if bv, ok := base[k]; ok {
                if bv, ok := bv.(map[string]interface{}); ok {
                    out[k] = MergeMaps(bv, v)
                    continue
                }
            }
        }
        // Else, overwrite
        out[k] = v
    }
    
    return out
}
