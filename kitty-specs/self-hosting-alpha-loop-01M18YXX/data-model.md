# Data Model: Operational Self-Hosting Alpha

## Purpose

This mission adds identity-continuity facts and local replication state while
retaining existing proposal and governance entities. Identity relationships
describe continuity among signing actors; policy alone grants project roles.

## Protocol entities

### Actor

- **Identity**: Lowercase SHA-256 fingerprint of one Ed25519 public key.
- **State**: One append-only sequence of exact-byte signed events.
- **Invariant**: One private key has one writer at a time; sequence and
  predecessor form a complete chain from event 1.
- **Relationships**: May authorize other actors as devices or successors and
  may accept authorizations targeting itself.

### Identity Authorization

- **Identity**: Full event ID.
- **Signer**: Existing actor creating the outgoing relationship.
- **Attributes**:
  - relationship: `device` or `successor`;
  - target actor fingerprint;
  - target raw public key;
  - optional human-readable explanation.
- **Validation**:
  - target fingerprint matches the target public key;
  - target differs from signer;
  - relationship is recognized;
  - all ordinary actor-chain and signature rules pass.
- **State**: `pending`, `accepted`, or `conflicting` in projection.

### Identity Acceptance

- **Identity**: Full event ID.
- **Signer**: The exact target actor named by an Identity Authorization.
- **Attributes**: Subject authorization event ID.
- **Validation**:
  - subject is an available Identity Authorization;
  - signer actor and public key exactly match its target;
  - one acceptance cannot reinterpret the authorization relationship.
- **Effect**: Completes a device or successor edge; grants no project role.

### Identity Projection

- **Inputs**: All verified Identity Authorizations and Identity Acceptances in
  the accepted selected histories.
- **Per-edge state**: pending, accepted, conflicting, or dependency-missing.
- **Per-actor state**: active, retired-by-successor, ambiguous-successor, or
  unrelated.
- **Invariants**:
  - accepted `successor` edges are directed predecessor-to-successor;
  - cycles and multiple accepted successors make succession ambiguous;
  - no timestamp, delivery order, display name, or sequence across actors
    chooses a winner;
  - device edges do not retire either actor;
  - projection never mutates policy roles.

### Policy Amendment

Not a new event type. It is a Candidate whose head contains policy bytes that
differ from its base.

- **Identity**: Candidate event ID.
- **Base policy**: Exact bytes and digest at the signed base commit; governs the
  amendment itself.
- **Head policy**: Exact proposed bytes and digest; governs only later
  candidates whose base contains it.
- **Derived change set**:
  - added and removed maintainers;
  - added and removed trusted reviewers;
  - author-approval change;
  - required-approval and required-acceptance changes;
  - per-pipeline added/removed trusted runners and result thresholds.
- **Invariants**: Both policies validate independently; the head cannot
  authorize evidence for its own candidate.

### Candidate, Evidence, Decision, and Merge Fact

Existing entities retain their version-0 shapes and identities.

- Candidate binds exact base and head commits and has a matching code ref.
- Reviews and run evidence reference one exact candidate.
- Acceptance decisions sign the base-policy digest and qualifying evidence.
- Merge facts sign candidate head, result commit, policy, and decisions.
- Identity continuity never changes which actors qualify; the base policy's
  explicit actor lists remain authoritative.

## Local private entities

### Identity Keyring

- **Location boundary**: Private repository-local state below `.git/nh`.
- **Attributes**:
  - schema version;
  - active actor fingerprint;
  - actor records containing display name, public key, private key, and local
    lifecycle state;
  - optional in-progress rotation transaction with predecessor, target, and
    completed event IDs.
- **Invariants**:
  - private key files are owner-readable only;
  - each stored key pair recomputes its actor fingerprint;
  - active actor names exactly one usable non-retired record;
  - switching the active actor uses atomic local replacement;
  - migration from the legacy identity file preserves key bytes and actor ID;
  - no tracked or signed artifact contains private fields.

### Replication Selection

- **Location boundary**: Untracked repository-local per-remote state below
  `.git/nh`.
- **Identity**: Remote name.
- **Attributes**:
  - selected full actor fingerprints;
  - selected full candidate event IDs;
  - explicit all-ref mode, when chosen;
  - positive event-count, object-count, attachment-size, object-size, and
    reachable-byte budgets.
