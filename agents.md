# Agents & Developer Guide

This document provides context for AI Agents and developers working on the Atlas codebase. It outlines the architecture, conventions, and key design decisions.

## Architecture

Atlas is built in Go and follows a modular architecture inspired by Clean Architecture principles.

- **`cmd/atlas`**: Entry point. Handles CLI flag parsing and TUI initialization.
- **`internal/app`**: Orchestration layer (Use Cases). Connects CLI/TUI to domain logic.
- **`internal/config`**: Configuration data models (`atlas.yaml`) and validation.
- **`internal/helm`**: Helm specific logic.
  - **`fetch`**: Component for chart fetching and vendoring.
  - **`render`**: Pure-Go Helm rendering engine.
  - **`oci`**: OCI registry client.
  - **`repo`**: HTTP Helm repo index parsing.
- **`internal/ui`**: TUI implementation using [Bubble Tea](https://github.com/charmbracelet/bubbletea).
- **`internal/workspace`**: path resolution logic. Defines where files live.
- **`internal/kust`**: Kustomize patching logic.
- **`internal/normalize`**: YAML normalization and sorting logic.

## Key Conventions

### 1. Filesystem & Paths
- **Workspace Resolution**: Always use `workspace.Resolve()` to determine paths for a release. Do not hardcode paths.
- **Atomic Writes**: Use `fs.AtomicWrite` for all file modifications to ensure data integrity.
- **Vendoring**: Charts are fully vendored into `chart/`. This allows `atlas` to work offline after fetching and provides a clear audit trail.

### 2. Lock File (`internal/lock`)
- `atlas/lock.yaml` is the source of truth for reproducibility.
- It stores resolved versions and content hashes (`VendorHash`) of the `chart/` directory.
- **Dirty Detection**: `status` command compares the current hash of `chart/` with the lock file to detect manual tampering.

### 3. Verification & Validation
- **Normalization**: All output is strictly normalized (sorted by Namespace, Kind, Name).
- **Namespace Enforcement**: Resources are validated to ensure they belong to the release's declared namespace. Cluster-scoped resources are rejected by default in v1.

### 4. TUI Architecture
- **Model-View-Update**: We use the ELM architecture provided by Bubble Tea.
- **State**: The main `Model` holds the `App` instance and sub-models (e.g., `dashboard`, `releaseform`).
- **Styles**: All UI styles are centralized in `internal/ui/styles`.

## Common Tasks

### adding a new CLI command
1. Define flags in `cmd/atlas/main.go`.
2. Implement the logic method in `internal/app/app.go` (or a specific file like `drift.go`).
3. Wire them up in the `switch` statement in `main.go`.
4. Be sure to expose the logic via `App` struct if it needs to be accessed by TUI.

### Adding a new TUI view
1. Create a new package in `internal/ui/components/`.
2. Implement `Model`, `Update`, `View`.
3. Add the component to the main `ui.Model`.
4. Handle state switching in `ui.Model.Update`.

## Testing
- **Unit Tests**: Place `_test.go` files next to the code.
- **Integration/E2E**: Use `verify.sh` for full system verification. It simulates a user workflow (scaffolding a repo, adding releases, rendering).
