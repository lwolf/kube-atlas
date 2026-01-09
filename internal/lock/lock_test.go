package lock

import (
    "os"
    "path/filepath"
    "testing"
)

func TestHashTree(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "atlas-lock-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Case 1: Empty dir
    hash1, err := HashTree(tmpDir)
    if err != nil {
        t.Fatalf("HashTree failed: %v", err)
    }

    // Case 2: One file
    fileA := filepath.Join(tmpDir, "a.txt")
    if err := os.WriteFile(fileA, []byte("foo"), 0644); err != nil {
        t.Fatal(err)
    }
    hash2, err := HashTree(tmpDir)
    if err != nil {
        t.Fatal(err)
    }
    if hash1 == hash2 {
        t.Error("Hash should change after adding file")
    }

    // Case 3: Modify file
    if err := os.WriteFile(fileA, []byte("bar"), 0644); err != nil {
        t.Fatal(err)
    }
    hash3, err := HashTree(tmpDir)
    if err != nil {
        t.Fatal(err)
    }
    if hash2 == hash3 {
        t.Error("Hash should change after modifying file")
    }

    // Case 4: Stability (revert)
    if err := os.WriteFile(fileA, []byte("foo"), 0644); err != nil {
        t.Fatal(err)
    }
    hash4, err := HashTree(tmpDir)
    if err != nil {
        t.Fatal(err)
    }
    if hash2 != hash4 {
        t.Error("Hash should be deterministic (same content = same hash)")
    }
}
