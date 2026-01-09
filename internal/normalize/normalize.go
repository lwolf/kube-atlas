package normalize

import (
    "fmt"
    "sort"
    
    "sigs.k8s.io/yaml"
)

// Normalize parses, sorts, and checks namespace of the given documents.
// It returns a single multi-doc YAML string.
// If defaultNamespace is not empty, it is injected into namespaced resources that lack a namespace.
func Normalize(docs []string, allowedNamespaces []string, defaultNamespace string) (string, error) {
    if len(docs) == 0 {
        return "", nil
    }

    // 1. Parse all to objects
    var objects []K8sObject
    for i, doc := range docs {
        if doc == "" {
            continue
        }
        var obj map[string]interface{}
        if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
            return "", fmt.Errorf("failed to parse doc %d: %w", i, err)
        }
        
        meta, _ := obj["metadata"].(map[string]interface{})
        gvk, _ := obj["apiVersion"].(string) // simplified kind+version
        kind, _ := obj["kind"].(string)
        name, _ := meta["name"].(string)
        ns, _ := meta["namespace"].(string)
        
        objects = append(objects, K8sObject{
            Original: doc,
            Data:     obj,
            Group:    gvk,
            Kind:     kind,
            Name:     name,
            Namespace: ns,
        })
    }

    // 2. Validate Scopes (Namespaces)
    allowed := make(map[string]bool)
    for _, n := range allowedNamespaces {
        allowed[n] = true
    }

    for i := range objects {
        o := &objects[i]
        
        if isClusterScoped(o.Kind) {
            // Cluster scoped objects strictly NOT allowed in v1 namespaced release output?
            // "Enforce: cluster-scoped objects are rejected (v1)"
            return "", fmt.Errorf("cluster-scoped object %s/%s not allowed in namespaced release output", o.Kind, o.Name)
        }

        if o.Namespace == "" {
            if defaultNamespace != "" {
                o.Namespace = defaultNamespace
                // Update Data map
                if o.Data["metadata"] == nil {
                    o.Data["metadata"] = make(map[string]interface{})
                }
                meta := o.Data["metadata"].(map[string]interface{})
                meta["namespace"] = defaultNamespace
            } else {
                 return "", fmt.Errorf("object %s/%s missing namespace (and strictly namespaced output required)", o.Kind, o.Name)
            }
        }
        
        // Validate against allowed
        if !allowed[o.Namespace] {
            return "", fmt.Errorf("object %s/%s has disallowed namespace %q", o.Kind, o.Name, o.Namespace)
        }
    }

    // 3. Sort
    // Sort Criteria:
    // 1. Namespace
    // 2. Kind (CRD first, Namespace second, then others)
    // 3. Name
    sort.Slice(objects, func(i, j int) bool {
        a, b := objects[i], objects[j]
        
        if a.Namespace != b.Namespace {
            return a.Namespace < b.Namespace
        }
        
        scoreA := kindScore(a.Kind)
        scoreB := kindScore(b.Kind)
        if scoreA != scoreB {
            return scoreA < scoreB
        }
        
        if a.Kind != b.Kind {
            return a.Kind < b.Kind
        }
        
        return a.Name < b.Name
    })

    // 4. Emit
    var out string
    for i, o := range objects {
        b, err := yaml.Marshal(o.Data)
        if err != nil {
            return "", err
        }
        
        if i > 0 {
            out += "---\n"
        }
        out += string(b)
    }

    return out, nil
}

type K8sObject struct {
    Original string
    Data     map[string]interface{}
    Group    string
    Kind     string
    Name     string
    Namespace string
}

func isClusterScoped(kind string) bool {
    // Whitelist known cluster scoped kinds to REJECT them consistently?
    // Or invert: assume namespaced unless...
    // The requirement is REJECT cluster-scoped.
    switch kind {
    case "CustomResourceDefinition", "Namespace", "ClusterRole", "ClusterRoleBinding", "PersistentVolume", "StorageClass":
        return true
    }
    return false
}

func kindScore(kind string) int {
    switch kind {
    case "CustomResourceDefinition":
        return 0
    case "Namespace":
        return 1
    case "ServiceAccount":
        return 2
    default:
        return 10
    }
}
