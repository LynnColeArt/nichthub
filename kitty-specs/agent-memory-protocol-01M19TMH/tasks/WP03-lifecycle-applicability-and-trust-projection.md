---
work_package_id: "WP03"
title: "Lifecycle, Applicability, and Trust Projection"
dependencies: ["WP01", "WP02"]
requirement_refs:
  - FR-003
  - FR-004
  - FR-008
  - FR-010
  - FR-012
  - FR-013
  - FR-014
  - FR-015
  - FR-016
  - FR-017
  - FR-022
  - NFR-001
  - NFR-005
  - NFR-010
  - C-004
  - C-010
subtasks: ["T011", "T012", "T013", "T014", "T015"]
owned_files:
  - "memory_projection.go"
  - "memory_projection_test.go"
  - "policy.go"
  - "policy_test.go"
authoritative_surface: "memory_projection.go"
create_intent:
  - "memory_projection.go"
  - "memory_projection_test.go"
execution_mode: "code_change"
agent_profile: "implementer-ivan"
role: "implementer"
agent: "codex"
model: ""
---

# Work Package Prompt: WP03 – Lifecycle, Applicability, and Trust Projection

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Build the pure, order-independent projection that turns verified memory
envelopes into explicit lifecycle, applicability, evidence, and policy-trust
classifications. Preserve every exact edge and unresolved dependency without
ever collapsing attribution, availability, applicability, policy, or semantic
truth into one trusted flag.

## Context

WP01 defines the strict `nh-memory/0` envelope and WP02 supplies independently
verified memory streams. This package consumes those verified values; it must
not sign records, traverse unaccepted refs, mutate Git state, fetch missing
objects, execute recalled content, or change the existing collaboration event
wire format.

Read `spec.md`, `plan.md`, `data-model.md`, `research.md`, and
`contracts/memory-wire-v0.md` before implementation. Also read the provenance
and missing-dependency requirements in `contracts/memory-cli-v0.md` and
`contracts/memory-replication-v0.md`. Treat projection as a set function:
timestamps may break presentation ties but never establish causality, all
public IDs remain full, and every returned collection has a stable sort order.

The new deep module is `memory_projection.go`; tests belong beside it in
`memory_projection_test.go`. Extend the existing `PolicyDocument` parser and
tests narrowly in `policy.go` and `policy_test.go`. No other source file is in
this work package's ownership map. Downstream WP04 consumes projected rows for
its disposable index, WP05 consumes them for recall, and WP06 consumes exact
dependency classifications for quarantine and shallow recovery.

### Subtask T011: Validate lifecycle relationships and missing targets

**Purpose**

Define one fail-closed relationship-validation pass for `supersede`, `retract`,
and `challenge` envelopes after stream and signature verification, while
retaining honest dependency facts when an exact target is unavailable.

**Steps**

1. Start in `memory_projection_test.go` with red table tests built from WP01
   envelopes and WP02 verified-stream outputs; do not bypass their canonical
   ID, actor, or signature rules with unrelated test-only representations.
2. Build the target lookup from every record-producing envelope: both
   `record` and the complete replacement record carried by `supersede`.
3. Require every lifecycle edge to name a full exact memory ID and distinguish
   a target absent from the selected verified set from a target that is present
   but is not a record-producing envelope.
4. Enforce same-author ownership for `supersede` and `retract` using the signed
   actor on the source and target; never infer ownership from stream position,
   actor display name, or delivery order.
5. Enforce the version-0 cross-author challenge rule without letting the
   challenger impersonate, rewrite, delete, or reclassify the target author.
6. Reject self-targeting lifecycle facts and any malformed relationship shape
   that should already have been excluded by WP01, keeping defense in depth at
   the projection boundary.
7. Preserve a missing target as a typed dependency containing the owning
   lifecycle memory ID, its stream ID, the operation, and the exact missing
   target ID; do not manufacture a placeholder target record.
