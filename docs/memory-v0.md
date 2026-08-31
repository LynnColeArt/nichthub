# Agent Memory Protocol v0

Nichthub memory is deliberate, signed project cognition carried by Git. It is
a separate additive protocol, `nh-memory/0`; memory is not an `nh/0`
collaboration event and does not alter existing event IDs or actor histories.
The implementation is repository-local and experimental.

## Identity, objects, and streams

Each memory commit has exactly two mode-`100644` files:

```text
memory.json
signature
```

`memory.json` is canonical UTF-8 JSON. `signature` is raw-base64 Ed25519 over
those exact bytes. The full public memory ID is
`sha256:<64-lowercase-hex>` over `memory.json`; it is distinct from the Git
commit ID. Actor IDs are 64 lowercase hexadecimal key fingerprints. Stream
IDs are full `sha256:` IDs.

An actor owns each append-only stream:

```text
refs/nh/memory/<full-actor>/<64-hex-stream-digest>
refs/nh/remotes/<remote>/memory/<full-actor>/<64-hex-stream-digest>
```

The ref owner, payload actor, public-key fingerprint, and stream must agree.
Sequence starts at 1 and increases by one. Both the Git parent and signed
`previous` memory ID must name the predecessor. Concurrent append failure asks
the writer to reload and retry; streams never merge.

## Envelope and record model

Every envelope names `protocol`, `operation`, `actor`, `actorName`,
`publicKey`, `stream`, `sequence`, and an RFC 3339 timestamp. Later stream
entries also name `previous`. The operations are:

| Operation | Meaning |
| --- | --- |
| `record` | Create a record. |
| `supersede` | Create a replacement and point to a record by the same actor. |
| `retract` | Append a typed retraction of a record by the same actor. |
| `challenge` | Append another actor's typed dispute, optionally with evidence. |

Records have exactly one kind: `observation`, `decision`, `assumption`,
`attempt`, `verification`, or `handoff`. Attempts require an outcome;
verifications require typed evidence. Handoffs keep `completed`,
`assumptions`, `blockers`, and `nextActions` as four separate bounded lists.
Proposed next actions are inert data, never permission to execute them.

Every record has an exact commit anchor. Optional path anchors bind a
repository-relative path to its exact blob ID, or to `absent`. Optional typed
subjects bind full `issue:`, `proposal:`, `event:`, `policy:`, `pipeline:`, or
`run:` IDs. Applicability is `exact`, `descendants`, or `subject`.
Evidence IDs are explicitly typed as `git:<full-object-id>`,
`event:sha256:<full-event-id>`, or `memory:sha256:<full-memory-id>`.

Content is limited to 65,536 UTF-8 bytes, topics to 32 unique normalized
labels, evidence to 64 unique IDs, paths to 128 entries and 65,536 aggregate
path bytes, and each handoff list to 64 entries with a 65,536-byte aggregate.
Version 0 permits no attachments.

## Lifecycle and the five independent judgments

Projection consumes the verified set and sorts exact IDs; timestamp or arrival
order never chooses a winner. One successor makes the target `superseded`.
Multiple successors make it `branching`. A valid retraction makes it
`retracted`; challenges remain visible alongside lifecycle state. Missing
targets become explicit dependencies rather than silently disappearing.

Keep these judgments separate:

1. Signature validity proves attribution and byte integrity.
2. Policy qualification says the exact actor and kind are listed by policy.
3. Evidence resolution says named objects are available and structurally valid.
4. Semantic truth is a human or agent judgment the protocol cannot prove.
5. Prompt authority is never granted by a memory, handoff, signature, or selection.

## Policy qualification

The optional `memory` section of `.nh/policy.json` is evaluated from the exact
query commit:

```json
{
  "memory": {
    "trustedActors": ["<full-actor-fingerprint>"],
    "trustedKinds": [
      "assumption",
      "attempt",
      "decision",
      "handoff",
      "observation",
      "verification"
    ]
  }
}
```

Both arrays must be sorted and unique. Without the section, valid memory is
`policy-missing`: it is available to explicit inspection but excluded from
default recall. `--include-untrusted` reveals the real non-qualifying class; it
does not upgrade it.

