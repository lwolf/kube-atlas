# TUI Improvements Work Plan

## Context

### Original Request
User reported 6 major usability issues with the kube-atlas TUI:
1. TUI not using full screen - should adapt to occupy all usable space
2. Cannot "enter" selected deployment from list
3. Cannot edit selected deployment
4. Cannot add/edit chart repositories
5. Cannot search charts in added registries and create release from selected
6. Stuck when errors occur - have to kill entire app

### Interview Summary
**Key Discussions**:
- Confirmed Bubbletea TUI framework usage
- User prefers "enter" shows details view with edit/view options
- Chart search should be integrated into add release workflow
- Error recovery should provide "go back" option

**Research Findings**:
- Bubbletea fullscreen: Use `tea.WithAltScreen()` + handle `tea.WindowSizeMsg` for responsive layout
- Current navigation: 'a' (add release), 'r' (repos), 'd' (dashboard), 'q' (quit)
- Dashboard has table selection via `SelectedRelease()` but no "enter" handler
- No edit workflows exist yet (need to add `EditRelease`/`EditRepository` to app layer)
- Errors cause app termination instead of graceful recovery
- Repo list has no selection tracking (unlike dashboard)

### State Machine & Message Contracts
**New States to Add**:
- `StateReleaseDetails` - Shows release info with action buttons
- `StateEditRelease` - Edit form pre-populated with release data
- `StateEditRepo` - Edit form pre-populated with repository data
- `StateChartSearch` - Chart search interface within release creation

**New Message Types**:
- `ReleaseDetailsMsg{Release: config.Release}` - Navigate to details view
- `EditReleaseMsg{Release: config.Release}` - Navigate to edit form
- `EditRepoMsg{Repo: config.Repository}` - Navigate to repo edit
- `ChartSearchMsg{}` - Enter chart search mode
- `ErrorRecoveryMsg{}` - Clear error and return to previous state

### Metis Review
**Identified Gaps** (addressed):
- Added explicit navigation patterns for "enter" actions
- Specified error recovery mechanisms
- Defined chart search integration approach
- Added guardrails for scope control

---

## Work Objectives

### Core Objective
Fix all 6 reported TUI usability issues to create a professional, user-friendly terminal interface for Helm chart management.

### Concrete Deliverables
- Fullscreen TUI experience
- Release details/edit views accessible via "enter"
- Repository editing capability
- Chart search integrated into release creation
- Graceful error handling with recovery options

### Definition of Done
- [ ] All 6 original issues resolved
- [ ] TUI runs in fullscreen mode
- [ ] Can navigate to release details with enter key
- [ ] Can edit existing releases and repositories
- [ ] Can search charts during release creation
- [ ] Error states provide recovery options instead of app termination

### Must Have
- Fullscreen mode enabled
- "Enter" navigation on selected releases
- Edit functionality for releases and repositories
- Chart search in add release workflow
- Error recovery with back navigation

### Must NOT Have (Guardrails)
- No major architectural changes to existing Bubbletea framework
- No removal of existing CLI functionality
- No changes to underlying Helm/chart management logic
- No breaking changes to existing navigation patterns

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: NO (only unit tests for UI components)
- **User wants tests**: YES (TDD for new components)
- **Framework**: Go testing with Bubbletea test patterns

### If TDD Enabled

Each TODO follows RED-GREEN-REFACTOR:

