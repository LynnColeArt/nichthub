# WP02 Review Feedback — Cycle 1

## Verdict

Rejected. The production implementation has a coherent repository-native
shape and all executed quality gates pass, but the work package's mandatory
adversarial acceptance surface is incomplete. WP02 cannot be approved until
the tests prove the hostile Git/ref and isolation invariants enumerated by
T008–T010 and the Definition of Done.

## Findings

### 1. Required corruption and continuity matrix is incomplete (blocking)

`memory_store_test.go` currently covers an extra tree entry, a wrong ref owner,
a wrong stream, one signed-`previous` mismatch, one merge commit, and one
invalid Ed25519 signature. It does not exercise the other independently
required failure modes from T009/T010:

- missing `memory.json` and missing `signature`;
- nested paths, non-`100644` modes/types, and duplicate-like malformed trees;
- invalid or noncanonical raw-base64 and unknown/invalid envelope bytes;
- unavailable/wrong/skipped predecessors and root-with-previous cases;
- zero, gap, duplicate, and decreasing sequences;
- actor/public-key mismatch and cross-stream grafting.

Add one-invariant-per-fixture tests through `loadMemoryStreamAt` or
`collectMemories`. For every rejected append/load case, snapshot all
`refs/nh/*` before the operation and assert exact equality afterward. The tests
must invoke the production loader/collector rather than assert synthetic
literal structures.

### 2. Accepted-stream and deterministic collection guarantees are not proven (blocking)

The collection test uses one memory duplicated through local and accepted
refs. A single output cannot prove deterministic protocol ordering or
equivalence across source discovery order. There is also no test for:

- local-only, accepted-only, and mixed source sets containing multiple actors,
  streams, and sequences in different ref-creation orders;
- malformed refs inside the exact local/accepted memory namespaces while
  unrelated accepted actor/proposal refs are ignored;
- `loadMemoryStreamAt` against an explicit Git directory, the reuse seam WP06
  requires for quarantine validation;
- the replication-pending guard preventing an accepted memory from becoming
  readable;
- quarantine refs and unreferenced objects being excluded as canonical input.

Add black-box Git fixtures that cover these paths and compare the full ordered
`StoredMemory` identity sequence, not merely its length or first ID.

### 3. Collaboration and privacy isolation evidence is incomplete (blocking)

The happy-path test proves that valid memory does not alter one collaboration
event. T010 additionally requires the inverse failure boundary: after a memory
ref is corrupted or removed, `collectEvents` must still return byte-identical
payloads, IDs, count, ordering, and actor/proposal refs while only the
memory-specific call reports failure. Add that fixture, plus an explicit actor
with no collaboration ref.

Also place unique sentinel values in environment variables, keyring/private
state, the working tree, and `.git/nh/memory/index-v0.json`, then prove neither
collection nor diagnostics expose those sentinels. Existing index and
working-tree files alone do not establish keyring/environment isolation or
diagnostic redaction.

### 4. Requirement-level test traceability is incomplete (blocking)

The WP declares FR-006, FR-014–FR-017, FR-023, NFR-005, NFR-010, C-003, and
C-006, but `memory_store_test.go` only stores `record` envelopes and does not
show that the storage layer accepts and round-trips every WP01-valid operation
without prematurely resolving lifecycle targets. Add storage round trips for
`supersede`, `retract`, and `challenge`, including a missing target that remains
cryptographically valid at this layer. Keep lifecycle projection assertions in
WP03.

For this WP's portion of NFR-005, prove deterministic storage collection under
different ref insertion/enumeration orders; WP03 will own full lifecycle
projection convergence.

## Quality-gate evidence

The following commands passed during independent review:

- `go test ./... -run 'TestMemory(Stream|Store|Ref)' -count=10`
- `go test ./...`
- `go test -race ./... -run 'TestMemory(Stream|Store|Ref)'`
- `go vet ./...`
- `go build ./...`
- `git diff --check kitty/mission-agent-memory-protocol-01M19TMH..HEAD`

Focused coverage reached the production paths that exist, but several error
branches remained unexecuted; passing gates do not substitute for the explicit
acceptance matrix above.

## Anti-pattern checklist

1. Dead code: N/A — WP02 adds internal seams intentionally consumed by later
   dependent WPs; no new exported function is orphaned.
2. Synthetic-fixture tests: PASS — existing tests invoke production paths.
3. Silent empty return: PASS — the only empty collection return is the specified
   collaboration-only/no-memory result.
4. FR coverage: FAIL — findings 2–4 leave declared behaviors unproved.
5. Frozen surface: PASS — commit `92894f7` changes only
   `memory_store.go` and `memory_store_test.go`.
6. Locked decisions: PASS — no service, network, Docker, mutable database,
   collaboration-ref coupling, unsigned merge, or private-index authority was
   introduced.
7. Shared-file ownership: PASS — both changed files are exclusive to WP02 in
   `lanes.json`.
8. Production fragility: N/A — Go implementation adds no panic/raise-style
   request path; failures are returned explicitly.

## Re-review acceptance

Re-run every gate above after adding the missing fixtures. Approval requires
all explicit T006–T010 validation bullets to have direct production-path
evidence; do not weaken the mission contracts or move lifecycle policy into
`memory_store.go` to satisfy these tests.
