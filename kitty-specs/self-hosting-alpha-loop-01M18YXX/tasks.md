# Work Packages: Operational Self-Hosting Alpha

**Mission**: `self-hosting-alpha-loop-01M18YXX`  
**Planning branch**: `feat/self-hosting-alpha-loop`  
**Merge target**: `feat/self-hosting-alpha-loop`  
**Generated**: 2026-08-30T17:26:50Z

## Execution Summary

This mission is divided into seven cohesive work packages. The critical path is
identity keyring -> signed continuity -> selected replication -> shallow
recovery -> operational acceptance -> public proof. Policy inspection is an
independent foundation package and joins the critical path at operational
acceptance.

Subtask completion is event-sourced. The rows below are references, not
checkboxes; implementations record completion with `spec-kitty agent tasks
mark-status`.

## Dependency Graph

```text
WP01 Canonical Identity Keyring ──> WP02 Signed Identity Continuity ──┐
                                                                     ├─> WP04 Selected Replication
WP03 Policy Amendment Inspection ────────────────────────────────────┘

WP04 ──> WP05 Shallow Recovery ──> WP06 Operational Acceptance ──> WP07 Living/Public Proof
WP03 ────────────────────────────────────────────────┘
```

## Subtask Index

| ID | Description | Work package | Parallel |
| --- | --- | --- | --- |
| T001 | Write failing migration and signer-compatibility tests | WP01 | No |
| T002 | Implement private keyring records and atomic local storage | WP01 | No |
| T003 | Migrate the legacy identity file through the active-identity facade | WP01 | No |
| T004 | Implement active actor switching and recoverable rotation state | WP01 | No |
| T005 | Prove permissions, idempotency, secret isolation, and compatibility | WP01 | No |
| T006 | Write failing identity event and projection tests | WP02 | No |
| T007 | Add canonical identity authorization and acceptance event validation | WP02 | No |
| T008 | Implement deterministic identity continuity projection | WP02 | No |
| T009 | Implement identity inspection, authorization, acceptance, and rotation commands | WP02 | No |
| T010 | Prove ambiguity visibility and strict policy-authority separation | WP02 | No |
| T011 | Write black-box policy show/check contract tests | WP03 | Yes |
| T012 | Unify exact policy loading from commits and draft files | WP03 | Yes |
| T013 | Implement deterministic structural policy comparison | WP03 | Yes |
| T014 | Implement policy show/check command handlers and proposal diagnostics | WP03 | Yes |
| T015 | Prove lockout rejection and exact base-policy authorization | WP03 | Yes |
| T016 | Write selected-sync, isolation, and budget boundary tests | WP04 | No |
| T017 | Implement private per-remote replication selections | WP04 | No |
| T018 | Fetch exact selected refs into generated bare quarantine | WP04 | No |
| T019 | Measure and enforce event, attachment, object, and byte budgets | WP04 | No |
| T020 | Validate selections independently and classify missing dependencies | WP04 | No |
| T021 | Promote objects and accepted refs atomically | WP04 | No |
| T022 | Route sync and replication commands through the transaction boundary | WP04 | No |
| T023 | Write depth-limited fail-closed and recovery tests | WP05 | No |
| T024 | Implement exact shallow dependency-gap classification | WP05 | No |
| T025 | Guard policy, candidate, CI, decision, and merge verification | WP05 | No |
| T026 | Route explicit selected recovery through quarantine | WP05 | No |
| T027 | Prove bounded retry without global unshallow or trust expansion | WP05 | No |
| T028 | Build a subprocess-level multi-repository acceptance harness | WP06 | No |
| T029 | Exercise policy amendment under the old policy | WP06 | No |
| T030 | Exercise distinct actors and planned rotation | WP06 | No |
| T031 | Exercise selected replication, hostile input, and every budget boundary | WP06 | No |
| T032 | Exercise shallow detection and bounded recovery | WP06 | No |
| T033 | Exercise role-distinct governance, compatibility, and three-run repeatability | WP06 | No |
| T034 | Update canonical protocol and governance documentation | WP07 | No |
| T035 | Document identity continuity, keyring safety, and rotation limits | WP07 | No |
| T036 | Document selected replication, quarantine, budgets, and threat boundaries | WP07 | No |
| T037 | Update README, host compatibility, and public operator flow | WP07 | No |
| T038 | Record deterministic offline verification evidence | WP07 | No |
| T039 | Perform and record the staged public self-hosting proof | WP07 | No |

## Phase 1 — Independent Foundations

### WP01 — Canonical Identity Keyring

- **Prompt**: `tasks/WP01-canonical-identity-keyring.md`
- **Priority**: P1
- **Dependencies**: none
- **Goal**: Replace the single private identity file with one canonical,
  permission-safe keyring while preserving every existing actor and signature.
- **Independent test**: Initialize a legacy repository, append an event, migrate
  it, and append again with the same actor/key bytes; interrupt and retry every
  local write boundary without changing the active signer prematurely.
