package resources

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoad(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "resources-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Case 1: No namespace -> injected
    doc1 := `apiVersion: v1
kind: Service
metadata:
  name: svc1
`
    os.WriteFile(filepath.Join(tmpDir, "svc.yaml"), []byte(doc1), 0644)

    // Case 2: Matching namespace -> ok
    doc2 := `apiVersion: v1
kind: Service
metadata:
  name: svc2
  namespace: my-ns
`
    os.WriteFile(filepath.Join(tmpDir, "svc2.yaml"), []byte(doc2), 0644)

    // Case 3: Mismatch -> error
    doc3 := `apiVersion: v1
kind: Service
metadata:
  name: svc3
  namespace: other-ns
`
    os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte(doc3), 0644)

    // Test Valid load
    // We only load svc.yaml and svc2.yaml first
    os.Remove(filepath.Join(tmpDir, "bad.yaml"))
    
    docs, err := Load(tmpDir, "my-ns")
    if err != nil {
        t.Fatalf("Load valid failed: %v", err)
    }
    if len(docs) != 2 {
        t.Errorf("Expected 2 docs, got %d", len(docs))
    }
    
    // Test Mismatch
    os.WriteFile(filepath.Join(tmpDir, "bad.yaml"), []byte(doc3), 0644)
    _, err = Load(tmpDir, "my-ns")
    if err == nil {
        t.Fatal("Expected error for mismatch namespace, got nil")
    }
}
