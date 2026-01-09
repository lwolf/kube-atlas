package chartfs

import (
    "io/fs"
    "os"
    "path/filepath"
)

// EffectiveFS implements a filesystem that merges an overlay directory on top of a base directory.
// It prioritizes files in the overlay directory.
type EffectiveFS struct {
    base    string
    overlay string
}

func New(base, overlay string) *EffectiveFS {
    return &EffectiveFS{
        base:    base,
        overlay: overlay,
    }
}

// Open opens the named file.
func (f *EffectiveFS) Open(name string) (fs.File, error) {
    // Check overlay first
    path := filepath.Join(f.overlay, name)
    if _, err := os.Stat(path); err == nil {
        return os.Open(path)
    }

    // Fallback to base
    return os.Open(filepath.Join(f.base, name))
}

// Stat returns the FileInfo for the named file.
func (f *EffectiveFS) Stat(name string) (os.FileInfo, error) {
    // Check overlay first
    path := filepath.Join(f.overlay, name)
    if info, err := os.Stat(path); err == nil {
        return info, nil
    }

    // Fallback to base
    return os.Stat(filepath.Join(f.base, name))
}

// ReadFile reads the named file and returns the contents.
func (f *EffectiveFS) ReadFile(name string) ([]byte, error) {
    // Check overlay first
    path := filepath.Join(f.overlay, name)
    if _, err := os.Stat(path); err == nil {
        return os.ReadFile(path)
    }

    // Fallback to base
    return os.ReadFile(filepath.Join(f.base, name))
}

// WalkDir walks the file tree rooted at root, calling fn for each file or directory in the tree, including root.
// It merges the views of base and overlay.
func (f *EffectiveFS) WalkDir(root string, fn fs.WalkDirFunc) error {
    // We need to union the file sets.
    // Simpler approach for v1: just walk base, and if overlay has it, use overlay. 
    // BUT we also need to walk files ONLY in overlay.
    
    seen := make(map[string]bool)
    
    // 1. Walk overlay
    overlayRoot := filepath.Join(f.overlay, root)
    err := filepath.WalkDir(overlayRoot, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            if os.IsNotExist(err) && path == overlayRoot {
                return nil // Overlay might not exist or be empty, ignore
            }
            return err
        }
        
        rel, err := filepath.Rel(f.overlay, path)
        if err != nil {
            return err
        }
        
        seen[rel] = true
        return fn(rel, d, nil)
    })
    if err != nil {
         return err
    }

    // 2. Walk base
    baseRoot := filepath.Join(f.base, root)
    err = filepath.WalkDir(baseRoot, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
             return err
        }
        
        rel, err := filepath.Rel(f.base, path)
        if err != nil {
            return err
        }
        
        if seen[rel] {
            return nil // Already visited via overlay
        }
        
        return fn(rel, d, nil)
    })
    
    return err
}
