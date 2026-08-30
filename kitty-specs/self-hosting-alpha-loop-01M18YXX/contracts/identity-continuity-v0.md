# Identity Continuity Contract v0

## Status

Experimental contract for the Operational Self-Hosting Alpha mission. It adds
event kinds under the existing `nh/0` experiment without changing the bytes or
identities of existing event kinds.

## Actor invariant

An actor remains exactly one Ed25519 public key fingerprint with exactly one
append-only event chain. A device, person, or key lineage is a projection over
multiple actors; it is never implemented by copying a private actor key.

## Authorization event

Kind: `identity.authorize`

```json
{
  "protocol": "nh/0",
  "kind": "identity.authorize",
  "actor": "<authorizing actor fingerprint>",
  "actorName": "Existing device",
  "publicKey": "<authorizing raw-base64 public key>",
  "sequence": 8,
  "timestamp": "2026-08-30T00:00:00Z",
  "previous": "<authorizing actor previous event ID>",
  "relationship": "device",
  "targetActor": "<target actor fingerprint>",
  "targetKey": "<target raw-base64 public key>",
  "body": "Optional explanation"
}
```

`relationship` is exactly `device` or `successor`. `targetActor` must equal the
fingerprint derived from `targetKey` and must differ from `actor`.

## Acceptance event

Kind: `identity.accept`

```json
{
  "protocol": "nh/0",
  "kind": "identity.accept",
  "actor": "<target actor fingerprint>",
  "actorName": "New device",
  "publicKey": "<target raw-base64 public key>",
  "sequence": 1,
  "timestamp": "2026-08-30T00:01:00Z",
  "subject": "<full identity.authorize event ID>"
}
```

The acceptance signer must exactly match the authorization's `targetActor` and
`targetKey`. Its ordinary actor-chain `previous` field is present when it is not
the target actor's first event.

## Projection

- Authorization without acceptance: pending.
- Matching authorization and acceptance: accepted edge.
- Accepted `device` edge: both actors remain active.
- Exactly one accepted acyclic `successor` edge from an actor: predecessor is
  retired-by-successor in the identity projection.
- Multiple accepted outgoing successors, a successor cycle, or contradictory
  target material: ambiguous/conflicting; no winner is inferred.
- Multiple acceptances of the same exact authorization by the named target do
  not create additional authority; all remain inspectable signed facts.
- Timestamp and cross-actor delivery order never resolve a conflict.

## Governance separation

Identity projection is descriptive. Policy evaluation continues to compare the
event signer's exact actor fingerprint with explicit policy actor lists. A
device or successor whose actor is not listed cannot satisfy a maintainer,
reviewer, runner, or decision requirement.

## Local keyring

Private state is stored below `.git/nh` with owner-only permissions:

```text
.git/nh/
├── active
├── identities/
│   └── <actor>.json
└── rotation.json        # present only while a local rotation is incomplete
```

`active` contains one full actor fingerprint. Each identity record retains the
existing actor/name/publicKey/privateKey values and a local lifecycle marker.
The legacy `.git/nh/identity.json` migrates atomically without changing key
bytes. Private records and transaction state are never Git objects.

## Command behavior

```text
nh identity show
nh identity list
nh identity public
nh identity authorize --relationship device|successor \
  --actor <full-target-actor> --public-key <raw-base64-target-key>
nh identity accept <full-authorization-event-id>
nh identity rotate [--name NAME]
```

- `public` emits only name, actor, and public key.
- `authorize` refuses shortened actor IDs and self-targets.
- `accept` requires the active actor to be the exact target.
- `rotate` generates a distinct local actor, signs both sides without sharing
  private material, persists retry state, and switches `active` only after both
  events are durable.
- A retired actor can be inspected explicitly but is not selected for ordinary
  new events by default.

## Failure and recovery

- Failure before authorization durability: no rotation exists; retry starts
  cleanly.
- Authorization durable but acceptance missing: pending relation; retry creates
  or resolves the exact acceptance.
- Both events durable but active pointer unchanged: retry verifies the event IDs
  and completes the local atomic pointer switch.
- Competing successor facts fetched later: projection becomes ambiguous, but
  existing event signatures and policy decisions remain unchanged.
- Lost predecessor key: outside this contract; do not fabricate acceptance or
  recovery authority.
