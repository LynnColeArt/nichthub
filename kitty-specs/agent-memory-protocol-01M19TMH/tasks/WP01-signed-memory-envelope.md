---
work_package_id: WP01
title: Signed Memory Envelope
dependencies: []
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-004
- FR-005
- FR-010
- FR-011
- FR-021
- NFR-001
- NFR-002
- NFR-004
- NFR-009
- NFR-010
- NFR-011
- C-002
- C-004
- C-005
- C-007
planning_base_branch: feat/agent-memory-protocol
merge_target_branch: feat/agent-memory-protocol
branch_strategy: Planning artifacts for this mission were generated on feat/agent-memory-protocol. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/agent-memory-protocol unless the human explicitly redirects the landing branch.
subtasks:
- T001
- T002
- T003
- T004
- T005
history: []
agent_profile: implementer-ivan
authoritative_surface: memory_event
create_intent:
- memory_event.go
- memory_event_test.go
execution_mode: code_change
model: ''
owned_files:
- memory_event.go
- memory_event_test.go
role: implementer
tags: []
tracker_refs: []
---

# Work Package Prompt: WP01 – Signed Memory Envelope

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Establish the version-0 `nh-memory/0` wire foundation as a new, strictly
decoded, bounded, canonically encoded, Ed25519-signed envelope without changing
the existing `nh/0` collaboration `Event` type, payload bytes, or public IDs.
Deliver reusable record, anchor, applicability, evidence, identifier, signing,
and verification primitives for every later memory work package.

## Context

This is the foundation for memory storage, projection, indexing, CLI, and replication.
Canonical memory is an immutable signed Git claim proving authorship/integrity,
not truth, applicability, evidence resolution, trust, authority, or permission.

Read `spec.md`, `plan.md`, `data-model.md`, `research.md`, and all three contracts.
Reuse identity/ID idioms from `identity.go`, `identity_keyring.go`, and `event.go`,
but keep changes inside the owned files. If reuse needs an out-of-map edit, add
a narrow memory-specific helper and leave consolidation for later.

Do not add fields/kinds to `Event`, call collaboration append/collection, or add
memory refs/storage. WP02 owns streams; WP03 owns lifecycle, applicability,
evidence availability, and trust. WP01 validates canonical signed shape only.

Run implementation through:

```sh
spec-kitty agent action implement WP01 --agent <name>
```

### Subtask T001: Freeze backward-compatible wire fixtures and constants

**Purpose**

Lock the additive `nh-memory/0` byte contract and every version-0 record-shape
limit while proving the prior collaboration wire remains byte/ID stable.

**Steps**

1. Create `memory_event_test.go` with deterministic fixture identities,
   timestamps, actors, stream IDs, Git OIDs, subjects, and typed evidence IDs.
2. Add a complete valid `record` envelope fixture whose JSON field order,
   omitted fields, empty-list representation, and nested-object layout are
   asserted against exact literal bytes.
3. Assert the fixture's full `sha256:<64-lowercase-hex>` memory ID, not merely
   its length or prefix, so later field reordering is detected.
4. Add one deterministic fixture for each operation: `record`, `supersede`,
   `retract`, and `challenge`; include both record-producing operations in ID
   goldens.
5. Cover all six record kinds in stable fixtures: `observation`, `decision`,
   `assumption`, `attempt`, `verification`, and `handoff`.
6. Freeze named constants for protocol version, content bytes, topic count,
   evidence count, path count/bytes, handoff list count/entry bytes, and any
   bounded reason/outcome text introduced by the implementation.
7. Use the specification's exact public bounds: 65,536 UTF-8 content bytes,
   32 normalized topics, and 64 unique typed evidence references.
8. Choose explicit conservative finite version-0 path and handoff limits where
   the contract delegates values to implementation; give them descriptive
   names and lock the selected values in tests rather than using magic numbers.
9. Preserve the distinction between nil/omitted optional fields and required
   empty arrays, especially the four handoff lists; fixture bytes are the
   authority once this package lands.
