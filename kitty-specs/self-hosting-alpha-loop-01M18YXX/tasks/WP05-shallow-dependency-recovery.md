---
work_package_id: WP05
title: Shallow Dependency Recovery
dependencies:
- WP04
requirement_refs:
- FR-013
- FR-014
- FR-015
- FR-017
- FR-018
- FR-019
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T023
- T024
- T025
- T026
- T027
phase: Phase 3 - Exact History Recovery
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: shallow.go
create_intent:
- shallow.go
- shallow_test.go
execution_mode: code_change
model: ''
owned_files:
- shallow.go
- shallow_test.go
- proposal.go
- ci.go
- governance.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP05 – Shallow Dependency Recovery

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Make shallow Git history an explicit integrity boundary rather than a silent
source of incomplete trust decisions. A shallow repository is not inherently
invalid: operations proceed when every exact dependency is present. When a
required object/fact is absent, the operation stops, names the exact gap, and
offers only an explicit selected recovery through WP04's quarantine/budget
path.

The package succeeds when policy inspection, candidate evaluation, CI request
and result verification, decisions, and merges cannot advance across an
unresolved shallow boundary, while explicit bounded recovery retrieves only
the selected supplying refs and restarts verification from clean accepted
state.

## Context & Constraints

Read [research decision D-007](../research.md), the [selected replication
contract](../contracts/selected-replication-v0.md), and the Shallow Dependency
Gap model in [data-model.md](../data-model.md). WP04 is the only fetch/promotion
implementation; this package must call it rather than create another Git fetch
path.

Never run global `git fetch --unshallow`, silently deepen the primary branch,
or add an actor/candidate to local selection. Missing dependency is distinct
from invalid data. Every trust-sensitive operation restarts after recovery;
no partial policy/evidence result may be reused.

## Branch Strategy

- **Strategy**: execution lane based on completed WP04; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Work in the allocated lane workspace. Keep the resolver in `shallow.go` and
make small guard integrations in the declared domain files. Do not bypass the
replication transaction for convenience.

## Subtasks & Detailed Guidance

### T023 – Write depth-limited fail-closed and recovery tests

- Begin in `shallow_test.go` with a real repository whose branch history is
  cloned using `--depth 1` over a transport that honors depth (use `file://`
  where necessary).
- Construct independent gaps for actor predecessor, candidate event/code ref,
  base commit, policy blob, pipeline definition, run request, decision, and
  merge ancestry.
- First prove a shallow clone with all exact custom-ref objects succeeds; do
  not reject solely because `.git/shallow` exists.
- For each actual gap, assert the trust-sensitive command exits non-zero before
  appending an event, changing an accepted ref, creating a merge commit, or
  otherwise advancing state.
- Assert diagnostics contain full missing IDs, kind/boundary, selected remote,
  supplying ref or selection action when known, and a bounded recovery command.
- Capture selected IDs and refs before/after recovery to prove no unrelated
  actor/candidate entered accepted projection.

### T024 – Implement exact shallow dependency-gap classification

- Add `shallow.go` as a focused resolver over exact required Git objects and
  event facts. Detect repository shallow status through Git plumbing, not by
  guessing path layout.
- Define a structured gap with operation, kind, exact missing ID, owning
  actor/candidate when known, selected remote, required supplying ref, and
  recovery guidance.
- Distinguish: object absent in a shallow repository, object absent despite a
  complete repository, selected fact missing, and available object/fact that
  fails validation.
- Preserve the original lower-level error as diagnostic context without
  collapsing every Git failure into a shallow gap.
- Resolve dependencies by immutable IDs and signed relationships only. Never
  search titles, display names, prefixes, timestamps, or successors as
  substitutes.
- Render deterministic human-readable errors suitable for black-box tests;
  exclude temporary paths and credentials.

### T025 – Guard every trust-sensitive verification boundary

- Inventory exact reads in `proposal.go`, `ci.go`, and `governance.go` for
  candidate base/head, policy, pipeline, request/result, decision evidence,
  merge head/result, and ancestry.
- Before accepting evidence or appending a trust-sensitive event, resolve the
  full dependency set and return a gap if any required object/fact is absent.
- Preserve existing invalid-data errors when the required object is present
  but malformed, mismatched, or unauthorized.
- Ensure policy always comes from the signed candidate base. Missing base
  policy cannot fall back to `HEAD`, working tree, head policy, or a similarly
  named revision.
- Ensure proposal code availability and ancestry checks do not treat shallow
  omission as a legitimate mismatch or conflict resolution.
- Keep read-only inspection successful when it can prove all requested facts
  from objects already present.

### T026 – Route explicit recovery through quarantine

- Implement a narrow recovery request that takes the existing local selection,
  missing full IDs, selected remote, and explicit operator consent from
  `--recover-shallow`.
- Resolve only supplying refs already selected or instruct the operator to add
  the exact missing selection. The recovery function cannot mutate selection.
- Invoke WP04's quarantine transaction with the same positive budgets and
  validate-before-promotion rules as ordinary synchronization.
- If recovery promotes required objects/refs successfully, discard every
  partial domain result and rerun the original operation from accepted state.
- If any budget, validation, relationship, object copy, or ref transaction
  fails, return non-zero with the previous accepted state intact.
- Keep recovery idempotent; an already-present exact dependency is a no-op
  followed by fresh verification.

### T027 – Prove bounded retry without trust expansion

- Test recovery with selected and unselected supplying actors/candidates. The
  selected case may recover; the unselected case must provide the exact
  selection action and remain blocked.
- Test each applicable budget during recovery and confirm over-budget objects
  do not become accepted roots.
- Inspect Git commands or observable refs to prove no `--unshallow`, wildcard
  broadening, implicit branch deepening, or unrelated selection occurs.
- Interrupt recovery before fetch completion, after measurement, after object
  copy, and before ref transaction; prior accepted state must remain valid.
- Retry successful recovery and confirm identical accepted refs and projection.
- Run existing conflict-revision and governance tests to ensure shallow guards
  do not change complete-repository semantics.

## Test Strategy

Required commands:

```bash
go test ./... -run 'Test.*(Shallow|DependencyGap|Recover)'
go test ./...
go test -race ./...
go vet ./...
```

Use actual depth-limited clones and local transports. Assert observable events,
refs, commits, and diagnostics rather than internal Git command ordering,
except for the explicit prohibition against global unshallow/trust broadening.

## Risks & Mitigations

- **False shallow classification**: first distinguish Git/object absence from
  malformed present data and unrelated command failure.
- **Implicit trust expansion**: require a saved exact selection and never edit
  it during recovery.
- **Partial decision reuse**: restart the whole operation after promotion.
- **Second fetch implementation**: call the replication transaction; keep
  `shallow.go` limited to classification and orchestration.
- **Overblocking**: allow shallow repositories when all exact dependencies are
  already available through branch/custom refs.

## Review Guidance

Reviewers should execute the same operation in a complete clone, a depth-one
clone with custom refs supplying everything, and a depth-one clone missing one
exact dependency. Only the last should block. Then recover it under a tiny
selection and inspect that no unrelated refs appeared.

Reject any fallback to `HEAD`, global unshallow, silent selection mutation,
generic "fetch more history" advice without the missing ID/ref, or operation
that emits a signed fact before the gap is resolved.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.
- 2026-08-30T18:32:20Z – codex – Accepted incoming handoff from WP03: proposal.go may already contain the narrow policy-amendment diagnostic call before proposal mutation. Preserve or deliberately supersede it while implementing shallow recovery; do not silently drop the diagnostic behavior.
