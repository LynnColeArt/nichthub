# Memory Replication Contract v0

## Selection

Saved replication selection version 1 gains an optional `memories` array of
sorted unique full stream IDs. CLI selection uses repeated `--memory` flags.
An explicit selection may combine actor, proposal, and memory selectors.

`--all` remains mutually exclusive with exact selectors and discovers all
advertised Nichthub actor, proposal, and memory refs. An explicit actor or
proposal selection imports zero memory unless a memory stream is also selected.

## Advertisement and ownership

Only refs matching this exact form are memory candidates:

```text
refs/nh/memory/<full-actor-fingerprint>/<64-hex-stream-digest>
```

One full stream ID must resolve to one owner/ref in a remote advertisement.
Ambiguous, malformed, duplicate, or owner-mismatched advertisement fails that
memory request without being interpreted as another selector kind.

## Quarantine validation

Each selected memory stream is fetched into the isolated transaction Git
directory and measured using the configured event, object, individual-object,
attachment, and total-byte budgets. Version 0 memory has no attachments, but
the existing total/object budget remains mandatory.

Before promotion, validation proves:

1. exact ref owner and stream ID agreement;
2. strict two-file commit trees;
3. payload hash and Ed25519 signature validity;
4. linear Git parents and signed sequence/previous continuity;
5. every record/lifecycle bound and kind-specific shape;
6. lifecycle target status and exact anchor/evidence dependency status;
7. compatibility with the already-accepted projection used by the transaction.

No quarantine ref is a recall source. Promotion uses transaction-safe accepted
refs only after validation.

## Independent outcomes

Memory uses `kind=memory` and the full stream ID in `ReplicationOutcome`. An
invalid, over-budget, dependency-missing, or interrupted memory request cannot
roll back or suppress an independently valid actor, proposal, or other memory
request. Previously accepted refs remain unchanged on failure.

Shared Git objects do not merge trust decisions. Pre-existing accepted objects
are never falsely denied, and one transaction cannot clear another pending
transaction's denial/anchor state.

## Shallow recovery

When an otherwise valid selected stream references an unavailable exact anchor,
evidence object, lifecycle target, or predecessor, diagnostics contain:

- the owning memory and stream IDs;
- the exact missing full ID and dependency kind;
- the selected remote, when known;
- the exact selector update required, when not selected;
- `nh sync <remote> --recover-shallow` only when the repository is actually
  shallow and the saved selection authorizes bounded recovery.

Absence never becomes evidence resolution. Malformed, wrong-type, or signature
failures remain ordinary invalid-data failures rather than shallow gaps.

## Publication

`nh sync` publishes every local `refs/nh/memory/*/*` ref after local actor and
proposal refs. Publication failure reports its phase safely and never changes
local canonical state or rewrites a stream.
