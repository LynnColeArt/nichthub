# Agent Memory Protocol Quickstart

This is the target operator flow for the mission. Commands and JSON fields are
acceptance contracts until implementation lands.

## 1. Record an explicit decision

```bash
nh memory record \
  --kind decision \
  --at HEAD \
  --applies descendants \
  --topic architecture \
  --evidence git:$(git rev-parse HEAD) \
  --content "Keep memory streams separate from collaboration actor chains."
```

The command prints the full memory ID, full stream ID, actor fingerprint, and
exact anchor. It records only the supplied fields; it does not inspect chat,
terminal history, environment variables, or unrelated working-tree content.

For adapters, submit the equivalent versioned JSON without shell interpolation:

```bash
nh memory record --input request.json --json
```

## 2. Record an agent handoff

```bash
nh memory handoff --at HEAD --applies descendants --input handoff.json --json
```

`handoff.json` separates completed work, assumptions, blockers, and proposed
next actions. Next actions are inert claims and grant no authority to execute.

## 3. Amend memory without rewriting history

```bash
nh memory supersede sha256:<full-memory-id> \
  --kind decision --at HEAD --applies descendants \
  --content "Replacement decision with current rationale."

nh memory retract sha256:<full-memory-id> --reason incorrect

nh memory challenge sha256:<another-actor-memory-id> \
  --reason evidence-mismatch \
  --evidence event:sha256:<full-event-id>
```

Supersession and retraction require the target author. A challenge preserves
both authors' facts and never deletes or silently overrides the target.

## 4. Add governed default-recall policy

Amend `.nh/policy.json` through the ordinary Nichthub proposal/review/CI/
decision path:

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

Without this section, signed memory remains inspectable but does not qualify for
default trusted recall.

## 5. Recall bounded inert data

```bash
nh memory recall --at HEAD --topic architecture --query "stream isolation" --json
```

The response contains at most 20 records and 65,536 encoded content bytes by
default. Every item includes the full memory/actor/stream IDs, signature status,
anchor, applicability, lifecycle edges, evidence status, trust classification,
content digest, and nested `data.content`.

To investigate excluded claims explicitly:

```bash
nh memory recall --at HEAD --include-untrusted --lifecycle all --json
```

This exposes classification; it does not upgrade trust.

## 6. Rebuild the local index

```bash
nh memory index rebuild
nh memory index verify
```

Deleting `.git/nh/memory/index-v0.json` loses no canonical memory. A rebuild
uses accepted signed refs and exact local policy bytes without network access.

## 7. Select and synchronize a memory stream

```bash
nh replication select origin \
  --memory sha256:<full-stream-id> \
  --max-events 10000 \
  --max-objects 30000 \
  --max-total-bytes 134217728

nh sync origin
```

Memory imports are quarantined and reported independently. A malformed or
over-budget memory stream cannot hide already-valid issues, proposals, reviews,
CI, or another independently valid memory stream.

If recall reports a shallow-missing exact anchor or evidence object, inspect its
full ID and explicit recovery selector before running:

```bash
nh sync origin --recover-shallow
```

## 8. Fresh-clone verification

In a fresh clone with no private keys or copied index:

```bash
nh replication select origin --memory sha256:<full-stream-id>
nh sync origin
nh memory index rebuild
nh memory recall --at HEAD --json
```

The clone must reconstruct the same memory IDs, lifecycle edges, exact filters,
and inert content envelope. It cannot author as the original actor.
