---
work_package_id: "WP04"
title: "Rebuildable Exact and Lexical Index"
dependencies: ["WP03"]
requirement_refs:
  - FR-008
  - FR-009
  - FR-019
  - FR-020
  - NFR-003
  - NFR-007
  - NFR-008
  - NFR-009
  - NFR-011
  - C-008
subtasks: ["T016", "T017", "T018", "T019", "T020"]
owned_files:
  - "memory_index.go"
  - "memory_index_test.go"
authoritative_surface: "memory_index.go"
create_intent:
  - "memory_index.go"
  - "memory_index_test.go"
execution_mode: "code_change"
task_type: "implement"
agent_profile: "implementer-ivan"
role: "implementer"
agent: "codex"
model: ""
---

# Work Package Prompt: WP04 – Rebuildable Exact and Lexical Index

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Implement a deterministic, private, disposable version-0 memory index that is
rebuilt solely from verified local and accepted memory sources plus an exact
policy input. Provide fast exact filters and deterministic Unicode lexical
lookup without making the index, its tokenization, or any optional ranking a
canonical claim about memory identity, lifecycle, trust, truth, or authority.

## Context

WP01 defines strictly validated signed memory envelopes, WP02 owns independent
local and accepted stream collection, and WP03 derives order-independent
record projections with lifecycle, applicability, evidence, policy trust, and
full provenance. WP04 consumes those verified/projection interfaces and stores
only a reproducible cache. It must not enumerate quarantine refs, fetch the
network, repair or append canonical memory, inspect collaboration refs as
memory, execute recalled content, or read actor private keys.

Read `plan.md` IC-05, `data-model.md` under “Disposable local entities” and
“Recall model,” research decision D7, and `contracts/memory-cli-v0.md`. The
version-0 private path is exactly:

```text
.git/nh/memory/index-v0.json
```

Use the repository's resolved Git directory rather than assuming `.git` is a
directory; linked worktrees must resolve to their own private administrative
area correctly. Reuse existing owner-only private-directory and atomic-write
helpers where their contracts fit. The persisted JSON has no wall-clock field,
random value, host path, map-order dependency, or other nondeterministic byte.

The source fingerprint commits the cache to the sorted exact local and accepted
memory ref names and head OIDs plus the exact policy digest used by WP03. It
must change when any of those inputs changes and must not include quarantine
refs, working-tree state, credentials, index bytes, or network state. A missing,
incompatible, corrupt, or stale cache is disposable and may be rebuilt locally;
it can never override or mutate verified sources.

Exact filtering covers the WP05 recall request dimensions needed from indexed
rows: commit/applicability context, exact subject, path, normalized topic,
memory kind, full actor ID, lifecycle state, and trust class. Lexical lookup is
an additional deterministic intersection over normalized Unicode word tokens.
Filtering and stable ordering happen before WP05 applies cursor, record-count,
and encoded-content bounds; this WP should expose deterministic candidate rows
and avoid inventing a second public recall envelope.

### Subtask T016: Define deterministic index schema, source fingerprint, and private path

**Purpose**: Freeze the internal `MemoryIndexV0` representation and the source
identity rules that make the cache reproducible, stale-detectable, private, and
unambiguously noncanonical.

**Steps**:

1. In `memory_index.go`, define the version-0 persisted index and narrow helper
   types for projected records and token postings.
   - Include exactly a format version, source fingerprint, deterministically
     sorted record rows, and deterministic token-to-memory-ID postings.
   - Preserve every indexed field needed to return the WP03 projection without
     recomputing claims from prose: full memory/stream/actor IDs, anchor and
     paths, applicability, kind/topics, lifecycle plus edge IDs, evidence,
     trust, signature status, content digest, and inert record data.
   - Do not add `builtAt`, access time, host identity, random salt, semantic
     score, summary, private material, or a canonical/truth boolean.

2. Define one private-path resolver for `index-v0.json` below the repository's
   resolved Git directory at `nh/memory/`.
   - Support normal repositories and linked worktrees.
   - Reject unsafe or unresolved administrative paths rather than falling back
     to a tracked working-tree file.
   - Reuse owner-only private-state conventions already present in Nichthub.

