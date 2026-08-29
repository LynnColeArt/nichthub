---
work_package_id: WP01
title: Signed Revision Event Foundation
dependencies: []
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-010
- FR-013
- FR-015
planning_base_branch: chore/spec-kitty-bootstrap
merge_target_branch: chore/spec-kitty-bootstrap
branch_strategy: Planning artifacts for this mission were generated on chore/spec-kitty-bootstrap. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into chore/spec-kitty-bootstrap unless the human explicitly redirects the landing branch.
subtasks:
- T001
- T002
- T003
- T004
- T005
phase: Phase 1 - Protocol Foundation
history:
- at: '2026-08-29T17:18:13Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: ''
authoritative_surface: event.go
create_intent:
- revision_event_test.go
execution_mode: code_change
model: ''
owned_files:
- event.go
- store.go
- event_test.go
- revision_event_test.go
role: ''
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP01 – Signed Revision Event Foundation

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `agent_profile` from frontmatter
- **Role**: `role` from frontmatter
- **Agent/tool**: `agent` from frontmatter

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## ⚠️ IMPORTANT: Review Feedback

Before implementation, inspect the WP status and `review_ref` through Spec Kitty.
Address every recorded feedback item and append progress to the Activity Log in
chronological order.

## Objectives & Success Criteria

Introduce `proposal.revise` as a signed `nh/0` event without weakening existing
event or actor-chain validation. A valid revision must name one exact proposal
predecessor, carry a distinct exact base/head, and be signed by the predecessor
author. Missing predecessors, wrong kinds, unauthorized signers, self-links,
cycles, and malformed commit IDs fail closed with actionable errors.

Completion requires red-first tests, unchanged behavior for legacy event kinds,
and no timestamp/delivery-order rule. A received revision is not invalidated
merely because a merge fact is also present; merge knowledge is a local command
and governance concern handled by later packages.

## Context & Constraints

Read `.kittify/charter/charter.md`, the mission `spec.md`, `plan.md`,
`data-model.md`, `research.md`, and `contracts/proposal-revision-v0.md` before
editing. Apply NH-001 through NH-005: signed immutable Git-native state, exact
evidence, hostile-input validation, and synchronized protocol documentation.

The repository uses one append-only actor ref per identity. Do not permit forked
actor sequences or add multi-device writer semantics. Do not add dependencies,
services, Docker, a mutable latest ref, or a new ref namespace.

## Branch Strategy

- **Strategy**: Use the Spec Kitty execution worktree allocated for WP01 in
  `lanes.json`; do not implement in the planning checkout.
- **Planning base branch**: `chore/spec-kitty-bootstrap`
- **Merge target branch**: `chore/spec-kitty-bootstrap`
- **Implementation command**: `spec-kitty agent action implement WP01 --agent <name>`

## Subtasks & Detailed Guidance

### T001 – Add red wire-shape tests for `proposal.revise`

- **Purpose**: Lock the additive event contract before implementation.
- **Steps**:
  1. Extend `event_test.go` or add focused cases in `revision_event_test.go` that
     build and sign a revision with `subject`, `base`, `head`, and optional body.
  2. Assert round-trip verification preserves exact fields and ID semantics.
  3. Add table cases for missing/invalid subject, missing or equal base/head,
     and unexpected proposal title semantics according to the contract.
  4. Run the focused tests and record the expected red failure before code.
- **Files**: `event_test.go`, `revision_event_test.go`.
- **Validation**: Failures must be caused by unsupported/missing revision
  behavior, not broken fixtures.

### T002 – Implement proposal-candidate classification and content validation

- **Purpose**: Give every command one narrow definition of a proposal candidate.
- **Steps**:
  1. Add `proposal.revise` to `validateEventContent` with exact subject and
     distinct valid Git OID checks.
  2. Add a small helper such as `isProposalKind` covering only
     `proposal.open` and `proposal.revise`.
  3. Keep `proposal.open` validation byte-for-byte compatible and retain the
     default unknown-kind rejection.
  4. Avoid adding a new Event field: `subject` is the signed predecessor link.
- **Files**: `event.go`.
- **Validation**: T001 becomes green; existing signed event tests remain green.

### T003 – Add red cross-event authorization and graph-safety tests

- **Purpose**: Specify the trust boundary over otherwise well-shaped events.
- **Steps**:
  1. Build stored event fixtures for open → revision and revision → revision.
  2. Add negative cases for missing predecessor, issue predecessor, a revision
     signed by another actor, direct self-link, and multi-node cycle.
  3. Add sibling cases with two same-author successors naming one predecessor.
  4. Shuffle input event order and assert identical pass/fail outcomes.
  5. Include a revision plus later merge fact and assert both valid facts remain
     accepted for projection.
- **Files**: `revision_event_test.go`.
- **Validation**: Tests fail specifically because relationship validation has
  not yet generalized proposals or checked the revision graph.

### T004 – Validate revision relationships

- **Purpose**: Reject untrusted lineage before it reaches any projection.
- **Steps**:
  1. Build the full event-ID map before iterating relationships.
  2. Accept a revision predecessor only when it is an available proposal
     candidate and has the same actor.
  3. Generalize review, run request, decision, and merge relationship checks to
     accept either candidate kind while preserving exact head/policy checks.
  4. Validate the entire predecessor graph with explicit visited/active state;
     errors must include the involved short event ID.
  5. Do not use timestamps, slice order, or mutable refs as graph authority.
- **Files**: `store.go`.
- **Validation**: All T003 cases turn green in multiple input orders.

### T005 – Prove fail-closed and legacy compatibility behavior

- **Purpose**: Prevent the extension from broadening accepted hostile input.
- **Steps**:
  1. Retain unknown-kind rejection and signature/public-key/actor checks.
  2. Add regression cases covering existing issue, proposal, review, run,
     decision, and merge relationships without revisions.
  3. Assert invalid revisions cannot become valid subjects for downstream
     review, run, decision, or merge events.
  4. Run `go test -race ./...` and `git diff --check` before handoff.
- **Files**: `event_test.go`, `revision_event_test.go`.
- **Validation**: New hostile cases fail closed; all pre-mission tests pass.

## Test Strategy

Use deterministic in-memory event fixtures where relationship behavior is the
unit under test and existing temporary Git repositories where stored-event load
or signature behavior matters. Do not weaken production validation to make
fixtures easier. Run focused red/green commands first, then:

```sh
gofmt -w event.go store.go event_test.go revision_event_test.go
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

## Risks & Mitigations

- **Retroactive ordering**: never reject a received revision based solely on a
  merge fact elsewhere in the set; no authoritative cross-actor order exists.
- **Cycle validation duplication**: keep graph integrity here and derived query
  behavior in WP02; expose only the small helper needed downstream.
- **Compatibility drift**: table-test every pre-existing relationship kind.
- **Hostile diagnostics**: sanitize names/body and print stable IDs, not raw
  untrusted payloads.

## Review Guidance

Reject the package if it relies on timestamps, changes actor-chain fork rules,
adds a dependency/ref namespace, permits non-author successors, or treats a
revision as valid without its predecessor. Verify the later-merge test preserves
both facts and that all legacy tests run unchanged.

## Activity Log

- 2026-08-29T17:18:13Z – system – Prompt created.

Append new entries at the end as
`YYYY-MM-DDTHH:MM:SSZ – agent_id – action`. Status changes use
`spec-kitty agent tasks move-task`; subtask completion uses
`spec-kitty agent tasks mark-status`.
