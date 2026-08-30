# Policy Amendment CLI Contract v0

## Principle

A policy amendment is an ordinary proposal containing a Git change to
`.nh/policy.json`. No policy event or administrator side channel is introduced.
The proposal is always governed by the exact policy bytes from its signed base.

## Inspect

```text
nh policy show [REV]
```

Defaults to `HEAD`. On success it reports:

- resolved commit object ID;
- policy digest over exact bytes;
- maintainers and required acceptance count;
- trusted reviewers, approval threshold, and author-approval rule;
- each pipeline's required results and trusted runners.

Actor IDs are printed in full in trust-bearing output.

## Check an amendment

```text
nh policy check --base REV --head REV
nh policy check --base REV --file PATH
```

Exactly one of `--head` and `--file` is required. Both base and proposed policy
pass the canonical parser and validator used by candidate evaluation. `--file`
reads an explicit working-tree draft and does not modify it.

On success the command reports:

- resolved base commit and base policy digest;
- proposed commit when `--head` is used and proposed policy digest;
- added/removed maintainers and required-acceptance change;
- added/removed trusted reviewers, approval-threshold change, and
  author-approval change;
- added/removed pipelines;
- per-pipeline runner additions/removals and result-threshold changes;
- an explicit statement that the base policy governs the amendment candidate.

Output ordering is deterministic: pipeline names and actor fingerprints are
lexicographically sorted.

## Rejection

The check fails before proposal preparation when either policy is malformed,
unknown fields are present, no maintainer remains, an actor is duplicated or
malformed, or a configured threshold is unsatisfiable. Errors identify the
exact policy side (`base` or `proposed`) and violated rule without printing
private data.

## Proposal interaction

`nh proposal open` retains its existing interface. When base and head contain
different valid policy bytes, it additionally reports both full policy digests
and states that the base digest governs the candidate. It never treats the head
policy's actors as qualifying evidence for that candidate.

## Identity continuity interaction

The checker may display that an added actor participates in a verified device
or successor relationship when those facts are locally available. This is
informational only. Policy validity and authority depend on the exact actor ID,
not its relationship projection.