3. Define the source-fingerprint preimage and encoding explicitly.
   - Collect exact source pairs as `(canonical ref name, head Git OID)` from
     verified local and accepted memory refs only.
   - Sort pairs bytewise by ref name and then OID, reject duplicates, and add
     the exact policy digest with an explicit domain/version separator.
   - Hash unambiguous length-delimited or canonical encoded bytes with SHA-256
     and render a full digest; never concatenate ambiguous free-form strings.

4. Establish deterministic JSON encoding rules: stable struct field order,
   sorted slices at construction boundaries, canonical empty-list behavior,
   no maps whose iteration order can affect semantic or encoded output, and a
   single trailing-newline policy.

**Files**: `memory_index.go` (new, schema/path/fingerprint foundation, roughly
120–180 lines); `memory_index_test.go` (new, schema/path golden tests, roughly
100–150 lines at this stage).

**Validation**:

- Golden fingerprint cases prove ref-order independence and sensitivity to
  every ref name, head OID, and policy digest change.
- Path tests cover normal repositories and linked worktrees and prove no index
  appears in tracked files.
- Golden JSON tests prove byte stability and absence of timestamps, secrets,
  host-specific paths, and semantic/vector fields.
- Record completion with
  `spec-kitty agent tasks mark-status T016 --status done` only after evidence is
  present.

### Subtask T017: Rebuild entirely from verified local and accepted refs

**Purpose**: Build every index byte from canonical signed memory that passed
WP02 verification and WP03 projection, without network access or dependence on
an earlier index.

**Steps**:

1. Add a rebuild entry point whose explicit inputs include repository/Git
   context, the exact policy commit or digest required by WP03, and injectable
   verified collection/projection seams.
   - Start from scratch even when an old index exists.
   - Use WP02's local and accepted memory collectors rather than walking
     arbitrary Git refs or accepting caller-supplied unverified envelopes.
   - Ensure local and accepted copies of the same canonical stream/record are
     deterministically deduplicated without discarding source ref identity from
     the source fingerprint.

2. Project the complete verified source set through WP03.
   - Copy projection classifications and exact dependency/edge details; do not
     reinterpret natural language or infer truth, contradiction, relevance,
     trust, or authorization.
   - Preserve inactive, untrusted, inapplicable, and dependency-missing rows so
     explicit exact inspection can filter them later.
   - Reject invalid internal projection shapes rather than serializing a cache
     that cannot round-trip safely.

3. Normalize each index row and lexical token posting deterministically.
   - Sort rows by full memory ID as the storage order unless a stronger stable
     projection key is already frozen by WP03.
   - Sort and deduplicate every nested edge, topic, path, evidence, dependency,
     and posting list without mutating upstream canonical values.
   - Keep author content nested as inert data and never use it as executable,
     authorization, file-path, ref, environment, or network input.

4. Encode to a temporary sibling file, flush/close according to existing
   repository atomic-write practice, set owner-only permissions where the host
   supports them, and atomically replace the destination only after success.
   - On collection, projection, encoding, permission, or rename failure, leave
     canonical refs unchanged and never expose a partial valid-looking index.
   - Return safe bounded diagnostics without echoing hostile memory prose or
     secret-bearing process/environment data.

**Files**: `memory_index.go` (rebuild, normalization, atomic persistence,
roughly 180–260 additional lines); `memory_index_test.go` (source-selection and
failure fixtures, roughly 150–220 additional lines).

**Validation**:

- Fixtures prove rebuild includes verified local and accepted refs, deduplicates
  canonical records, and excludes quarantine, malformed, and collaboration refs.
- A rebuild succeeds with network disabled and with no prior index present.
- Injected collect/project/write/rename failures preserve canonical refs and
  never leave a readable partial index.
- Record completion with
  `spec-kitty agent tasks mark-status T017 --status done` after focused tests.

### Subtask T018: Verify corruption and staleness with byte-identical rebuilds

**Purpose**: Make cache verification fail closed and prove that deletion,
corruption, policy drift, or ref drift can only trigger rejection/rebuild—not a
change to canonical memory or projection semantics.

