# Hosted Git compatibility

Nichthub needs an ordinary Git remote to accept, advertise, and transfer
objects reachable through `refs/nh/*`. This document records direct
observations only; it does not infer support for providers or configurations
that were not tested. Hosting UI/API state is not evidence.

| Remote | Result | Observed |
| --- | --- | --- |
| Local bare Git repository | Actor/candidate/memory refs, exact selected fetch, quarantine, budgets, and fresh-clone reconstruction passed | 2026-08-31 |
| GitHub.com private repository over HTTPS | Actor ref push, advertisement, and two-actor reconstruction passed | 2026-08-29 |
| GitHub.com public repository over HTTPS | Two-stage actor/candidate publication, separate primary-branch advancement, exact selected fetch, quarantine, and identity-free depth-1 reconstruction passed | 2026-08-30 |
| GitLab | Not tested | — |
| Other hosted providers | Not tested | — |

## Local bare Git

Automated black-box acceptance creates a fresh local bare remote and uses only
ordinary Git transport. The three-run operational scenario and focused hostile
fixtures directly establish:

- separate actor and candidate refs are advertised and exact-fetched;
- actor and candidate objects cross the remote without a Nichthub service;
- explicit selections exclude unselected refs;
- valid selections can promote while invalid, mismatched, over-budget, and
  dependency-missing selections remain absent from accepted refs;
- an identity-free depth-1 clone reconstructs exact policy, identity,
  candidate, CI, review, decision, merge, and Git commit bindings;
- selected recovery can obtain an exact missing predecessor while the
  repository remains shallow and without importing an unselected actor.

On 2026-08-31, `TestOperationalAgentMemory` separately created two actors,
recorded all six memory kinds, superseded and retracted exact records, added a
cross-actor challenge, and published the two exact memory stream refs through
`nh sync`. A credential-disabled fresh clone initially had no local or accepted
memory refs, private identity, or index. After saving only those two full stream
selectors, it synchronized through quarantine, rebuilt
`.git/nh/memory/index-v0.json`, and recalled the same full IDs, ownership,
anchors, lifecycle edges, evidence, trust classes, content digests, and inert
handoff lists. Deleting and rebuilding the index produced byte-identical index
and recall JSON. Three consecutive offline runs passed.

That local-bare observation establishes ordinary Git transport and protocol
behavior, not public-host policy or retention. Hostile memory containing shell-
and tool-shaped prose, ANSI controls, newlines, and Unicode produced no marker
effect and remained nested under the recall data boundary.

The detailed run evidence and bounds are in
[self-hosting-alpha.md](self-hosting-alpha.md).

## GitHub.com private observation

The 2026-08-29 test used a temporary private repository under an authenticated
personal account. GitHub accepted actor refs outside the branch and tag
namespaces:

```text
refs/nh/actors/2f2534ef...
refs/nh/actors/e675ea17...
```

Ordinary `git push`, `git ls-remote`, and explicit Git fetch established:

1. a new actor ref was accepted and advertised unchanged;
2. a fresh clone fetched actor history only after Nichthub supplied an
   explicit custom refspec;
3. a second actor published a separate signed comment history;
4. the first clone fetched it and reconstructed both verified events;
5. every event commit and parent needed by those histories transferred.

That observation preceded candidate refs and selected quarantine, so it does
not claim those newer surfaces for the private-repository configuration.

## GitHub.com public operational-alpha observation

Before the operational-alpha public mutation on 2026-08-30, ordinary Git
advertisement for `https://github.com/LynnColeArt/nichthub.git` returned:

```text
refs/heads/main e8b0955a3d10b5677fb0dfaba42580f8f4473080
refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd 2b0e261037e0760f83536fedfad044e95deff1fe
refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290 327199e686dbe0bcc4cc4bf92f10237472b083a8
```

These values were obtained with `git ls-remote`, not a provider API. Stage 1
published the two actor histories and policy candidate before separately
advancing `main` to
`08291d6f363fa68ca66b84d922d46cc8302fc9a0`. Stage 2 advanced the two actor
histories and added the release candidate before separately advancing `main`.
The final ordinary-Git advertisement was:

```text
refs/heads/main 7e41574792be828ba507e5df7adda71662475483
refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd 4d9caba6548c50eba49ed89dc8ee45960226c01e
refs/nh/actors/10425b859ef6c86afe46e9586a1c26a9fe007f7a0aa38a46a3c316fcfc69a1db 8bc318bdd0a8a37e3fc55d692651811124516c14
refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290 327199e686dbe0bcc4cc4bf92f10237472b083a8
refs/nh/proposals/52aee3a8a6fd110321b3506e9408397158d02d3dfe63747cb445effa157e0a95 a3ef855ca062c191a6c7e5c14709de67bdfb4f1c
refs/nh/proposals/122a6dc9ba587f787897395e89c38bd48aef1f68dc5d7dc35121562a1073371d 07b64dfdb11dc376644e068f201ca998bb54defe
```

A fresh HTTPS clone at depth 1 saved exact selections for the two actors and
three proposals with budgets 10,000 events, 100,000 objects, 16,777,216 bytes
per object, 1,048,576 bytes per attachment, and 268,435,456 total bytes. It
promoted 24 verified events, reconstructed the release as merged with its
exact amended policy and role-distinct evidence, and remained shallow. It had
no private keyring, legacy identity record, or local public actor ref.

The complete event, policy, object, and commit record is in
[the public proof record](self-hosting-alpha.md).

## Clone behavior

An ordinary `git clone` and ordinary branch fetch do not import custom refs,
because Git configures a branch refspec by default. This is expected. After an
explicit selection, `nh sync` requests exact selected actor/candidate refs and
stores validated roots below `refs/nh/remotes/<remote>/*`.

The primary branch and collaboration refs are separate namespaces and separate
publication outcomes. A host accepting one does not imply that the other was
updated; verify both with ordinary Git:

```sh
git ls-remote --symref origin HEAD
git ls-remote origin 'refs/heads/main' 'refs/nh/actors/*' 'refs/nh/proposals/*'
git ls-remote origin 'refs/nh/memory/*/*'
```

## Limits not established by these observations

- GitHub Enterprise Server or organization-specific ref policies;
- very large actor/candidate counts, histories, or packs;
- very large memory-stream counts, histories, or packs;
- provider garbage-collection and long-term custom-ref retention;
- deletion, rollback, force-push, and global redaction behavior;
- partial-clone filters;
- portable hard pre-download quotas;
- providers other than the explicitly observed local Git and GitHub.com
  configurations.

A successful transport observation does not make the host a policy authority,
prove separate humans, or strengthen the local runner sandbox.