8. Treat missing as distinct from invalid: wrong-kind, actor-rule, self-edge,
   or malformed targets are validation errors and must not become recoverable
   shallow gaps.
9. Ensure one bad relationship cannot remove unrelated valid records or alter
   the collaboration projection; return precise scoped diagnostics suitable
   for WP06's independent memory outcome.
10. Sort accepted relationship edges and missing-target facts by stable full-ID
    keys before exposing them to lifecycle reduction or callers.
11. Cover a lifecycle fact arriving before its target, the same set after the
    target arrives, cross-stream targets, same-author records in separate
    streams, and unrelated records alongside each failure.
12. Keep natural-language disagreement out of validation: only signed explicit
    `target` relationships create canonical lifecycle edges.

**Files**

- Create `memory_projection.go` for relationship types, lookup construction,
  validation, exact dependency facts, and stable ordering (roughly 120–180
  lines for this first slice of the module).
- Create `memory_projection_test.go` with focused relationship fixtures and
  missing-versus-invalid assertions (roughly 180–260 lines initially).

**Validation**

- Run `go test ./... -run 'TestMemoryProjection.*Relationship'`.
- Assert all errors and gaps contain full safe IDs, never abbreviated IDs.
- Assert unrelated verified records survive every scoped invalid/missing case.
- Record T011 completion with
  `spec-kitty agent tasks mark-status T011 --status done` only after the focused
  tests pass.

### Subtask T012: Derive deterministic lifecycle state including branching

**Purpose**

Reduce the validated relationship set into auditable per-record lifecycle rows
whose summary states and edge lists converge byte-for-byte for every input
permutation.

**Steps**

1. Define a projection row for each record-producing memory with full memory,
   stream, and actor IDs plus lifecycle summary and sorted `challengers`,
   `successors`, and `retractions` lists.
2. Keep all relationship IDs even when the summary label has higher-precedence
   state; the label is a convenience and never erases the graph.
3. Mark a target `superseded` for one valid direct same-author successor and
   mark it `branching` when more than one valid direct successor exists.
4. Project each superseding replacement as its own record, independently active
   unless another explicit lifecycle edge changes its state.
5. Mark a target `retracted` for any valid same-author retraction regardless of
   whether a challenge or supersession was seen earlier or later.
6. Attach every valid challenge ID as dispute metadata without labeling the
   target false, untrusted, inactive, deleted, or semantically contradicted.
7. Represent unresolved lifecycle dependencies explicitly and apply the
   specified summary precedence: `dependency-missing`, `retracted`,
   `branching`, `superseded`, then `active`.
8. Derive solely from set membership and explicit edges. Never use slice input
   order, stream collection order, map iteration, sequence across streams, or
   timestamp comparison as lifecycle causality.
9. Sort final rows by full memory ID and every nested ID list
   lexicographically so canonical JSON fixtures are byte-identical.
10. Make the projection API pure over supplied verified envelopes and explicit
    dependency context: no Git commands, policy loads, logging, ref updates,
    or network access in lifecycle reduction.
11. Add fixtures for a linear supersession chain, two direct supersessors,
    retraction plus branching, challenge before/after retraction, and multiple
    challengers from separate actors.
12. Include a missing-target lifecycle fact beside complete branches and prove
    the missing edge does not erase or reorder complete record rows.

**Files**

- Expand `memory_projection.go` with immutable projection rows, graph indexes,
  lifecycle precedence, and stable output helpers (roughly 160–240 additional
  lines, keeping helpers narrow and unexported unless later WPs need them).
- Expand `memory_projection_test.go` with auditable expected rows and canonical
  JSON comparison fixtures (roughly 220–320 additional lines).

**Validation**

- Run `go test ./... -run 'TestMemoryProjection.*Lifecycle'`.
- Compare original, reverse, and deterministic shuffled inputs as encoded
  bytes, including branching and missing-dependency cases.
