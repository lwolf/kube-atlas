package policy

import (
    "context"
    "os"
    "path/filepath"
    "testing"
)

func TestChecker_Check(t *testing.T) {
    // Create dummy manifest
    tmpDir := t.TempDir()
    manifest := filepath.Join(tmpDir, "manifest.yaml")
    if err := os.WriteFile(manifest, []byte("ok"), 0644); err != nil {
        t.Fatal(err)
    }

    tests := []struct {
        name    string
        command []string
        wantErr bool
    }{
        {
            name:    "pass (grep)",
            command: []string{"grep", "ok"},
            wantErr: false,
        },
        {
            name:    "fail (grep)",
            command: []string{"grep", "fail"},
            wantErr: true,
        },
        {
            name:    "no command",
            command: []string{},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := New(tt.command)
            if err := c.Check(context.Background(), manifest); (err != nil) != tt.wantErr {
                t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
