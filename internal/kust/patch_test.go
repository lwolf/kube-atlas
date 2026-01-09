package kust

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestPatch(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "kust-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Setup patch dir
    kust := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources: []
patches:
- patch: |-
    - op: add
      path: /metadata/labels/test
      value: "true"
  target:
    kind: Service
    name: mysvc
`
    os.WriteFile(filepath.Join(tmpDir, "kustomization.yaml"), []byte(kust), 0644)
    
    baseDocs := []string{
        `apiVersion: v1
kind: Service
metadata:
  name: mysvc
  labels:
    app: foo
spec:
  ports:
  - port: 80
`,
    }

    p := New()
    out, err := p.Patch(baseDocs, tmpDir)
    if err != nil {
        t.Fatalf("Patch failed: %v", err)
    }

    // Check if label added
    // Naive check
    if !strings.Contains(out, "test: \"true\"") {
		t.Errorf("Label not found in output:\n%q", out)
	}
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && len(substr) > 0 && 
           (s == substr || 
            (len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
            // Actually just use strings.Contains
}
