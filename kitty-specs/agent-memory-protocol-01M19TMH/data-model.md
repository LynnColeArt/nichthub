# Data Model: Agent Memory Protocol

## Canonical entities

### MemoryEnvelope

One exact signed JSON payload stored in a memory-stream Git commit.

| Field | Type | Rules |
| --- | --- | --- |
| `protocol` | string | Exactly `nh-memory/0` |
| `operation` | enum | `record`, `supersede`, `retract`, `challenge` |
| `actor` | actor fingerprint | Full SHA-256 fingerprint matching `publicKey` |
| `actorName` | string | Display-only, control-safe on output |
| `publicKey` | base64 | Raw Ed25519 public key |
| `stream` | stream ID | Full `sha256:<64-hex>` |
| `sequence` | uint64 | Positive and contiguous within one stream |
| `timestamp` | RFC3339Nano | Signed presentation/tie-break value |
| `previous` | memory ID | Empty only for sequence 1; otherwise exact predecessor |
| `record` | MemoryRecord | Required for `record` and `supersede` |
| `target` | memory ID | Required for lifecycle operations |
| `reason` | enum/string | Required for `retract` and `challenge` |
| `evidence` | typed ID list | Lifecycle evidence; unique, maximum 64 |

The memory ID is `sha256:` plus SHA-256 of the exact encoded envelope bytes.
The Git commit ID is transport/storage identity and is not interchangeable with
the memory ID.

### MemoryRecord

The cognition claim carried by `record` or the replacement carried by
`supersede`.

| Field | Type | Rules |
| --- | --- | --- |
| `kind` | enum | `observation`, `decision`, `assumption`, `attempt`, `verification`, `handoff` |
| `content` | string | Valid UTF-8, nonblank, maximum 65,536 encoded bytes |
| `anchor` | MemoryAnchor | Required exact repository state |
| `applicability` | Applicability | Required explicit mode |
| `topics` | string list | Normalized, unique, sorted, maximum 32 |
| `evidence` | typed ID list | Unique, maximum 64; verification requires at least one |
| `attemptOutcome` | enum/string | Required only for `attempt` |
| `handoff` | HandoffFields | Required only for `handoff` |

`content` is always data. No field in this entity grants permission or changes
instruction priority.

### MemoryAnchor

| Field | Type | Rules |
| --- | --- | --- |
| `commit` | Git OID | Required SHA-1 or SHA-256 object ID |
| `paths` | PathAnchor list | Optional, unique normalized repository paths |
| `subject` | typed subject ID | Optional exact issue/proposal/event/policy/pipeline/run ID |

### PathAnchor

| Field | Type | Rules |
| --- | --- | --- |
| `path` | string | Repository-relative, slash-normalized, no traversal or control bytes |
| `blob` | Git OID or `absent` | Exact tree state at `MemoryAnchor.commit` |

The recorder verifies the path/blob pair before signing. An importer verifies it
when the commit is available and otherwise reports a missing anchor.

### Applicability

| Field | Type | Rules |
| --- | --- | --- |
| `mode` | enum | `exact`, `descendants`, `subject` |
| `subject` | typed subject ID | Required only for `subject` and must match the anchor subject |

### HandoffFields

| Field | Type | Rules |
| --- | --- | --- |
| `completed` | string list | Explicit completed statements |
| `assumptions` | string list | Explicit still-open assumptions |
| `blockers` | string list | Explicit blockers, possibly empty |
| `nextActions` | string list | Proposed actions only; never executable instructions |

Each list is bounded and control-safe when displayed. The whole record remains
subject to the content-size bound.

### TypedEvidenceID

A string with one of these forms:

```text
git:<full-git-object-id>
event:sha256:<64-hex>
memory:sha256:<64-hex>
```

The prefix determines resolution rules. Resolution status is derived and is not
stored as a truth claim in the signed record.

## Canonical stream model

### MemoryStream

| Attribute | Meaning |
| --- | --- |
| owner | Actor whose key signs every envelope |
| stream ID | Full content-addressed stream address |
| head | Git commit named by the stream ref |
| chain | Parent-linked commits ordered from sequence 1 |
| local ref | `refs/nh/memory/<actor>/<stream-digest>` |
| accepted ref | `refs/nh/remotes/<remote>/memory/<actor>/<stream-digest>` |

Invariants:

1. Every commit has exactly one parent except the first, which has none.
2. Envelope actor and stream match the ref.
3. Sequence increases by one and `previous` names the preceding memory ID.
4. The Git commit tree has only `memory.json` and `signature` in version 0.
5. Every payload verifies under its embedded actor public key.
6. One stream never depends on an actor collaboration ref for chain validity.

### Default stream derivation

```text
SHA256("nh-memory-stream-v0\x00" + actor + "\x00default")
```

The result is rendered as a full `sha256:` ID.

## Lifecycle projection

### MemoryProjection

One derived row per record-producing envelope.

| Field | Values |
| --- | --- |
| memory ID | Full payload digest |
| stream | Full stream ID |
| actor | Full actor fingerprint |
| lifecycle | `active`, `superseded`, `retracted`, `branching`, `dependency-missing` |
| challengers | Sorted full challenge IDs |
| successors | Sorted full superseding memory IDs |
| retractions | Sorted full retraction IDs |
| signature | `valid` (invalid envelopes never enter accepted projection) |
| applicability | `applicable`, `inapplicable`, `anchor-missing`, `anchor-invalid` |
| evidence | `resolved`, `missing`, `invalid` plus exact dependency details |
| trust | `qualified`, `actor-untrusted`, `kind-untrusted`, `policy-missing` |