- Assert challenges coexist with every summary state and never affect trust.
- Record T012 completion with
  `spec-kitty agent tasks mark-status T012 --status done` after convergence is
  demonstrated.

### Subtask T013: Resolve exact applicability and typed evidence with gaps

**Purpose**

Classify declared scope and supporting references against an explicit query
commit without inference, and expose missing dependencies separately from
invalid anchors or evidence.

**Steps**

1. Add an explicit read-only projection context carrying the requested commit,
   optional exact subject/path filters, accepted memory records, and the
   collaboration events available for exact evidence resolution.
2. For `exact` applicability, require equality between the record's anchor
   commit and requested commit; an existing different commit is inapplicable,
   not invalid.
3. For `descendants`, use Git ancestry between the exact anchor commit and
   requested commit. Reuse an existing exact ancestry helper when its missing
   and invalid distinctions match this contract; do not parse `git` prose.
4. For `subject`, require the applicability subject to equal the anchor subject
   and the requested exact subject. Never infer a subject or relevance from
   `content`, topics, path names, or a related event's prose.
5. Verify every path/blob pair against the tree at the anchor commit, including
   the explicit `absent` marker. A later changed, moved, or removed path does
   not retroactively invalidate a valid anchor pair.
6. Do not implement rename detection. Path matching is repository-relative,
   slash-normalized, exact, and tied to the declared anchor state.
7. Classify unavailable anchor commits or objects as `anchor-missing`; classify
   malformed, wrong-type, or mismatched path/blob claims as `anchor-invalid`;
   classify valid scope as `applicable` or `inapplicable` independently.
8. Resolve `git:` evidence as the exact full Git object ID, `event:` evidence
   through exact verified collaboration-event identity, and `memory:` evidence
   through exact accepted memory identity.
9. Return evidence as `resolved`, `missing`, or `invalid` with a sorted detail
   for every dependency including evidence type, exact requested ID, owning
   memory ID, and a stable reason code.
10. Never treat a syntactically valid but unavailable object as resolved, and
    never treat a wrong-type or malformed object as a shallow-recoverable gap.
11. Keep evidence availability separate from lifecycle, applicability, policy
    qualification, signature validity, and factual truth; no successful
    evidence lookup upgrades any other classification.
12. Exercise absent commits, non-ancestor commits, wrong blobs, explicit absent
    paths, unavailable events/memories, and mixed resolved/missing evidence.

**Files**

- Extend `memory_projection.go` with applicability/evidence enums, typed exact
  dependency details, and injectable read-only resolver seams (roughly 180–260
  lines; keep Git access outside pure lifecycle reduction).
- Extend `memory_projection_test.go` using temporary repositories and accepted
  event/memory fixtures (roughly 260–380 additional lines).

**Validation**

- Run `go test ./... -run 'TestMemoryProjection.*(Applicability|Evidence)'`.
- Assert each gap is stable, sorted, full-ID, and distinct from invalid data.
- Assert path changes after the anchor and renamed paths cause no inference.
- Record T013 completion with
  `spec-kitty agent tasks mark-status T013 --status done` after focused tests
  pass.

### Subtask T014: Add optional MemoryPolicy and exact trust classification

**Purpose**

Extend the existing strict repository policy with an optional memory section
and classify actor/kind eligibility at the exact requested policy commit while
preserving every legacy policy behavior.

**Steps**

1. Add a narrow `MemoryPolicy` model to `policy.go` with JSON fields
   `trustedActors` and `trustedKinds`, and add optional
   `PolicyDocument.Memory` using omission semantics compatible with legacy
   `.nh/policy.json` documents.
2. Keep `nh.policy/0`, the existing top-level field names, exact-byte digest,
   size limit, unknown-field rejection, trailing-JSON rejection, maintainer,
   proposal, and pipeline validation unchanged.
3. Validate every trusted actor as a full actor fingerprint and every trusted
   kind against WP01's complete version-0 memory-kind set; share the canonical
   kind validator rather than creating a drifting duplicate enum.
