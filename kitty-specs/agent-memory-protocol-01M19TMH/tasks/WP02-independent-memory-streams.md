---
work_package_id: "WP02"
title: "Independent Memory Streams"
dependencies: ["WP01"]
requirement_refs:
  - "FR-006"
  - "FR-014"
  - "FR-015"
  - "FR-016"
  - "FR-017"
  - "FR-023"
  - "NFR-005"
  - "NFR-010"
  - "C-003"
  - "C-006"
subtasks: ["T006", "T007", "T008", "T009", "T010"]
owned_files:
  - "memory_store.go"
  - "memory_store_test.go"
authoritative_surface: "memory_store.go"
create_intent:
  - "memory_store.go"
  - "memory_store_test.go"
execution_mode: "code_change"
task_type: "implement"
agent_profile: "implementer-ivan"
role: "implementer"
agent: "codex"
model: ""
---

# Work Package Prompt: WP02 – Independent Memory Streams

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Implement the repository-native storage boundary for signed memory envelopes:
safe actor-owned refs, deterministic default streams, linear two-file Git
commits appended with compare-and-swap, and deterministic collection of local
and accepted streams. Memory storage must remain independent of Nichthub's
collaboration event chains in both directions.

## Context

WP01 supplies the exact `nh-memory/0` envelope, canonical encoding, memory ID,
signing, and verification primitives. Use those primitives as the only payload
authority; do not duplicate wire validation in this package. The storage
contract adds Git/ref invariants around already verified envelopes.

Read `plan.md` IC-02, `data-model.md` under “Canonical stream model,” research
decisions D1, D2, and D10, and both `contracts/memory-wire-v0.md` and
`contracts/memory-replication-v0.md`. The required namespaces are:

```text
refs/nh/memory/<full-actor-fingerprint>/<64-hex-stream-digest>
refs/nh/remotes/<remote>/memory/<full-actor-fingerprint>/<64-hex-stream-digest>
```

The `sha256:` prefix is part of the semantic stream ID carried in the envelope,
but the local/accepted ref suffix is only its 64 lowercase hexadecimal digits.
Derive the default stream as SHA-256 over the exact domain-separated bytes
`nh-memory-stream-v0\x00<actor>\x00default`, rendered as a full `sha256:` ID.

Use `store.go` as a mechanical precedent for object creation and CAS updates,
`git.go` for Git execution, `quarantine.go` for accepted-ref grammar and pending
acceptance boundaries, and WP01 for memory identity. Do not edit those files in
this WP. The only owned and planned-new files are `memory_store.go` and
`memory_store_test.go`.

No index, policy, CLI routing, lifecycle projection, publication, selection,
or quarantine promotion is implemented here. WP03 consumes verified memories
for lifecycle projection, and WP06 later reuses this stream validator inside
quarantine. This WP must expose narrow package-level seams usable by both.

### Subtask T006: Define safe ref grammar and deterministic default stream

**Purpose**

Create the address boundary that converts validated semantic actor/stream IDs
to Git refs and parses local or accepted refs without accepting ref injection,
ambiguous ownership, or shortened trust-bearing IDs.

**Steps**

1. Start with table-driven red tests in `memory_store_test.go` for the exact
   local and accepted ref forms documented above.
2. Cover full valid actor fingerprints and full `sha256:` stream IDs, including
   leading zeroes in the digest; no short-ID convenience belongs in storage.
3. Define a focused default-stream helper using `crypto/sha256` and the exact
   byte concatenation `nh-memory-stream-v0`, NUL, actor, NUL, `default`.
4. Assert the derivation against a fixed golden actor/vector computed in the
   test independently enough to catch a changed domain separator or delimiter.
5. Validate the actor with the existing actor-fingerprint validator before it
   reaches ref construction.
6. Validate a stream as exactly `sha256:` plus 64 lowercase hexadecimal digits;
   reject uppercase, truncation, extra prefixes, whitespace, and empty input.