**Steps**:

1. Implement strict index loading and verification.
   - Require version 0, strict JSON decoding, valid UTF-8, known fields only,
     full IDs, normalized/sorted unique collections, and internally consistent
     record/posting membership.
   - Reject truncation, trailing JSON, duplicate semantic entries, missing
     records, orphan token postings, altered content digests, and unsupported
     versions with typed safe errors.

2. Recompute the live source fingerprint from verified local/accepted refs and
   exact policy bytes before treating a persisted cache as usable.
   - Classify missing, corrupt, incompatible, and stale indexes distinctly.
   - Ref-name/head changes and policy-digest changes are stale even if current
     matching queries would happen to return the same records.
   - Never resolve disagreement in favor of cached rows.

3. Rebuild the derived representation independently and compare deterministic
   encoded bytes or a complete normalized structure.
   - Verification must detect altered membership, projection classifications,
     anchors, lifecycle edges, evidence details, trust, inert data, and tokens.
   - Provide a rebuild path for missing/stale/corrupt caches, but keep verify as
     a read-only diagnostic unless its API explicitly requests repair.

4. Add two-clean-rebuild and permutation fixtures.
   - Delete the index, rebuild, retain exact bytes, delete again, rebuild from
     the same sources, and require byte-for-byte equality.
   - Present verified source refs and projection rows in multiple orders and
     require identical bytes and query results.
   - Confirm no wall-clock metadata, temporary filename, or map order leaks
     into persisted output.

**Files**: `memory_index.go` (strict load/verify/staleness paths, roughly
120–180 additional lines); `memory_index_test.go` (corruption and reproducible
rebuild matrix, roughly 180–260 additional lines).

**Validation**:

- Focused tests mutate each persisted field family and prove verification
  rejects the cache without altering any memory ref or policy.
- Source and policy changes invalidate the old cache deterministically.
- Two clean rebuilds and input-order permutations are byte-identical.
- Record completion with
  `spec-kitty agent tasks mark-status T018 --status done` after all checks pass.

### Subtask T019: Implement exact filters and deterministic Unicode lexical lookup

**Purpose**: Expose a deterministic candidate-query API for WP05 recall using
exact indexed classifications first and a portable, embedding-free lexical
intersection second.

**Steps**:

1. Define a narrow internal query type or accept WP03/WP05-compatible filter
   values without defining a competing public `RecallRequestV0` wire contract.
   - Support exact requested commit/applicability context, subject, path,
     normalized topics, kinds, full actor IDs, lifecycle states, and trust
     classes.
   - Treat multiple values within one exact category consistently as a set and
     document whether the category uses union; topics and lexical terms must
     use the specification's required intersection behavior.
   - Reject invalid/empty identifiers and unnormalized values at the boundary.

2. Apply exact filters to projected rows before lexical matching.
   - Use WP03's explicit applicability result/resolver for the requested commit;
     do not infer relevance from content or guess path renames.
   - Match subject IDs, repository slash paths, topics, kinds, full actors,
     lifecycle states, and trust classes exactly—never by prefix or short ID.
   - Preserve explicit inspection of inactive/untrusted rows without changing
     their stored classifications.

3. Implement one deterministic Unicode-aware tokenizer with Go's standard
   library and use it for both indexed content/topics and query strings.
   - Decode valid UTF-8, lowercase deterministically, split words on a clearly
     documented Unicode letter/number boundary, discard empty tokens, sort and
     deduplicate query terms, and avoid locale/host-dependent behavior.
   - Define behavior for combining marks, punctuation, non-Latin scripts,
     emoji, repeated words, and case variants in tests.
   - Intersect sorted posting lists; every normalized query token must match.

4. Return candidate rows in one stable order suitable for WP05's later cursor
   and bounds processing.
   - Follow the plan's ordering inputs—anchor relevance class, lifecycle class,
     signed timestamp, then full memory ID—using explicit total-order tie breaks.
   - Query results must be identical whether read after rebuild or from a
     verified persisted index and must not mutate index state.