4. Require the policy lists to have deterministic unique membership and reject
   duplicates with field-specific errors. Do not rewrite policy bytes merely
   to sort input or change the digest semantics.
5. Treat an absent memory section as a valid legacy policy and classify all
   memory as `policy-missing` for default recall; do not fail repository or
   collaboration policy loading.
6. For a present memory policy, classify actor membership and kind membership
   independently, then return exactly one stable trust class:
   `qualified`, `actor-untrusted`, or `kind-untrusted` according to a documented
   deterministic precedence when both dimensions fail.
7. Load policy from the exact commit supplied by the projection caller through
   the existing `loadPolicy` byte path. Never consult mutable working-tree
   policy or `HEAD` implicitly after the caller selects a commit.
8. Expose the policy digest alongside projection context for downstream index
   source fingerprints, but do not make that digest signature, lifecycle,
   evidence, applicability, truth, or authorization.
9. Ensure explicit untrusted inspection in WP05 can retain the real trust class
   without mutating policy or relabeling a record `qualified`.
10. Add legacy decode/round-trip fixtures with no `memory`, policies containing
    empty valid memory lists, and policies containing sorted trusted values.
11. Add failures for unknown memory fields, invalid actors/kinds, duplicates,
    and trailing JSON while asserting existing policy diagnostics remain
    stable enough for current callers.
12. Run the existing proposal/governance policy tests to prove the additive
    field does not change collaboration evaluation or public event IDs.

**Files**

- Modify existing `policy.go` for the optional type, validation, and a small
  pure trust-classification helper (roughly 50–90 added lines).
- Modify existing `policy_test.go` for legacy and memory-policy cases (roughly
  120–180 added lines). Do not create replacement policy modules.

**Validation**

- Run `go test ./... -run 'Test.*Policy'`.
- Confirm a byte-for-byte legacy policy fixture parses with the same digest and
  drives the same collaboration results.
- Confirm missing policy, actor failure, and kind failure stay distinct.
- Record T014 completion with
  `spec-kitty agent tasks mark-status T014 --status done` only after focused
  and legacy regressions pass.

### Subtask T015: Prove convergence and classification separation

**Purpose**

Lock the complete projection contract with permutation, exact-policy,
ancestry/path, and orthogonal-classification tests that downstream recall and
replication can safely reuse.

**Steps**

1. Build one representative corpus containing linear and branching
   supersession, same-author retraction, cross-actor challenges, a missing
   target, exact and descendant anchors, path/blob pairs, and all evidence
   prefixes.
2. Project the identical verified set in original, reversed, and at least one
   deterministic shuffled order; marshal the complete result and require
   byte-identical output for all permutations.
3. Include equal timestamps and deliberately misleading timestamps to prove
   signed time changes only documented presentation ordering, never lifecycle
   graph meaning.
4. Create two committed policies that qualify different actors or kinds.
   Project against each exact policy commit after moving `HEAD` and prove the
   selected commit alone determines trust and policy digest.
5. Test exact applicability at the anchor, descendant applicability at a child
   and unrelated commit, plus unavailable and wrong-type anchor objects.
6. Test a path whose blob matches at the anchor and changes, disappears, or is
   renamed later; require exact anchor validation and zero rename inference.
7. Construct rows where only one dimension changes at a time: lifecycle,
   applicability, evidence, actor trust, kind trust, and policy presence.
8. Assert each changed dimension leaves signature validity and every unrelated
   classification byte-for-byte unchanged.
9. Include a valid signature with unresolved evidence and an untrusted actor
   with resolved evidence to prove neither condition is summarized as truth.
10. Verify a challenge cannot change trust or evidence, a retraction cannot
    change applicability, and a policy change cannot alter historical
    lifecycle edges.
11. Assert missing dependency details are full, typed, sorted, and preserved
    beside projected rows rather than hidden in logs or generic errors.
