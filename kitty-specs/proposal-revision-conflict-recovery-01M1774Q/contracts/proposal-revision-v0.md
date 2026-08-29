# Contract: Proposal Revision v0

This contract extends the experimental `nh/0` event language. It is not a
stable-version compatibility promise.

## Signed Payload

```json
{
  "protocol": "nh/0",
  "kind": "proposal.revise",
  "actor": "<predecessor-author fingerprint>",
  "actorName": "Alice",
  "publicKey": "<raw-base64 Ed25519 public key>",
  "sequence": 8,
  "timestamp": "2026-08-29T18:00:00Z",
  "previous": "sha256:<author's prior event>",
  "subject": "sha256:<direct predecessor proposal>",
  "body": "Resolved against the new target base.",
  "base": "<exact Git commit OID>",
  "head": "<exact Git commit OID>"
}
```

Empty optional fields are omitted under the existing event encoder. The event
ID and signature use the existing exact-byte SHA-256 and Ed25519 rules.

## Validation Contract

A client accepts the relationship only when all conditions hold:

1. the payload, actor identity, signature, sequence, timestamp, and previous
   link pass existing event validation;
2. `subject` is a valid event ID resolving to a verified `proposal.open` or
   `proposal.revise`;
3. `base` and `head` are valid, distinct Git commit OIDs;
4. the revision actor equals the predecessor actor;
5. adding the predecessor edge does not produce a self-link or cycle; and
6. the immutable proposal code ref, when available, points to the signed head.

Relationship validation uses the complete locally verified event set, not
timestamp or arrival order. Invalid relationships do not influence lineage,
readiness, acceptance, or merge eligibility.

Merge state is deliberately not a reception-validity condition. A local create
command refuses an already known merged predecessor, but a receiver preserves
an otherwise valid revision alongside a merge fact because cross-actor event
order cannot be proven from delivery order or timestamps.

## CLI Contract

### Create a revision

```text
nh proposal revise PREDECESSOR --base REV --head REV [--body TEXT]
```

- `PREDECESSOR` must resolve unambiguously to an exact verified proposal.
- The loaded identity must be the predecessor author.
- The predecessor must not be locally known as merged.
- `REV` arguments resolve through Git to exact commit IDs and must differ.
- Success prints the new revision ID, predecessor ID, and exact code range.
- If event append succeeds but code-ref publication fails, the error prints the
  created revision ID and does not change any existing proposal ref.

### Inspect lineage

```text
nh proposal list
nh proposal show PROPOSAL
nh proposal status PROPOSAL
```

For a revision, output includes its exact predecessor. For any lineage member,
show/status includes known exact successors, siblings, merged members, and
merge-conflict state when present. Stable collections sort by full proposal ID.
The words `latest`, `current revision`, and implicit winner are not used.

### Evidence and governance

```text
nh review PROPOSAL ...
nh run request PROPOSAL PIPELINE
nh decide PROPOSAL --accept
nh merge PROPOSAL
```

All commands still require an explicit proposal ID. Reviews, runs, decisions,
policy, and code checks apply only to that exact ID. A new acceptance or merge
fails when the candidate is superseded or another lineage member is known
merged. Errors include the exact blocking proposal IDs.

Rejection remains allowed as an immutable exact-proposal fact. Previously
created evidence and acceptance facts remain inspectable after supersession.

## Sync Contract

No refspec is added. Revision events travel on the author's existing actor ref,
and revision code travels at the existing per-proposal ref:

```text
refs/nh/actors/<actor>
refs/nh/proposals/<revision event SHA-256 without prefix>
```

The existing `refs/nh/remotes/<remote>/...` projection applies unchanged.

## Compatibility Contract

- A history with no `proposal.revise` events has unchanged behavior.
- An older client encountering `proposal.revise` rejects the unknown kind and
  therefore cannot silently perform a lineage-unsafe merge.
- A newer client accepts both original and revision proposal candidates at all
  exact-proposal command boundaries.
- Same-private-key concurrent publication from disconnected clones remains
  unsupported because actor histories are single-writer linear chains.