7. Render only the digest component in the ref, never the literal `sha256:`
   prefix, because a colon is not part of the contract's ref path.
8. Parse local refs into exact actor plus restored full stream ID; reject extra
   path components and noncanonical spelling.
9. Parse accepted refs into remote, actor, and restored full stream ID, reusing
   the existing remote-name rules rather than creating a looser validator.
10. Reject traversal and Git-ref metacharacter families: `/`, `..`, `.lock`,
    `@{`, backslash, controls, spaces, newline, and unexpected separators.
11. Treat `git check-ref-format` as optional defense in depth only after the
    semantic grammar passes; semantic IDs remain the protocol authority.
12. Ensure a malformed ref underneath the exact memory namespace is surfaced
    as malformed memory, not silently reclassified as collaboration data.
13. Keep local and accepted parsing distinct so callers cannot accidentally
    append to an accepted remote namespace.
14. Return safe diagnostics naming the invalid public field/ref without
    reflecting control bytes unsafely or exposing any private material.

**Files**

- Create `memory_store.go` for stream constants and ref helpers, approximately
  80–130 lines at this stage.
- Create `memory_store_test.go` for grammar matrices and the golden default
  derivation, approximately 120–180 lines at this stage.

**Validation**

- Focused grammar tests accept only the two canonical namespaces.
- A valid ref round-trips actor and full stream ID exactly.
- The same actor always derives the same default stream; different actors do
  not collide in representative fixtures.
- Every hostile segment fixture fails before any Git object or ref mutation.

### Subtask T007: Append two-file Git commits with compare-and-swap

**Purpose**

Persist a WP01-signed envelope as one canonical Git commit and advance exactly
one local memory ref atomically, without merge commits, blind overwrites, or
dependence on the actor's collaboration history.

**Steps**

1. Specify first-append, ordinary append, stale-head, and retry behavior with
   temporary Git repositories before implementing the writer.
2. Accept a validated/signed WP01 memory value plus signer identity through the
   narrowest interface compatible with WP01's landed API.
3. Resolve only the local memory ref derived from the envelope actor and stream;
   never read `refs/nh/actors/<actor>` to determine sequence or predecessor.
4. Read the expected current stream head once and use it as both the parent
   basis and the expected old OID for the final ref transaction.
5. For a nonempty stream, load and verify the current head before constructing
   the next commit; reject an unreadable or owner-mismatched head.
6. Require sequence 1 and empty `previous` for the first envelope.
7. Require sequence `head.sequence + 1` and `previous == head.memoryID` for
   every subsequent envelope before writing the new commit.
8. Write the exact WP01 canonical payload bytes as blob `memory.json`.
9. Encode the exact Ed25519 signature as raw-base64 blob `signature`, matching
   the frozen wire contract and existing event-storage convention.
10. Build a tree containing exactly two `100644 blob` entries named
    `memory.json` and `signature`, in deterministic order.
11. Create a root commit with no parent for sequence 1 and exactly one parent
    equal to the previous stream head for every later sequence.
12. Use safe deterministic commit metadata analogous to `appendEvent`; the Git
    commit ID is transport identity and must never replace the memory ID.
13. Advance the local ref with `git update-ref <ref> <new> <expected-old>`, using
    the empty old value for first creation across SHA-1 and SHA-256 repos.
14. On CAS failure, return a reload-and-retry diagnostic and leave the winning
    ref unchanged; an unreachable object written before failure is acceptable.
15. Never synthesize a merge, renumber a prebuilt envelope, overwrite a ref,
    or silently retry with new signed bytes inside the storage function.
16. Return a stored-memory value containing full memory ID, commit ID, envelope,
    exact payload, and signature for later projection/replication consumers.

**Files**

- Extend `memory_store.go` with append orchestration and Git object creation,
  approximately 120–180 additional lines.
- Extend `memory_store_test.go` with real-repository CAS and tree fixtures,
  approximately 140–220 additional lines.

