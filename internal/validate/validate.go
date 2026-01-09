package validate

import (
    "fmt"

    "sigs.k8s.io/yaml"
)

// Validate checks for structural correctness and duplicate IDs.
func Validate(docs []string) error {
    seen := make(map[string]bool)

    for i, doc := range docs {
        if doc == "" {
            continue
        }

        var obj map[string]interface{}
        if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
            return fmt.Errorf("doc %d: failed to parse yaml: %w", i, err)
        }

        // Basic fields
        apiVersion, _ := obj["apiVersion"].(string)
        kind, _ := obj["kind"].(string)
        meta, ok := obj["metadata"].(map[string]interface{})
        if !ok {
            return fmt.Errorf("doc %d: missing metadata", i)
        }
        name, _ := meta["name"].(string)
        ns, _ := meta["namespace"].(string)

        if apiVersion == "" {
            return fmt.Errorf("doc %d: missing apiVersion", i)
        }
        if kind == "" {
            return fmt.Errorf("doc %d: missing kind", i)
        }
        if name == "" {
            return fmt.Errorf("doc %d: missing metadata.name", i)
        }

        // Check duplicate
        // ID = group/kind/namespace/name
        // Group is implicit in apiVersion usually, but for uniqueness apiVersion+kind is safer or close enough.
        id := fmt.Sprintf("%s/%s/%s/%s", apiVersion, kind, ns, name)
        if seen[id] {
            return fmt.Errorf("duplicate resource found: %s", id)
        }
        seen[id] = true
    }

    return nil
}
