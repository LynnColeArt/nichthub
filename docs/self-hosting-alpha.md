# Operational self-hosting alpha record

This record has two evidence layers:

1. deterministic offline black-box acceptance over disposable repositories;
2. a staged public landing verified through ordinary Git transport.

The protocol and threat boundaries are defined in
[protocol-v0](protocol-v0.md), [identity-v0](identity-v0.md),
[governance-v0](governance-v0.md), and
[replication-v0](replication-v0.md). Full IDs are used for every public
trust-bearing value. Short IDs shown by some inspection commands are display
conveniences only.

## Offline black-box evidence

Observed on 2026-08-30 with the release-candidate source in this repository:

```sh
go test ./... -run TestOperationalSelfHostingAlpha -count=3 -v
```

The complete three-count command passed in 45.89 seconds. Its three full
two-actor operational scenarios completed in 9.13, 9.01, and 8.98 seconds,
each below the 120-second bound.

### Topology and governance assertions

Every run generated a fresh author clone, reviewer clone, identity-free
depth-1 verifier clone, and local bare Git remote. The test generated distinct
Ed25519 actors independently; it failed if actor or public-key material was
reused or if the reviewer received a legacy identity file.

The synthetic baseline policy trusted only the author actor, allowed author
approval, and had no required pipeline. The policy amendment added the second
actor as trusted reviewer and runner, disabled author approval, and required
one passing `test` result. The test asserted:

- both full base/head policy digests were reported;
- the second actor's signed review of the amendment was valid but contributed
  zero trusted approvals under the old base policy;
- the amendment decision signed only the old-policy author's qualifying
  review, its exact base-policy digest, and no new-actor evidence;
- a later candidate based on the amended policy remained blocked after its
  author's review and run request;
- it became ready only after the second actor supplied the exact review and
  passing run result;
- decision and merge facts bound the exact candidate, pipeline definition,
  proposed head, policy digest, evidence IDs, and resulting Git commit.

The automated runner used the explicitly authorized host backend only for a
controlled synthetic `printf` pipeline. The live public proof records the
default Bubblewrap result separately.

### Selection, budgets, and isolation

The main scenario saved these positive per-remote budgets:

```text
max-events: 10000
max-objects: 100000
max-object-bytes: 16777216
max-attachment-bytes: 1048576
max-total-bytes: 268435456
```

Focused boundary tests measured every dimension at one below, exactly equal,
and one above the configured limit: measured-above-limit rejected; equality
and measured-below-limit promoted.

The hostile remote scenario selected one valid actor, one two-event actor over
an event limit of one, one ref whose root was not an event commit, and one
well-formed ref name pointing at another actor's chain. It also published an
unselected valid actor. The result was:

| Category | Accepted projection result |
| --- | --- |
| Independently valid selected actor | Promoted and usable |
| Selected over-budget actor | Rejected; no accepted ref |
| Selected non-event root | Structurally invalid; no accepted ref |
| Selected actor/ref mismatch | Structurally invalid; no accepted ref |
| Unselected actor | Not requested and absent from accepted refs |
| Candidate selected without supplying actor history | Dependency missing with full candidate ID and exact actor-selection action |

Crash tests interrupted before object copy, after object copy, before accepted
ref commit, after accepted ref commit, and before completion recording. Missing,
corrupt, unsafe-mode, symlinked, oversized, or internally inconsistent private
anchors/receipts kept trust operations fail-closed. A retry reconciled the
exact durable transaction rather than treating copied objects as accepted.

### Shallow recovery and identity projection

The depth-limited fixture detected an exact missing actor predecessor, reported
the full event/actor IDs and `nh sync origin --recover-shallow`, promoted only
the selected supplier through quarantine, replayed the original operation,
and remained a shallow repository. An available but unselected supplier was
not added implicitly; recovery refused until the saved exact selection named
it.

The main identity-free verifier reconstructed the amendment, role-distinct
candidate, reviews, run request/result, decisions, merges, and planned
successor authorization/acceptance without any private identity records or
local actor ref. A separate two-actor successor-cycle fixture converged on
`ambiguous` for both actors while policy continued to list only its explicit
maintainer.

### Stable pre-identity event IDs

The deterministic compatibility fixture proves the new optional identity
fields do not change existing payload bytes or IDs:

