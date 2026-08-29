---
work_package_id: WP02
title: Deterministic Revision Lineage Projection
dependencies:
- WP01
requirement_refs:
- FR-004
- FR-007
- FR-008
- FR-010
- FR-011
- FR-012
planning_base_branch: chore/spec-kitty-bootstrap
merge_target_branch: chore/spec-kitty-bootstrap
branch_strategy: Planning artifacts for this mission were generated on chore/spec-kitty-bootstrap. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into chore/spec-kitty-bootstrap unless the human explicitly redirects the landing branch.
subtasks:
- T006
- T007
- T008
- T009
phase: Phase 2 - Domain Projection
history:
- at: '2026-08-29T17:18:13Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: lineage.go
create_intent:
- lineage.go
- lineage_test.go
execution_mode: code_change
model: ''
owned_files:
- lineage.go
- lineage_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP02 – Deterministic Revision Lineage Projection

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `agent_profile` from frontmatter
- **Role**: `role` from frontmatter
- **Agent/tool**: `agent` from frontmatter

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## ⚠️ IMPORTANT: Review Feedback

Inspect Spec Kitty status and any `review_ref` before work. Treat every feedback
item as required and append chronological Activity Log entries.

## Objectives & Success Criteria

Create a single derived lineage module that turns verified proposal candidates
and merge facts into deterministic roots, predecessors, successors, siblings,
ancestors, descendants, and safety states. The same event set must return the
same result regardless of input order. No API may select a winner or latest
revision.

The module must remain cycle-safe even though WP01 rejects cycles, expose exact
IDs for command diagnostics, and operate in O(events + lineage edges) for index
construction and lineage traversal.

## Context & Constraints

Read the charter and all mission design artifacts. WP01 supplies
`proposal.revise`, proposal-candidate classification, and trusted relationship
validation. This package consumes only verified `StoredEvent` values; it does
not load Git, sign events, mutate refs, or decide policy.

Use domain vocabulary exactly: predecessor, successor, sibling, lineage,
superseded, lineage closed, and merge conflict. “Latest” and “current revision”
are prohibited because there is no total order.

## Branch Strategy

- **Strategy**: Work only in the WP02 execution worktree from `lanes.json`.
- **Planning base branch**: `chore/spec-kitty-bootstrap`
- **Merge target branch**: `chore/spec-kitty-bootstrap`
- **Dependency**: WP01 must be integrated into this lane's resolved base.
- **Implementation command**: `spec-kitty agent action implement WP02 --agent <name>`

## Subtasks & Detailed Guidance

### T006 – Add red deterministic lineage-query tests

- **Purpose**: Specify the graph API through observable domain results.
- **Steps**:
  1. Create fixtures for one root, a linear revision chain, two siblings, and a
     sibling with its own successor.
  2. Assert direct predecessor/successors, sibling membership, root, complete
     lineage membership, ancestors, and descendants.
  3. Require full-ID sorting for every returned set; timestamps must not alter
     semantics.
  4. Run every case across original, reversed, and deterministically shuffled
     event input.
  5. Cover unknown proposal IDs with explicit errors rather than empty results
     that could mask caller mistakes.
- **Files**: `lineage_test.go`.
- **Validation**: Tests fail because the lineage module does not exist.

### T007 – Implement the derived lineage index and traversal API

- **Purpose**: Concentrate graph knowledge behind a narrow deep-module surface.
- **Steps**:
  1. Index proposal candidates by ID and revision predecessor edges.
  2. Store direct successors in full-ID sorted order after construction.
  3. Provide query methods needed by later CLI/governance code without exposing
     mutable internal maps.
  4. Use iterative or guarded recursive traversal with visited sets.
  5. Return exact IDs/StoredEvents consistently so callers do not re-resolve or
     re-sort graph facts.
  6. Keep construction pure: no Git commands, output, global variables, or ref
     mutation.
- **Files**: `lineage.go`.
- **Validation**: T006 core topology tests pass.

### T008 – Derive terminal and conflict states

- **Purpose**: Give all user and governance surfaces one safety interpretation.
- **Steps**:
  1. Index valid `proposal.merged` events by exact proposal subject.
  2. Derive `superseded` only for an unmerged candidate with direct successors.
  3. Derive `lineage closed` when another member of the same lineage is merged.
  4. Derive `merge conflict` when two or more distinct lineage candidates have
     merge facts; return every merged candidate ID in stable order.
  5. Preserve a candidate's own merged state even if successors or competing
     merges exist.
  6. Add combinations for accepted-but-unmerged history; decisions do not
     affect lineage topology.
- **Files**: `lineage.go`, `lineage_test.go`.
- **Validation**: State definitions match `data-model.md` and never delete facts.

### T009 – Prove convergence and representative performance

- **Purpose**: Enforce NFR-001 and provide evidence toward NFR-003.
- **Steps**:
  1. Generate a deterministic fixture approximating 10,000 events, 1,000
     proposals, and 100 revision links without sleeping or network access.
  2. Compare canonical query results across multiple permutations.
  3. Add a benchmark for index build plus representative root/sibling/state
     queries; if a wall-clock assertion is used, isolate it from `-race` and
     avoid a flaky threshold.
  4. Confirm no quadratic full-event scans occur per proposal query.
  5. Run focused benchmark output for review evidence and the full test suite.
- **Files**: `lineage_test.go`.
- **Validation**: Deterministic results are exact; representative operations are
  comfortably below the two-second project goal on the development machine.

## Test Strategy

Tests should call the lineage module directly with verified-event fixtures.
Favor exact expected ID slices and state structs over matching formatted CLI
text. Include hostile cycle defense as a secondary guard even though the store
rejects cycles first.

```sh
gofmt -w lineage.go lineage_test.go
go test -race ./...
go test -run Lineage -bench Lineage -benchmem
go vet ./...
go build ./...
git diff --check
```

## Risks & Mitigations

- **Hidden ordering**: sort by full ID only and permute input in tests.
- **Shallow abstraction**: callers must not need internal maps or repeat walks;
  add narrow query methods for actual later use.
- **Quadratic status listing**: build merge and adjacency indexes once.
- **Conflicting facts**: return all merge IDs/members; never choose one.

## Review Guidance

Review method names against the ubiquitous language and verify no method exposes
“latest.” Inspect complexity of list/status usage, stable ordering, and cycle
guards. Require exact convergence tests and ensure the module performs no Git or
CLI I/O.

## Activity Log

- 2026-08-29T17:18:13Z – system – Prompt created.

Append later entries chronologically. Use Spec Kitty commands for WP and subtask
status rather than editing frontmatter or task reference rows.
