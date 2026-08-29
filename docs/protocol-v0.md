# Nichthub protocol v0 experiment

This document records the format implemented by the initial prototype. It is a
testable sketch, not a compatibility promise.

## Identity

An actor owns an Ed25519 key pair. Its actor identifier is the lowercase hex
SHA-256 digest of the 32-byte public key:

```text
actor = hex(sha256(ed25519_public_key))
```

The prototype includes the raw-base64 public key in every event. A future
version may replace this repetition with identity events or key documents.

## Event payload

An event is UTF-8 JSON produced without insignificant whitespace. Version 0
uses this fixed field order:

```json
{
  "protocol": "nh/0",
  "kind": "issue.open",
  "actor": "<public-key fingerprint>",
  "actorName": "Alice",
  "publicKey": "<raw-base64 Ed25519 public key>",
  "sequence": 1,
  "timestamp": "2026-08-29T12:00:00Z",
  "title": "An issue title",
  "body": "Optional body"
}
```

Optional empty fields are omitted. `previous` identifies the preceding event
from the same actor. `subject` identifies the event that a comment or other
dependent event concerns. Proposal events add `base` and `head` Git commit
object IDs. Review events add a `verdict` of `approve` or `request-changes`.
Run events use `pipeline`, `definition`, `commit`, `outcome`, `exitCode`,
`durationMs`, `log`, `backend`, `platform`, and `runner` as described below.
Governance events use `policy` and an ordered set of `evidence` event IDs.

The implemented event kinds are:

```text
issue.open
issue.comment
proposal.open
review.submit
run.request
run.result
proposal.decision
proposal.merged
```

The event identifier is independent of Git's configured object hash:

```text
event_id = "sha256:" + hex(sha256(exact_event_payload_bytes))
```

The signature is Ed25519 over the exact event payload bytes. Verification does
not reserialize the JSON.

## Git representation

Every event is represented by a Git commit whose tree contains:

```text
event.json    exact signed payload bytes
signature     raw-base64 Ed25519 signature
```

A `run.result` tree additionally contains `log.txt`. Its exact bytes hash to
the signed `log` event field. Therefore logs travel with the runner's actor
history and can be verified without trusting the remote.

The commit's parent is the preceding commit in that actor's chain. Keeping the
relationship in the Git commit graph makes all earlier events reachable during
fetch and garbage collection. The signed `previous` field independently links
the protocol-level event IDs.

An actor publishes the head of its chain at:

```text
refs/nh/actors/<actor>
```

A proposal also makes its signed `head` commit reachable at an immutable ref
derived from the proposal event ID:

```text
refs/nh/proposals/<event SHA-256 without the "sha256:" prefix>
```

This separate ref is necessary because mentioning a Git object ID inside event
JSON does not make that object reachable to Git's fetch or garbage-collection
machinery. A reviewer must verify that the proposal ref points to the `head`
recorded in the signed event before reviewing it.

Fetched refs are stored locally at:

```text
refs/nh/remotes/<remote>/actors/<actor>
refs/nh/remotes/<remote>/proposals/<proposal>
```

Each actor chain must begin at sequence 1, increase by one, and refer to the
previous event ID. Two different events at the same actor sequence constitute
a fork and are rejected by the current projection.

## Run requests and results

A `run.request` references a proposal through `subject` and binds:

```text
pipeline     repository-local pipeline name
definition   SHA-256 ID of the exact pipeline JSON bytes
commit       proposed Git head commit
```

A `run.result` references the request through `subject` and repeats all three
bindings. Clients reject a result if any repeated value differs from the signed
request. The result also records `passed` or `failed`, its exit code, duration,
the SHA-256 ID of its attached log, and signed claims about its backend,
platform, and runner implementation.

These events are attestations, not proofs of honest execution. A signature says
which runner made the claim and protects its contents. Project policy decides
which runner identities and execution environments count.

## Governance events

A `proposal.decision` references a proposal, the SHA-256 ID of the exact base
policy bytes, and an `accept` or `reject` verdict. Accept decisions include the
signed review and run-result event IDs that satisfy the policy. Rejections
require an explanation.

A `proposal.merged` event binds the proposal, original proposed head, resulting
Git commit, base policy digest, and acceptance-decision IDs. It is emitted only
after Git creates the merge commit.

## Synchronization

For remote `origin`, the fetch refspec is:

```text
+refs/nh/actors/*:refs/nh/remotes/origin/actors/*
```

The client pushes its current identity's actor ref and locally created proposal
refs. No server-side Nichthub process participates.

## Known unanswered questions

- Which hosted Git services accept and faithfully retain the custom ref space?
- How should one identity safely append from multiple devices?
- How are key rotation and identity recovery expressed?
- How do clients limit spam and resource exhaustion before fetching objects?
- How should deletion requests coexist with immutable replicated history?
- Which JSON canonicalization rules should a stable cross-language version use?
- How should extensions be negotiated without fragmenting projections?
