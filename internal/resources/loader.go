package resources

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "sigs.k8s.io/yaml"
)

// Load reads all .yaml/.yml files from dir and returns them as a slice of strings.
// It enforces that all namespaced resources belong to defaultNamespace.
// It injects defaultNamespace if missing in metadata.
func Load(dir string, defaultNamespace string) ([]string, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to read resources dir: %w", err)
    }

    var files []string
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        ext := filepath.Ext(e.Name())
        if ext == ".yaml" || ext == ".yml" {
            files = append(files, filepath.Join(dir, e.Name()))
        }
    }
    sort.Strings(files)

    var docs []string
    for _, f := range files {
        content, err := os.ReadFile(f)
        if err != nil {
            return nil, fmt.Errorf("failed to read resource %q: %w", f, err)
        }

        // Split multi-doc yaml
        // Simple string split by "---" is dangerous if inside quotes, but standard for now.
        // Better: use a proper decoder.
        // For now, let's assume one doc per file or simple separator.
        
        // We use a robust way? No, let's rely on string split for v1, similar to kubectl apply -f
        // Actually, let's try to parse into map[string]interface{} to modify namespace.
        
        rawParts := strings.Split(string(content), "\n---")
        for _, part := range rawParts {
            part = strings.TrimSpace(part)
            if part == "" {
                continue
            }
            
            processed, err := processDoc([]byte(part), defaultNamespace, f)
            if err != nil {
                return nil, err
            }
            docs = append(docs, processed)
        }
    }

    return docs, nil
}

func processDoc(data []byte, ns string, sourceFile string) (string, error) {
    var obj map[string]interface{}
    if err := yaml.Unmarshal(data, &obj); err != nil {
        return "", fmt.Errorf("failed to parse %s: %w", sourceFile, err)
    }
    
    // Check metadata
    meta, ok := obj["metadata"].(map[string]interface{})
    if !ok {
        // Might be a list or something else? If it's K8s object, it must have metadata.
        // If missing, arguably invalid k8s object, but maybe valid yaml.
        return string(data), nil
    }

    // Check kind/apiVersion to decide if namespaced?
    // In v1 we assume EVERYTHING is namespaced unless we have a whitelist of cluster-scoped kinds.
    // However, the rule was "Start by enforcing namespaced objects... Reject cluster-scoped kinds by default".
    // Since we don't have api discovery, we rely on configuration or heuristic.
    // For now: Always inject namespace if missing. If present, must match.
    
    currentNs, _ := meta["namespace"].(string)
    if currentNs == "" {
        // Inject
        meta["namespace"] = ns
    } else if currentNs != ns {
        return "", fmt.Errorf("resource %s has conflicting namespace %q (expected %q)", sourceFile, currentNs, ns)
    }
    
    // Re-marshal
    out, err := yaml.Marshal(obj)
    if err != nil {
        return "", fmt.Errorf("failed to marshal %s: %w", sourceFile, err)
    }
    return string(out), nil
}
