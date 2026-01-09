package chartfs

import (
    "os"
    "path/filepath"
    "testing"
)

func TestEffectiveFS(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "chartfs-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    base := filepath.Join(tmpDir, "base")
    overlay := filepath.Join(tmpDir, "overlay")
    os.Mkdir(base, 0755)
    os.Mkdir(overlay, 0755)

    // Setup:
    // base/a.txt
    // base/b.txt
    // overlay/b.txt (override)
    // overlay/c.txt (new)

    os.WriteFile(filepath.Join(base, "a.txt"), []byte("base-a"), 0644)
    os.WriteFile(filepath.Join(base, "b.txt"), []byte("base-b"), 0644)
    os.WriteFile(filepath.Join(overlay, "b.txt"), []byte("overlay-b"), 0644)
    os.WriteFile(filepath.Join(overlay, "c.txt"), []byte("overlay-c"), 0644)

    fsys := New(base, overlay)

    // Test ReadFile
    // a.txt -> base
    content, err := fsys.ReadFile("a.txt")
    if err != nil {
        t.Errorf("failed to read a.txt: %v", err)
    }
    if string(content) != "base-a" {
        t.Errorf("a.txt: got %q, want 'base-a'", content)
    }

    // b.txt -> overlay
    content, err = fsys.ReadFile("b.txt")
     if err != nil {
        t.Errorf("failed to read b.txt: %v", err)
    }
    if string(content) != "overlay-b" {
        t.Errorf("b.txt: got %q, want 'overlay-b'", content)
    }

    // c.txt -> overlay
    content, err = fsys.ReadFile("c.txt")
     if err != nil {
        t.Errorf("failed to read c.txt: %v", err)
    }
    if string(content) != "overlay-c" {
        t.Errorf("c.txt: got %q, want 'overlay-c'", content)
    }
}
