package normalize

import (
    "strings"
    "testing"
)

func TestNormalize(t *testing.T) {
    docs := []string{
        `apiVersion: v1
kind: Service
metadata:
  name: svc-b
  namespace: default
`,
        `apiVersion: v1
kind: Service
metadata:
  name: svc-a
  namespace: default
`,
        `apiVersion: v1
kind: Service
metadata:
  name: svc-no-ns
`,
    }
    
    // Inject default
    out, err := Normalize(docs, []string{"default"}, "default")
    if err != nil {
        t.Fatalf("Normalize failed: %v", err)
    }
    
    // Check order: svc-a before svc-b. svc-no-ns should be there (sorted by name? check kindScore equal)
    // svc-a, svc-b, svc-no-ns -> a < b < n
    idxA := strings.Index(out, "name: svc-a")
    idxB := strings.Index(out, "name: svc-b")
    idxN := strings.Index(out, "name: svc-no-ns")
    
    if idxA == -1 || idxB == -1 || idxN == -1 {
         t.Fatal("Missing services in output")
    }
    if idxA > idxB {
        t.Error("Expected svc-a before svc-b")
    }
}

func TestNormalize_RejectCluster(t *testing.T) {
    docs := []string{
        `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cr
`,
    }
    
    _, err := Normalize(docs, []string{"default"}, "default")
    if err == nil {
        t.Error("Expected error for ClusterRole, got nil")
    }
}
