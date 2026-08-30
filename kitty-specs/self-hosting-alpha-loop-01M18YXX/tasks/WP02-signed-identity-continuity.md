---
work_package_id: WP02
title: Signed Identity Continuity
dependencies:
- WP01
requirement_refs:
- FR-005
- FR-006
- FR-007
- FR-008
- FR-018
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: execution lane based on completed WP01; merge through Spec Kitty into feat/self-hosting-alpha-loop
subtasks:
- T006
- T007
- T008
- T009
- T010
phase: Phase 1 - Independent Foundations
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
authoritative_surface: identity_continuity.go
create_intent:
- identity_continuity.go
- identity_continuity_test.go
execution_mode: code_change
model: ''
owned_files:
- event.go
- identity_continuity.go
- identity_continuity_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP02 – Signed Identity Continuity

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Add repository-native, mutually signed continuity facts among distinct actors.
An authorization proves intent by an existing actor; an acceptance proves
control of the exact target key. Projection must expose pending, accepted,
retired, dependency-missing, cyclic, and competing states deterministically
without choosing a global identity winner or granting policy authority.

Completion requires exact contract behavior for `identity.authorize` and
`identity.accept`, usable identity commands, recoverable local planned
rotation, and unchanged bytes/IDs for every pre-existing event kind.

## Context & Constraints

Read the [identity contract](../contracts/identity-continuity-v0.md), [data
model](../data-model.md), [research decisions D-002/D-003](../research.md), and
the keyring behavior delivered by WP01.

An actor remains one Ed25519 public key and one single-writer history. A person,
device set, or succession lineage is only a derived projection. Timestamps,
display names, arrival order, and cross-actor sequence values cannot resolve
conflicts. Policy continues to compare exact actor fingerprints and must ignore
continuity edges for authorization.

WP04 later connects continuity validation to accepted remote projection.
Design public projection functions over supplied verified events so they are
testable now and reusable there without another interpretation.

## Branch Strategy

- **Strategy**: execution lane based on completed WP01; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Use the workspace assigned in `lanes.json`. Do not edit WP01 files unless an
unavoidable integration correction is documented in the Activity Log; prefer
the keyring interfaces already merged from the dependency.

## Subtasks & Detailed Guidance

### T006 – Write failing identity event and projection tests

- Start in `identity_continuity_test.go` with disposable actor keys and exact
  signed payloads.
- Specify authorization validation for relationship, target actor, target raw
  public key, self-target rejection, and optional body.
- Specify acceptance validation for a full authorization event ID and an exact
  signer/target match.
- Cover pending edges, accepted device edges, one accepted successor, duplicate
  acceptances, missing subjects, competing successors, longer chains, and
  cycles.
- Shuffle verified input order repeatedly and assert byte-for-byte identical
  projection values and rendering order.
- Capture existing event payload fixtures before adding fields and assert their
  event IDs do not change.

### T007 – Add canonical authorization and acceptance event validation

- Extend `Event` only with optional JSON fields required by the contract:
  relationship, target actor, and target key. Reuse `Subject` for acceptance.
- Existing kinds must continue omitting the new zero-value fields, preserving
  exact JSON bytes and event IDs.
- In `validateEventContent`, accept exactly `device` and `successor`, require a
  full valid target actor, decode an Ed25519 public key, recompute its actor,
  and reject self-targets.
- For acceptance, require a full subject ID; cross-event target/signature
  validation belongs in the projector/relationship validator where the
  authorization is available.
- Keep ordinary signature, actor-chain, sequence, and predecessor validation
  unchanged.
- Return errors that identify the violated fact without printing private data
  or trusting display names.

### T008 – Implement deterministic identity continuity projection

- Create focused public-data types in `identity_continuity.go` for edges,
  per-edge state, per-actor state, and explicit conflicts/missing dependencies.
- Index authorizations and acceptances by full immutable event ID. An
  acceptance completes only the exact authorization it references and only
  when its actor/public key exactly matches the target.
- Treat repeated matching acceptances as separately inspectable facts but one
  logical completed edge; do not multiply authority or retirement effects.
- Device edges never retire either actor. Exactly one accepted acyclic
  successor can mark a predecessor retired in this descriptive projection.
- Multiple accepted outgoing successors, contradictory target material, or
  any successor cycle produce an explicit ambiguous/conflicting state.
- Sort every projected collection by stable full IDs. Never use timestamps or
  map iteration as a semantic tie-breaker.
- Expose missing authorization subjects distinctly so WP04 can report an exact
  additional selection rather than classifying a valid partial set as corrupt.

### T009 – Implement identity commands and planned rotation

- Implement `identity show`, `list`, and `public` behavior in
  `identity_continuity.go`, consuming WP01's keyring interface.
- `public` may emit only display name, full actor, and raw public key. `list`
  must distinguish active, locally retired, and other stored records without
  exposing secrets.
- Implement authorization flags exactly as contracted; reject shortened IDs,
  malformed keys, self-targets, and unknown relationships before appending.
- Implement acceptance by resolving the exact authorization from verified
  available events and requiring the active signer to be its exact target.
- Implement planned rotation as a retryable local transaction: generate or
  reuse a distinct target record, append predecessor authorization, append
  successor acceptance using the explicit target signer, then switch active.
- If a failure occurs before both event facts are durable, leave the old actor
  active. Retry must reuse recorded keys/event IDs rather than create siblings.
- Top-level `main.go` routing is owned by WP06; keep command functions ready for
  that integration without editing the router here.

### T010 – Prove ambiguity visibility and policy-authority separation

- Add fixtures for incomplete authorization, acceptance by the wrong actor,
  replay across another authorization, competing accepted successors, two-
  actor and longer cycles, and fetched-later conflicts.
- Assert no projection result silently selects a successor in ambiguous cases.
- Use existing `evaluateProposal` behavior to prove a valid device/successor
  actor not named in the base policy cannot satisfy reviewer, runner,
  maintainer, decision, or merge requirements.
- Prove adding a continuity fact does not modify policy bytes or policy digest.
- Verify the predecessor history remains readable after local rotation and the
  successor begins its own chain at sequence one.
- Preserve all signed facts even when the derived relationship is conflicting.

## Test Strategy

Run focused red/green cycles and then all gates:

```bash
go test ./... -run 'Test.*(IdentityContinuity|IdentityAuthorize|IdentityAccept|Rotation)'
go test ./...
go test -race ./...
go vet ./...
```

Tests should construct signed facts through the same event encoding seam as
production. Use order permutations to verify deterministic projection. Avoid
mocking signatures or policy evaluation: those are the properties at risk.

## Risks & Mitigations

- **Implicit authority**: keep projection types free of policy-role fields and
  prove governance still checks explicit actor lists.
- **Ambiguous succession**: surface all edges/conflicts and refuse a winner.
- **Crash across histories**: persist rotation state and switch active only
  after both signed events are durable.
- **Payload compatibility**: optional fields must remain omitted for old kinds;
  retain golden ID assertions.
- **Missing dependency confusion**: return a typed/structured missing subject
  result rather than a generic invalid-signature error.

## Review Guidance

Review the contract field by field. Independently shuffle event order and
construct a competing successor cycle. Confirm both clones derive the same
ambiguous state, all signatures remain inspectable, and policy evaluation is
unchanged. Verify a rotation interrupted after each write retains the old
default signer until mutual completion.

Reject any implementation that copies a private key between actors, treats a
name as identity, selects a successor using time/order, or makes continuity a
shortcut into project roles.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.

