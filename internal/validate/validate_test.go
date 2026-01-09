package validate

import (
    "testing"
)

func TestValidate(t *testing.T) {
    valid := []string{
        `apiVersion: v1
kind: Service
metadata:
  name: s1
  namespace: n1
`,
        `apiVersion: v1
kind: Pod
metadata:
  name: p1
  namespace: n1
`,
    }
    if err := Validate(valid); err != nil {
        t.Errorf("Validate(valid) failed: %v", err)
    }

    invalid := []string{
        `apiVersion: v1
kind: Service
metadata:
  name: s1
  namespace: n1
`,
        `apiVersion: v1
kind: Service
metadata:
  name: s1
  namespace: n1
`,
    }
    if err := Validate(invalid); err == nil {
        t.Error("Validate(duplicate) expected error, got nil")
    }

    missingName := []string{
        `apiVersion: v1
kind: Service
metadata:
  namespace: n1
`,
    }
    if err := Validate(missingName); err == nil {
        t.Error("Validate(missingName) expected error, got nil")
    }
}
