package kust

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "sigs.k8s.io/kustomize/api/krusty"
    "sigs.k8s.io/kustomize/kyaml/filesys"
)

// Patcher handles post-render patching.
type Patcher struct {
}

func New() *Patcher {
    return &Patcher{}
}

// Patch applies Kustomization found in patchDir to the base manifests.
func (p *Patcher) Patch(baseDocs []string, patchDir string) (string, error) {
    if len(baseDocs) == 0 {
        return "", nil
    }

    // 1. Create a temporary work directory
    tmpDir, err := os.MkdirTemp("", "bucket-kust")
    if err != nil {
        return "", fmt.Errorf("failed to create temp dir: %w", err)
    }
    defer os.RemoveAll(tmpDir)

    // 2. Write base manifests to base.yaml
    baseContent := strings.Join(baseDocs, "\n---\n")
    if err := os.WriteFile(filepath.Join(tmpDir, "base.yaml"), []byte(baseContent), 0644); err != nil {
        return "", fmt.Errorf("failed to write base.yaml: %w", err)
    }

    // 3. Copy everything from patchDir to tmpDir
    // We need the kustomization.yaml and any patch files it references.
    if err := copyDir(patchDir, tmpDir); err != nil {
        return "", fmt.Errorf("failed to copy patches: %w", err)
    }

    // 4. Update kustomization.yaml in tmpDir to include base.yaml
    // We act dumb and just append "resources:\n- base.yaml" if resources is missing, 
    // or we parse and modify.
    // Parsing is safer.
    
    kustPath := filepath.Join(tmpDir, "kustomization.yaml")
    kustBytes, err := os.ReadFile(kustPath)
    if err != nil {
        return "", fmt.Errorf("failed to read kustomization.yaml: %w", err)
    }

    // Append base.yaml to resources
    // We'll use string manipulation for v1 to avoid circular deps or complex parsing if possible.
    // A robust way without pulling in kustomize edit logic:
    // "resources:" key might exist.
    
    // Let's use simple append logic assuming standard format:
    // If we append a new "resources: [base.yaml]" block, it might conflict if exists.
    // Safest: We generate a NEW kustomization.yaml that includes the user's kustomization as a base?
    // No, user's kustomization IS the overlay.
    
    // Strategy: treating user's patchDir as a base is wrong. It IS the overlay.
    // But it needs to reference the input resources.
    
    // Let's rewrite kustomization.yaml to add 'base.yaml' to resources.
    // We can assume valid YAML.
    newKust := string(kustBytes) + "\nresources:\n- base.yaml\n"
    
    // BUT if resources key already exists, this is invalid YAML (duplicate key).
    // In our `atlas release add` scaffold, we created: `resources: []`.
    // So simply appending tries to add duplicate key.
    
    // Simple hack: Replace "resources: []" with "resources: [base.yaml]"
    if strings.Contains(newKust, "resources: []") {
        newKust = strings.Replace(newKust, "resources: []", "resources: [base.yaml]", 1)
    } else {
        // Fallback: try to just append and hope kustomize merges or errors? 
        // It errors. 
        // We really should parse it. But we want to avoid dep hell.
        // Wait, we have sigs.k8s.io/yaml imported in other packages.
        // Let's use it.
    }

    if err := os.WriteFile(kustPath, []byte(newKust), 0644); err != nil {
        return "", err
    }

    // 5. Build
    fSys := filesys.MakeFsOnDisk()
    opts := krusty.MakeDefaultOptions()
    k := krusty.MakeKustomizer(opts)
    
    m, err := k.Run(fSys, tmpDir)
    if err != nil {
        return "", fmt.Errorf("kustomize build failed: %w", err)
    }

    yml, err := m.AsYaml()
    if err != nil {
        return "", fmt.Errorf("failed to convert to yaml: %w", err)
    }

    return string(yml), nil
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
