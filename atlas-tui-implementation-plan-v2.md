# Atlas TUI (kube-atlas rewrite) — Step-by-step implementation plan (Go)

**Objective:** Rewrite `kube-atlas` from scratch as a **TUI-first** Go application that manages Helm repositories (HTTP + OCI), vendors charts, renders deterministic final Kubernetes YAML, and maintains **overlays** (chart-level) and **patches** (post-render Kustomize), with strong emphasis on **simplicity, minimal dependencies, testability, and stability**.

This plan assumes **Argo CD** consumes the rendered YAML directories directly (Directory source), and the tool commits both inputs and final outputs.

---

## Change log (this revision)
- Adds an explicit **“Add release”** workflow (CLI + TUI wizard) including scaffolding.
- Specifies **requirements for `release-id`** (format, uniqueness, mapping to Helm release name and filesystem paths).
- Removes any per-release rendered scratch directories (all ephemeral artifacts live under `atlas/cache/`).

---

## Scope decisions (confirmed)
- **Pure-Go** Helm render and **pure-Go** Kustomize patching.
- **GitOps repository** contains both:
  - **Inputs**: vendored charts, overlays, values, patches, extra resources
  - **Outputs**: final rendered multi-doc YAML (apply-able by any tool)
- **Argo CD** is the CD target; repository layout supports **namespace-scoped paths**.
- **Hooks excluded** from final output.
- **No `atlas/repos/` directory**; repositories defined inline in `atlas.yaml`.
- **No per-release rendered scratch directory**; only `atlas/cache/` is ephemeral/ignored.
- Add **`resources/`** per release for extra objects (e.g., cert-manager Certificate).
- **One repo = one cluster**.
- **One release = one namespace target** (duplicating vendored chart per release is acceptable in v1).

---

## Release ID requirements (must be enforced)

### Definition
`release-id` is the stable identifier for a single release instance (one chart deployment into one namespace). It is used in:
- directory paths under `atlas/releases/namespaces/<namespace>/<release-id>/...`
- output path `atlas/rendered/namespaces/<namespace>/<release-id>.yaml`
- default Helm release name (unless overridden)

### Constraints
Enforce the following constraints on `release-id`:

1. **Format (DNS-label compatible)**  
   - Regex: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
   - Lowercase only.
   - Hyphens allowed, but not at beginning or end.

2. **Length**  
   - Recommended max: 53 characters (keeps headroom when charts append suffixes).
   - Hard max: 63 characters (DNS label max). If you choose hard max 63, ensure file paths remain reasonable.

3. **Uniqueness**  
   - **Globally unique within the repo** (recommended for simplicity).  
     This prevents collisions in tooling, logging, and status tables.

4. **Reserved IDs**  
   Disallow (case-insensitive) the following to avoid directory/UX ambiguity:
   - `cluster`, `namespaces`, `releases`, `rendered`, `cache`, `lock`
   - `.` or `..`
   - any value that would create hidden paths (starts with `.`)

