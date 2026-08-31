# Research: Agent Memory Protocol

## Purpose

This document resolves the implementation questions that must be stable before
planning. The mission specification remains authoritative for product intent.
These decisions choose a version-0 representation that fits Nichthub's current
Go and Git architecture without making memory part of the collaboration event
chain or turning recalled text into authority.

## Existing foundations

The operational alpha already supplies the necessary lower layers:

- Ed25519 actor identities and a permission-safe local keyring.
- Content-addressed signed JSON facts stored in Git commit trees.
- Exact full-ID validation and deterministic relationship projection.
- Independently selected refs, quarantine validation, positive budgets,
  transaction-safe promotion, and explicit shallow recovery.
- Policy documents loaded from exact commits and governed through signed
  proposal evidence.

The collaboration implementation currently assumes one actor event chain at
`refs/nh/actors/<actor>`. Memory cannot be appended to that chain: doing so
would make collaboration-only replication depend on the presence and size of
memory history, violating FR-006, FR-023, and C-003.

## Decisions

### D1. Use a separate signed memory envelope and separate Git refs

Canonical memory uses protocol `nh-memory/0`, not additional fields on the
existing `nh/0` collaboration `Event`. Each stream is an append-only Git commit
chain whose tree contains exactly `memory.json` and `signature`. The payload is
canonical Go `encoding/json` output and its public ID is the SHA-256 digest of
the exact payload bytes.

Local stream refs use:

```text
refs/nh/memory/<actor>/<stream-digest>
```

Accepted remote refs use:

```text
refs/nh/remotes/<remote>/memory/<actor>/<stream-digest>
```

Quarantine refs remain transaction-local and never become a recall source.
This mirrors the proven event storage technique while making collaboration and
memory availability independent.

### D2. Streams are explicit, actor-owned, and have a deterministic default

A stream ID is a full `sha256:<64-hex>` identifier. The default stream ID is
the SHA-256 digest of the domain-separated bytes
`nh-memory-stream-v0\x00<actor>\x00default`. Adapters may provide another full
stream ID to partition memory by purpose. The actor and stream ID are present
in every signed payload, and every event in one stream must have the same actor
and stream ID.

This avoids a stream-registry service and gives simple CLI use a stable stream,
while allowing explicit partitioning. A stream ID is an address, not a trust
claim or secret.

### D3. Model records and lifecycle facts in one stream protocol

The signed envelope has an operation discriminator:

- `record`: observation, decision, assumption, attempt, verification, or
  handoff.
- `supersede`: same-author replacement record naming one earlier record.
- `retract`: same-author lifecycle fact naming one earlier record and reason.
- `challenge`: cross-author dispute naming one record, typed reason, and
  optional evidence.

`supersede` contains a complete replacement record, so the successor can be
recalled without a mutable side document. `retract` and `challenge` do not
copy or modify target content. Lifecycle edges name full memory IDs.

Projection is a pure function over the verified set, independent of delivery
order. Same-author rules are checked against the target record. Multiple
supersessors are preserved and classified as branching. Challenges never
erase a target. Missing targets remain explicit dependency-missing facts and
cannot qualify for default active recall.

### D4. Use explicit typed anchors and evidence

Every record requires an exact Git commit anchor. Optional anchors are encoded
as structured values rather than parsed from prose:

- repository paths paired with the exact blob ID (or an explicit absent marker)
  at the anchor commit;
- one exact Nichthub subject ID;
- exact policy, pipeline-definition, run, collaboration-event, memory, or Git
  object evidence references.

Applicability is one of `exact`, `descendants`, or `subject`. `subject`
requires an exact subject; the other modes do not infer a subject from text.
Descendant applicability is checked with Git ancestry against the requested
commit. A missing anchor or evidence object is reported separately from an
invalid one and includes the exact ID and recovery guidance.

Typed evidence prevents ambiguity between Git object IDs and `sha256:`
Nichthub IDs. Version 0 uses strings with the prefixes `git:`, `event:`, and
`memory:` and validates the identifier following the prefix.

### D5. Extend policy with an optional memory section

`PolicyDocument` gains an optional `memory` object containing sorted unique
`trustedActors` and `trustedKinds`. Absence means no actor-authored memory is
eligible for default trusted recall. It does not break existing policy files or
change any collaboration event ID.

Trust classification remains multi-dimensional:

- signature validity is established before projection;
- actor and kind policy qualification are computed at the requested policy
  commit;
- anchor applicability and evidence resolution are separate fields;
- lifecycle state is separate from all of the above.

The default recall view includes only active, applicable, policy-qualified,
evidence-resolved records. Explicit inspection flags may include other classes
but never relabel them trusted.

### D6. Make recall JSON the safety boundary

The stable machine interface is a versioned JSON request and response. Human
flags are translated into the same internal request. Every recall result puts
author text only in a nested `data.content` field and includes a constant
warning that content is untrusted inert data. There is no field named
`instruction`, `command`, `tool`, or `authorization`, and recall performs no
execution or adapter callback.

Default bounds are 20 records and 65,536 encoded content bytes. Both bounds
are positive, enforced after deterministic ordering, and reflected in
`matched`, `returned`, `truncated`, and an opaque deterministic continuation
cursor. Cursors bind the normalized query and last sort key so they cannot be
reused silently for a different query.

Ordering is deterministic: anchor relevance class, lifecycle class, signed
timestamp, then full memory ID. Timestamps are presentation and tie-breaking
data only; they never establish causality.

### D7. Keep the index private, deterministic, and disposable

