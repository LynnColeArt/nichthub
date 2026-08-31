---
work_package_id: WP06
title: Selected Memory Replication and Recovery
dependencies:
- WP01
- WP02
- WP03
requirement_refs:
- FR-006
- FR-007
- FR-022
- FR-023
- FR-024
- NFR-006
- NFR-009
- NFR-010
- NFR-011
- NFR-012
- C-003
planning_base_branch: feat/agent-memory-protocol
merge_target_branch: feat/agent-memory-protocol
branch_strategy: Planning artifacts for this mission were generated on feat/agent-memory-protocol. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/agent-memory-protocol unless the human explicitly redirects the landing branch.
subtasks:
- T026
- T027
- T028
- T029
- T030
history: []
agent_profile: implementer-ivan
authoritative_surface: replication.go
create_intent:
- memory_replication_test.go
execution_mode: code_change
model: ''
owned_files:
- replication.go
- quarantine.go
- shallow.go
- commands.go
- memory_replication_test.go
role: implementer
tags: []
tracker_refs: []
---

# Work Package Prompt: WP06 – Selected Memory Replication and Recovery

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Extend Nichthub's selected, quarantined replication transaction with memory as
a third independently failing selector kind. Synchronize exact signed memory
streams, report honest dependency gaps, and prove fresh-clone convergence
without making collaboration availability depend on memory validity or volume.

## Context

The operational alpha already owns per-remote selections, direct refspecs,
separate bare quarantine repositories, positive budgets, exact validation,
durable pending-acceptance anchors, atomic accepted-ref promotion, publication,
and selected shallow recovery. Extend those existing seams narrowly; do not
create a second replication engine or bypass its transaction receipts.

WP01 defines strict bounded `nh-memory/0` payloads, WP02 owns memory ref grammar
and verified linear streams, and WP03 supplies deterministic lifecycle,
anchor, evidence, applicability, and dependency classifications. Read
`spec.md`, `plan.md`, `data-model.md`, `research.md`, all three contracts, and
the existing `docs/replication-v0.md` before implementation. In source, trace
`ReplicationSelection`, `replicationRequest`, `ReplicationOutcome`,
`runReplicationTransaction`, `resolveReplicationRequests`, budget measurement,
promotion receipts, `ShallowDependencyGap`, `recoverSelectedShallow`, and
`publishLocalFacts` before editing them.

Memory uses local refs `refs/nh/memory/<actor>/<stream-digest>` and accepted
refs `refs/nh/remotes/<remote>/memory/<actor>/<stream-digest>`. A quarantine
object or ref is never a recall source. An explicit actor or proposal selector
imports no memory; compatibility `--all` discovers memory after upgrade. Keep
selection separate from policy trust, signature validity separate from
evidence resolution, and missing dependencies separate from malformed data.

Only the five owned files may change. `memory_replication_test.go` is the sole
new file; the other four are existing integration surfaces. Preserve every
existing collaboration payload, public event ID, accepted ref, saved selection
meaning, and command behavior unless this prompt explicitly adds memory.

### Subtask T026: Add optional memory selectors with strict full-ID validation

**Purpose**

Make exact memory streams selectable per remote while keeping every existing
version-1 actor/proposal selection valid and behaviorally unchanged.

**Steps**

1. Add `Memories []string` to `ReplicationSelection` with JSON name
   `memories` and `omitempty`; do not bump `replicationSelectionVersion`.
2. Decode legacy saved JSON with no `memories` as an empty memory selection and
   preserve its current actor, proposal, `all`, budget, and display meaning.
3. Validate each memory selector with WP02's canonical full stream-ID validator;
   reject short, uppercase, malformed, whitespace-padded, or ambiguous IDs.
4. Require memory selectors to be unique, sort them before persistence and
   display, and keep all diagnostics on full safe public IDs.
5. Preserve actor/proposal duplicate and cross-kind rules exactly; adding the
   optional field must not weaken or reinterpret their validation.
6. Extend the exact-selection nonempty rule so any actor, proposal, or memory
   is sufficient, while a truly empty non-`all` selection still fails.