| Existing kind | Exact fixture event ID |
| --- | --- |
| `issue.open` | `sha256:86d61551b5233d9bf0ae98eb506abb53a3406d7622c7c66b1a34c2f244812873` |
| `issue.comment` | `sha256:47d4e38ecb1b55d5bd81a72de95dc6ab6c7040115991a196b65d6aa71c032ace` |
| `proposal.open` | `sha256:52eb04da80dcc2cb36389f47b516fcdc1d2e948fd5de7c992990a3d03d931e88` |
| `proposal.revise` | `sha256:a884e31c401c361ca39236ef9c2c759b3b7d1c582ef4c3d4da71671fe016f424` |
| `review.submit` | `sha256:6bc412d82f000130dd59acdd40358fd852fa1714bef5291762c3995eba31850f` |
| `run.request` | `sha256:13f9262d57acb402cb46682d925d7904ef58bf7d77a8a942d4168b00d3c2df6a` |
| `run.result` | `sha256:7f3acaeb8e611fe6d8f0e1c9b8514885d6874f5141278e822274ba11f5fe1b17` |
| `proposal.decision` | `sha256:3e6d8f2c4ba428fc04a5cca304a316a20876f91b9ca52138bd7e35e881c4fcec` |
| `proposal.merged` | `sha256:a92000b45474fadd1cb39a5f18821ef34d7c1bac76ed1713083f2d39c86dc7f8` |

The same focused test also preserves the exact standalone legacy fixture ID
`sha256:cc324cc49ad14cb8f75e4ff6f112396d966af50546937770d5adb7b8e979f091`.

### Isolation claims for automated proof

The automated proof used zero public network mutations, zero hosting-provider
API calls, zero Docker operations, and zero copied private identity. It reset
home, credential-helper, askpass, token, and Nichthub test environment surfaces
for subprocesses. Run-specific random actor/event IDs are intentionally not
recorded; the assertions above and stable fixture IDs are reproducible.

## Public staged proof

The exact public record is inserted here only after the documented workflow
passes unchanged against a fresh disposable remote and the explicit public ref
targets are reviewed. No public mutation had occurred when this structure was
created.

### Stage 1: identity continuity and old-policy amendment

The final record will list both actor IDs, authorization/acceptance events,
base and head policy digests, policy candidate and code ref, run request/result,
review, decision, merge event, policy commit, and resulting merge commit. It
will explicitly identify the second actor's non-qualifying amendment evidence.

### Stage 2: role-distinct release candidate

The final record will list the later candidate, exact amended base, code head,
author non-qualifying review, run request, distinct actor result/review,
decision, merge event, resulting commit, and separate collaboration-ref and
primary-branch publication observations.

### Fresh public depth-limited reconstruction

The final record will include the exact selection command, advertised refs,
accepted refs, shallow state, recovered identifiers, primary branch target,
and absence of private identity in a fresh depth-1 clone.

## Verification method

The public verification uses only ordinary Git and the release-candidate `nh`
binary. The final exact values replace the descriptive stage text above.

```sh
git ls-remote --symref origin HEAD
git ls-remote origin 'refs/heads/main' 'refs/nh/actors/*' 'refs/nh/proposals/*'

git clone --depth 1 https://github.com/LynnColeArt/nichthub.git verify
cd verify
nh replication select origin \
  --actor <full-maintainer-actor-from-public-record> \
  --actor <full-reviewer-runner-actor-from-public-record> \
  --proposal <full-policy-candidate-from-public-record> \
  --proposal <full-release-candidate-from-public-record> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin
nh proposal status <full-release-candidate-from-public-record>
nh proposal show <full-release-candidate-from-public-record>
nh identity list
nh log
git rev-parse --is-shallow-repository
git rev-parse origin/main
git for-each-ref --format='%(refname) %(objectname)' refs/nh/remotes
```

If an inspection command records a shallow dependency gap, run
`nh sync origin --recover-shallow`; it uses only an already saved exact
supplier and does not globally unshallow. A successful initial selected sync
may already contain every dependency, in which case recovery is unnecessary
and explicitly calling it would report that no shallow recovery is required.

## Limits of the proof

Two distinct signing keys prove distinct actor histories, not distinct humans.
The public host is evidence of ordinary Git ref/object transport, not of its UI
or collaboration API. Promotion budgets do not prevent standard Git from
downloading a selected pack before measurement. Planned rotation requires the
old key. Public immutable facts cannot carry a global erasure guarantee.