The version-0 index is a strict JSON projection below
`.git/nh/memory/index-v0.json`, written atomically with owner-only permissions.
It contains verified record metadata plus normalized lexical tokens, but no
private key material and no additional canonical claims. Its source fingerprint
is the sorted list of accepted memory ref names and object IDs plus the policy
digest.

`nh memory index rebuild` reconstructs it entirely from accepted refs.
`nh memory index verify` compares its source fingerprint and derived record
projection; mismatch causes a fail-closed diagnostic and supports rebuilding.
Recall may rebuild a missing index locally but never fetch network data.

Tokenization is deterministic Unicode-aware lowercase word splitting using the
Go standard library. Exact filters are applied before lexical matching.
Embeddings and model-generated summaries are outside version 0.

### D8. Extend selected replication as a third independent selector kind

`ReplicationSelection` gains `memories`, a sorted unique list of full stream
IDs. `--all` continues to mean all advertised Nichthub refs for compatibility.
Explicit selection can combine actors, proposals, and memory streams, but each
selection is validated, measured, quarantined, promoted, and reported
independently.

Memory stream discovery examines only `refs/nh/memory/*/*`. Validation checks
ref ownership, signatures, payload limits, stream-chain continuity, lifecycle
relationships available within the selected projection, and exact dependency
status. Memory failure does not remove independently valid actor or proposal
outcomes from a transaction.

Publication pushes local memory refs after actor and proposal refs. Failure is
reported as a publication-phase error and does not rewrite local canonical
state.

### D9. Enforce bounds at construction and ingestion

The same constants govern local recording and hostile import:

- content: at most 65,536 UTF-8 bytes and valid UTF-8;
- evidence: at most 64 unique typed references;
- topics: at most 32 unique normalized labels;
- paths: bounded count and encoded size;
- one payload and signature per memory commit; no arbitrary attachments in v0.

One-below, exact-limit, and one-above tests cover the required public limits.
Replication budgets remain additional transaction limits rather than replacing
record-shape limits.

### D10. Build around deep modules instead of spreading memory cases

Implementation should isolate memory behavior behind these modules:

- `memory_event.go`: wire envelope, canonical validation, signing, verification.
- `memory_store.go`: refs, append/load, stream-chain validation, collection.
- `memory_projection.go`: lifecycle, applicability, evidence, trust projection.
- `memory_index.go`: deterministic disposable index and lexical matching.
- `memory_commands.go`: record, lifecycle, recall, show, and index CLI.
- narrow additions to policy and replication modules for memory policy and
  selector transport.

The existing collaboration `Event` wire shape must not change. Shared helpers
may be extracted only when their invariants remain identical.

## Rejected alternatives

### Put memory in `refs/nh/actors/<actor>`

Rejected because it couples collaboration availability to memory volume and
means a collaboration-only clone must ingest memory commits.

### Commit mutable Markdown summaries to the ordinary branch

Rejected because authorship, exact correction lineage, selective replication,
and deterministic convergence would depend on branch merge behavior rather
than signed protocol facts.

### Make SQLite, a vector database, or embeddings canonical

Rejected because they add platform or vendor dependencies and cannot be
independently reconstructed from Git alone. A later optional projection may use
them without changing the protocol.

### Automatically summarize chats or terminal history

Rejected because it violates deliberate capture and secret-isolation
requirements. Adapters must submit explicit structured record requests.

### Treat a signature or evidence reference as truth

Rejected because it collapses provenance, policy, availability, lifecycle, and
semantic correctness. The recall envelope exposes these properties separately.

## Verification strategy

Testing proceeds through public boundaries wherever observable behavior is at
stake:

1. Wire fixtures prove stable IDs, signatures, strict decoding, and every
   record bound.
2. Permutation tests prove lifecycle convergence, branching, challenge, and
   missing-dependency behavior.
3. CLI tests prove deliberate record input, JSON envelopes, inert hostile text,
   exact filters, byte/count bounds, and continuation.
4. Replication tests use temporary bare remotes for memory-only,
   collaboration-only, mixed, invalid, over-budget, and shallow cases.
5. Index tests delete, rebuild twice, corrupt, verify, and compare byte-identical
   projections.
6. A 10,000-record benchmark-style acceptance test checks rebuild under 30
   seconds and exact/lexical recall under one second on the development host.
7. A fresh-clone black-box scenario proves identical IDs, lifecycle, and recall
   without keys or copied indexes.
8. Existing `go test -race ./...`, `go vet ./...`, and `go build ./...` remain
   mandatory regression gates.

## Risks and open questions

- **Policy bootstrap:** existing repositories have no memory policy, so default
  recall initially returns no trusted records. Documentation must make the
  explicit inspection and governed amendment path obvious.
- **Evidence breadth:** version 0 supports exact typed references; it does not
  infer whether arbitrary external facts are true or relevant.
- **Path ancestry:** rename inference is intentionally absent. A path anchor is
  evaluated exactly at its commit, and descendant applicability concerns the
  memory record, not a guessed rename history.
- **Retention:** Git can retain unreachable rejected objects. Retraction and
  local index deletion are not redaction guarantees.
- **Cursor stability:** changing accepted refs or policy invalidates a cursor;
  the response must say so rather than silently continuing a different result
  set.
- **Scale:** the standard-library JSON index should comfortably meet the
  specified 10,000-record alpha target. If measurements disagree, optimization
  stays inside the disposable index module and cannot alter canonical ordering
  or identity.

No unresolved product decision blocks planning. The main implementation risk is
keeping replication changes narrow enough that memory failures cannot affect
the mature collaboration path; that must be isolated in its own work package
and proven with mixed hostile transactions.
