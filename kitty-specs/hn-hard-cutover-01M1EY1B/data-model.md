# Data Model: `hn` Hard Cutover

## Namespace Surface

| Surface | Active value | Legacy value | Legacy disposition |
|---|---|---|---|
| Executable / CLI | `hn` | `nh` | Not shipped |
| Collaboration protocol | `hn/0` | `nh/0` | Rejected |
| Memory protocol | `hn-memory/0` | `nh-memory/0` | Rejected |
| Pipeline schema | `hn.pipeline/0` | `nh.pipeline/0` | Rejected |
| Policy schema | `hn.policy/0` | `nh.policy/0` | Rejected |
| Git ref root | `refs/hn/` | `refs/nh/` | Ignored, historically retained |
| Private state root | `.git/hn/` | `.git/nh/` | Ignored, never migrated |
| Repository config | `.hn/` | `.nh/` | Ignored, frozen for transition evidence |
| Environment prefix | `HN_` | `NH_` | Ignored |
| Runner label | `hn/<version>` | `nh/<version>` | Old result not accepted by new policy evidence |

## Active Git namespace

```text
refs/hn/
├── actors/<actor-fingerprint>
├── proposals/<candidate-event-digest>
├── memory/<actor-fingerprint>/<stream-digest>
├── remotes/<remote>/{actors,proposals,memory}/...
└── quarantine/{actors,proposals,memory}/...
```

All readers, writers, validators, fetch refspecs, push refspecs, transaction
records, and shallow-recovery anchors share this root. There is no fallback
root.

## Active private namespace

```text
.git/hn/
├── identities/
├── identity.json
├── memory/index-v0.json
└── replication/
    ├── selections/
    ├── transactions/
    ├── anchors/
    └── quarantine/
```

Existing permission, symlink, bounded-read, atomic-write, and fsync invariants
remain unchanged.

## Active repository configuration

```text
.hn/
├── policy.json
├── pipelines/<pipeline>.json
└── actions/<executable>
```

The active policy's maintainer, reviewer, runner, and memory trust lists contain
new `hn` actor fingerprints. The frozen `.nh/` tree is outside this model.

## Cutover Identity

Attributes:

- `actor`: SHA-256 fingerprint derived from an Ed25519 public key.
- `name`: local display name.
- `public_key`: base64 public key exposed for authorization.
- `private_key`: local keyring material below `.git/hn/`, never committed.
- `sequence`: begins independently at one for the actor's first `hn/0` event.

Relationships:

- Active policy names actor fingerprints.
- Actor refs contain signed collaboration chains.
- Memory refs are actor-owned but use the distinct `hn-memory/0` envelope.
- Legacy and active identities have no continuity edge.

## Transition Boundary

The transition candidate is simultaneously:

1. The head commit named by the final legacy `nh/0` proposal and merge event.
2. The first public code version whose runtime recognizes only `hn` state.
3. The policy base from which fresh `hn` governance begins.

This relationship is documentary and Git-object-based; the new runtime does not
parse legacy events to establish it.
