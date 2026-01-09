package repo

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestLoadIndex(t *testing.T) {
    // Serve a mock index.yaml
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/index.yaml" {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        
        w.Write([]byte(`apiVersion: v1
entries:
  nginx:
    - name: nginx
      version: 1.2.3
      urls: [https://example.com/nginx-1.2.3.tgz]
    - name: nginx
      version: 1.2.4
      urls: [https://example.com/nginx-1.2.4.tgz]
`))
    }))
    defer ts.Close()

    idx, err := LoadIndex(ts.URL)
    if err != nil {
        t.Fatalf("LoadIndex failed: %v", err)
    }

    if len(idx.Entries) != 1 {
        t.Errorf("Expected 1 entry, got %d", len(idx.Entries))
    }
    
    // Test resolution
    v, err := idx.GetVersion("nginx", "^1.2.0")
    if err != nil {
        t.Fatalf("GetVersion failed: %v", err)
    }
    if v.Version != "1.2.4" {
        t.Errorf("Expected 1.2.4, got %s", v.Version)
    }

    v, err = idx.GetVersion("nginx", "1.2.3")
    if err != nil {
         t.Fatalf("GetVersion failed: %v", err)
    }
     if v.Version != "1.2.3" {
        t.Errorf("Expected 1.2.3, got %s", v.Version)
    }
}