**Files**: `memory_index.go` (filter/token/posting/query logic, roughly 160–240
additional lines); `memory_index_test.go` (exact-filter and multilingual token
table tests, roughly 180–260 additional lines).

**Validation**:

- Table tests cover every exact dimension alone and in combination, including
  full-ID rejection, path-at-anchor semantics, multi-topic intersection, and
  explicit untrusted/inactive inspection.
- Unicode fixtures prove stable results across case, punctuation, accents,
  combining marks, scripts, emoji, and repeated terms without external APIs.
- Query order, index source order, and Go map order cannot change result bytes.
- Record completion with
  `spec-kitty agent tasks mark-status T019 --status done` after focused tests.

### Subtask T020: Prove 10k performance, permissions, isolation, and disposable recovery

**Purpose**: Demonstrate that the real version-0 index meets the alpha scale and
security properties under public tests, including clean deletion and rebuild,
without introducing a canonical cache dependency.

**Steps**:

1. Build a deterministic 10,000-record corpus through verified in-memory or
   temporary-repository fixtures representative of all record kinds, actors,
   lifecycle states, paths, topics, trust classes, and Unicode lexical terms.
   - Exercise the actual rebuild and query implementations; do not benchmark a
     mock posting map that bypasses verification/projection/encoding.
   - Use fixed source data and no private keys, access tokens, environment
     dumps, prompts, terminal transcripts, or external network/service calls.

2. Measure a clean rebuild against the NFR-008 target of under 30 seconds on
   the development host and indexed exact plus lexical recall under 1 second
   p95.
   - Keep timing assertions isolated enough to diagnose slow CI hosts while
     preserving a required acceptance proof for the stated development host.
   - Add allocation/complexity guard evidence where useful to prevent obvious
     quadratic scans or repeated full JSON decoding per query.

3. Test filesystem safety and failure recovery.
   - On Unix, require the index file and created private directories to be
     owner-only; on other platforms, assert the strongest portable helper
     contract without pretending POSIX modes are universal.
   - Inject permission-denied directory/file/atomic-replace cases and require
     bounded safe errors, no partial cache, and unchanged canonical sources.
   - Resolve linked-worktree paths and reject any tracked-index placement.

4. Test secret and inert-content isolation.
   - Seed sentinel key/token/environment/transcript values outside explicit
     memory content and prove they do not appear in index bytes or diagnostics.
   - Include hostile instruction-like author text and prove indexing/tokenizing
     has no execution, tool, network, ref mutation, policy mutation, or
     authorization side effect.

5. Delete and rebuild the index in ordinary and fresh local state.
   - Prove exact-filter and lexical membership/order are identical before
     deletion, after clean rebuild one, and after clean rebuild two.
   - Prove recall candidates recover solely from verified refs and exact policy
     with no copied index, semantic model, service, Docker, or vendor API.

**Files**: `memory_index_test.go` (10k corpus, timing, permission, secret,
hostile-content, delete/rebuild, and failure-injection tests, roughly 250–400
additional lines); minimal `memory_index.go` support only where tests require it.

**Validation**:

- `go test ./... -run TestMemoryIndex` covers correctness, isolation, and the
  reproducible delete/rebuild matrix.
- The 10k acceptance fixture records rebuild duration and p95 exact/lexical
  query duration and meets the stated 30-second/1-second development targets.
- `go test -race ./...`, `go vet ./...`, `go build ./...`, and
  `git diff --check` pass.
- Record completion with
  `spec-kitty agent tasks mark-status T020 --status done` only after all gates.

## Definition of Done

- T016 has event-sourced `done` evidence for the strict schema, exact private
  path, deterministic source fingerprint, and byte-stable encoding rules.
- T017 has event-sourced `done` evidence that rebuild starts from verified
  local/accepted refs and exact policy only, excludes quarantine and network
  state, preserves full projection fields, and writes atomically.
- T018 has event-sourced `done` evidence for strict corruption detection,
  ref/policy staleness, read-only verification, and byte-identical clean rebuilds.
- T019 has event-sourced `done` evidence for every exact filter, Unicode-aware
  lexical intersection, full-ID behavior, and stable total ordering.