- **Estimated prompt size**: ~260 lines

Included subtasks:

T001 Write failing migration and signer-compatibility tests (WP01)
T002 Implement private keyring records and atomic local storage (WP01)
T003 Migrate the legacy identity file through the active-identity facade (WP01)
T004 Implement active actor switching and recoverable rotation state (WP01)
T005 Prove permissions, idempotency, secret isolation, and compatibility (WP01)

Implementation sketch: begin with observable migration and permission tests;
introduce one private persistence module; make `loadIdentity` and
`createIdentity` delegate through it; retain the legacy file until the new
record and active pointer are durable; keep retired keys inspectable but never
implicitly active.

Risks: actor fingerprint drift, partial migration, permissive modes, private
key leakage, and active-pointer changes before rotation completion.

### WP02 — Signed Identity Continuity

- **Prompt**: `tasks/WP02-signed-identity-continuity.md`
- **Priority**: P1
- **Dependencies**: WP01
- **Goal**: Add mutually signed device/successor facts and a deterministic,
  ambiguity-preserving identity projection that never grants project roles.
- **Independent test**: Two distinct actors authorize and accept device and
  successor relationships; incomplete, replayed, cyclic, and competing claims
  yield the same visible projection in every event order.
- **Estimated prompt size**: ~300 lines

Included subtasks:

T006 Write failing identity event and projection tests (WP02)
T007 Add canonical identity authorization and acceptance event validation (WP02)
T008 Implement deterministic identity continuity projection (WP02)
T009 Implement identity inspection, authorization, acceptance, and rotation commands (WP02)
T010 Prove ambiguity visibility and strict policy-authority separation (WP02)

Implementation sketch: extend the event envelope without changing existing
payloads, validate target keys and exact subjects, project relationships from
verified facts only, then connect the keyring's recoverable rotation state to
the two event appends and final active-pointer switch.

Risks: hidden tie-breaking, implicit trust inheritance, duplicate acceptance,
cycles, and a crash between actor-chain writes.

### WP03 — Policy Amendment Inspection

- **Prompt**: `tasks/WP03-policy-amendment-inspection.md`
- **Priority**: P1
- **Dependencies**: none
- **Goal**: Expose exact policy provenance, validation, and deterministic
  before/after changes while retaining base-policy-only candidate authority.
- **Independent test**: Compare a committed base with both a committed head and
  a working-tree draft; observe exact digests and sorted changes; reject an
  unsatisfiable proposal before any candidate ref or event is created.
- **Estimated prompt size**: ~260 lines

Included subtasks:

T011 Write black-box policy show/check contract tests (WP03)
T012 Unify exact policy loading from commits and draft files (WP03)
T013 Implement deterministic structural policy comparison (WP03)
T014 Implement policy show/check command handlers and proposal diagnostics (WP03)
T015 Prove lockout rejection and exact base-policy authorization (WP03)

Implementation sketch: keep `policy.go` authoritative, add source-aware
loading and comparison values, render stable full-ID output in a focused
command module, and strengthen proposal diagnostics without introducing a
policy event or mutating a draft.

Risks: validator duplication, unstable map ordering, misleading
self-authorization language, or changing existing evaluation behavior.

## Phase 2 — Hostile-Input Transport

### WP04 — Selected Quarantined Replication

- **Prompt**: `tasks/WP04-selected-quarantined-replication.md`
- **Priority**: P1
- **Dependencies**: WP02, WP03
- **Goal**: Replace direct wildcard import with persisted exact selection,
  separate quarantine, positive budgets, per-selection validation, failure
  isolation, and atomic accepted-ref promotion.
- **Independent test**: Synchronize valid, invalid, over-budget, missing-
  dependency, and unselected refs from one real bare remote; only independent
  valid selections become accepted projection roots.
- **Estimated prompt size**: ~390 lines

Included subtasks:

T016 Write selected-sync, isolation, and budget boundary tests (WP04)
T017 Implement private per-remote replication selections (WP04)
T018 Fetch exact selected refs into generated bare quarantine (WP04)
T019 Measure and enforce event, attachment, object, and byte budgets (WP04)
T020 Validate selections independently and classify missing dependencies (WP04)
T021 Promote objects and accepted refs atomically (WP04)
T022 Route sync and replication commands through the transaction boundary (WP04)

Implementation sketch: persist full selectors and limits under `.git/nh`,
construct exact refspecs, fetch to a generated bare repository, measure before
projection, validate each actor/candidate against accepted plus selected facts,
copy only promotable objects, and update accepted remote refs with one Git ref
transaction. Existing sync syntax remains, but never directly exposes fetched
wildcard state.

Risks: temporary pack exhaustion, accidental unselected fetch, invalid actor
poisoning, dangling promoted refs, partial ref updates, and compatibility drift.

## Phase 3 — Exact History Recovery

### WP05 — Shallow Dependency Recovery

