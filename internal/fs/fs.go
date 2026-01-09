package fs

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
)

// FS is an abstraction for filesystem operations to allow testing.
// It embeds fs.StatFS for read operations.
type FS interface {
    fs.StatFS
    MkdirAll(path string, perm os.FileMode) error
    WriteFile(name string, data []byte, perm os.FileMode) error
    Rename(oldpath, newpath string) error
    Remove(name string) error
}

// RealFS implements FS using the os package.
type RealFS struct{}

func (RealFS) Open(name string) (fs.File, error) { return os.Open(name) }
func (RealFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (RealFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (RealFS) WriteFile(name string, data []byte, perm os.FileMode) error {
    return os.WriteFile(name, data, perm)
}
func (RealFS) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }
func (RealFS) Remove(name string) error             { return os.Remove(name) }

// AtomicWrite writes data to a temporary file in the same directory as the target,
// then renames it to the target file.
func AtomicWrite(fsys FS, path string, data []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    if err := fsys.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory %q: %w", dir, err)
    }

    // Create temp file
    // We use a pattern that is likely unique and ignored
    tmpPath := path + ".tmp"
    
    // Write data to temp file
    if err := fsys.WriteFile(tmpPath, data, perm); err != nil {
        return fmt.Errorf("failed to write temp file %q: %w", tmpPath, err)
    }

    // Rename temp to target
    if err := fsys.Rename(tmpPath, path); err != nil {
        // Try to clean up temp file
        _ = fsys.Remove(tmpPath)
        return fmt.Errorf("failed to rename %q to %q: %w", tmpPath, path, err)
    }

    return nil
}
