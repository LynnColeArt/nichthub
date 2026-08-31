# Governance v0

Nichthub policy answers which verified signed claims count for a project. It is
a normal versioned file at `.nh/policy.json`:

```json
{
  "version": "nh.policy/0",
  "maintainers": ["<full-actor-fingerprint>"],
  "proposals": {
    "requiredApprovals": 1,
    "requiredAccepts": 1,
    "trustedReviewers": ["<full-actor-fingerprint>"],
    "allowAuthorApproval": false
  },
  "pipelines": {
    "test": {
      "requiredResults": 1,
      "trustedRunners": ["<full-actor-fingerprint>"]
    }
  }
}
```

Actor values are complete lowercase SHA-256 fingerprints of Ed25519 public
keys. Unknown JSON fields, duplicate or malformed actors, an empty maintainer
list, unsatisfiable thresholds, invalid pipeline names, and unsupported policy
versions are rejected. Policy files are limited to 1 MiB.

## Inspect policy

```sh
nh policy show [REV]
```

The revision defaults to `HEAD` and resolves to one full commit object ID. The
command runs the canonical parser and validator and prints:

- the resolved commit and full `sha256:<64-hex>` digest of the exact policy
  bytes;
- every maintainer and the required acceptance count;
- every trusted reviewer, approval threshold, and author-approval rule;
- pipelines in lexical order, with required results and full trusted-runner
  actor IDs.

Trust-bearing actor IDs are never shortened in this output.

## Validate an amendment

```sh
nh policy check --base REV --head REV
nh policy check --base REV --file PATH
```

Exactly one of `--head` and `--file` is required. Both the base and proposed
policy go through the same canonical validation used during candidate
evaluation. `--file` reads an explicit working-tree draft and does not mutate
it. The command prints resolved commit IDs when applicable, both full policy
digests, and a deterministic comparison of:

- exact-byte change;
- maintainers and required acceptances;
- trusted reviewers, required approvals, and author approval;
- added and removed pipelines;
- required-result and trusted-runner changes for every pipeline.

Pipeline names and actor IDs are sorted. A malformed side is identified as
`base` or `proposed`, and validation fails before a candidate is prepared.

## An amendment is an ordinary candidate

A policy amendment introduces no administrator event or mutable control plane.
It is a normal immutable proposal candidate whose `head` changes
`.nh/policy.json`:

```sh
nh policy check --base main --file .nh/policy.json
git add .nh/policy.json
git commit -m "amend collaboration policy"
nh policy check --base main --head HEAD
nh proposal open --base main --head HEAD \
  --body "Governed only by the signed base policy." \
  "Amend collaboration policy"
```

When valid policy bytes differ, `nh proposal open` prints both full digests and
states which base digest governs. The proposed policy becomes relevant only to
later candidates whose signed base commit contains those bytes. A candidate
cannot add a permissive actor or lower a threshold and then use that proposed
authority to accept itself.

## Immutable evaluation context

A candidate signs both a base and a head commit. Its governance policy is
always loaded from:

```text
<signed-base-commit>:.nh/policy.json
```

The SHA-256 protocol ID of those exact bytes is signed by decisions and merge
facts. Policy authority comes only from the explicit actor fingerprints in
that base document.

Identity-continuity projection and replication selection are deliberately
non-authoritative inputs:

- an accepted device/successor relation may explain an actor's relationship,
  but cannot transfer or inherit a role;
- a selected and cryptographically valid actor history is visible, but its
  evidence counts only if that exact actor appears in the applicable base
  policy;
- a successor, related device, or similarly named actor is never substituted
  for a listed actor.

## Readiness and exact evidence

`nh proposal status <full-candidate-id>` independently derives readiness from
accepted repository data:

1. the candidate code ref exists and equals the signed head;
2. the configured number of distinct trusted reviewers currently approve;
3. author approval counts only when the base policy permits it;
4. every required pipeline has enough distinct passing trusted runners;
5. each run request was signed by a base-policy maintainer and binds the exact
   candidate, head, pipeline name, and pipeline-definition digest.

A reviewer's latest review is current. Change requests are displayed but are
not a protocol-level veto. A runner's latest exact result is current for a
pipeline. Evidence for a predecessor or sibling candidate never qualifies a
revision, even when code or lineage is related.

Before publishing an acceptance, the client reloads the verified event set and
re-evaluates the exact base policy. An accept decision signs the exact review
and passing run-result IDs selected as evidence. A rejection requires an
explanation and does not require readiness.

Only actors listed as base-policy maintainers can publish qualifying decisions.
A maintainer's latest decision is current; a current rejection blocks merge.
The required number of distinct current acceptances must be present.

## Merge and conflict recovery

`nh merge <full-candidate-id>` requires:

- a clean worktree on a named branch;
- a current branch descended from the signed base;
- an accepted candidate ref matching the signed head;
- enough base-policy acceptance decisions and no current rejection;
- the active actor to be a base-policy maintainer.

Immediately before changing Git state, merge reloads facts, policy, evidence,
and lineage gates. Git performs a `--no-ff` merge. A successful merge emits a
signed `proposal.merged` event binding the candidate, proposed head, resulting
merge commit, exact base-policy digest, and acceptance decision IDs.

The merge commit and merge event exist locally before either publication step.
Publish collaboration refs with `nh sync`; publish the primary branch with an
ordinary explicit `git push`. Either action can be retried without rewriting
the signed fact.

On a Git conflict, Nichthub aborts the merge and restores the previously clean
worktree. Resolve the code with Git and publish a new immutable
`proposal.revise` candidate with explicit base and head commits. The original,
siblings, and their evidence remain inspectable. The revision requires new
exact review, CI, and acceptance evidence.

There is no timestamp or delivery-order winner among siblings. A candidate
with successors is superseded; a lineage with a merged member is closed; and
multiple valid merge facts produce a visible merge conflict rather than an
invented winner.
