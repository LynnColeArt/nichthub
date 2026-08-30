# Identity continuity and local keyring v0

Identity has two deliberately separate surfaces:

| Public signed facts | Private local state |
| --- | --- |
| Actor public keys and fingerprints | Private signing keys |
| `identity.authorize` and `identity.accept` events | Active-actor pointer |
| Actor refs under `refs/nh/actors/*` | Rotation journals and replication selections |
| Deterministic continuity projection | Files below the repository's `.git/nh/` |

Only the left column is stored as Git objects or transported. Never copy
`.git/nh`, an identity record, or a private key to establish a second device.

## Actor invariant

An actor is the lowercase hexadecimal SHA-256 fingerprint of exactly one
Ed25519 public key. It has one append-only event chain and one private signer.
Different devices initialize distinct actors. A person, device group, or key
lineage is a projection over signed relationships among those actors—not a
shared actor key and not a policy principal created by inference.

`nh identity public` emits only display name, full actor fingerprint, and
raw-base64 public key. `nh identity show` emits the active name, actor, and
actor ref; it never emits key material.

## Mutual relationship facts

The authorizing actor publishes:

```sh
nh identity authorize \
  --relationship device \
  --actor <full-64-hex-target-actor> \
  --public-key <raw-base64-target-public-key>
```

The target actor independently publishes:

```sh
nh identity accept <full-sha256-authorization-event-id>
```

`authorize` rejects shortened/malformed actors, unsupported relationship
values, public keys whose fingerprint differs from the named actor, and
self-targets. `accept` resolves the full signed authorization and requires the
active actor and public key to match its exact target.

`device` keeps both actors active in the projection. `successor` expresses
planned replacement: exactly one accepted acyclic successor retires the
predecessor in the identity projection. Neither relationship changes project
policy.

## Separate-device operator flow

In each clone, initialize independently:

```sh
# Existing device
nh init --name "Existing device"
nh identity public

# New device, in a separate clone
nh init --name "Review device"
nh identity public
```

Exchange only the printed public fields. On the existing device, authorize the
new actor and publish collaboration refs:

```sh
nh identity authorize --relationship device \
  --actor <full-new-device-actor> \
  --public-key <new-device-public-key>
nh sync origin
```

Save an exact selection in the new clone, synchronize the authorizing actor,
and accept the full authorization event ID:

```sh
nh replication select origin \
  --actor <full-existing-device-actor> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin
nh identity accept <full-sha256-authorization-event-id>
nh sync origin
```

Then add both full actor selectors in each clone, synchronize, and inspect:

```sh
nh replication select origin \
  --actor <full-existing-device-actor> \
  --actor <full-new-device-actor> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin
nh identity list
```

The relationship can now be reconstructed from public facts in an
identity-free clone. To give the new actor a reviewer, runner, maintainer, or
decision role, amend `.nh/policy.json` through the ordinary base-governed
candidate workflow.

## Deterministic projection

Projection validates the signed payload of every identity event rather than
trusting decoded fields supplied by a caller. Output ordering is derived from
IDs, never arrival order.

- Authorization without matching acceptance is `pending`.
- One or more matching acceptances produce one `accepted` edge; duplicate
  acceptance facts remain visible and add no authority.
- An acceptance whose authorization is unavailable is reported with both full
  acceptance and subject IDs. No relation is inferred.
- A signer or public-key mismatch is a conflict.
- Multiple accepted outgoing successors are `competing-successors`; the
  predecessor becomes ambiguous and no successor wins.
- An accepted successor cycle marks its members ambiguous and reports every
  involved actor/event ID.
- A conflict fetched later changes the derived projection, but never changes
  existing payloads, signatures, decisions, or base-policy authority.
- Timestamps and cross-actor delivery order never resolve a conflict.

## Local keyring and permissions

Current local state is rooted below the repository's private Git directory:

```text
.git/nh/
├── active
├── identities/
│   └── <full-actor-fingerprint>.json
├── identity.json             # optional legacy migration source
├── rotation.json             # only while rotation is incomplete
└── rotation-command.json     # only while rotation/retry is incomplete
```

The actual Git directory may be worktree-specific; the paths above are logical
and must never be committed or copied into documentation evidence. Private
directories must be regular directories with mode `0700`. Private records,
the active pointer, and transaction journals must be regular non-symlink files
with mode `0600`. Unknown fields, trailing JSON, oversized state, wrong actor
paths, mismatched key pairs, symlinks, and unsafe modes fail closed.

Writes use a mode-`0600` temporary file, file sync, atomic rename, and directory
sync. Legacy `.git/nh/identity.json` migration recomputes and preserves the
same actor/public/private key bytes, writes the actor record durably, and then
creates the active pointer. Retrying does not generate a replacement actor.
The legacy source may remain on disk; once `active` exists, ordinary loading
uses the keyring record it names.

## Planned rotation

```sh
nh identity rotate --name "Rotated device"
nh identity show
nh identity list
nh sync origin
```

Rotation creates a distinct target actor and a `successor` authorization from
the predecessor plus an acceptance from the target. The predecessor remains
active until both exact signed events are durable and verified. Only then does
the active pointer switch. Completed rotation journals are removed.

If interrupted, the private journals bind the predecessor, target key, and any
durable event IDs. A retry verifies those exact signed events, appends only a
missing side, and completes the pointer switch. It refuses replacement event
IDs, changed target material, competing matching facts, or a changed active
predecessor. The predecessor history remains public and inspectable.

Rotation does not give the successor any inherited project role. A separate
policy amendment is required if the successor should become authoritative.

## Explicit deferrals

The operational alpha does not implement:

- lost-key recovery or rotation after the predecessor key is unavailable;
- compromise response or revocation with global erasure;
- social, organizational, or threshold recovery;
- concurrent disconnected writers sharing one actor/private key;
- implicit policy migration to devices or successors;
- a claim that distinct keys prove distinct people.

Published continuity facts are immutable and may be superseded or made
ambiguous by later facts. Nichthub cannot remove copies already replicated by
others.
