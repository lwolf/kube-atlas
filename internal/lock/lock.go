package lock

import (
    "crypto/sha256"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type Lockfile struct {
    Releases map[string]ReleaseLock `yaml:"releases"`
}

type ReleaseLock struct {
    Resolved   ResolvedLock `yaml:"resolved"`
    VendorHash string       `yaml:"vendorHash"`
    // Optional hashes for other dirs can be added here
}

type ResolvedLock struct {
    Chart   string `yaml:"chart"`
    Version string `yaml:"version"`
    Digest  string `yaml:"digest,omitempty"`
}

func (l *Lockfile) Get(id string) (ReleaseLock, bool) {
    if l.Releases == nil {
        return ReleaseLock{}, false
    }
    r, ok := l.Releases[id]
    return r, ok
}

func (l *Lockfile) Update(id string, r ReleaseLock) {
    if l.Releases == nil {
        l.Releases = make(map[string]ReleaseLock)
    }
    l.Releases[id] = r
}

func (l *Lockfile) IsDirty(id, dir string) (bool, error) {
    r, ok := l.Get(id)
    if !ok {
        return true, nil // Unknown is effectively dirty or needs fetch
    }
    return r.IsDirty(dir)
}

// Load reads the lockfile from disk.
func Load(path string) (*Lockfile, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return &Lockfile{Releases: make(map[string]ReleaseLock)}, nil
        }
        return nil, fmt.Errorf("failed to read lockfile: %w", err)
    }

    var l Lockfile
    if err := yaml.Unmarshal(data, &l); err != nil {
        return nil, fmt.Errorf("failed to parse lockfile: %w", err)
    }
    
    if l.Releases == nil {
        l.Releases = make(map[string]ReleaseLock)
    }

    return &l, nil
}

// Save writes the lockfile to disk.
func (l *Lockfile) Save(path string) error {
    data, err := yaml.Marshal(l)
    if err != nil {
        return fmt.Errorf("failed to marshal lockfile: %w", err)
    }
    return os.WriteFile(path, data, 0644)
}

// HashTree computes a stable hash of a directory's contents.
// It walks the directory, sorts files by path, and hashes (path + content).
func HashTree(dir string) (string, error) {
    h := sha256.New()
    
    // We need to walk and collect all files first to sort them, 
    // although fs.WalkDir is usually lexical, we want to be explicit.
    // However, filepath.WalkDir is lexical.
    
    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(dir, path)
        if err != nil {
            return err
        }
        
        // Use forward slashes for stability across OS
        relPath = filepath.ToSlash(relPath)

        // Write path to hash
        if _, err := fmt.Fprintf(h, "FILE %s\n", relPath); err != nil {
            return err
        }

        // Write content
        f, err := os.Open(path)
        if err != nil {
            return err
        }
        defer f.Close()

        if _, err := io.Copy(h, f); err != nil {
            return err
        }
        
        // Separator
        if _, err := io.WriteString(h, "\n"); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return "", err
    }

    return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// IsDirty checks if the directory hash matches the recorded hash.
func (r ReleaseLock) IsDirty(dir string) (bool, error) {
    if r.VendorHash == "" {
        return true, nil
    }
    
    currentHash, err := HashTree(dir)
    if err != nil {
        return false, err
    }
    
    return currentHash != r.VendorHash, nil
}