7. Extend `--all` mutual exclusion to memory, actor, and proposal selectors;
   do not silently discard any exact selector when `--all` is present.
8. Add repeated `--memory STREAM` handling to `nh replication select` using the
   existing repeated-string flag mechanism and update usage text precisely.
9. Print selected memories deterministically as `memory: <full-stream-id>` in
   `nh replication show`, after current actor/proposal lines.
10. Keep selection files private, atomic, local-only, and byte-compatible with
    saved version-1 documents that omit the optional field.
11. Ensure a no-selection compatibility load still returns bounded `All=true`;
    its new discovery behavior belongs to T027, not selection persistence.
12. Add focused tests for legacy decode/show, memory-only selection, mixed
    selection, duplicates, malformed/short IDs, empty selection, and `--all`.

**Files**

- Modify `replication.go` for the additive model, validation, CLI flag, usage,
  stable persistence, and display behavior (roughly 35–70 added lines).
- Add the selection compatibility cases to `memory_replication_test.go`
  (roughly 80–120 lines for this slice).

**Validation**

- Run `go test ./... -run 'TestMemoryReplication.*Selection'`.
- Run the existing `TestReplicationSelectionRoundTripAndRejectsAmbiguity` and
  `TestReplicationCompatibilitySelectionIsBounded` unchanged.
- Record completion with
  `spec-kitty agent tasks mark-status T026 --status done` only after both new
  and legacy selection tests pass.

### Subtask T027: Advertise, discover, request, and publish exact memory refs

**Purpose**

Translate selected stream IDs into exact owner-bound remote refs and publish
local memory streams without confusing memory with actor or proposal roots.

**Steps**

1. Add a `memory` replication kind constant and exact accepted-memory ref
   builder/parser matching the contract's remote/owner/stream grammar.
2. Extend remote advertisement to request only the actor, proposal, and memory
   namespaces; never enumerate arbitrary remote refs or primary branches.
3. Parse memory candidates only from
   `refs/nh/memory/<full-actor>/<64-hex-stream-digest>` and reconstruct the
   canonical full `sha256:` stream ID without accepting alternate spelling.
4. Detect malformed, duplicate, ambiguous, and owner-conflicting advertised
   refs as scoped memory-request failures, not actor/proposal requests.
5. Require one explicitly selected stream ID to resolve to exactly one owner
   and exact source ref; a missing match reports the full stream and remote.
6. Under compatibility `--all`, discover every well-formed advertised memory
   ref in stable ref order and pass it through ordinary quarantine validation.
7. Keep explicit actor/proposal selection independent: do not discover or
   request any memory unless `--memory` or compatibility `--all` authorizes it.
8. Create memory quarantine destinations under a transaction-local namespace
   that includes the exact owner and stream digest and cannot collide with
   actor or proposal destinations.
9. Extend request keys and destination dispatch so the same textual digest in
   different selector kinds cannot overwrite another request or outcome.
10. Extend promotion-ref parsing and durable receipt validation to recognize
    only exact accepted-memory refs while retaining prior record versions.
11. Publish every local `refs/nh/memory/*/*` after local actor and proposal refs,
    using exact push refspecs and deterministic enumeration.
12. Report memory publication inspection/push failures as a safe publication
    phase error; never rewrite local canonical streams or mask import results.

**Files**

- Modify `quarantine.go` for memory advertisement, exact request construction,
  quarantine destinations, accepted refs, and receipt validation.
- Modify `commands.go` only for post-actor/proposal local memory publication.
- Extend `memory_replication_test.go` with explicit, all, malformed,
  unselected, ambiguous-owner, and publication cases.

**Validation**

- Run `go test ./... -run 'TestMemoryReplication.*(Discovery|Request|Publish)'`.
- Inspect test remotes to prove explicit actor/proposal selection requested
  zero memory and `--all` admitted only exact well-formed memory refs.
- Record completion with
  `spec-kitty agent tasks mark-status T027 --status done` after focused tests
  and existing sync/publication tests pass.