**Validation**

- `git ls-tree` shows exactly the two required regular blobs.
- `git show -s --format=%P` shows zero parents for the root and one exact parent
  for each successor.
- Two appenders built from one observed head yield one success and one explicit
  CAS failure, with no ref rollback and no merge commit.
- A memory append succeeds when no collaboration actor ref exists.

### Subtask T008: Load and collect local and accepted memory streams independently

**Purpose**

Provide reusable read paths for one stored memory, one stream, and the complete
verified set reachable from canonical local/accepted memory refs, while keeping
memory availability independent from collaboration refs and private indexes.

**Steps**

1. Add a loader equivalent in responsibility to `loadStoredEventAt`, but for
   WP01 memory envelopes and the exact two-file memory tree.
2. Support an explicit Git directory parameter so WP06 can validate the same
   canonical stream shape inside quarantine without copying interpretation.
3. Read `memory.json` and `signature` as bounded blobs and decode raw-base64
   strictly; pass exact payload/signature bytes to WP01 verification.
4. Do not perform lifecycle-target resolution here; a cryptographically valid
   lifecycle envelope may arrive before its target and belongs in WP03.
5. Add a stream loader that walks one head oldest-first and returns every
   stored memory with its commit identity preserved.
6. Keep stream loading independent from repository `HEAD`, the working tree,
   policy files, actor keyring files, and collaboration event collection.
7. Enumerate canonical sources only beneath `refs/nh/memory` and accepted
   `refs/nh/remotes/<remote>/memory`; ignore actor/proposal refs as other kinds.
8. Parse every memory ref with T006 helpers and retain the expected owner and
   stream identity beside the advertised head for T009 validation.
9. Fail closed on malformed refs inside an exact memory namespace; safely skip
   unrelated accepted actor/proposal namespaces.
10. Ensure accepted commits still pass the existing replication-pending guard
    before they become readable from the main repository.
11. Never treat quarantine refs, unreferenced Git objects, saved selections,
    `.git/nh/memory/index-v0.json`, or any other private state as a source.
12. Deduplicate a commit/memory reached through multiple canonical refs without
    dropping distinct stream-owner validation for each source.
13. Return deterministic ordering using stable protocol keys, not Git ref
    enumeration order, map iteration, wall-clock load time, or delivery order.
14. Keep the collection API memory-specific; it must not call `collectEvents`,
    and `collectEvents` must not call it.
15. Preserve exact, actionable errors for unreadable commits and invalid
    signatures without echoing payload content or filesystem-private values.

**Files**

- Extend `memory_store.go` with stored-memory, stream-source, load, and collect
  seams, approximately 140–210 additional lines.
- Extend `memory_store_test.go` with local/accepted/deduplicated source fixtures,
  approximately 160–240 additional lines.

**Validation**

- Local-only, accepted-only, and mixed canonical refs return the same verified
  records in deterministic order when their source sets are equivalent.
- Memory collection works in a repository containing zero actor refs.
- Collaboration-only repositories return an empty memory set without error.
- Private index/quarantine/unreferenced fixtures never enter collection.

### Subtask T009: Verify parent, sequence, previous, owner, and stream continuity

**Purpose**

Make an accepted/local ref trustworthy as one actor-owned linear memory stream
by validating both the Git parent chain and the signed protocol chain against
the ref-derived owner and stream identity.

**Steps**

1. Validate each source stream independently before any of its envelopes are
   returned as accepted input to later projection.
2. Require at least one commit for every advertised stream head.
3. Inspect commit parents explicitly rather than assuming `rev-list` output
   proves linearity.
4. Require the first chronological commit to have no parents and every later
   commit to have exactly one parent equal to the prior commit ID.
5. Reject merge commits, parent skips, graft-like visible discontinuities, and
   any head whose required predecessor is unavailable.