12. Retain a collaboration-only baseline: with no memory refs and no memory
    policy, existing tests and event projections behave exactly as before.
13. Run the race detector after focused tests to catch shared mutable maps or
    lazy sorting that could make concurrent projection nondeterministic.
14. Keep fixture helpers deterministic and local; no network, model API,
    service, Docker, environment secret, or copied private index is allowed.

**Files**

- Finish `memory_projection_test.go` with permutation and separation suites
  (roughly 350–500 total lines for the final file, favor table fixtures over
  repeated setup).
- Update `policy_test.go` only where exact-at-commit integration needs policy
  fixtures; production changes remain within declared owned files.

**Validation**

- Run `go test ./... -run 'TestMemoryProjection'` at least three times.
- Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and
  `go build ./...`.
- Run `git diff --check` and confirm no existing collaboration wire fixture or
  public event ID changed.
- Record T015 completion with
  `spec-kitty agent tasks mark-status T015 --status done` only after all gates
  pass.

## Definition of Done

- T011 has an event-sourced `done` status and relationship validation enforces
  exact targets, same-author correction, cross-author challenge, and explicit
  missing-versus-invalid dependency behavior.
- T012 has an event-sourced `done` status and all lifecycle rows, edge lists,
  branching states, and precedence outcomes converge byte-for-byte under input
  permutation.
- T013 has an event-sourced `done` status and exact/descendant/subject
  applicability, path/blob validation, and all three typed evidence namespaces
  return explicit resolved, missing, or invalid classifications.
- T014 has an event-sourced `done` status and optional `MemoryPolicy` parsing,
  validation, exact-commit qualification, and legacy policy compatibility are
  proven by focused and existing tests.
- T015 has an event-sourced `done` status and the integrated separation suite,
  collaboration-only regression, race detector, vet, build, and diff checks
  all pass.
- Projection output includes full memory/stream/actor IDs, anchor,
  applicability, lifecycle plus all edge IDs, evidence details, trust class,
  and content digest for every record-facing row required by downstream WPs.
- No code path infers truth, contradiction, relevance, authorization, or
  instruction priority from memory prose, signatures, evidence, or policy.
- Only `memory_projection.go`, `memory_projection_test.go`, `policy.go`, and
  `policy_test.go` are changed, with one-line rationale recorded before any
  necessary out-of-map edit.

## Risks

- Map iteration or arrival order can leak into nested edge/dependency slices;
  sort at construction boundaries and compare fully encoded fixtures.
- A missing object can be mislabeled invalid or resolved, especially in
  shallow repositories; use typed resolver outcomes and preserve exact gaps.
- Policy lookup can accidentally use mutable `HEAD`; pass and test the exact
  policy commit and exact-byte digest explicitly.
- A convenience boolean can collapse orthogonal trust properties; model and
  test signature, lifecycle, applicability, evidence, and policy separately.
- Path ancestry can drift into rename heuristics; validate the declared path at
  the anchor commit only and keep descendant scope commit-based.
- Adding `Memory` can break strict legacy policy decoding or IDs; keep it
  optional and run the complete pre-memory policy/collaboration suite.

## Reviewer Guidance

Reviewers should trace every lifecycle summary back to preserved sorted edge
IDs and verify the result is invariant under input order. Pay special attention
to same-author correction rules, cross-author challenge behavior, exact
missing-versus-invalid distinctions, branch precedence, and whether successors
remain independently projected.

For applicability and evidence, check exact commit/object semantics, no rename
or prose inference, and typed gap details. For policy, verify exact-commit
loading, legacy JSON compatibility, unchanged exact-byte digest behavior, and
that actor/kind qualification never changes signature, lifecycle, evidence,
applicability, truth, or authorization. Reject any hidden network access,
mutable global cache, truncated trust-bearing ID, or collaboration wire change.

Implementation command:

`spec-kitty agent action implement WP03 --agent <name>`
