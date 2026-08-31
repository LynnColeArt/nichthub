# Memory Wire Contract v0

## Canonical object

Each memory stream commit contains exactly:

```text
100644 memory.json
100644 signature
```

`memory.json` is a strict JSON `MemoryEnvelope` with `protocol` equal to
`nh-memory/0`. Unknown fields, duplicate semantic values, invalid UTF-8, invalid
IDs, and kind-inapplicable fields are rejected. `signature` is raw-base64
Ed25519 over the exact `memory.json` bytes.

The public memory ID is `sha256:` plus SHA-256 of the exact JSON bytes. The Git
commit ID is storage identity, not memory identity.

## Operations

| Operation | Required payload | Relationship rule |
| --- | --- | --- |
| `record` | Complete MemoryRecord | Creates a new record |
| `supersede` | Target plus complete replacement MemoryRecord | Target must be a record by the same actor |
| `retract` | Target plus nonblank typed reason | Target must be a record by the same actor |
| `challenge` | Target plus typed reason; optional evidence | Target may be another actor's record |

Every envelope also requires actor/public-key binding, full stream ID, positive
sequence, signed timestamp, and exact previous memory ID except at sequence 1.

## Record kinds

`observation`, `decision`, `assumption`, `attempt`, `verification`, and
`handoff` are the complete v0 set. Attempt records require an outcome.
Verification records require at least one typed evidence ID. Handoffs require
the four explicitly separated lists defined in `data-model.md`.

## Bounds

- content: maximum 65,536 UTF-8 bytes;
- topics: maximum 32 normalized unique labels;
- evidence: maximum 64 unique typed IDs;
- no arbitrary attachments in version 0;
- path anchors and handoff list entries have implementation constants documented
  publicly and covered at one-below, exact, and one-above boundaries.

## Stream refs

```text
refs/nh/memory/<actor>/<stream-digest>
refs/nh/remotes/<remote>/memory/<actor>/<stream-digest>
```

The ref actor, payload actor, signing key fingerprint, and every stream ID in the
chain must agree. A stream has one linear Git parent chain and one linear signed
`previous`/`sequence` chain. Concurrent append compare-and-swap failure asks the
caller to reload and retry; it never creates an unsigned merge.

## Lifecycle convergence

Projection consumes a set of verified envelopes, not delivery order. It sorts
all edges by full ID. Supersession/retraction author constraints are verified
against targets. Branching supersession, retraction, challenges, and missing
targets remain auditable simultaneously; a summary state never erases edge IDs.

Signatures prove attribution and payload integrity only. They do not prove
truth, applicability, evidence validity, current policy qualification, or
authorization.
