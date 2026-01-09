package render

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/lwolf/kube-atlas/internal/config"
)

func TestRender(t *testing.T) {
    // Create a dummy chart
    tmpDir, err := os.MkdirTemp("", "render-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    chartContent := `name: test-chart
apiVersion: v2
version: 0.1.0
`
    os.WriteFile(filepath.Join(tmpDir, "Chart.yaml"), []byte(chartContent), 0644)
    
    os.Mkdir(filepath.Join(tmpDir, "templates"), 0755)
    
    // Deployment
    tpl := `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-cm
  namespace: {{ .Release.Namespace }}
data:
  foo: {{ .Values.foo }}
`
    os.WriteFile(filepath.Join(tmpDir, "templates", "cm.yaml"), []byte(tpl), 0644)

    // Hook
    hook := `apiVersion: v1
kind: Job
metadata:
  name: test-hook
  annotations:
    "helm.sh/hook": pre-install
`
    os.WriteFile(filepath.Join(tmpDir, "templates", "hook.yaml"), []byte(hook), 0644)

    rel := config.Release{
        ID:        "my-release",
        Namespace: "my-ns",
    }
    vals := map[string]interface{}{
        "foo": "bar",
    }
    
    e := New()
    docs, err := e.Render(rel, tmpDir, vals)
    if err != nil {
        t.Fatalf("Render failed: %v", err)
    }
    
    // Should contain 1 doc (CM), hook excluded
    if len(docs) != 1 {
        t.Errorf("Expected 1 doc, got %d", len(docs))
    }
    
    // Check content
    // Expect: name: my-release-cm
    // Expect: foo: bar
}