**Task Structure:**
1. **RED**: Write failing test first
   - Test file: `[component]_test.go`
   - Test command: `go test ./internal/ui/...`
   - Expected: FAIL (test exists, implementation doesn't)
2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./internal/ui/...`
   - Expected: PASS
3. **REFACTOR**: Clean up while keeping green
   - Command: `go test ./internal/ui/...`
   - Expected: PASS (still)

---

## Task Flow

```
1. Enable fullscreen → 2. Add release details view → 3. Add edit release → 4. Add repo editing → 5. Add chart search → 6. Fix error handling
                      ↓
                7. Integration testing
```

## Parallelization

| Group | Tasks | Reason |
|-------|-------|--------|
| A | 2, 3, 4 | Independent UI components |

| Task | Depends On | Reason |
|------|------------|--------|
| 5 | 2 | Chart search integrated into release workflow |
| 6 | All | Error handling affects all screens |
| 7 | All | Integration testing after all features |

---

## TODOs

- [ ] 1. Enable fullscreen mode

  **What to do**:
  - Add `tea.WithAltScreen()` to program initialization in `cmd/atlas/main.go`
  - Handle `tea.WindowSizeMsg` in `internal/ui/model.go` to resize components responsively
  - Update table heights dynamically (currently hardcoded to 10)

  **Must NOT do**:
  - Change any other program options

  **Parallelizable**: YES

  **References**:

  **Pattern References** (existing code to follow):
  - `cmd/atlas/main.go:75-78` - Current program initialization pattern
  - `internal/ui/dashboard/dashboard.go:29-35` - Table creation with hardcoded height

  **API/Type References** (contracts to implement against):
  - Bubbletea docs: `tea.WithAltScreen()` and `tea.WindowSizeMsg`

  **Test References** (testing patterns to follow):
  - `internal/ui/ui_test.go` - Existing TUI test patterns

  **Acceptance Criteria**:

  **Manual Execution Verification**:
  - [ ] Run `atlas` command
  - [ ] Verify TUI occupies entire terminal window (no visible background)
  - [ ] Resize terminal window
  - [ ] Verify tables and components adjust to new size
  - [ ] Verify terminal state restored on exit (no leftover artifacts)

  **Commit**: YES
  - Message: `feat: enable fullscreen mode with responsive layout`
  - Files: `cmd/atlas/main.go`, `internal/ui/model.go`
  - Pre-commit: `go test ./cmd/atlas/...`

- [ ] 2. Add release details view

  **What to do**:
  - Create new `releasedetails` component in `internal/ui/components/`
  - Add `StateReleaseDetails` to `internal/ui/model.go:19`
  - Add `ReleaseDetailsMsg{Release: config.Release}` message type
  - In `internal/ui/model.go`, handle Enter key in `StateDashboard` by calling `m.dashboard.SelectedRelease()` and emitting `ReleaseDetailsMsg`
  - Display release info with action buttons (edit, view YAML, back)
  - Handle button navigation with tab/enter, emit `EditReleaseMsg` for edit button

  **Must NOT do**:
  - Implement edit functionality yet (separate task)

  **Parallelizable**: YES

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/components/releaseform/model.go` - Form component pattern
  - `internal/ui/model.go:90-98` - State transition pattern
  - `internal/ui/dashboard/dashboard.go:64` - `SelectedRelease()` method

  **API/Type References** (contracts to implement against):
  - `internal/config/config.go:Release` - Release data structure
  - Bubbletea docs: Key handling patterns

  **Test References** (testing patterns to follow):
  - `internal/ui/ui_test.go:48-65` - State transition testing

  **Acceptance Criteria**:

  **If TDD (tests enabled):**
  - [ ] Test file created: `internal/ui/components/releasedetails/model_test.go`
  - [ ] Test covers: Enter key navigation from dashboard
  - [ ] `go test ./internal/ui/components/releasedetails` → PASS

  **Manual Execution Verification**:
  - [ ] Start TUI and navigate to dashboard
  - [ ] Select a release with arrow keys
  - [ ] Press Enter key
  - [ ] Verify details view shows with release information and action buttons

  **Commit**: YES
  - Message: `feat: add release details view accessible via enter key`
  - Files: `internal/ui/components/releasedetails/`, `internal/ui/dashboard/dashboard.go`, `internal/ui/model.go`
  - Pre-commit: `go test ./internal/ui/...`

- [ ] 3. Add release editing capability

  **What to do**:
  - Create `releaseedit` component (pre-populated form)
  - Add `StateEditRelease` to `internal/ui/model.go:19`
  - Add `EditReleaseMsg{Release: config.Release}` message type
  - Add "Edit" button handler in release details view to emit `EditReleaseMsg`
  - Implement `app.EditRelease()` method in `internal/app/app.go` (similar to `AddRelease` but updates existing)
  - **Edit Semantics**: Release ID is immutable (cannot change directory structure). Chart/version changes remain config-only (no auto fetch/render). Validate via `internal/config/config.go:87`, save via `(*Config).Save` at `internal/config/config.go:76`, then reload config via `internal/ui/model.go:156` pattern

  **Must NOT do**:
  - Change existing add release workflow
  - Allow editing release ID

  **Parallelizable**: YES (with 4)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/components/releaseform/model.go` - Form structure to adapt
  - `internal/ui/model.go:90-98` - Result message handling pattern
  - `internal/app/release.go:23` - `AddRelease` implementation pattern

  **API/Type References** (contracts to implement against):
  - `internal/config/config.go:87` - Validation rules
  - `internal/config/config.go:76` - Save method
  - `internal/ui/model.go:156` - Config reload pattern

  **Test References** (testing patterns to follow):
  - `internal/ui/components/releaseform/model.go:18-24` - Message types pattern

  **Acceptance Criteria**:

  **If TDD (tests enabled):**
  - [ ] Test file created: `internal/ui/components/releaseedit/model_test.go`
  - [ ] Test covers: Form pre-population and update flow
  - [ ] `go test ./internal/ui/components/releaseedit` → PASS

  **Manual Execution Verification**:
  - [ ] Navigate to release details view
  - [ ] Click "Edit" button
  - [ ] Verify form is pre-populated with current values
  - [ ] Verify ID field is disabled/not editable
  - [ ] Make changes to chart/version and submit
  - [ ] Verify release is updated in config (no auto fetch/render)

  **Commit**: YES
  - Message: `feat: add release editing capability`
  - Files: `internal/ui/components/releaseedit/`, `internal/app/app.go`, `internal/ui/components/releasedetails/model.go`
  - Pre-commit: `go test ./internal/ui/... ./internal/app/...`

- [ ] 4. Add repository editing capability

  **What to do**:
  - Create `repoedit` component (pre-populated form)
  - Add `StateEditRepo` to `internal/ui/model.go:19`
  - Add `EditRepoMsg{Repo: config.Repository}` message type
  - **Prerequisite**: Update `internal/ui/components/repolist/model.go` to track repositories and selection (add `repos []config.Repository` field and `SelectedRepository()` method similar to dashboard)
  - Add "enter" key handler in repo list to emit `EditRepoMsg` with selected repo
  - Implement `app.EditRepository()` method in `internal/app/app.go` (similar to `AddRepository` but updates existing)
  - Use same validation/save/reload pattern as release editing

  **Must NOT do**:
  - Change existing add repo workflow

  **Parallelizable**: YES (with 3)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/components/repoform/model.go` - Form structure to adapt
  - `internal/ui/components/repolist/model.go` - List component pattern (needs modification)
  - `internal/ui/dashboard/dashboard.go:64` - Selection tracking pattern to replicate

  **API/Type References** (contracts to implement against):
  - `internal/app/repo.go:16` - `AddRepository` implementation pattern
  - `internal/config/config.go:Repository` - Repository data structure

  **Test References** (testing patterns to follow):
  - `internal/ui/components/repoform/model.go` - Form message types

  **Acceptance Criteria**:

  **If TDD (tests enabled):**
  - [ ] Test file created: `internal/ui/components/repoedit/model_test.go`
  - [ ] Test covers: Form pre-population and update flow
  - [ ] `go test ./internal/ui/components/repoedit` → PASS

  **Manual Execution Verification**:
  - [ ] Navigate to repositories view
  - [ ] Select a repository with arrow keys
  - [ ] Press Enter key
  - [ ] Verify edit form is pre-populated
  - [ ] Make changes and submit
  - [ ] Verify repository is updated in config

  **Commit**: YES
  - Message: `feat: add repository editing capability`
  - Files: `internal/ui/components/repoedit/`, `internal/app/app.go`, `internal/ui/components/repolist/model.go`
  - Pre-commit: `go test ./internal/ui/... ./internal/app/...`

- [ ] 5. Add chart search to release creation

  **What to do**:
  - Create `chartsearch` component with search input and results table
  - Add `StateChartSearch` to `internal/ui/model.go:19`
  - Add `ChartSearchMsg{}` message type
  - Integrate into release creation workflow (add "Search Charts" button after basic form, before submit)
  - **Backend Requirements**: Add `ListCharts(repo config.Repository, query string)` to `internal/helm/repo` layer that aggregates charts across repos with caching + error handling. Use async `tea.Cmd` with loading indicator for network calls.
  - Query configured Helm repositories for charts matching search
  - Allow selection of chart and auto-populate version in parent form

  **Must NOT do**:
  - Change existing manual chart entry option

  **Parallelizable**: NO (depends on 2)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/components/releaseform/model.go` - Workflow integration pattern
  - `internal/helm/repo/index.go:70` - Existing chart resolution pattern

  **API/Type References** (contracts to implement against):
  - `internal/helm/repo/index.go` - Chart index querying
  - `internal/config/config.go:Repository` - Repository configuration

  **Test References** (testing patterns to follow):
  - `internal/helm/repo/index_test.go` - Repository testing patterns

  **Acceptance Criteria**:

  **If TDD (tests enabled):**
  - [ ] Test file created: `internal/ui/components/chartsearch/model_test.go`
  - [ ] Test covers: Search functionality and result selection
  - [ ] `go test ./internal/ui/components/chartsearch` → PASS

  **Manual Execution Verification**:
  - [ ] Start add release workflow
  - [ ] Fill basic fields (ID, namespace)
  - [ ] Click "Search Charts" button
  - [ ] Type search query
  - [ ] Wait for results (loading indicator shown)
  - [ ] Select chart from results
  - [ ] Verify chart and version auto-populated in release form

  **Commit**: YES
  - Message: `feat: add chart search integrated into release creation`
  - Files: `internal/ui/components/chartsearch/`, `internal/ui/components/releaseform/model.go`, `internal/helm/repo/`
  - Pre-commit: `go test ./internal/ui/... ./internal/helm/repo/...`

- [ ] 6. Fix error handling and recovery

  **What to do**:
  - Add `prevState State` field to `internal/ui/model.go:28` to track previous state
  - Modify error display to show recovery options instead of terminating
  - Add `ErrorRecoveryMsg{}` message type
  - On any error, render error panel with `b: back`, `esc: back`, `q: quit` options
  - Handle `b`/`esc` keys to emit `ErrorRecoveryMsg`, which clears `m.err` and returns to `m.prevState`
  - Update all state transitions to set `m.prevState = m.state` before changing state
  - Remove fatal exits from TUI operations (errors set `m.err` instead)

  **Must NOT do**:
  - Change CLI error behavior (only TUI)

  **Parallelizable**: NO (depends on all)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/model.go:169-171` - Current error display pattern
  - `internal/ui/components/releaseform/model.go:113-114` - Cancel pattern
  - `internal/ui/model.go:68` - Error state handling

  **API/Type References** (contracts to implement against):
  - Bubbletea docs: Program lifecycle and error handling

  **Test References** (testing patterns to follow):
  - `internal/ui/ui_test.go` - State transition testing

  **Acceptance Criteria**:

  **Manual Execution Verification**:
  - [ ] Trigger error in any workflow (invalid release ID, etc.)
  - [ ] Verify error message displayed with "Back" option
  - [ ] Press 'b' or 'esc' key
  - [ ] Verify TUI returns to previous screen (error cleared)
  - [ ] Verify TUI continues running (no termination)

  **Commit**: YES
  - Message: `feat: add error recovery with back navigation`
  - Files: `internal/ui/model.go`, error handling in all components
  - Pre-commit: `go test ./internal/ui/...`

- [ ] 7. Integration testing and polish

  **What to do**:
  - Test complete user workflows end-to-end
  - Verify navigation consistency across all screens
  - Check keyboard shortcuts and help text
  - Polish UI styling and responsiveness

  **Must NOT do**:
  - Add new features

  **Parallelizable**: NO (depends on all)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/ui/styles/` - Styling patterns
  - `internal/ui/model.go` - Navigation patterns

  **Acceptance Criteria**:

  **Manual Execution Verification**:
  - [ ] Complete add release workflow (with chart search)
  - [ ] Complete edit release workflow
  - [ ] Complete edit repository workflow
  - [ ] Verify all error recovery works
  - [ ] Test fullscreen and window resizing
  - [ ] Verify all keyboard shortcuts work consistently

  **Commit**: YES
  - Message: `feat: integration testing and UI polish`
  - Files: Various UI component files
  - Pre-commit: `go test ./internal/ui/...`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat: enable fullscreen mode in TUI` | cmd/atlas/main.go | TUI occupies full screen |
| 2 | `feat: add release details view accessible via enter key` | internal/ui/components/releasedetails/, internal/ui/dashboard/, internal/ui/model.go | Enter navigates to details |
| 3 | `feat: add release editing capability` | internal/ui/components/releaseedit/, internal/app/app.go | Can edit existing releases |
| 4 | `feat: add repository editing capability` | internal/ui/components/repoedit/, internal/app/app.go | Can edit existing repos |
| 5 | `feat: add chart search integrated into release creation` | internal/ui/components/chartsearch/, internal/ui/components/releaseform/ | Chart search in add workflow |
| 6 | `feat: add error recovery with back navigation` | internal/ui/model.go, various components | Errors show back option |
| 7 | `feat: integration testing and UI polish` | various | All workflows work end-to-end |

---

## Success Criteria

### Verification Commands
```bash
# Test fullscreen
atlas  # Should occupy full terminal

# Test navigation (requires releases in atlas.yaml)
# - Arrow keys to select
# - Enter to view details
# - Edit workflows
# - Error recovery

# Run all UI tests
go test ./internal/ui/...
```

### Final Checklist
- [ ] Fullscreen mode enabled
- [ ] Enter key navigates to release details
- [ ] Edit functionality works for releases and repos
- [ ] Chart search integrated into release creation
- [ ] Error states provide recovery instead of termination
- [ ] All keyboard shortcuts consistent
- [ ] UI responsive and polished