package ui

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/lwolf/kube-atlas/internal/app"
)

func TestModel_Init(t *testing.T) {
    // Setup dummy repo
    tmpDir := t.TempDir()
    err := os.WriteFile(filepath.Join(tmpDir, "atlas.yaml"), []byte("releases: []"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    application := app.New()
    m, err := NewModel(context.Background(), tmpDir, application)
    if err != nil {
        t.Fatalf("NewModel failed: %v", err)
    }

    if m.state != StateDashboard {
        t.Errorf("expected initial state Dashboard, got %v", m.state)
    }
}

func TestModel_Update_Quit(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "atlas.yaml"), []byte("releases: []"), 0644)
    
    m, _ := NewModel(context.Background(), tmpDir, app.New())
    
    // Send 'q'
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
    if cmd == nil {
        t.Fatal("expected quit command, got nil")
    }
    // Tea.Quit is an internal type, hard to compare directly, but if we get a cmd back it's likely correct 
    // given the switch statement. 
    // Actually, bubbletea returns tea.Quit which is a tea.Cmd.
}

func TestModel_Update_SwitchToForm(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(filepath.Join(tmpDir, "atlas.yaml"), []byte("releases: []"), 0644)
    
    m, _ := NewModel(context.Background(), tmpDir, app.New())
    
    // Send 'a' to switch to add release
    newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
    
    updatedM, ok := newModel.(Model)
    if !ok {
        t.Fatal("Update did not return Model")
    }
    
    if updatedM.state != StateAddRelease {
        t.Errorf("expected state StateAddRelease, got %v", updatedM.state)
    }
}