- **Invariants**:
  - duplicate, shortened, malformed, or negative selections are rejected;
  - selection affects replication only, never policy authorization;
  - shallow recovery cannot add actors or candidates silently.

### Replication Transaction

- **Identity**: Generated local transaction ID.
- **State machine**:

```text
created -> fetched -> measured -> structurally_validated
        -> relationships_validated -> promoted -> complete
        \-> rejected
```

- **Attributes**:
  - remote and exact requested refs;
  - quarantine repository path;
  - advertised and fetched object IDs;
  - per-selection event/object/byte measurements;
  - validation outcomes and missing dependencies;
  - accepted atomic ref-update set.
- **Invariants**:
  - no accepted remote-tracking ref points to transaction objects before
    promotion;
  - rejected transactions cannot alter accepted refs;
  - promotion updates the validated ref set atomically;
  - valid independent selections can promote when another selection fails;
  - no private credentials or paths enter signed events or logs.

### Quarantined Selection

- **Identity**: Remote plus actor or candidate full ID within a transaction.
- **State**: fetched, over-budget, structurally-invalid, dependency-missing,
  relationship-invalid, or promotable.
- **Measurements**: actor events, reachable Git objects, total reachable bytes,
  largest object, attachments and sizes.
- **Outcome**: Atomic promotion target or exact rejection report.

### Shallow Dependency Gap

- **Identity**: Operation plus missing object or event ID.
- **Kinds**: actor predecessor, candidate event, proposal code ref, base commit,
  policy blob, pipeline definition, run request, decision, merge ancestor.
- **Attributes**: selected remote, owning actor/candidate when known, shallow
  status, required ref, estimated or measured budget data, recovery command.
- **Invariant**: A trust-sensitive operation cannot advance while a required
  gap remains unresolved.

## Public verification entity

### Operational Verification Record

Documentation rather than a new protocol object.

- **Policy-amendment stage**:
  - base/head policy digests;
  - candidate, CI, review, decision, merge, and Git commit IDs;
  - actor authorized under the old policy.
- **Role-distinct stage**:
  - author, reviewer, runner, and maintainer actor IDs;
  - candidate and exact evidence IDs;
  - amended base-policy digest;
  - resulting merge event and Git commit.
- **Identity stage**:
  - authorization and acceptance IDs;
  - device or successor actor IDs;
  - projected continuity and explicit policy-authority result.
- **Replication/shallow observations**:
  - exact selected refs and budgets;
  - promoted/rejected selections;
  - shallow gaps and recovery;
  - fresh-clone reconstructed identifiers;
  - absence of local private identity.

## Relationships

```text
Actor A -- Identity Authorization --> Actor B
Actor B -- Identity Acceptance ----> authorization event
                    |
                    v
          Identity Projection only
                    |
                    x  (no implicit authority)
                    |
Base Policy ---- explicit actor lists ----> qualifying governance claims

Policy Amendment Candidate
  |-- governed by --> base policy digest
  `-- installs ----> head policy for later candidate bases

Replication Selection
  -> Replication Transaction
  -> Quarantined Selections
  -> budget + signature + chain + relationship validation
  -> atomic accepted remote refs
  -> existing domain projection
```

## Lifecycle examples

### Planned rotation

1. A successor actor key is generated locally or on another device.
2. The predecessor signs an authorization naming its actor and public key with
   relationship `successor`.
3. The successor signs an acceptance referencing that authorization.
4. All clones project the predecessor as retired only when both facts are
   available and unambiguous.
5. The local keyring switches its active pointer after both events are durable.
6. A separate policy amendment is required wherever the successor should gain
   a project role.

### Selected synchronization

1. The participant records full actor/candidate IDs and budgets for a remote.
2. A transaction requests only their exact refs into a quarantine repository.
3. Each selection is measured and structurally validated.
4. Relationships are checked against accepted facts plus the selected set.
5. Valid independent selections are promoted atomically; invalid and missing
   dependency selections remain rejected with exact diagnostics.
6. Existing projections consume only accepted remote-tracking refs.

### Shallow operation

1. A command resolves every exact object required by its trust decision.
2. If all objects exist, shallow repository status is irrelevant.
3. If an object is missing at a shallow boundary, the operation emits a
   Shallow Dependency Gap and stops.
4. With explicit approval, the selected ref is fetched through quarantine
   under the same budgets.
5. Verification restarts from accepted state; no prior partial decision is
   reused.
