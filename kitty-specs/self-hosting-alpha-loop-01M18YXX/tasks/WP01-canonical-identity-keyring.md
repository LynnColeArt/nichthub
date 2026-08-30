---
work_package_id: WP01
title: Canonical Identity Keyring
dependencies: []
requirement_refs:
- FR-005
- FR-006
- FR-018
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T001
- T002
- T003
- T004
- T005
phase: Phase 1 - Independent Foundations
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: identity.go
create_intent:
- identity_keyring.go
- identity_keyring_test.go
execution_mode: code_change
model: ''
owned_files:
- identity.go
- identity_keyring.go
- identity_keyring_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP01 – Canonical Identity Keyring

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Replace `.git/nh/identity.json` as the active private-state authority with a
versioned keyring containing one record per actor and one atomic active-actor
pointer. Existing repositories must migrate automatically without changing a
single key byte, actor fingerprint, signed payload, or actor ref.

This package is complete when:

- a new repository initializes directly into the keyring layout;
- a valid legacy identity migrates exactly once and remains usable;
- migration is idempotent and recoverable at every durable boundary;
- every stored keypair is validated before it can become active;
- private records and transaction state are owner-only;
- a retired/non-active key is never chosen by ordinary `loadIdentity` calls;
- current issue, proposal, review, CI, decision, and merge tests still pass.

## Context & Constraints

Read the mission [spec](../spec.md), [plan](../plan.md), [data model](../data-model.md),
[research](../research.md), and [identity contract](../contracts/identity-continuity-v0.md).
The project charter requires signed immutable history, test-first changes,
targeted staging, and no mandatory service or new Go dependency.

The existing `Identity` JSON shape remains useful as the record shape. An actor
is still the digest of exactly one Ed25519 public key. The private state below
`.git/nh` is not protocol data, must never be added to Git, and must never be
printed by public inspection commands.

Do not implement public identity-continuity events here; WP02 owns them. This
package supplies the local persistence and active-signer interface WP02 uses.
Do not implement lost-key recovery, shared-key writers, or policy authority
transfer.

## Branch Strategy

- **Strategy**: execution lane from `feat/self-hosting-alpha-loop`; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

The runtime allocates the real execution workspace from `lanes.json`. Work only
inside that workspace. Commit through the Spec Kitty safe-commit flow, staging
only the files declared in `owned_files`.

## Subtasks & Detailed Guidance

### T001 – Write failing migration and signer-compatibility tests

- Start with tests in `identity_keyring_test.go`; production changes follow.
- Build real temporary Git repositories and set repository-local Git identity.
- Create a legacy `.git/nh/identity.json` using the current implementation,
  sign an event, and record actor, public/private encodings, event ID, and ref.
- Specify the target layout: `.git/nh/active`,
  `.git/nh/identities/<actor>.json`, and optional `.git/nh/rotation.json`.
- Assert migration returns the same signer and permits a second event on the
  same actor chain with the correct predecessor.
- Exercise missing, malformed, mismatched, truncated, and symlink-sensitive
  local files as hostile private state; failures must not switch the signer.
- Keep tests behavioral: inspect modes, bytes, actor IDs, signatures, and refs,
  not temporary filenames or internal helper call order.

### T002 – Implement private keyring records and atomic local storage

- Add `identity_keyring.go` as the sole module that reads or writes private
  identity records, the active pointer, and local rotation transaction state.
- Define a small versioned persisted schema. Reject unknown versions and
  malformed actors/keys before returning a signer.
- Reuse the existing actor/keypair checks and centralize them so legacy and
  keyring records cannot diverge semantically.
- Create directories with `0700` and private files with `0600`. On existing
  files, detect and reject unsafe non-owner permissions where portable rather
  than silently blessing them.
- Use same-directory temporary files, flush/close, chmod, and atomic rename for
  record/pointer replacement. Avoid process-global working-directory changes.
