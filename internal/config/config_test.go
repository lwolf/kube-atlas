package config

import (
    "testing"
)

func TestLoad_Valid(t *testing.T) {
    // We'll create a temp valid file
    cfg, err := Load("../../testdata/config/valid.yaml")
    if err != nil {
        t.Fatalf("Load(valid.yaml) returned error: %v", err)
    }
    
    if len(cfg.Releases) != 1 {
        t.Errorf("Expected 1 release, got %d", len(cfg.Releases))
    }
    if cfg.Releases[0].ID != "nginx-public" {
        t.Errorf("Expected release ID 'nginx-public', got %q", cfg.Releases[0].ID)
    }
}

func TestValidateReleaseID(t *testing.T) {
    tests := []struct {
        id      string
        wantErr bool
    }{
        {"valid-id", false},
        {"v", false},
        {"123", false},
        {"Valid-id", true}, // Uppercase
        {"-invalid", true},
        {"invalid-", true},
        {"invalid.", true},
        {"cluster", true}, // Reserved
        {"cache", true}, // Reserved
        {"this-is-a-very-long-id-that-exceeds-the-maximum-length-of-sixty-three-characters", true},
    }

    for _, tt := range tests {
        t.Run(tt.id, func(t *testing.T) {
            err := validateReleaseID(tt.id)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateReleaseID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
            }
        })
    }
}