### Subtask T028: Quarantine, budget, validate, and atomically promote streams

**Purpose**

Admit hostile memory only after per-stream measurement and complete structural,
cryptographic, continuity, and dependency validation, with independent outcome
and atomic accepted-ref semantics.

**Steps**

1. Fetch each exact memory ref only into the generated separate bare
   quarantine repository and confirm its root equals the advertised OID.
2. Measure before trust-bearing validation using the existing per-selection
   event/commit, object, individual-object, attachment, and total-byte budgets.
3. Count memory-stream commits against `max-events`; version-0 has zero valid
   attachments, but object and total limits still apply without exemption.
4. Discount only objects reachable from that exact memory stream's previous
   accepted ref, preserving the existing per-selection budget semantics.
5. Invoke WP02's strict stream validator against the quarantine Git directory:
   exact owner/stream agreement, two-file trees, signatures, parents, sequence,
   `previous`, record bounds, and kind-specific shape must all pass.
6. Treat any extra tree entry, attachment, wrong owner, wrong stream, invalid
   signature, discontinuity, or malformed available lifecycle target as a
   scoped invalid memory outcome, never as a recoverable gap.
7. Build dependency context from already accepted memory plus promotable
   transaction candidates and use WP03's exact lifecycle/anchor/evidence
   classifications without reading quarantine as accepted state.
8. Preserve exact absent targets, anchors, and evidence as dependency-missing
   outcomes with typed details; do not label them evidence-resolved.
9. Ensure one failed memory affects only requests that explicitly depend on it;
   unrelated actor, proposal, and memory outcomes remain independently valid.
10. Copy and verify all objects for the promotable set, then include accepted
    memory refs in the existing single expected-old-value update-ref transaction.
11. Preserve every failed stream's previous accepted ref byte-for-byte and
    keep accepted-memory enumeration limited to promoted/local verified roots.
12. Include memory roots and objects in pending anchors, validated receipts,
    acceptance reconciliation, cleanup, and recovered shallow-boundary release.

**Files**

- Modify `quarantine.go` for measurement dispatch, WP02/WP03 validation,
  dependency propagation, independent outcomes, and promotion assembly.
- Modify `shallow.go` narrowly where durable transaction records parse accepted
  memory promotions and acceptance state.
- Add strict hostile-stream tests to `memory_replication_test.go`.

**Validation**

- Run `go test ./... -run 'TestMemoryReplication.*(Quarantine|Validation|Promotion)'`.
- Assert every memory outcome uses `kind=memory`, the full stream ID, and one
  stable status/diagnostic without exposing paths, credentials, or content.
- Record completion with
  `spec-kitty agent tasks mark-status T028 --status done` only after focused
  tests prove no quarantine ref enters normal projection.

### Subtask T029: Prove mixed isolation, hostile budgets, and crash transactions

**Purpose**

Lock the security boundary through real-remotes tests combining actor,
proposal, and memory outcomes across budget rejection and every durable
transaction interruption point.

**Steps**

1. Create `memory_replication_test.go` around real temporary repositories, two
   independent actors, and an ordinary local bare remote; avoid transport mocks.
2. Publish valid actor and proposal histories plus multiple valid memory
   streams, an invalid-signature stream, an owner mismatch, and a discontinuity.
3. Run one mixed exact selection and assert each actor/proposal/memory receives
   an independent outcome keyed by kind plus full ID.
4. Prove an invalid, missing, or over-budget memory cannot suppress or alter
   independently valid actor/proposal events or an unrelated valid memory.
5. Snapshot collaboration event IDs and encoded payloads before import and
   require byte-identical projection after every memory-only failure.
6. Exercise memory commit count, object count, largest object, and total bytes
   at one below, exactly at, and one above each measured value.
7. Prove the attachment measurement remains zero for valid version-0 memory
   and that an attempted attachment fails strict tree validation.
8. Confirm pre-existing accepted memory objects are discounted only for the
   same stream, while shared/unreferenced objects do not weaken another stream.