6. Validate every commit tree as exactly two unique `100644 blob` entries named
   `memory.json` and `signature`; reject trees, symlinks, extras, and omissions.
7. Require every verified envelope actor to equal the full actor fingerprint
   parsed from its source ref.
8. Require the embedded public-key fingerprint validation from WP01 to remain
   successful; a display name never participates in ownership.
9. Require every envelope stream to equal the full stream ID parsed from the
   source ref, including on the first commit.
10. Require chronological sequences to be exactly 1, 2, 3, … with no zero,
    gap, duplicate, decrease, or cross-ref continuation.
11. Require empty `previous` only at sequence 1; every successor must name the
    exact WP01 memory ID of the immediately preceding envelope.
12. Validate signed previous continuity separately from Git parent continuity;
    neither one substitutes for the other.
13. Reject a commit valid under another actor/stream when grafted beneath this
    ref, even if its signature and tree are otherwise valid.
14. Return errors that identify the source ref, sequence/commit, and violated
    invariant using safe public IDs suitable for WP06 diagnostics.
15. Keep lifecycle same-author/cross-author target rules out of this validator;
    those set-level, order-independent rules belong to WP03.
16. Expose validation at an explicit Git directory so the same rules can guard
    local refs now and quarantined streams later.

**Files**

- Extend `memory_store.go` with exact tree and chain validation, approximately
  100–160 additional lines.
- Extend `memory_store_test.go` with one-invariant-per-fixture corruption tests,
  approximately 200–300 additional lines.

**Validation**

- A canonical multi-record stream validates in both SHA-1-default fixtures and
  any available SHA-256-repository fixture without assuming OID length.
- Mutating only Git parents or only signed `previous` fails distinctly.
- Owner, stream, and sequence corruption each fail at the storage boundary.
- The validator has no policy, lifecycle projection, or network side effects.

### Subtask T010: Prove concurrency, corruption, ref safety, privacy, and isolation

**Purpose**

Close the storage threat model with adversarial tests demonstrating that stale
writers, malformed Git data, hostile ref components, and private-state residue
cannot corrupt a stream or alter collaboration behavior.

**Steps**

1. Build a deterministic stale-writer fixture where two valid next envelopes
   share one observed parent and expected old ref value.
2. Assert exactly one CAS update succeeds, the loser receives reload/retry
   guidance, and the final ref names the winner without a merge.
3. Retry from the winning head with a newly sequenced/signed envelope and prove
   normal linear progress resumes.
4. Construct commits with missing `memory.json`, missing `signature`, an extra
   blob, a nested path, wrong mode/type, duplicate-like malformed tree input,
   invalid base64, invalid signature, and unknown/invalid envelope bytes.
5. Construct separate parent-chain failures: merge parent, wrong parent, missing
   predecessor, and signed `previous` mismatch.
6. Construct actor/stream/sequence failures: wrong ref owner, wrong key owner,
   wrong embedded stream, zero/gap/duplicate sequence, and cross-stream graft.
7. Exercise ref injection strings containing controls, newlines, traversal,
   slash expansion, `.lock`, `@{`, whitespace, uppercase digest, and short IDs.
8. Snapshot every `refs/nh` value before rejected append/load cases and assert
   failures do not mutate any canonical ref.
9. Place plausible memory JSON and secrets in working-tree files, environment
   variables, keyring/private-state files, and `.git/nh/memory/index-v0.json`;
   assert storage never scans or publishes them.
10. Confirm diagnostics do not echo private-key bytes, tokens, environment
    values, or unbounded hostile payload text.
11. Record collaboration events, snapshot their exact payload bytes, IDs, and
    `collectEvents` projection, then append and collect memory.
12. Assert collaboration bytes, IDs, event count, ordering, and actor/proposal
    refs remain byte-for-byte unchanged after memory operations.
13. Create valid memory for an actor with no `refs/nh/actors/<actor>` ref and
    prove append/load/collect still succeeds.