- T020 has event-sourced `done` evidence for the real 10,000-record performance
  target, owner-only permissions, failure atomicity, secret isolation, inert
  hostile content, and delete/rebuild recovery.
- Deleting `.git/nh/memory/index-v0.json` loses no canonical data and the next
  explicit rebuild reconstructs identical membership, classifications, token
  postings, and deterministic query results without network access.
- A stale, corrupt, unsupported, or semantically inconsistent index is never
  preferred over verified signed sources and cannot mutate refs or policy.
- No index field or query result collapses signature, lifecycle, applicability,
  evidence, policy trust, semantic truth, or authority into one flag.
- No model API, vector database, hosting-provider API, mandatory service,
  Docker operation, new dependency, or platform-specific runtime is introduced.
- Only `memory_index.go` and `memory_index_test.go` are changed, with any
  unavoidable out-of-map edit preceded by a recorded one-line rationale.
- Completion is recorded with separate event-sourced commands, never prompt
  checkboxes:

```sh
spec-kitty agent tasks mark-status T016 --status done
spec-kitty agent tasks mark-status T017 --status done
spec-kitty agent tasks mark-status T018 --status done
spec-kitty agent tasks mark-status T019 --status done
spec-kitty agent tasks mark-status T020 --status done
```

## Risks

- **Hidden authority**: a convenient cache may become the source of truth;
  recompute fingerprints/derived rows and always resolve disagreement for refs.
- **Nondeterministic bytes**: maps, timestamps, temporary paths, or input order
  can leak into JSON; sort every collection and compare complete rebuild bytes.
- **Incomplete fingerprint**: omitting ref names, head OIDs, or policy digest
  permits stale trust/projection; freeze and test the complete preimage.
- **Unverified-source ingestion**: broad ref enumeration can admit quarantine
  or collaboration state; consume WP02's verified collectors only.
- **Token drift**: locale libraries or inconsistent normalization can change
  matches; use one standard-library tokenizer for indexing and queries.
- **Quadratic retrieval**: scanning 10,000 full records for every term can miss
  p95 targets; intersect sorted postings after selective exact filters.
- **Permission portability**: POSIX mode assertions can be misleading on other
  hosts; enforce Unix owner-only mode and test the portable helper contract.
- **Sensitive cache content**: the index necessarily repeats deliberately
  published memory prose; keep it private, never add ambient secrets, and do
  not describe deletion as erasure of canonical Git objects.
- **Worktree path confusion**: writing beneath the working tree may publish the
  cache accidentally; resolve the actual Git administrative directory.
- **Scope bleed into WP05**: provide internal candidates only; cursor encoding,
  recall envelopes, content-byte bounds, and CLI routing belong to WP05.

## Reviewer Guidance

Start by deleting the cache and tracing rebuild inputs. Every row must come
from a WP02-verified local or accepted stream and WP03 projection under exact
policy bytes. Confirm quarantine refs, arbitrary refs, collaboration events,
working-tree files, prior index contents, environment values, and network state
cannot enter the rebuild. Inspect source-fingerprint test vectors and require a
change for every source ref/head or policy change.

Compare the complete bytes of two clean rebuilds, then corrupt each field family
and verify the cache loses every disagreement with canonical sources. Check
strict decoding, atomic replacement, owner-only state, linked-worktree path
resolution, and permission-failure behavior. Deleting the index must never make
memory unrecoverable or mutate refs, policy, signatures, or projection.

For query semantics, verify exact filters precede lexical matching, IDs never
use prefixes, topic/query terms intersect as specified, Unicode behavior is
portable, and the result order is total and stable. Re-run the actual 10,000-
record corpus and inspect whether measurements exercise rebuild, persistence,
load/verify, postings, and combined exact/lexical lookup rather than mocks.

Finally, search index bytes and diagnostics for the sentinel secrets and inspect
hostile instruction-like text paths. Reject any execution hook, implicit fetch,
semantic truth/ranking claim, mutable canonical database, public cache path, or
change outside the two owned files without recorded rationale.

## Implementation Command

```bash
spec-kitty agent action implement WP04 --agent <name>
```
