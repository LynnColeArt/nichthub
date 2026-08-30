---
work_package_id: WP04
title: Selected Quarantined Replication
dependencies:
- WP02
- WP03
requirement_refs:
- FR-009
- FR-010
- FR-011
- FR-012
- FR-013
- FR-017
- FR-018
- FR-019
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T016
- T017
- T018
- T019
- T020
- T021
- T022
phase: Phase 2 - Hostile-Input Transport
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
authoritative_surface: replication.go
create_intent:
- replication.go
- quarantine.go
- replication_test.go
execution_mode: code_change
model: ''
owned_files:
- commands.go
- store.go
- git.go
- replication.go
- quarantine.go
- replication_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP04 – Selected Quarantined Replication

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Replace direct wildcard fetch into accepted collaboration refs with an explicit
local selection and validate-before-promotion transaction. Remote refs and Git
objects are hostile input. They become visible to `collectEvents` only after
exact selection, positive-budget measurement, structural and relationship
validation, object transfer, and atomic accepted-ref update.

The package succeeds when one synchronization can accept independently valid
selected histories while isolating invalid, over-budget, dependency-missing,
and unselected histories. Existing `nh sync [REMOTE]` remains usable through a
documented bounded compatibility-all mode, but no path performs direct
unvalidated wildcard import.

## Context & Constraints

Read the [selected replication contract](../contracts/selected-replication-v0.md),
[research decisions D-004 through D-006](../research.md), and the transaction
model in [data-model.md](../data-model.md). Consume WP02's identity relationship
validation and WP03's exact-ID conventions; do not infer policy authority from
selection or continuity.

Selection is private, local, per remote, and full-ID only. The hard budget
promise begins before promotion/retention, not before network transfer:
portable Git does not provide a hard pre-download pack byte ceiling. Direct
refspecs must ensure unselected histories are not requested.

This is the only accepted-remote mutation path. A separate generated bare Git
repository is a required security boundary, not an optional implementation
detail. Do not use Docker, a service, a database, or a new Go module.

## Branch Strategy

- **Strategy**: execution lane based on completed WP02 and WP03; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Use the lane workspace from `lanes.json`. The package intentionally owns the
existing sync/storage/Git seams so replication has one coherent mutation
boundary. Do not edit `main.go`; WP06 adds top-level routing after integration.

## Subtasks & Detailed Guidance

### T016 – Write selected-sync, isolation, and budget boundary tests

- Begin in `replication_test.go` with multiple real temporary repositories and
  one ordinary bare remote. Avoid transport mocks.
- Publish at least two valid actor histories, one malformed actor ref/history,
  one valid candidate ref, one conflicting/mismatched candidate ref, and an
  unselected history.
- Specify exact saved selection/show behavior, full-ID rejection, duplicate
  rejection, `--all` exclusivity, and positive limits.
- Assert newly fetched quarantine state is absent from the main repository's
  accepted remote-tracking refs until promotion.
- Exercise event count, reachable object count, maximum object bytes, maximum
  attachment bytes, and total reachable bytes one below, exactly at, and one
  above each configured limit.
- In one transaction combine valid and invalid selections; prove valid
  independent selections remain usable and failed refs preserve their old
  accepted values.
- Capture full failing selection and dependency IDs in diagnostics without
  depending on temporary quarantine paths.

### T017 – Implement private per-remote replication selections

- Add a versioned local selection model in `replication.go` stored beneath the
  resolved `.git/nh` directory. Use a safe encoded remote name or validated
  filename mapping; prevent path traversal and alias collisions.
- Persist sorted, deduplicated full actor fingerprints and full proposal event
  IDs plus explicit all-ref mode and every positive budget.
- Require at least one selector or explicit `--all`. Make `--all` mutually
  exclusive with actor/proposal values.
- Reject shortened IDs even if they are currently unique. Trust configuration
  must never change meaning when more refs arrive.
- Write atomically with private owner-only permissions and never track or
  publish the selection file.
- Provide deterministic `select` and `show` command functions, defaulting the
  remote to `origin`, without conflating selection with trusted policy actors.
- Define conservative positive compatibility defaults for repositories with
  no explicit saved selection; once saved, the selection is authoritative.

### T018 – Fetch exact selected refs into generated bare quarantine

- Add `quarantine.go` to create a fresh bare Git repository for one transaction
  outside the main accepted ref namespace. Resolve all paths from Git rather
  than relying on the process working directory.
- For explicit selectors, construct only exact actor and proposal source
  refspecs. Confirm advertised refs resolve to the expected object type and
  target namespace; a missing selected ref is an exact per-selection failure.
- Compatibility-all may inspect advertisements and request the documented
  actor/proposal namespaces, but still fetches into quarantine and applies all
  validation/budgets.
- Never add quarantine alternates or refs that make its objects implicitly
  visible to ordinary main-repository revision walking before promotion.
- Record transaction state and selection outcomes locally for recovery or
  diagnostics, excluding credentials and host-private paths from user-facing
  output and signed facts.
- Ensure cleanup is safe and bounded to the resolved generated transaction
  directory. A failure to clean temporary state must not promote refs.

### T019 – Measure and enforce all promotion budgets