10. Add a collaboration compatibility fixture in this new test file that
    marshals representative existing `Event` values and checks exact legacy
    payload bytes and exact existing IDs.
11. Run the existing
    `TestIdentityFieldsPreserveEveryExistingEventKindID` fixture as part of the
    focused gate; do not edit its expected bytes or IDs to accommodate memory.
12. Confirm red-first failures are caused by absent memory types/functions,
    not by changing or weakening existing event fixtures.

**Files**

- Create `memory_event_test.go` for golden payload, ID, constant, and
  compatibility fixtures; expect roughly 250–400 lines by the end of WP01.
- Create `memory_event.go` only after the wire expectations are explicit;
  constants and types added by later subtasks remain in this file.
- Do not modify `event.go`, `event_test.go`, or
  `identity_continuity_test.go`; they are compatibility inputs, not owned
  surfaces.

**Validation**

- `go test ./... -run 'TestMemory(Event|Wire)|TestIdentityFieldsPreserve'`
  initially fails for missing memory behavior, then passes after T002–T004.
- Exact JSON byte and digest assertions fail on field order, tag, omission, or
  normalization drift.
- Existing collaboration fixture IDs remain unchanged without updating any legacy golden.

### Subtask T002: Implement strict envelope and nested record types

**Purpose**

Define an unambiguous signed envelope whose optionality distinguishes absent
operation-inapplicable fields from present but invalid zero values.

**Steps**

1. In `memory_event.go`, define `MemoryEnvelope` with protocol, operation,
   actor, actor name, public key, stream, sequence, timestamp, previous,
   record, target, reason, and lifecycle evidence fields in canonical order.
2. Represent `record` as a pointer or equivalent presence-aware value so
   `record`/`supersede` can require it and `retract`/`challenge` can prohibit it.
3. Define `MemoryRecord` with kind, content, anchor, applicability, topics,
   evidence, attempt outcome, and optional handoff data in the order frozen by
   T001.
4. Define `MemoryAnchor`, `PathAnchor`, `Applicability`, and `HandoffFields`
   directly from `data-model.md`; keep author prose in data fields and do not
   introduce instruction, command, tool, authorization, or execution fields.
5. Use JSON tags that make the T001 fixture canonical. Apply `omitempty` only
   to genuinely optional fields; do not let it erase a required empty handoff
   list or blur absent versus invalid nested objects.
6. Define closed version-0 constants/enums for the four operations, six record
   kinds, three applicability modes, and any explicitly selected
   attempt-outcome or lifecycle-reason vocabulary.
7. Where the artifacts permit a typed nonblank string rather than a closed
   enum, validate and bound it without inventing semantic truth or authority.
8. Add a strict decoder used by verification: require valid UTF-8 payload
   bytes, reject unknown fields at every nested level, reject trailing JSON,
   and reject multiple top-level values.
9. Re-marshal the decoded envelope with Go `encoding/json` and require exact
   byte equality with the supplied payload. This rejects whitespace, field
   reordering, duplicate object keys, and alternate encodings that are valid
   JSON but not the canonical wire form.
10. Keep strict decoding memory-specific; do not retrofit it onto legacy
    collaboration events because doing so could change accepted historical
    payloads.
11. Keep types free of Git repository access, environment reads, callbacks,
    model APIs, or mutable package state. This layer is a pure wire boundary.
12. Add round-trip tests for every nested type, including empty optional
    collections, explicitly present empty required collections, and hostile
    unknown fields nested below `record`, `anchor`, and `handoff`.

**Files**

- Create and implement `memory_event.go`; expect roughly 300–500 lines for
  types, constants, strict decode, validation, IDs, and crypto by WP completion.
- Extend only `memory_event_test.go` with shape and strict-decoding tests.
- Do not create storage, projection, policy, CLI, or replication types in this
  package; later WPs own those concerns.

**Validation**

- Canonical fixtures decode and re-encode to identical bytes.
- Unknown/nested fields, duplicate keys, trailing tokens, invalid UTF-8, and noncanonical bytes fail closed.
- `go test ./... -run 'TestMemory(EventTypes|StrictDecode|CanonicalJSON)'` passes without changing collaboration tests.

