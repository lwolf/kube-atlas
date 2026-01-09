package policy

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
)

type Checker struct {
    Command []string
}


// New returns a new Checker.
func New(command []string) *Checker {
    return &Checker{
        Command: command,
    }
}

// Check executes the configured command against the manifest file.
// It assumes the manifest file path is the last argument to the command.
func (c *Checker) Check(ctx context.Context, manifestPath string) error {
    if len(c.Command) == 0 {
        return nil
    }

    args := append(c.Command[1:], manifestPath)
    cmd := exec.CommandContext(ctx, c.Command[0], args...)
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("policy check failed for %s: %s\n%s", manifestPath, stdout.String(), stderr.String())
    }
    return nil
}
