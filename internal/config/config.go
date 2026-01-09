package config

import (
    "errors"
    "fmt"
    "os"
    "regexp"
    "strings"

    "gopkg.in/yaml.v3"
)

const (
    APIVersion = "atlas/v1"
    Kind       = "AtlasConfig"
)

var (
    // DNS-label compatible regex for release IDs
    // Lowercase only, hyphens allowed but not at start/end.
    // Max length is checked separately.
    releaseIDRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
    
    // DNS-label compatible regex for namespaces
    namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

type Config struct {
    APIVersion   string       `yaml:"apiVersion"`
    Kind         string       `yaml:"kind"`
    Repositories []Repository `yaml:"repositories,omitempty"`
    Policy       PolicyConfig `yaml:"policy,omitempty"`
    Releases     []Release    `yaml:"releases"`
}

type Repository struct {
    Name string `yaml:"name"`
    URL  string `yaml:"url"`
}

type PolicyConfig struct {
    Command []string `yaml:"command"`
}

type Release struct {
    ID           string   `yaml:"id"`
    Namespace    string   `yaml:"namespace"`
    Chart        string   `yaml:"chart"`
    Version      string   `yaml:"version"`
    Values       []string `yaml:"values,omitempty"`
    OverlayDir   string   `yaml:"overlayDir,omitempty"`
    PatchesDir   string   `yaml:"patchesDir,omitempty"`
    ResourcesDir string   `yaml:"resourcesDir,omitempty"`
}

// Load reads and validates the config from a file.
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    return &cfg, nil
}

// Save writes the config to a file.
func (c *Config) Save(path string) error {
    data, err := yaml.Marshal(c)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    // TODO: Use atomic write if we had a FS dependency injected,
    // but Config shouldn't really depend on FS.
    // For now, os.WriteFile is fine as config changes are rare and manual mostly.
    return os.WriteFile(path, data, 0644)
}

// Validate checks for structural and logical errors in the config.
func (c *Config) Validate() error {
    if c.APIVersion != APIVersion {
        return fmt.Errorf("unsupported apiVersion: %q (expected %q)", c.APIVersion, APIVersion)
    }
    if c.Kind != Kind {
        return fmt.Errorf("unsupported kind: %q (expected %q)", c.Kind, Kind)
    }

    releaseIDs := make(map[string]bool)
    
    for i, r := range c.Releases {
        if r.ID == "" {
            return fmt.Errorf("releases[%d].id is required", i)
        }
        if err := validateReleaseID(r.ID); err != nil {
            return fmt.Errorf("releases[%d].id %q is invalid: %w", i, r.ID, err)
        }
        if releaseIDs[r.ID] {
            return fmt.Errorf("duplicate release id: %q", r.ID)
        }
        releaseIDs[r.ID] = true

        if r.Namespace == "" {
            return fmt.Errorf("release %q: namespace is required", r.ID)
        }
        if !namespaceRegex.MatchString(r.Namespace) {
            return fmt.Errorf("release %q: namespace %q must be a valid DNS label", r.ID, r.Namespace)
        }
        if len(r.Namespace) > 63 {
             return fmt.Errorf("release %q: namespace %q must be <= 63 characters", r.ID, r.Namespace)
        }

        if r.Chart == "" {
            return fmt.Errorf("release %q: chart is required", r.ID)
        }
        if r.Version == "" {
            return fmt.Errorf("release %q: version is required", r.ID)
        }
    }

    return nil
}

func validateReleaseID(id string) error {
    if len(id) > 63 {
        return errors.New("must be <= 63 characters")
    }
    if !releaseIDRegex.MatchString(id) {
        return errors.New("must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character")
    }
    
    // Check reserved words
    reserved := []string{"cluster", "namespaces", "releases", "rendered", "cache", "lock", ".", ".."}
    for _, r := range reserved {
        if strings.EqualFold(id, r) {
            return fmt.Errorf("id %q is reserved", id)
        }
    }

    return nil
}