9. Reuse the production interruption hooks at after-fetch, after-measure,
   after-copy, immediately-before-promote, and durable receipt boundaries.
10. At every pre-promotion interruption, assert accepted actor, proposal, and
    memory refs retain their exact old values and projection remains fail-closed.
11. At post-copy interruption, assert residue is never accepted; retry must
    reconcile pending anchors/receipts and converge or report partial success.
12. Inject ref-transaction and completion-record failures and verify all-or-old
    atomic promotion, truthful diagnostics, and idempotent retry behavior.
13. Test two transactions with shared objects so one transaction cannot clear
    another transaction's pending denial or acceptance state.
14. Keep fixture data free of keys, tokens, environment values, raw transcripts,
    external services, model APIs, Docker, and nondeterministic network access.

**Files**

- Build the mixed hostile and failure-injection suite in the new
  `memory_replication_test.go` (roughly 500–800 lines total by this subtask).
- Modify production seams only in the already owned replication files; do not
  create test-only bypasses for quarantine, validation, or promotion.

**Validation**

- Run `go test ./... -run 'TestMemoryReplication.*(Mixed|Budget|Transaction)'`
  repeatedly and compare accepted refs before/after each injected failure.
- Run the existing replication and shallow interruption suites unchanged.
- Record completion with
  `spec-kitty agent tasks mark-status T029 --status done` after crash-boundary
  tests prove collaboration isolation and transactional retry.

### Subtask T030: Report exact shallow gaps and prove fresh-clone recovery

**Purpose**

Classify unavailable memory dependencies with exact selected recovery actions,
then prove a credential-disabled fresh clone converges from canonical remote
refs without copied keys, indexes, services, or vendor state.

**Steps**

1. Add memory dependency kinds for stream predecessor, lifecycle target, Git
   anchor/path object, typed evidence, and supplying memory stream as needed.
2. Populate every `ShallowDependencyGap` with operation, owning full memory and
   stream IDs, exact missing full ID, kind, selected remote, required ref when
   known, and one precise recovery action.
3. Classify only unavailable otherwise-valid objects/facts as shallow gaps;
   malformed IDs, wrong object types, signature failures, owner mismatches, and
   invalid relationships remain ordinary invalid-data failures.
4. When a supplying stream is not selected, direct the operator to add the
   exact `--memory <full-stream-id>` while preserving existing selectors and
   budgets; never silently edit the selection.
5. Offer `nh sync <remote> --recover-shallow` only when the repository is
   actually shallow and the explicitly saved selection authorizes the exact
   supplying stream/ref under its existing budgets.
6. Extend recovery subset construction so a recorded memory gap selects only
   its exact authorized memory stream plus required selected dependencies.
7. Bind recovery to unchanged saved selection bytes, route the fetch through
   quarantine/budgets/validation/promotion, then rerun the complete original
   verification before clearing the durable gap.
8. Preserve the gap and accepted refs after interruption, invalid recovery
   data, selection drift, or budget failure; retry remains deterministic.
9. Build a publisher corpus with separate collaboration and memory roots,
   lifecycle links, anchors, and typed evidence, then publish to a bare remote.
10. Clone the project branch freshly with Git credentials/config disabled and
    no copied `.git/nh`, actor private keys, index, embeddings, or adapter state.
11. Save an exact memory selection, synchronize through the public command,
    and compare full memory IDs, stream ownership, anchors, lifecycle edges,
    dependency status, and exact filtered projection to the publisher.
12. Exercise an initially shallow/missing dependency, assert the full gap and
    exact recovery guidance, recover it explicitly, and require convergence.
13. Prove memory-only sync needs no local signer and a collaboration-only clone
    reproduces its prior projection with zero selected memory refs.
14. Use local-path transport only; no credential helper, SSH agent, global Git config, service, model API, or Docker may participate.

**Files**

- Modify `shallow.go` for memory gap kinds, guidance, recovery subsets, and
  complete-verification replay.
- Extend `memory_replication_test.go` with shallow and credential-disabled
  fresh-clone convergence scenarios.

