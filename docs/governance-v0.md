# Governance v0

Nichthub policy answers which signed claims count for a project. It is a normal
versioned file at `.nh/policy.json`:

```json
{
  "version": "nh.policy/0",
  "maintainers": ["<actor>"],
  "proposals": {
    "requiredApprovals": 1,
    "requiredAccepts": 1,
    "trustedReviewers": ["<actor>"],
    "allowAuthorApproval": false
  },
  "pipelines": {
    "test": {
      "requiredResults": 1,
      "trustedRunners": ["<actor>"]
    }
  }
}
```

Actor values are full SHA-256 fingerprints of Ed25519 public keys. Duplicate,
shortened, or malformed fingerprints are rejected.

## Immutable evaluation context

A proposal signs both a base and head Git commit. Its governance policy is
always loaded from:

```text
<signed base commit>:.nh/policy.json
```

The exact policy bytes receive a SHA-256 protocol ID. Decisions sign that ID.
A proposal may change policy for later proposals, but cannot use its proposed
policy to authorize itself.

## Readiness

`nh proposal status` independently derives readiness from repository data:

1. the proposal code ref exists and matches its signed head;
2. the configured number of distinct trusted reviewers currently approve;
3. author approval is included only if policy permits it;
4. every required pipeline has enough distinct passing trusted runners;
5. run requests were signed by maintainers and bind the exact pipeline bytes in
   the proposed commit.

Change requests are displayed but are not a protocol-level veto. A reviewer's
latest review is their current review state.

## Decisions

Only policy maintainers can publish `accept` or `reject` decisions. An accept
is refused locally until readiness requirements pass. It signs the exact
approval and run-result event IDs used as evidence. Other implementations can
verify this evidence without trusting the deciding client.

A maintainer's latest decision is current. A current rejection blocks merging.
The configured number of distinct current maintainer acceptances is required.

Rejections require explanatory text but do not require readiness evidence.

## Merge

`nh merge` requires:

- a clean worktree on a named target branch;
- a target branch descended from the proposal's signed base;
- an available proposal code ref matching the signed head;
- enough signed acceptance decisions and no current rejection;
- a maintainer identity under the proposal's base policy.

Git performs a `--no-ff` merge. On conflict, Nichthub automatically aborts the
merge and restores the previously clean worktree. After success, Nichthub emits
a signed `proposal.merged` event containing the proposal head, resulting merge
commit, policy digest, and acceptance-decision evidence.

There is an unavoidable transaction boundary: Git creates the merge commit
before Nichthub can sign the merge event. If event creation fails afterward,
the CLI reports the resulting commit explicitly for recovery.
