# Data Model: Self-Hosting Alpha Loop

## Purpose

This mission introduces no new protocol entity by default. It composes existing
signed events and Git facts into a bounded verification record for one public
self-hosting run.

## Existing protocol entities

### Actor History

- **Identity**: Actor fingerprint.
- **State**: Ordered signed events with sequence, previous event ID, timestamp,
  actor name, and public key.
- **Invariant**: A history begins at sequence 1 and advances by exactly one;
  fetched forks or invalid signatures fail verification.
- **Relationship**: Contains every collaboration event produced by one actor.

### Issue

- **Identity**: Full `sha256:` event ID of an `issue.open` event.
- **Attributes**: Actor, timestamp, title, body.
- **Relationship**: May have zero or more `issue.comment` events.
- **Mission use**: Names and explains the self-hosting work. Version 0 does not
  encode a signed issue-to-candidate relationship.

### Candidate

- **Identity**: Full event ID of `proposal.open` or `proposal.revise`.
- **Attributes**: Author, base commit, head commit, title or inherited title,
  optional description, optional predecessor.
- **Invariant**: The candidate is immutable and its content-bearing proposal
  ref must resolve to its signed head.
- **Relationships**: Receives reviews; owns run requests and decisions; may be
  represented by one merge fact; revisions form a preserved lineage.

### Run Request

- **Identity**: Full `run.request` event ID.
- **Attributes**: Candidate ID, pipeline name, exact pipeline-definition
  digest, candidate head commit.
- **Invariant**: All bindings match the exact candidate and repository bytes.
- **Relationship**: Has current results keyed by runner actor.

### Run Result

- **Identity**: Full `run.result` event ID.
- **Attributes**: Request ID, repeated candidate bindings, outcome, exit code,
  duration, log digest, backend, platform, runner version.
- **Invariant**: Attached log bytes match the signed digest; repeated request
  bindings match exactly.

### Review

- **Identity**: Full `review.submit` event ID.
- **Attributes**: Candidate ID, reviewer actor, verdict, optional body.
- **Invariant**: Only the latest review per actor is current for a candidate;
  evidence never transfers between lineage members.

### Decision

- **Identity**: Full `proposal.decision` event ID.
- **Attributes**: Candidate ID, maintainer actor, verdict, base-policy digest,
  ordered qualifying evidence IDs, optional body.
- **Invariant**: An acceptance can be created only when the exact candidate is
  ready under the policy bytes from its signed base.

### Merge Fact

- **Identity**: Full `proposal.merged` event ID.
- **Attributes**: Candidate ID, candidate head, resulting merge commit,
  base-policy digest, acceptance decision IDs.
- **Invariant**: It records an observed Git result and never substitutes a
  sibling or predecessor's evidence.

## Mission verification entity

### Self-Hosting Verification Record

This is documentation, not a new signed protocol object.

- **Identity**: Mission slug plus the inaugural issue ID.
- **Public facts**:
  - remote address;
  - issue ID;
  - candidate ID and code-ref name;
  - run-request and passing-result IDs;
  - review ID;
  - acceptance-decision ID;
  - merge-event ID;
  - candidate base and head commits;
  - resulting merge commit;
  - published primary-branch commit.
- **Verification observations**:
  - custom refs advertised by the remote;
  - event count and exact IDs reconstructed in a fresh clone;
  - proposal head/ref equality;
  - policy readiness and merged state;
  - absence of copied private identity state.
- **Limitations**: Role independence, host longevity, moderation, multi-device
  append, and provider-neutral breadth are explicitly not claimed.

## Relationships

```text
Issue (human-correlated in v0)
  |
  v
Candidate --< Review
  |
  +--< Run Request --< Run Result
  |
  +--< Decision --references--> Review + Run Result
  |
  +---- Merge Fact --references--> Decision
          |
          v
     resulting Git commit

Self-Hosting Verification Record references every full identity above.
```

## Lifecycle

1. The issue is opened and published.
2. A candidate binds the feature branch head to its main-branch base and makes
   the code reachable.
3. A request binds the exact candidate pipeline; a trusted runner produces a
   result.
4. A trusted reviewer publishes a verdict.
5. A maintainer decision signs the evidence that satisfies the base policy.
6. The accepted candidate is merged locally and a merge fact records the
   resulting commit.
7. Collaboration refs and the primary branch are published separately.
8. A fresh clone synchronizes and independently verifies the record.

Failure does not move an entity backward or alter it. A new run result,
review, decision, or proposal revision supersedes prior current state only
through an additional signed fact under the existing projection rules.