- Enumerate each selected ref's reachable object graph in quarantine using Git
  plumbing with explicit object IDs and types. Avoid parsing localized prose.
- Measure unique reachable object count, total uncompressed object bytes, and
  largest individual object bytes deterministically.
- Parse actor event commits to count events and inspect supported attachments;
  enforce the maximum event attachment size before relationship projection.
- Specify whether shared objects count per selection or transaction in one
  canonical function consistent with the contract's per-selected-ref wording;
  tests must make the behavior unambiguous.
- Reject non-positive budgets at configuration time and stop validation work
  for a selection as soon as a hard measured boundary is known exceeded.
- Report budget name, configured value, measured value, and full selection ID.
- Do not claim these checks cap bytes already transferred into quarantine.

### T020 – Validate selections independently and classify dependencies

- Validate Git object types, event tree shape, exact signature, actor chain,
  supported attachments, proposal code-ref/head binding, and all existing
  event relationships before marking a selection promotable.
- Refactor `store.go` so validation can operate over explicit accepted facts
  plus a transaction's candidate facts, rather than only one global ref walk.
- Validate actor chains per actor first. An invalid actor must not prevent an
  unrelated selected actor from reaching promotion.
- Run WP02 identity-continuity validation over accepted plus selected facts;
  preserve ambiguous valid relationships as visible projection, while
  malformed claims remain invalid.
- Distinguish absent referenced facts from malformed available facts. Return
  the referencing event kind/ID, missing full ID, owning selection, and exact
  additional actor/proposal selection when derivable.
- Determine dependency closure without substituting prefixes, related actors,
  titles, successors, or delivery order.
- A selection depending on a failed selection stays unpromoted, but unrelated
  valid selections can still form an atomic promotable set.

### T021 – Promote objects and accepted refs atomically

- Before ref mutation, copy/import every object reachable from the promotable
  set into the main object database and verify each required object can be read
  there by exact ID.
- Build accepted destinations exactly under
  `refs/nh/remotes/<remote>/actors/<actor>` and
  `refs/nh/remotes/<remote>/proposals/<candidate-hash>`.
- Use one `git update-ref --stdin` transaction (or equivalent atomic Git
  plumbing) with expected old values for the consistent promotable update set.
- Never advance a rejected selection's accepted ref. When a ref previously
  existed, preserve it exactly after failure.
- If object copy or ref transaction fails, return non-zero and leave all
  accepted refs at their pre-transaction values.
- After success, make `collectEvents` consume only local actor refs and these
  accepted remote refs. Quarantine refs must not enter normal projection.
- Prove promoted refs never point at unavailable objects even after quarantine
  cleanup.

### T022 – Route sync and replication commands through the boundary

- Replace `cmdSync`'s direct wildcard fetch with the replication transaction.
  Accept `nh sync [REMOTE] [--recover-shallow]`; WP05 supplies recovery logic,
  so preserve a clear hook and reject unsupported recovery until available.
- Add `cmdReplication` with `select` and `show` subcommands and exact usage from
  the contract. WP06 will add the top-level router entry.
- Keep publication separate from import: publish only the active local actor
  and locally authored proposal refs allowed by existing rules.
- Do not let an invalid remote import prevent safe publication reporting from
  identifying which phase failed. Never describe primary-branch publication
  as part of `nh sync`.
- After promotion, collect accepted events and report verified count plus
  explicit per-selection successes/failures.
- Preserve identity-free read behavior: absence of a local private key is not a
  sync error and publishes nothing.
- Re-run all existing sync/proposal-ref tests and update them only for the new
  validate-before-promotion contract.

## Test Strategy

Required commands:

```bash
go test ./... -run 'Test.*(Replication|Selection|Quarantine|Budget|Sync)'
go test ./...
go test -race ./...
go vet ./...
```

Tests must use local-path remotes and real Git objects. Test each budget at the
three exact boundaries and inspect accepted refs/object availability after
every success and failure. Avoid asserting internal quarantine directory names
or Git pack layout.

## Risks & Mitigations

- **Invalid-state poisoning**: keep fetched roots in a separate repository and
  validate actors independently.
- **Dangling accepted refs**: copy and verify all objects before one atomic ref
  transaction.
- **Unbounded selection**: positive defaults and limits are mandatory even for
  compatibility-all.
- **Transport overclaim**: document that standard Git may transfer a selected
  pack before post-fetch measurement rejects it.
- **Trust conflation**: selection admits facts for inspection; only policy
  makes verified claims qualify.
- **Path/remote injection**: validate ref components and avoid constructing
  shell commands; continue using `exec.Command` argument vectors.

## Review Guidance

Review through a disposable bare remote containing one good actor, one bad
actor, one oversized graph, one missing dependency, and one unselected actor.
Inspect refs before and after sync. The good independent ref must promote; all
others must remain absent or at their previous accepted values.

Verify no direct fetch target writes beneath `refs/nh/remotes`, every configured
budget has one-below/exact/one-above coverage, and `collectEvents` ignores
quarantine/unaccepted roots. Reject any global validation failure that disables
an unrelated accepted actor or any promotion that is not atomic.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.

