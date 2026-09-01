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
home, credential-helper, askpass, token, and Hubnot test environment surfaces
for subprocesses. Run-specific random actor/event IDs are intentionally not
recorded; the assertions above and stable fixture IDs are reproducible.

## Public staged proof

Observed on 2026-08-30 against the public ordinary-Git endpoint
`https://github.com/LynnColeArt/hubnot.git`. The same two-stage sequence first
passed against a fresh disposable bare remote. The public run then used the
reviewed explicit source and destination refs below. No wildcard, deletion,
force push, hosting-provider API, or hosting-provider UI supplied evidence.

The initial advertised public state was:

| Ref | Initial object ID |
| --- | --- |
| `refs/heads/main` | `e8b0955a3d10b5677fb0dfaba42580f8f4473080` |
| `refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd` | `2b0e261037e0760f83536fedfad044e95deff1fe` |
| `refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290` | `327199e686dbe0bcc4cc4bf92f10237472b083a8` |

The proof created a distinct reviewer/runner actor and accepted its device
relationship without changing project authority:

| Fact | Event ID | Event commit |
| --- | --- | --- |
| Maintainer actor | `36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd` | initial actor ref above |
| Reviewer/runner actor | `10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db` | accepted actor head listed below |
| Device authorization | `sha256:a46e7813c7ea38265bb5ca690e9481405d60e18ed738b6fd94580bcab6c67f8e` | `7854e2ad2278e345bb3db4b9e4dd372698bc1fa3` |
| Device acceptance | `sha256:ca429d39922dd3ce169d2d94a43542dd399451f49a0d140639f7f83ae2fa410f` | `144da8e22f9bc531861e5c2578e0d020fc8fb641` |

These are public signing histories and continuity claims. The private keys,
active-identity selection, replication selections, and transaction records
remained local and were not published. `nh sync` publishes every local public
actor ref required by the signed history, including predecessor histories;
continuity is not authority.

### Stage 1: identity continuity and old-policy amendment

The candidate changed the policy from
`sha256:f58220c462cc5faaa1448e7c0e6e548f4ce10705b1352ce6ed0a7a5377eb874c`
to
`sha256:e4f68ce6e0d26212683b983a5a99c08a34b34dba8b5cb390fbde0a4d484f0d05`.
The proposed policy/code commit was
`a3ef855ca062c191a6c7e5c14709de67bdfb4f1c`.

| Fact | Event ID | Event commit |
| --- | --- | --- |
| Policy candidate | `sha256:52aee3a8a6fd110321b3506e9408397158d02d3dfe63747cb445effa157e0a95` | `d439bcd11b20b2be5e99cd62cce790fd32bb3075` |
| New actor review, valid but `0/1` under the old policy | `sha256:ac1304c5463228c3bb82a9e7a100ced52469347205a3d76a313539daea8c3e4d` | `55842fb308c3f78ad028f0d3be843beb83756209` |
| Old-policy qualifying review | `sha256:586c31d21080e9ca0205213872490e122340e58d85a4c7c246233d63f32070ae` | `fa3fc5b153edfc9240caa23d067b90c4b6fef4f1` |
| `test` request | `sha256:8c3df609474d349379746ca7a4afd5590774c0987f41b6981c53e96a6e91a48b` | `f11b05ea5db56e43b665b3728c415fb1616005cc` |
| Sandbox pass | `sha256:76403d577fa470807f135dc79c39b18c95a475509e72db7e9d6992d57bf6c532` | `0608aa7eedc834b5ed4603bb722665adc25a5abd` |
| Old-policy decision | `sha256:8cb000550496225eff39ea2a08fe3b11471f9fe7af8c56ce55408e5a66365b74` | `68866a9ea894bfc77d6710e7c83c7834989c1f38` |
| Merge fact | `sha256:871f7bd12e69f0282c2415512310e07ea4efc038b5c7c333b553e312b074b676` | `c7f59ff9b35260a3fdd524f0936ea0d630a65190` |

The run used pipeline definition
`sha256:73874909dd6e6365f6417e56b2bc50292fbebe5446b35a6ab37c550ee875d494`.
The merge advanced `main` from the initial commit to
`08291d6f363fa68ca66b84d922d46cc8302fc9a0`. The merge fact binds that
result to the proposed head and the old base-policy digest; the newly accepted
actor's review did not qualify under that base policy.

Stage 1 published and then separately verified:

| Destination ref | Advertised object ID |
| --- | --- |
| `refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd` | `c7f59ff9b35260a3fdd524f0936ea0d630a65190` |
| `refs/nh/actors/10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db` | `55842fb308c3f78ad028f0d3be843beb83756209` |
| `refs/nh/proposals/52aee3a8a6fd110321b3506e9408397158d02d3dfe63747cb445effa157e0a95` | `a3ef855ca062c191a6c7e5c14709de67bdfb4f1c` |
| `refs/heads/main` | `08291d6f363fa68ca66b84d922d46cc8302fc9a0` |

### Stage 2: role-distinct release candidate

The release candidate used the stage-1 merge as its base and the
metadata-free product/test/documentation commit
`07b64dfdb11dc376644e068f201ca998bb54defe` as its proposed head. Its exact
base policy was the amended digest above.