5. **Filesystem safety**
   - Must not contain `/`, `\`, spaces, or control characters.
   - Must not end with `.` (Windows compatibility).

### Mapping to Helm release name
Default mapping (v1):
- Helm release name = `release-id`

Optional override (v1.1+ if needed):
- `helmReleaseName` in release config; must satisfy Helm name constraints and be stable.

### Tests
- Table-driven tests for valid/invalid release-id
- Ensure uniqueness enforcement in config validation and `release add` flow

---

## Deliverables
1. `atlas` binary:
   - CLI commands (for CI and automation)
   - TUI (for interactive management)
2. Opinionated repo layout under `atlas/`
3. Deterministic renderer producing:
   - `atlas/rendered/namespaces/<ns>/<release-id>.yaml`
   - `atlas/rendered/cluster/*.yaml` (cluster-scoped buckets)
4. `atlas/lock.yaml` tracking resolved chart versions/digests and vendor “dirty” state.
5. GitHub Actions workflow as a prerequisite quality gate.

---

## Repository layout (v1 contract)

At repo root:
- `atlas.yaml` (committed)
- `.gitignore` includes `atlas/cache/**`

Under `atlas/`:
```
atlas/
  lock.yaml

  releases/
    namespaces/
      <namespace>/
        <release-id>/
          chart/          # vendored chart source + deps (committed, tool-managed)
          overlay/        # chart overlay (committed)
          values/         # values files (committed)
          patches/        # kustomize overlay for post-render patching (committed)
          resources/      # extra raw manifests (committed)

  rendered/
    namespaces/
      <namespace>/
        <release-id>.yaml  # final multi-doc YAML (committed)
    cluster/
      crds.yaml
      rbac.yaml
      other.yaml

  cache/          # ephemeral (NOT committed)
```

**Rule:** `atlas/rendered/**` contains only final YAML. No kustomization files there.  
**Rule:** Everything under `atlas/cache/**` must be ignored and never referenced by GitOps.

---

## High-level architecture (testable core + thin UI)

### Packages (recommended)
- `internal/config` — parsing/validation of `atlas.yaml` and `atlas/lock.yaml`
- `internal/workspace` — path resolution, file conventions
- `internal/fs` — filesystem abstraction + atomic writes
- `internal/helm` — chart fetch (HTTP/OCI), dependency handling, render
- `internal/kust` — post-render kustomize patching (pure-Go)
- `internal/resources` — loading & validating `resources/`
- `internal/normalize` — parse YAML docs, classify scope, stable sort, canonical emit
- `internal/diff` — unified diff generation for rendered output
- `internal/validate` — best-effort validation (schema optional)
- `internal/policy` — exec-based policy runner (e.g., conftest)
- `internal/drift` — best-effort drift detection (build-tag optional if you want to avoid client-go by default)
- `internal/app` — orchestration services used by CLI + TUI
- `internal/tui` — terminal UI using `tcell` (or similar minimal library)
- `cmd/atlas` — CLI entry point

### Dependency principle
Minimise external deps, but accept:
- Helm v3 SDK (`helm.sh/helm/v3/...`)
- Kustomize Go API (`sigs.k8s.io/kustomize/...`)
- `tcell` for TUI
- Small semver lib (or implement minimal constraints if you prefer)

---

## Quality gates (prerequisite)

### Step 0 — Create GitHub Actions workflow (required before feature work)
**Files**
- `.github/workflows/ci.yml`

**Workflow requirements**
- Trigger on `push` and `pull_request`
- Go version pinned (e.g., `1.23.x` or current stable in your org)
- Steps:
  - `go test ./...`
  - `go vet ./...`
  - `gofmt` check (fail if diff)
  - optional: `staticcheck` (skip if you want fewer deps)

**Verification**
- PR that only adds workflow + minimal `main.go` should pass CI.

**Tests**
- N/A (pipeline-level)

---

# Step-by-step implementation plan (each step testable)

Each step below should be implemented as a PR. Every PR must:
- include tests (unit and/or golden)
- keep the build green on GitHub Actions
- avoid large refactors that obscure review

---

## Step 1 — Skeleton binary + module layout
**Goal:** A compiling `atlas` binary with structured packages.

**Implementation**
- Initialise `go.mod`
- Add `cmd/atlas/main.go` with subcommand scaffolding:
  - `atlas version`
- Add minimal `internal/app` with `App` struct

**Verification**
- `go test ./...` passes
- `atlas version` prints version info

**Tests**
- Unit test for version formatting (if you implement build info)

---

## Step 2 — Config model (`atlas.yaml`) and validation
**Goal:** Load config deterministically and validate required fields.

**Implementation**
- Define `atlas.yaml` schema:
  - `apiVersion`, `kind`
  - `paths` (optional)
  - `repositories[]`
  - `releases[]` with:
    - `id`, `namespace`
    - `chart` (repo name or OCI ref, chart name, version/constraint)
    - optional `values[]` (default to release values dir)
    - optional `overlayDir`, `patchesDir`, `resourcesDir`
- Implement strict validation:
  - **release-id constraints** (format, reserved, length)
  - release IDs unique (global)
  - namespace non-empty and DNS-label compatible
  - repo references exist (for HTTP repos)
- Implement `config.Load(path)`.

**Verification**
- Add `testdata/config/valid.yaml`
- Add `testdata/config/invalid-*.yaml`
- `atlas validate-config` (CLI) returns non-zero on invalid config

**Tests**
- Table-driven tests for parsing and validation errors
- Dedicated tests for release-id validation matrix

---

## Step 3 — Workspace conventions + path resolution
**Goal:** Resolve all paths from config consistently (support limited overrides).

**Implementation**
- `workspace.Resolve(config, repoRoot)` returns:
  - release root dir
  - chart dir, overlay dir, values dir, patches dir, resources dir
  - rendered output path (`atlas/rendered/namespaces/<ns>/<release-id>.yaml`)
- Enforce defaults if directories omitted.
- Ensure directories are created when needed (non-destructive).

**Verification**
- `atlas doctor` prints resolved paths for each release

**Tests**
- Snapshot tests: given config, resolved paths match expected

---

## Step 4 — Add release (CLI + TUI wizard) + scaffolding
**Goal:** Provide a safe, low-friction way to add a release, and ensure manual edits remain possible.

### CLI command
Introduce:
- `atlas release add --id <id> --namespace <ns> --chart <repo/chart|oci:...> --version <v|constraint> [--no-fetch] [--no-render]`

Examples:
```bash
atlas release add --id nginx-public --namespace web --chart bitnami/nginx --version 15.5.2
atlas release add --id nginx-public --namespace web --chart oci:ghcr.io/myorg/charts/nginx --version 15.5.2
```

### Behaviour
1. Validate inputs (release-id constraints, namespace constraints, repo existence if needed).
2. Mutate `atlas.yaml` by adding a `releases[]` entry (preserve file as best-effort).
3. Create release directories (idempotent):
   - `overlay/`
   - `values/` (create `00-base.yaml`)
   - `patches/` (create a minimal `kustomization.yaml`)
   - `resources/`
4. Unless `--no-fetch`, run `atlas fetch --release <id>`.
5. Unless `--no-render`, run `atlas render --release <id>` to produce final YAML.

### Minimal generated files
- `values/00-base.yaml`:
  - empty YAML document with comments explaining ordering
- `patches/kustomization.yaml`:
  - a minimal valid file (no resources required because the tool supplies the base in-memory)

### TUI wizard (initial)
- Dashboard action: “Add release”
- Steps:
  - enter release-id
  - enter namespace
  - choose repo, chart, version
  - confirm, then create + (fetch+render)

**Verification**
- After `release add`, all directories exist, config updated, and rendered YAML exists (unless skipped).

**Tests**
- Unit tests: config mutation (adds release, enforces uniqueness/reserved IDs)
- FS tests: scaffolding created, re-run is idempotent
- Integration test (offline): use a test HTTP repo served by `httptest`, run add (with fetch/render enabled) and assert outputs exist

---

## Step 5 — Filesystem abstraction + atomic writes
**Goal:** Make rendering safe and testable.

**Implementation**
- `fs.FS` interface:
  - `ReadFile`, `WriteFile`, `MkdirAll`, `WalkDir`, `Stat`, `Rename`, etc.
- `fs.AtomicWrite(path, bytes)`:
  - write temp file in same dir
  - fsync (best-effort)
  - rename
- Implement real `OSFS` and in-memory `MemFS` for tests.

**Verification**
- `AtomicWrite` never leaves partial output when interrupted (best-effort simulation)

**Tests**
- Unit tests with `MemFS`
- Failure injection tests (simulate rename failure)

---

## Step 6 — Lock file model + hashing (dirty detection)
**Goal:** Reproducibility and “dirty” semantics.

**Implementation**
- Define `atlas/lock.yaml` with:
  - per-release `resolved` (repo, chart, version, digest)
  - `vendorHash` (hash of `chart/` directory contents)
  - optional `overlayHash`, `valuesHash`, `patchesHash`, `resourcesHash` (for UX; not required for correctness)
- Implement `HashTree(dir)`:
  - walk files in lexical order
  - hash path + file bytes
- Dirty rules:
  - chart dirty if current `HashTree(chartDir)` != `vendorHash`
- CLI:
  - `atlas status` shows dirty/clean per release

**Verification**
- Modify a file under `chart/` and `atlas status` shows dirty.

**Tests**
- Hash determinism tests
- Dirty detection tests

---

## Step 7 — Helm repo index (HTTP) fetching + chart version resolution
**Goal:** Support HTTP(S) Helm repositories for chart browsing and resolution.

**Implementation**
- `helm/repo` service:
  - fetch `index.yaml`
  - parse chart entries
  - resolve version constraints (or pinned)
- Cache downloaded index in `atlas/cache/helm/http/...` (ignored)
- Repos defined inline in `atlas.yaml`

**Verification**
- `atlas repo list`
- `atlas repo refresh`
- `atlas chart search nginx`
- `atlas chart versions bitnami/nginx`

**Tests**
- `httptest.Server` serving `index.yaml` from `testdata/helmrepo/`
- Unit tests for semver resolution

---

## Step 8 — OCI chart fetch (minimal v1)
**Goal:** Fetch charts from OCI registries.

**Implementation**
- Implement OCI pull for charts:
  - resolve `oci://...` reference + tag
  - download blob/layers to cache
  - extract chart archive
- Keep auth best-effort (use standard OCI credential sources if supported by chosen library).

**Verification**
- Provide an integration test that is skipped unless env vars configured, and unit tests for parsing.
- CLI: `atlas fetch --release <id>` works for OCI repos when configured locally.

**Tests**
- Unit tests for ref parsing and cache path selection
- Optional integration test behind `-run OCI` + env gate

---

## Step 9 — Vendor chart into `chart/` + dependency vendoring
**Goal:** Create committed chart tree per release and update lock.

**Implementation**
- `atlas fetch`:
  - if release is dirty: refuse unless `--force` or `--rebase`
  - download chart, extract to temp in `atlas/cache/tmp`
  - vendor into `atlas/releases/.../chart` (atomic replace)
  - run dependency vendoring via Helm SDK into `chart/charts`
  - compute `vendorHash`, update lock with digest/version
- `--rebase`:
  - overwrite `chart/` from upstream
  - preserve `overlay/`, `values/`, `patches/`, `resources/`

**Verification**
- Run `atlas fetch --release <id>` and see populated `chart/` + updated lock

**Tests**
- Serve `.tgz` from `httptest.Server`, verify extraction
- Dirty refusal test
- Rebase preserves overlay/patches/resources directories

---

## Step 10 — Overlay merge (chart-level)
**Goal:** Allow overriding chart files without editing base.

**Implementation**
- Create `helm/chartfs` effective filesystem:
  - base chart dir + overlay dir merged (overlay wins)
  - optional `.atlas-remove` file listing paths to delete from base
- Ensure render reads from effective fs, not by mutating base chart.

**Verification**
- Add an override template file in `overlay/templates/...` and confirm it affects render.

**Tests**
- Unit tests for merge logic
- Delete marker tests (if implemented)

---

## Step 11 — Values loading/merging
**Goal:** Deterministic values application.

**Implementation**
- Values sources:
  - chart default values
  - files in `values/` dir (lexical order) OR `release.values[]` explicit list
- Merge semantics:
  - YAML map merge (later wins)
  - arrays: replace by default (Helm-like)
- Validate values YAML syntax

**Verification**
- Changing `values/10-foo.yaml` changes render output deterministically.

**Tests**
- Golden tests for merged values
- Invalid YAML test returns error

---

## Step 12 — Helm render engine (pure-Go) + hook exclusion
**Goal:** Generate raw manifest stream from Helm.

**Implementation**
- Use Helm SDK to render templates:
  - set release name = release ID (or configurable)
  - set namespace from release
- Filter out hook resources:
  - remove docs that include annotation `helm.sh/hook`
- Ensure CRDs inclusion toggle supported (v1: include if `includeCRDs: true`)

**Verification**
- `atlas render --release <id>` produces output in memory and writes final file.

**Tests**
- Minimal test chart fixtures under `testdata/charts/`
- Hook filtering test chart fixture
- Golden output tests (pre-kustomize)

---

## Step 13 — Extra resources (`resources/`) inclusion
**Goal:** Support adding objects like `cert-manager Certificate` without chart overlays.

**Implementation**
- Load all `.yaml/.yml` files in `resources/` (lexical order)
- Parse docs, enforce namespace:
  - if namespaced and namespace missing: inject release namespace
  - if present and different: error
- Reject cluster-scoped kinds by default (configurable later)
- Concatenate to Helm output before kustomize

**Verification**
- Drop a YAML file into `resources/` manually and rerender; it appears in final YAML.

**Tests**
- Namespace injection test
- Namespace mismatch error test
- Duplicate ID conflict test (with Helm producing same object)

---

## Step 14 — Post-render Kustomize patching (pure-Go)
**Goal:** Apply patches in `patches/` to the combined manifest stream.

**Implementation**
- Build a Kustomize target in-memory:
  - base = generated manifest stream
  - overlay = `patches/` directory (must include `kustomization.yaml`)
- Execute kustomize build, return patched YAML stream

**Verification**
- A patch that adds a label changes final output.

**Tests**
- Fixture with patches; golden output test

---

## Step 15 — Normalisation, scope enforcement, deterministic final YAML
**Goal:** Produce stable output and enforce namespace path correctness.

**Implementation**
- Parse YAML docs into unstructured objects
- Classify:
  - namespaced vs cluster-scoped
- Enforce:
  - namespaced objects for this release namespace only
  - cluster-scoped objects are rejected (v1) OR routed to cluster bucket only if enabled (future)
- Best-effort stable ordering:
  - CRDs first (if allowed)
  - Namespaces next
  - then sort by (group, version, kind, namespace, name)
- Canonical YAML emitter:
  - stable map key ordering (best-effort)
  - stable doc separators
- Write final YAML atomically to:
  - `atlas/rendered/namespaces/<ns>/<release-id>.yaml`

**Verification**
- Running `atlas render` twice yields byte-identical output.

**Tests**
- Idempotence test (same output across runs)
- Ordering test
- Cluster-scoped rejection test

---

## Step 16 — CLI commands: fetch/render/status/diff
**Goal:** CI-friendly automation surface.

**Implementation**
- `atlas fetch --release <id>|--all`
- `atlas render --release <id>|--all`
- `atlas status` (dirty + rendered presence)
- `atlas diff --release <id>|--all`:
  - diff committed file vs newly rendered candidate
  - exit codes: 0 no diff, 1 diff, 2+ error

**Verification**
- `atlas diff` shows changes before you commit.

**Tests**
- Diff service unit tests (known strings)
- CLI argument parsing tests (lightweight)

---

## Step 17 — Validation (best-effort)
**Goal:** Detect common issues before Argo CD.

**Implementation**
- Validate:
  - YAML parse
  - required fields apiVersion/kind/metadata.name
  - duplicate resource IDs
  - namespace mismatches
- Optional schema validation:
  - v1: implement as “plugin hook” (exec kubeconform) rather than heavy embedding

**Verification**
- `atlas validate --release <id>` returns non-zero on invalid manifests.

**Tests**
- Invalid manifest fixtures
- Duplicate detection tests

---

## Step 18 — Policy checks (exec-based)
**Goal:** Integrate Conftest/OPA-like checks without deep deps.

**Implementation**
- Configurable policy commands in `atlas.yaml`:
  - e.g. `conftest test -p policy -`
- Tool feeds rendered YAML to stdin
- Capture output, return status and diagnostics

**Verification**
- `atlas policy --release <id>` reports pass/fail

**Tests**
- Use a fake script in tests to simulate pass/fail

---

## Step 19 — Drift detection (best-effort, optional build tag)
**Goal:** Compare rendered desired state vs live cluster.

**Implementation**
- Use kubeconfig if available
- Fetch live objects for the same IDs (dynamic client)
- Normalise both (strip status, known defaulted fields best-effort) and compare
- Report drift summary

**Dependency strategy**
- Consider build tag `drift` to avoid pulling `client-go` into minimal builds:
  - default build: drift disabled
  - `go test -tags=drift ./...` for drift-enabled builds

**Verification**
- On a dev cluster, drift reports differences for a modified live object.

**Tests**
- Fake client tests (no real cluster required)

---

## Step 20 — TUI foundation (dashboard)
**Goal:** Minimal, stable TUI with a dashboard and action execution.

**Implementation**
- Use `tcell` (or similar) and implement small widgets:
  - table, status bar, log pane
- Dashboard:
  - list releases with status columns:
    - dirty, fetched, rendered, diff, validate, policy, drift
  - actions: Fetch, Render, Diff, Validate, Policy, Drift
- Inject services so UI remains thin and testable.

**Verification**
- Run `atlas tui` and trigger fetch/render from UI.

**Tests**
- Unit tests for state transitions (no terminal)
- Service mocks to assert calls

---

## Step 21 — TUI usability helpers (overlay/patch/resource creation)
**Goal:** High-leverage workflows without large dependencies.

**Implementation**
- Overlay helper: “take file into overlay”:
  - copies selected file from `chart/` into `overlay/`, then opens `$EDITOR`
- Resource helper:
  - create from template (e.g., Certificate), or import file
- Patch helper (v1):
  - generate strategic-merge patch template for a selected object and ensure `kustomization.yaml` references it
- Optional v1.1:
  - open object in editor, generate JSON6902 diff automatically

**Verification**
- Create an overlay/patch/resource via TUI and observe final YAML change.

**Tests**
- Copy/template generation tests (filesystem abstraction)
- Kustomization update tests

---

# Suggested `.gitignore` (minimum)
```
atlas/cache/**
```

---

# Appendix: Minimal GitHub Actions workflow (template)

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23.x'
          check-latest: true

      - name: Go fmt (check)
        run: |
          fmt_out=$(gofmt -l .)
          if [ -n "$fmt_out" ]; then
            echo "gofmt needed on:"
            echo "$fmt_out"
            exit 1
          fi

      - name: Go vet
        run: go vet ./...

      - name: Go test
        run: go test ./...
```

---

# Appendix: Release inputs (example)

```
atlas/releases/namespaces/web/nginx-public/
  chart/            (vendored)
  overlay/          (overrides)
  values/
    00-base.yaml
  patches/
    kustomization.yaml
    patches/
      add-label.yaml
  resources/
    certificate.yaml
```

Final output:
- `atlas/rendered/namespaces/web/nginx-public.yaml`