### Subtask T003: Validate anchors, applicability, topics, evidence, and kind-specific bounds

**Purpose**

Reject malformed or ambiguous claims on local signing and hostile ingestion.
Validation establishes shape, never existence, resolution, truth, or trust.

**Steps**

1. Add one central envelope validator and narrow nested validators so local
   encoding and hostile verification enforce the same rules and constants.
2. Require protocol `nh-memory/0`, a supported operation, full actor/stream IDs, a positive sequence, and an RFC3339Nano timestamp.
3. Require sequence 1 to have no `previous`; above 1 requires a full memory ID.
   Do not inspect the stream head here; WP02 validates continuity.
4. Enforce exact operation shape: `record` has only a record; `supersede` has
   target and replacement; `retract`/`challenge` have target and typed reason,
   with optional lifecycle evidence only on `challenge`.
5. Prohibit operation-inapplicable record, target, reason, and evidence fields.
6. Require content to be valid UTF-8, nonblank, and at most 65,536 encoded bytes.
   Count bytes, not runes, and preserve command-like prose as inert data.
7. Require an exact full SHA-1 or SHA-256 Git commit OID on every anchor without
   resolving it or making ancestry claims in this layer.
8. Validate slash-based paths independent of host OS: reject absolute, backslash,
   empty/dot, traversal, non-normalized, control, duplicate, invalid-blob, and over-limit forms.
9. Permit the exact `absent` blob marker only where specified; every other path
   blob must be a full Git OID.
10. Validate typed subjects as documented issue/proposal/event/policy/pipeline/run classes
    plus full IDs; fixture one exact rendering and reject prose or short IDs.
11. Enforce `exact`, `descendants`, and `subject`; subject applicability requires
    a value identical to the anchor subject, while other modes prohibit it.
12. Use one deterministic standard-library topic normalization rule; reject blank/control,
    non-normalized, duplicate, unsorted, or more than 32 topics rather than mutating input.
13. Validate ordered unique `git:`, `event:`, and `memory:` evidence with full IDs;
    preserve order and reject duplicates or more than 64 entries.
14. Require record evidence for `verification` without resolving it; keep it distinct from challenge evidence.
15. Require bounded nonblank/frozen `attemptOutcome` only for `attempt` and prohibit it elsewhere.
16. Require `handoff` data only for `handoff`, with all four present, bounded,
    valid-UTF-8, nonblank-entry lists: completed, assumptions, blockers, and next actions.
17. Prohibit handoff data elsewhere; proposed next actions are neither executable nor authorized.
18. Test each rule with one-field mutations and bounded diagnostics that do not echo hostile text.

**Files**

- Add pure validation and normalization helpers to `memory_event.go`.
- Add table-driven mutation tests and helper builders to
  `memory_event_test.go`.
- Keep repository existence, Git ancestry, same-author target checks, evidence
  resolution, and policy trust outside this WP.

**Validation**

- `go test ./... -run 'TestMemory(EventValidation|Anchor|Applicability|Topics|Evidence|Kinds)'`
  covers every operation and record kind.
- Tests prove invalid fields fail identically on local encode and hostile
  verify paths.
- Error assertions match bounded field-oriented diagnostics, not entire author
  content, keys, signatures, or environment values.

### Subtask T004: Implement canonical encode, sign, verify, and full IDs

**Purpose**

Provide exact canonical bytes signed by the selected actor, their public memory
ID, and independent fail-closed structural/identity/signature verification.

**Steps**

1. Return `sha256:` plus lowercase SHA-256 hex of exact envelope bytes as memory ID; never use a Git commit ID.
2. Add a strict full memory-ID validator and full stream-ID validator. Reject
   short, uppercase, malformed, wrong-prefix, and wrong-length trust-bearing
   values.
3. Add deterministic default stream derivation from the exact domain-separated
   bytes `nh-memory-stream-v0\x00<actor>\x00default`, rendered as a full stream
   ID.