| Fact | Event ID | Event commit |
| --- | --- | --- |
| Release candidate | `sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d` | `96e23598774d2d6df8ccd8bf5ce8979bfe7e0928` |
| Author review, valid but `0/1` under the amended policy | `sha256:574edfdf45ce760b157e80111fd349db6fc252f9a2531f1461a3dcb415c54eca` | `90fd7a3563355c239c7fbc6fd5c0b5c0e26ca41c` |
| `test` request | `sha256:471862506580482a13fc39dc6fb1e59a85a4d3a9ca9e6b2beec3e42b355b7963` | `5d4bc4f14c6b711d06893217f296151a7be34b2b` |
| Distinct actor sandbox pass | `sha256:f6ea394efdf4010c6e627f5434965bf9173f0b51336747b0c5d431a1bedc6f89` | `7a331cb2f52d995bbc6a1a1a48f44401fbd99182` |
| Distinct actor qualifying review | `sha256:8737d3bd963a6a6d823cdf7a472485a6f94fc43197d4733f259cfa169fd522ce` | `8bc318bdd0a8a37e3fc55d692651811124516c14` |
| Amended-policy decision | `sha256:cb8ba0187d63de9716497a042236efbdaad6b2ab83fe666392d6a17d6fe72a38` | `1f5d287bc6313c90efd2a1d9a0ab9ab7b94d1548` |
| Merge fact | `sha256:e6ba7f9f028f72028c7b8ac286d74c5821e647e1a2fa332a5423d3525d91df77` | `4d9caba6548c50eba49ed89dc8ee45960226c01e` |

The distinct actor's result used the same exact pipeline definition, the
default `sandbox` backend, commit `07b64dfdb11dc376644e068f201ca998bb54defe`,
outcome `passed`, and duration 52,517 ms. After its review and result arrived,
status changed from `0/1` review and `0/1` pipeline to ready. The final merge
commit was `7e41574792be828ba507e5df7adda71662475483`.

Stage 2 advanced the actor histories and added the release ref first. Those
advertised OIDs were checked before `main` was advanced separately:

| Destination ref | Final advertised object ID |
| --- | --- |
| `refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd` | `4d9caba6548c50eba49ed89dc8ee45960226c01e` |
| `refs/nh/actors/10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db` | `8bc318bdd0a8a37e3fc55d692651811124516c14` |
| `refs/nh/proposals/122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d` | `07b64dfdb11dc376644e068f201ca998bb54defe` |
| `refs/heads/main` and `HEAD` | `7e41574792be828ba507e5df7adda71662475483` |

### Fresh public depth-limited reconstruction

An identity-free depth-1 HTTPS clone selected the two actors and all three
candidate histories needed by the maintainer chain. The pre-existing candidate
is required because it is referenced by the maintainer history; omitting it
correctly reports a missing dependency.

```sh
git clone --depth 1 https://github.com/LynnColeArt/hubnot.git verify
cd verify
nh replication select origin \
  --actor 36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd \
  --actor 10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db \
  --proposal sha256:4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290 \
  --proposal sha256:52aee3a8a6fd110321b3506e9408397158d02d3dfe63747cb445effa157e0a95 \
  --proposal sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin
nh proposal status sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d
nh identity list
nh log
git rev-parse --is-shallow-repository
git for-each-ref --format='%(refname) %(objectname)' refs/nh/remotes/origin
```

The selected sync promoted 24 verified events and five selected histories.
The release reconstructed as merged with one qualifying distinct review, one
trusted sandbox pass, and one acceptance decision. Both public actors and the
accepted device relationship reconstructed. `git rev-parse
--is-shallow-repository` remained `true`. The clone contained neither a
private identity keyring nor a legacy identity record, and it contained no
local `refs/nh/actors/*`; accepted public facts were under
`refs/nh/remotes/origin/*`.

Every selected dependency was present, so public recovery was unnecessary.
The separate offline fixture above exercises the exact
`nh sync origin --recover-shallow` path and confirms that recovery preserves
the shallow boundary.

## Verification method

The public verification uses only ordinary Git and the release-candidate `nh`
binary:

```sh
git ls-remote --symref origin HEAD
git ls-remote origin 'refs/heads/main' 'refs/nh/actors/*' 'refs/nh/proposals/*'

git clone --depth 1 https://github.com/LynnColeArt/hubnot.git verify
cd verify
nh replication select origin \
  --actor 36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd \
  --actor 10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db \
  --proposal sha256:4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290 \
  --proposal sha256:52aee3a8a6fd110321b3506e9408397158d02d3dfe63747cb445effa157e0a95 \
  --proposal sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin
nh proposal status sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d
nh proposal show sha256:122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d
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
and explicitly calling it exits zero after an idempotent accepted-projection
verification without writing to standard output or standard error.

## Limits of the proof

Two distinct signing keys prove distinct actor histories, not distinct humans.
The public host is evidence of ordinary Git ref/object transport, not of its UI
or collaboration API. Promotion budgets do not prevent standard Git from
downloading a selected pack before measurement. Planned rotation requires the
old key. Public immutable facts cannot carry a global erasure guarantee.
