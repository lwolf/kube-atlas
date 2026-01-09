package fs

import (
    "os"
    "path/filepath"
    "testing"
)

func TestAtomicWrite_RealFS(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "atlas-test-fs")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    fsys := RealFS{}
    target := filepath.Join(tmpDir, "file.txt")
    content := []byte("hello world")

    if err := AtomicWrite(fsys, target, content, 0644); err != nil {
        t.Fatalf("AtomicWrite failed: %v", err)
    }

    // Verify content
    read, err := os.ReadFile(target)
    if err != nil {
        t.Fatalf("Failed to read back file: %v", err)
    }
    if string(read) != string(content) {
        t.Errorf("Content mismatch: got %q, want %q", string(read), string(content))
    }

    // Verify permissions (best effort check, umask might affect it)
    info, err := os.Stat(target)
    if err != nil {
        t.Fatal(err)
    }
    // We expect at least the read bits we asked for, but mostly we check existence here.
    if info.Mode().Perm()&0400 == 0 {
        t.Errorf("Expected read permission, got %v", info.Mode())
    }
}