4. Add a constructor or narrow initializer that copies actor, actor name, and
   public key from an existing `Identity`, accepts explicit operation/stream
   sequence inputs, and emits a UTC RFC3339Nano timestamp.
5. Keep timestamp injectable or allow tests to construct envelopes directly so
   golden payloads never depend on wall-clock time.
6. Implement canonical encoding with `encoding/json` only after full envelope
   validation. Do not use indentation, HTML/template rendering, maps for wire
   objects, or a vendor-specific canonicalization library.
7. Before signing, require the envelope actor and embedded public key to match
   the supplied `Identity`; decode the raw-base64 Ed25519 key and verify the
   actor fingerprint binding.
8. Sign the exact encoded payload with the identity's existing private-key
   loader and return payload plus raw signature bytes in the same convention as
   existing event helpers. WP02 will store the signature file as raw-base64.
9. Verify payload validity and canonical bytes, decode and validate the embedded
   raw-base64 public key, recompute its actor fingerprint, verify the Ed25519
   signature over the exact bytes, and return the full memory ID.
10. Do not consult the local identity, keyring status, current policy, Git
    history, or remote state while verifying a received envelope.
11. Ensure actor display name is never used as cryptographic identity and is
    never interpolated into an unbounded error.
12. Add deterministic sign/verify round trips for every operation and record
    kind, plus tamper cases for payload, signature, actor, public key, stream,
    protocol, timestamp, sequence, and previous ID.
13. Assert two equivalent structs encode to the same bytes and ID, while any
    intentional signed field change produces a different ID.
14. Assert verification computes the ID from the supplied canonical payload
    bytes rather than from a second representation or transport commit.

**Files**

- Add ID, constructor, encode/sign, and verify helpers to `memory_event.go`.
- Add cryptographic round-trip, tamper, and deterministic-ID tests to
  `memory_event_test.go`.
- Reuse `Identity.publicKey`, `Identity.privateKey`, `actorForPublicKey`, and
  compatible validation idioms without altering their collaboration behavior.

**Validation**

- `go test ./... -run 'TestMemory(Encode|Sign|Verify|ID|DefaultStream)'` passes.
- Modified payloads/signatures never yield an accepted envelope or ID.
- Tests distinguish actor, memory, stream, evidence, and Git IDs so they cannot be substituted.

### Subtask T005: Add hostile-input and exact-boundary regression tests

**Purpose**

Prove exact bounds, inert hostile author text, and collaboration isolation; the
fixtures become the acceptance boundary reused by storage and quarantine WPs.

**Steps**

1. Build table-driven one-below, exact-limit, and one-above tests for the
   65,536-byte content limit using both single-byte and multibyte UTF-8 text.
2. Add 31/32/33-topic and 63/64/65-evidence cases with duplicate, non-normalized,
   unsorted, empty, wrong-prefix, short-ID, and invalid-hex mutations.
3. Test one-below/exact/one-above for each path and handoff count/byte limit,
   including any aggregate encoded-size limit.
4. Exercise bad sequence/previous combinations and malformed target, actor, stream,
   Git OID, subject, path, blob, timestamp, key, and signature values.
5. Test all forbidden operation-field combinations and every forbidden
   kind-specific field combination, not only missing required fields.
6. Verify rejection of raw invalid UTF-8, escaped controls, nested unknown fields,
   duplicate keys, trailing values, whitespace, and reordered signed fields.
7. Preserve prompt-, command-, tool-, control-, and repetition-like content only as
   JSON string data, with no callback, process, file, network, ref, or policy effect.
8. Name bounded fields in errors without echoing hostile content, secrets, environment, credentials, or raw JSON.
9. Add exact byte-and-ID compatibility assertions for representative prior
   `nh/0` collaboration payloads in `memory_event_test.go`.
10. Run the existing all-event-kind compatibility fixture and the complete
    repository suite; no existing collaboration golden may be regenerated.
11. Run race, vet, and build gates because these primitives underpin every later WP.
12. Format only the two owned Go files and perform a whitespace/error diff
    check before handing the package to review.

**Files**

- Complete `memory_event_test.go` with hostile, exact-boundary, crypto, inert
  content, and legacy-compatibility coverage.