- Make writes idempotent: the exact same actor record may be retried, while a
  different key under an existing actor path is a hard error.
- Never include private values in error text, logs, event bodies, or filenames
  other than the public actor fingerprint.

### T003 – Migrate the legacy identity through the active facade

- Keep `createIdentity` and `loadIdentity` as the compatibility facade used by
  existing commands, but route all new persistence through the keyring.
- For new initialization, reject only when a usable active identity already
  exists; do not mistake an interrupted rotation record for a valid signer.
- On legacy load, validate the complete old record first, then durably create
  the identical actor record, then atomically set the active pointer.
- Retain the legacy file until both new artifacts are durable. A retry after a
  crash must converge without generating another key or actor.
- Once the keyring is authoritative, ordinary loads must not fall back to a
  subsequently modified legacy file.
- Preserve current user-facing initialization behavior except for the private
  path it reports; do not print a private key or raw transaction state.

### T004 – Implement active actor switching and recoverable rotation state

- Provide narrow package-private operations to list public record metadata,
  retrieve a specific record for explicit local operations, inspect the active
  actor, and atomically switch the active actor.
- Model in-progress planned rotation with predecessor actor, target actor,
  relationship, and optional authorization/acceptance event IDs.
- Reject a pointer switch unless the target record validates and the caller
  supplies a completed transaction state. WP02 will prove the event facts and
  call this boundary.
- Ensure a transaction retry can distinguish: nothing durable, target key
  durable, authorization durable, acceptance durable, and pointer switched.
- Keep the predecessor record for historical inspection. Local lifecycle
  metadata may mark it retired, but must not modify its public actor history.
- Do not invent an automatic winner when multiple local records exist.

### T005 – Prove permissions, idempotency, secret isolation, and compatibility

- Add boundary tests for exact `0600` files and `0700` directories on supported
  Unix hosts, with portable skips only where permission semantics do not exist.
- Inject failures around durable steps through narrow test seams; assert the
  previous active signer remains usable until the final switch.
- Run migration twice and resume every partial state twice; the actor, ref, and
  number of records must remain stable.
- Scan tracked fixtures and captured command output for private-key encodings.
- Re-run existing identity and event tests to prove old actors still sign and
  verify without event-ID changes.
- Mark T001–T005 done through the task event log only after the focused tests
  and full suite pass.

## Test Strategy

Required checks:

```bash
go test ./... -run 'Test.*(Identity|Keyring|Migration|RotationState)'
go test ./...
go test -race ./...
go vet ./...
```

Use real temporary repositories for filesystem behavior. Never weaken a test
because an implementation detail changed; assert the public signer, file
permission, atomicity, and event-chain contracts. No test fixture may contain a
real project private key.

## Risks & Mitigations

- **Actor drift**: compare decoded key bytes and recomputed actor before and
  after migration, not merely JSON fields.
- **Partial migration**: order durable writes so the legacy signer stays
  authoritative until the new record and active pointer are valid.
- **Secret disclosure**: keep errors structural and use synthetic disposable
  keys in tests.
- **Concurrent local writes**: atomic rename protects files, while actor ref
  compare-and-swap remains the event-chain concurrency authority.
- **Scope creep**: expose only persistence and signer-selection primitives;
  leave public continuity semantics to WP02.

## Review Guidance

The reviewer should recreate a legacy repository independently, compare actor
and signature facts before/after migration, inspect filesystem modes, and
interrupt each transaction boundary. Reject the package if migration can
generate a replacement identity, if the active pointer can name an invalid or
incomplete actor, or if any private material appears in output or tracked data.

Verify that `identity_keyring.go` is the single private persistence authority
and that `identity.go` is a compatibility facade rather than a second storage
implementation.

## Activity Log

> Append entries in chronological order. Status is managed through
> `status.events.jsonl`, not by editing task checkboxes.

- 2026-08-30T17:26:50Z – system – Prompt created.