14. Corrupt or remove a memory ref and prove independently valid collaboration
    remains readable; the memory-specific call alone reports its failure.
15. Run focused tests repeatedly and under the race detector; avoid sleeps by
    coordinating stale writers through explicit expected-head fixtures.

**Files**

- Complete adversarial coverage in `memory_store_test.go`; no production file
  beyond `memory_store.go` may be changed for test hooks or convenience.
- Keep fixtures ephemeral and repository-local; do not add tracked golden files,
  private state, secrets, or new dependencies.

**Validation**

- `go test ./... -run 'TestMemory(Stream|Store|Ref)' -count=10` is stable.
- `go test -race ./... -run 'TestMemory(Stream|Store|Ref)'` reports no race.
- Existing collaboration tests remain green with identical golden IDs.
- `git diff --check` reports no prompt or implementation whitespace defects.

## Definition of Done

- `memory_store.go` is the sole implementation surface and
  `memory_store_test.go` is the sole test surface changed by this WP.
- T006 proves canonical local/accepted ref parsing, hostile-segment rejection,
  and the exact deterministic default-stream derivation.
- T007 proves strict two-blob commits, linear Git parents, and CAS behavior for
  first, ordinary, stale, and retry append paths.
- T008 proves deterministic verified collection from local and accepted memory
  refs only, with no collaboration, quarantine, index, or private-state source.
- T009 proves ref owner/stream, public-key owner, Git parents, sequence, signed
  previous, strict tree shape, and signature continuity independently.
- T010 proves concurrency safety, corruption rejection, ref-injection defense,
  secret/private-state isolation, and unchanged collaboration projection.
- No existing `Event` wire field, collaboration ref, event payload byte, or
  public event ID changes.
- No service, model API, provider API, Docker dependency, network call, mutable
  canonical database, or new Go dependency is introduced.
- Focused tests, `go test ./...`, `go test -race ./...`, `go vet ./...`,
  `go build ./...`, and `git diff --check` all pass.
- Completion evidence is recorded event-sourced, not by editing checkboxes:
  run `spec-kitty agent tasks mark-status T006 --status done` through
  `spec-kitty agent tasks mark-status T010 --status done` separately after each
  subtask's validation evidence exists.

## Risks

- **Ref/path confusion**: derive refs only from validated full semantic IDs,
  round-trip parse them, and keep local append refs distinct from accepted refs.
- **Concurrent lost update**: bind every ref advance to the observed old OID;
  expose CAS conflict instead of blind retry or merge.
- **Dual-chain ambiguity**: validate Git parent continuity and signed
  sequence/previous continuity separately, then require both.
- **Cross-stream grafting**: compare actor and stream in every envelope against
  the exact source ref, not only the stream head.
- **Premature coupling**: keep lifecycle target resolution in WP03 and
  replication selection/promotion in WP06 while exposing reusable `gitDir`
  validation seams.
- **Collaboration regression**: never modify or broaden `collectEvents`; prove
  event payloads, IDs, refs, and projections are unchanged by memory failures.
- **Private data leakage**: enumerate canonical refs only and sanitize errors;
  private indexes, keyrings, environment, and working-tree files are not input.

## Reviewer Guidance

Review against the wire contract rather than only the happy path. Independently
inspect a produced commit with `git ls-tree` and `git show`, force a stale CAS,
and alter Git-parent and signed-previous continuity one at a time. Confirm that
all ref components come from full validated IDs and that accepted-ref parsing
cannot be used as a local append target.

Require direct isolation evidence: memory must work without an actor
collaboration ref, collaboration must remain readable when memory is malformed,
and neither collector may call or enumerate the other's namespace. Reject any
implementation that treats a Git commit ID as a memory ID, trusts `rev-list`
without exact parent/tree checks, reads a private index as canonical, silently
repairs a stale signed envelope, or resolves lifecycle semantics in storage.

## Implementation Command

```bash
spec-kitty agent action implement WP02 --agent <name>
```