- Adjust `memory_event.go` only to correct behavior exposed by these tests;
  avoid relaxing validation to make fixtures pass.
- Do not edit legacy payload fixtures, generated mission artifacts, or files
  owned by later WPs.

**Validation**

Run the focused and full gates:

```sh
gofmt -w memory_event.go memory_event_test.go
go test ./... -run 'TestMemory|TestIdentityFieldsPreserve'
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

Exact limits pass, one-above and hostile crypto/canonical cases fail closed,
inert strings remain data, and collaboration payload bytes/IDs stay unchanged.

## Definition of Done

- Only the two `create_intent` implementation files are changed by WP01.
- The canonical envelope/nested types match the data model and exact-byte fixtures.
- All operations, kinds, anchors, applicability, topics, evidence, paths, outcomes, and handoffs validate fail closed.
- Signing/verification share canonical JSON validation, Ed25519 actor binding, and full IDs.
- One-below, exact, and one-above tests cover every record bound.
- Hostile JSON/UTF-8, unknown/duplicate fields, noncanonical bytes, tampering, and field smuggling are rejected.
- Tests prove memory content remains inert data and that the wire layer has no
  execution, authorization, network, policy, ref, or automatic-capture path.
- Exact collaboration payload bytes and public IDs remain unchanged, including
  the existing all-event-kind golden fixture.
- `go test -race ./...`, `go vet ./...`, `go build ./...`, and
  `git diff --check` pass from the WP implementation worktree.
- After each subtask is complete and its evidence is present, record completion
  through the event-sourced task surface:

```sh
spec-kitty agent tasks mark-status T001 --status done
spec-kitty agent tasks mark-status T002 --status done
spec-kitty agent tasks mark-status T003 --status done
spec-kitty agent tasks mark-status T004 --status done
spec-kitty agent tasks mark-status T005 --status done
```

## Risks

- **Canonical JSON ambiguity**: Go accepts alternate encodings/duplicate keys;
  compare decoded/re-marshaled bytes and use hostile fixtures.
- **Compatibility drift**: sharing or modifying `Event` could change historical
  collaboration IDs; mitigate with a separate envelope and immutable goldens.
- **UTF-8 replacement**: `encoding/json` can replace malformed string bytes;
  mitigate by rejecting invalid UTF-8 on raw payload bytes before decoding.
- **Semantic overreach**: syntax is not truth, applicability, evidence, or trust;
  keep those projections out of WP01 and terminology separate.
- **Host-path dependence**: `filepath` normalization differs across operating
  systems; use repository slash-path rules and cross-platform hostile cases.
- **Limit inconsistency**: encode/verify can diverge; route both through one
  validator and boundary table.
- **Error-data leakage**: hostile content, keys, or signatures could appear in
  diagnostics; report stable field names and safe full public IDs only.
- **Unspecified typed syntax**: freeze the narrowest interoperable subject,
  outcome, and reason choice in exact fixtures; reject permissive IDs.

## Reviewer Guidance

Review exact wire bytes first. Confirm field order/omission, distinct full-ID
namespaces, rejection of signed noncanonical JSON, exact-byte signatures/IDs,
and raw invalid-UTF-8 rejection before JSON replacement.

Trace each operation and each record kind through both local encode and hostile
verify. Pay particular attention to prohibited fields, nil versus empty lists,
topic normalization, evidence order/uniqueness, subject applicability matching,
path normalization, verification evidence, attempt outcome, and all four
handoff lists. Reject any implementation that resolves Git objects or lifecycle
targets in this layer or collapses provenance, trust, evidence, and truth into
one flag.

Finally, inspect the compatibility evidence: `Event` must be untouched, exact
legacy bytes/IDs must remain fixed, and collaboration tests must pass without
memory refs or migration. Re-run the complete DoD commands and verify the diff
contains only `memory_event.go` and `memory_event_test.go` plus Spec Kitty's
event-sourced status records created by the workflow.

Implementation entry point:

```sh
spec-kitty agent action implement WP01 --agent <name>
```