**Validation**

- Run `go test ./... -run 'TestMemoryReplication.*(Shallow|FreshClone)'`.
- Run credential-disabled tests with isolated Git config and assert no private
  identity or local memory index is created as a side effect of synchronization.
- Record completion with
  `spec-kitty agent tasks mark-status T030 --status done` only after exact gaps,
  bounded recovery, and publisher/clone convergence all pass.

## Definition of Done

- T026 has an event-sourced `done` record and optional sorted `memories` selectors,
  strict full stream-ID validation, legacy version-1 JSON, and exact-selection
  compatibility are proven.
- T027 has an event-sourced `done` record and explicit/`--all` discovery, exact
  owner-bound requests, accepted ref parsing, and post-collaboration memory
  publication match the contract.
- T028 has an event-sourced `done` record and every memory stream is fetched,
  measured, strictly validated, dependency-checked, copied, and promoted only
  through the existing quarantine transaction.
- T029 has an event-sourced `done` record and mixed actor/proposal/memory,
  hostile-budget, shared-object, receipt, and crash-boundary tests prove
  independent outcomes and all-or-old accepted refs.
- T030 has an event-sourced `done` record and exact typed gaps, explicit bounded
  recovery, credential-disabled fresh-clone convergence, and collaboration-only
  compatibility are demonstrated.
- Memory outcomes use `kind=memory` plus full stream IDs; invalid, missing,
  over-budget, interrupted, or unselected memory never changes independently
  valid collaboration projections or failed streams' accepted refs.
- Existing selection JSON without `memories`, collaboration event bytes/IDs,
  actor/proposal selection semantics, and bounded compatibility-all remain valid.
- No quarantine ref becomes a collection/recall source, no selection implies
  policy trust, and no absent dependency is reported resolved.
- Only `replication.go`, `quarantine.go`, `shallow.go`, `commands.go`, and the new
  `memory_replication_test.go` change; record a one-line rationale before any
  necessary out-of-map edit.
- `git diff --check`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and
  `go build ./...` all pass.

## Risks

- Stream-ID-only selection must resolve a unique remote owner/ref; fail the
  memory request on ambiguity instead of choosing by advertisement order.
- Reusing actor event-count logic can skip memory commit limits; count stream
  commits explicitly while retaining every existing budget dimension.
- Validation can accidentally read copied or quarantine objects as accepted;
  pass explicit Git directories and accepted/candidate projection contexts.
- One global transaction error can roll back unrelated collaboration; preserve
  per-request outcomes and propagate failure only along exact dependencies.
- Durable receipt parsers can reject older records after adding memory refs;
  retain prior versions and extend only current promotion grammar.
- Shallow recovery can overfetch or silently mutate selection; require exact
  saved authorization, unchanged bytes, narrow subsets, and full replay.
- Publishing memory before collaboration refs changes established phase order;
  keep actor, proposal, then memory publication with safe phase diagnostics.
- Fresh-clone tests can pass accidentally through global credentials or copied
  private state; isolate Git config and inspect the clone's private directories.

## Reviewer Guidance

Review this package as an extension of one mature transaction boundary, not as
a parallel memory sync implementation. Trace one exact memory selector from
saved JSON through remote advertisement, owner/ref resolution, quarantine,
per-stream measurement, WP02 verification, WP03 dependency classification,
object copy, durable receipts, atomic accepted ref, and publication. At each
step verify the full stream ID and owner remain bound and quarantine is never a
normal projection source.

Exercise a mixed remote containing a valid actor, valid proposal, valid memory,
invalid memory, over-budget memory, and dependency-missing memory. The three
valid independent outcomes must promote; each failed memory must retain its
old accepted ref and leave collaboration bytes unchanged. Inspect crash hooks,
pending anchors, shared objects, shallow gaps, and retry behavior. Reject short
IDs, global memory failure, trust/selection conflation, vague recovery advice,
credential leakage, or fresh-clone convergence that relies on copied keys or
indexes.

Implementation command:

`spec-kitty agent action implement WP06 --agent <name>`