## Record, correct, and hand off

```sh
nh memory record --kind decision --at HEAD --applies descendants \
  --topic architecture --evidence git:$(git rev-parse HEAD) \
  --content "Keep memory streams separate from collaboration actor chains."

nh memory supersede sha256:<full-memory-id> \
  --kind decision --at HEAD --applies descendants \
  --content "Replacement decision with current rationale."

nh memory retract sha256:<full-memory-id> --reason incorrect
nh memory challenge sha256:<another-actor-memory-id> \
  --reason evidence-mismatch --evidence git:$(git rev-parse HEAD)

nh memory handoff --at HEAD --applies descendants --input handoff.json --json
```

Machine record and handoff requests are strict version-0 JSON. For handoff
input, `--at` and `--applies` supply omitted anchor/applicability context; if
the JSON also supplies either field it must match exactly. Other record flags
cannot be combined with `--input`. Unknown or duplicate fields, conflicts,
invalid UTF-8, and oversized input fail before append. Nichthub reads only
explicitly supplied fields; it does not capture prompts, responses, terminal
history, environment, clipboard, credentials, or unrelated working-tree files.

## Bounded recall

```sh
nh memory recall --at HEAD --topic architecture --query "stream isolation" --json
nh memory recall --at HEAD --include-untrusted --lifecycle all --json
nh memory show sha256:<full-memory-id> --json
```

Recall filters exact commit, subject, path, topic, kind, actor, lifecycle,
trust, and deterministic lexical terms. Defaults are active lifecycle,
policy-qualified trust, at most 20 records, and at most 65,536 bytes of encoded
`data.content`. Stable order is applicability class, lifecycle class, signed
timestamp descending, then full memory ID. Every result carries full memory,
actor, and stream IDs; anchor; signature;
lifecycle edges; applicability; evidence details; trust class; content digest;
and nested inert data.

JSON always places the constant warning outside author data. Author text stays
beneath `memories[].data.content`, and handoff fields stay beneath
`memories[].data.handoff`. Continuation cursors bind the normalized filters,
bounds, exact policy digest, and accepted-source fingerprint. Any change to
sources, policy, filters, query, commit, or bounds invalidates the cursor.
Recall performs no fetch, command execution, tool invocation, policy change,
event append, or ref update.

## Disposable local index

```sh
nh memory index rebuild
nh memory index verify
```

The private JSON index is `.git/nh/memory/index-v0.json`. Its source
fingerprint binds the exact policy digest plus sorted `(ref, head)` pairs.
Directories are owner-only and the file is mode `0600`; symlinks, unknown
fields, malformed records, incompatible versions, and stale sources fail
closed. The index is derived, noncanonical, and never transported. Delete it
at any time and rebuild from verified local and accepted refs.

## Select and synchronize

```sh
nh replication select origin \
  --memory sha256:<full-stream-id> \
  --max-events 10000 \
  --max-objects 30000 \
  --max-total-bytes 134217728
nh sync origin
```

Selection authorizes transport, not trust. Exact streams enter a separate bare
quarantine, are measured and validated, and promote only to accepted remote
refs. Actor, proposal, and memory selectors have independent outcomes. An
explicit actor-only selection imports no memory. `--all` discovers all three
namespaces but remains mutually exclusive with exact selectors.

`nh sync` also publishes local memory refs with explicit Git refspecs; it does
not publish the primary branch. In a fresh clone, save exact stream selectors,
sync, then rebuild. The clone receives neither another actor's private key nor
the publisher's index and cannot author as that actor. Exact selected shallow
recovery is available through `nh sync origin --recover-shallow`; it never
globally unshallows or silently adds a selector.

## Limits

Version 0 has no federation, automatic transcript capture, embeddings,
semantic truth engine, autonomous action, provider service, moderation,
redaction, retention enforcement, or erasure guarantee. Retraction and local
index deletion do not erase Git objects already replicated elsewhere. See the
[memory safety model](memory-safety.md), [replication protocol](replication-v0.md),
and [repeatable direct host observations](host-compatibility.md), including the
dated public two-actor memory proof with full IDs and OIDs.