- **Prompt**: `tasks/WP05-shallow-dependency-recovery.md`
- **Priority**: P2
- **Dependencies**: WP04
- **Goal**: Fail closed on exact missing history and recover only explicitly
  selected supplying refs through the same bounded quarantine transaction.
- **Independent test**: Run trust-sensitive commands in a depth-one clone with
  missing base, policy, candidate, actor predecessor, pipeline, and ancestry;
  each blocks with an exact identifier, and an approved bounded recovery makes
  the unchanged operation succeed without global unshallow.
- **Estimated prompt size**: ~290 lines

Included subtasks:

T023 Write depth-limited fail-closed and recovery tests (WP05)
T024 Implement exact shallow dependency-gap classification (WP05)
T025 Guard policy, candidate, CI, decision, and merge verification (WP05)
T026 Route explicit selected recovery through quarantine (WP05)
T027 Prove bounded retry without global unshallow or trust expansion (WP05)

Implementation sketch: model missing dependencies separately from invalid
facts, resolve exact required objects at trust boundaries, derive a selected
supplying ref when possible, and restart the operation only after normal
quarantine validation and promotion complete.

Risks: treating absence as corruption, broadening selection, reusing partial
decisions, hidden `--unshallow`, and commands advancing with missing evidence.

## Phase 4 — Operational Proof

### WP06 — Operational Black-Box Acceptance

- **Prompt**: `tasks/WP06-operational-black-box-acceptance.md`
- **Priority**: P1
- **Dependencies**: WP01, WP02, WP03, WP04, WP05
- **Goal**: Prove the complete user-visible workflow through the compiled CLI,
  separate processes, real temporary repositories, and an ordinary disposable
  bare Git remote.
- **Independent test**: The full policy, identity, replication, shallow, CI,
  review, decision, and merge scenario passes three consecutive times within
  120 seconds each, without credentials, Docker, public mutation, or recursive
  test execution.
- **Estimated prompt size**: ~340 lines

Included subtasks:

T028 Build a subprocess-level multi-repository acceptance harness (WP06)
T029 Exercise policy amendment under the old policy (WP06)
T030 Exercise distinct actors and planned rotation (WP06)
T031 Exercise selected replication, hostile input, and every budget boundary (WP06)
T032 Exercise shallow detection and bounded recovery (WP06)
T033 Exercise role-distinct governance, compatibility, and three-run repeatability (WP06)

Implementation sketch: finish top-level routing, compile one test binary,
operate isolated author/reviewer/runner/verifier clones, generate a tiny
nonrecursive pipeline action, assert full IDs and accepted refs rather than
formatting trivia, and preserve the current public-event compatibility baseline.

Risks: recursive `go test`, shared working-directory state, timing flakes,
environment credential leakage, brittle output assertions, and accidental
network or public-remote mutation.

## Phase 5 — Living Protocol and Public Evidence

### WP07 — Living Protocol and Public Proof

- **Prompt**: `tasks/WP07-living-protocol-and-public-proof.md`
- **Priority**: P1
- **Dependencies**: WP06
- **Goal**: Align every operator/protocol/security document with shipped
  behavior, then perform and record the staged public self-hosting loop with
  exact reproducible identifiers.
- **Independent test**: Follow only the published documentation from a fresh
  shallow clone with no private identity and reconstruct every recorded actor,
  event, ref, policy, and Git object ID.
- **Estimated prompt size**: ~320 lines

Included subtasks:

T034 Update canonical protocol and governance documentation (WP07)
T035 Document identity continuity, keyring safety, and rotation limits (WP07)
T036 Document selected replication, quarantine, budgets, and threat boundaries (WP07)
T037 Update README, host compatibility, and public operator flow (WP07)
T038 Record deterministic offline verification evidence (WP07)
T039 Perform and record the staged public self-hosting proof (WP07)

Implementation sketch: document exact wire fields and CLI behavior without
overclaiming transport quotas or human independence; update compatibility
observations; rehearse the offline sequence; then use Nichthub facts and
ordinary Git publication for the two governed public stages and record the
full immutable evidence chain.

Risks: stale docs, leaked local paths or secrets, claiming a network quota Git
cannot enforce, confusing actor separation with organizational independence,
or publishing branch and collaboration refs inconsistently.

## Global Definition of Done

- Every FR-001 through FR-020 is mapped to at least one work package.
- Existing version-0 event payloads and public IDs remain readable unchanged.
- `go test ./...`, `go test -race ./...`, and `go vet ./...` pass.
- The offline operational scenario passes three consecutive times within its
  per-run limit.
- No tracked or signed artifact contains a private key, token, credential, or
  host-private path.
- No workflow requires GitHub collaboration APIs, a Nichthub service, Docker,
  shared actor keys, implicit authority transfer, or global unshallow.
- Protocol, threat-boundary, compatibility, operator, and public-verification
  documents agree with observable behavior.
