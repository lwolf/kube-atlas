# Atlas

Atlas is a GitOps-native tool for managing Kubernetes charts. It combines the power of Helm with a declarative, repository-centric workflow, offering both a CLI and a Terminal User Interface (TUI).

## Features

- **Declarative Configuration**: Define your releases in `atlas.yaml`.
- **Hybrid Workflow**: Use the CLI for CI/CD and the TUI for interactive management.
- **Pure-Go Assembly**: Fetches OCI charts, renders Helm templates, and applies Kustomize patches without external binary dependencies (mostly).
- **Deterministic Output**: Enforces strict normalization and namespace scoping for rendered manifests.
- **Drift Detection**: Detects differences between your rendered manifests and the live cluster state (wraps `kubectl diff`).
- **Policy Checks**: Enforce custom policies on your rendered manifests (e.g. using `grep`, `opa`, or `datree`).

## Installation

```bash
go install github.com/lwolf/kube-atlas/cmd/atlas@latest
```

## Usage

### TUI Mode

Simply run `atlas` without arguments to launch the interactive dashboard:

```bash
atlas
```

Use the dashboard to view release status, add new releases, and navigate your configuration.

### CLI Mode

**Add a Release**
```bash
atlas release add --id my-app --namespace default --chart oci://ghcr.io/stefanprodan/charts/podinfo --version 6.7.0
```

**Fetch Charts**
Downloads charts to the `chart/` directory and updates the lock file.
```bash
atlas fetch
```

**Render Manifests**
Renders charts to `atlas/rendered/`, applying any overlays or patches.
```bash
atlas render
```

**Check Status**
Checks for "dirty" states (modified values, lock mismatch, etc.).
```bash
atlas status
```

**Diff**
Shows changes between the currently rendered files and what would be generated now.
```bash
atlas diff
```

**Policy Check**
Run configured policy commands against rendered manifests.
```bash
atlas check
```

**Drift Detection**
Check if the rendered state matches the live cluster.
```bash
atlas drift
```

## Configuration (atlas.yaml)

```yaml
apiVersion: atlas/v1
kind: AtlasConfig
releases:
  - id: my-app
    namespace: default
    chart: oci://ghcr.io/stefanprodan/charts/podinfo
    version: 6.7.0
    # Optional overrides
    # overlayDir: ...
    # patchesDir: ...
    # values: [ ... ]
policy:
  command: ["grep", "-L", "namespace"] # Example: fail if 'namespace' not found
```

## Directory Structure

```
.
├── atlas.yaml              # Main configuration
├── atlas/
│   ├── lock.yaml           # Lock file (versions & hashes)
│   ├── releases/           # Source of truth for releases
│   │   └── namespaces/
│   │       └── default/
│   │           └── my-app/
│   │               ├── chart/      # Vendored chart content (do not edit)
│   │               ├── values/     # Value overrides
│   │               ├── patches/    # Kustomize patches
│   │               └── resources/  # Extra raw manifests
│   └── rendered/           # Final output (commit this!)
```