Projection rules:

- A valid same-author `supersede` edge makes its target `superseded`.
- More than one valid direct successor marks the target `branching`; every
  successor remains independently projected.
- A valid same-author retraction makes its target `retracted` regardless of
  later delivery order.
- A challenge adds dispute metadata but does not change the target to trusted,
  false, inactive, or deleted.
- Missing lifecycle targets create dependency-missing facts. They never cause
  an unrelated stream or collaboration event to disappear.
- If multiple states apply, the response preserves all edge IDs and uses the
  precedence `dependency-missing`, `retracted`, `branching`, `superseded`,
  `active` for the single summary label.

## Policy projection

### MemoryPolicy

Optional section of the repository policy:

| Field | Type | Meaning |
| --- | --- | --- |
| `trustedActors` | actor fingerprint list | Actors eligible for default recall |
| `trustedKinds` | memory-kind list | Kinds eligible for default recall |

Both lists are unique and sorted during validation. Missing `memory` policy is a
valid legacy policy and classifies all memory as non-qualifying by default.

Qualification is evaluated at an exact requested policy commit. It never
changes signature validity or historical lifecycle.

## Recall model

### RecallRequestV0

| Field | Type | Default/meaning |
| --- | --- | --- |
| `version` | integer | Exactly 0 |
| `atCommit` | Git OID | Current `HEAD` when omitted by human CLI |
| `subject` | typed ID | Optional exact filter |
| `path` | repository path | Optional exact path-at-anchor filter |
| `topic` | string list | Optional normalized intersection filter |
| `kind` | memory-kind list | Optional exact filter |
| `actor` | actor list | Optional full-ID filter |
| `lifecycle` | state list | Defaults to `active` |
| `trust` | trust class list | Defaults to `qualified` |
| `query` | string | Optional deterministic lexical terms |
| `maxRecords` | positive integer | Default 20 |
| `maxContentBytes` | positive integer | Default 65,536 |
| `cursor` | string | Optional opaque continuation token |
| `includeUntrusted` | bool | Explicit inspection, never reclassification |

### RecallEnvelopeV0

| Field | Meaning |
| --- | --- |
| `version` | Machine response version |
| `warning` | Constant inert-content warning |
| `queryDigest` | Digest of normalized request and accepted-source fingerprint |
| `matched` | Total deterministic matches before page bounds |
| `returned` | Records in this page |
| `truncated` | Whether more matching records remain |
| `nextCursor` | Query-bound continuation, omitted at end |
| `memories` | Ordered RecallItem list |
| `missingDependencies` | Sorted exact unresolved IDs and recovery guidance |

### RecallItem

Every item contains full memory ID, actor, stream, signature status, anchor,
applicability, lifecycle and all lifecycle edge IDs, evidence status, trust
classification, content digest, and nested `data`. `data.content` is the only
place author prose is returned.

## Disposable local entities

### MemoryIndexV0

Private file: `.git/nh/memory/index-v0.json`.

| Field | Meaning |
| --- | --- |
| `version` | Index format version |
| `sourceFingerprint` | Digest of sorted accepted memory refs/OIDs and policy digest |
| `builtAt` | Diagnostic timestamp, excluded from deterministic comparison or fixed from sources |
| `records` | Sorted projected record metadata and inert data |
| `tokens` | Deterministic token-to-sorted-memory-ID map |

The preferred design omits wall-clock `builtAt` so two clean rebuilds are
byte-identical. If a diagnostic build time is needed, it belongs in stderr or a
separate non-compared local status file.

### ReplicationSelection extension

| Field | Type | Meaning |
| --- | --- | --- |
| `memories` | stream ID list | Exact independently selected memory streams |

The existing actor and proposal fields retain their wire names and behavior.
Saved selections without `memories` decode unchanged.

### ReplicationOutcome extension

Memory outcomes use `kind=memory`, `id=<full-stream-id>`, and the same promoted,
invalid, over-budget, dependency-missing, and transaction diagnostics as other
selection kinds.

## State transitions

```text
explicit record input
  -> validate bounds/anchors/evidence
  -> sign exact envelope
  -> append stream commit/ref atomically
  -> index becomes stale (canonical state already durable)

selected remote ref
  -> quarantine fetch
  -> measure budgets
  -> verify stream/ref/signatures/shape
  -> resolve or record exact missing dependencies
  -> promote accepted memory ref atomically
  -> rebuild/mark stale local index

accepted refs + exact policy commit
  -> deterministic lifecycle projection
  -> applicability/evidence/trust classification
  -> exact filters and lexical match
  -> deterministic ordering and bounds
  -> inert RecallEnvelopeV0
```

## Compatibility invariants

- Existing `Event` JSON fields and IDs are unchanged.
- Existing actor and proposal refs are unchanged.
- Existing replication-selection JSON without `memories` remains valid.
- Collaboration collection does not enumerate memory refs.
- Memory collection does not require actor collaboration refs.
- Deleting every memory ref and local index restores collaboration-only
  behavior without data migration.
