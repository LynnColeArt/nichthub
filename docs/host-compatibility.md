# Hosted Git compatibility

Nichthub needs an ordinary Git remote to accept, advertise, and transfer
objects reachable through `refs/nh/*`. This document records direct
observations only; it does not infer support for providers or configurations
that were not tested. Hosting UI/API state is not evidence.

| Remote | Result | Observed |
| --- | --- | --- |
| Local bare Git repository | Actor/candidate refs, exact selected fetch, quarantine, budgets, and depth-limited reconstruction passed | 2026-08-30 |
| GitHub.com private repository over HTTPS | Actor ref push, advertisement, and two-actor reconstruction passed | 2026-08-29 |
| GitHub.com public repository over HTTPS | Existing actor and candidate refs advertised before the operational-alpha landing | 2026-08-30 |
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

## GitHub.com public baseline

Before the operational-alpha public mutation on 2026-08-30, ordinary Git
advertisement for `https://github.com/LynnColeArt/nichthub.git` returned:

```text
refs/heads/main e8b0955a3d10b5677fb0dfaba42580f8f4473080
refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd 2b0e261037e0760f83536fedfad044e95deff1fe
refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290 327199e686dbe0bcc4cc4bf92f10237472b083a8
```

These values were obtained with `git ls-remote`, not a provider API. The
staged operational-alpha result, including exact new refs and depth-limited
selected reconstruction, is recorded after it completes in this document and
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
```

## Limits not established by these observations

- GitHub Enterprise Server or organization-specific ref policies;
- very large actor/candidate counts, histories, or packs;
- provider garbage-collection and long-term custom-ref retention;
- deletion, rollback, force-push, and global redaction behavior;
- partial-clone filters;
- portable hard pre-download quotas;
- providers other than the explicitly observed local Git and GitHub.com
  configurations.

A successful transport observation does not make the host a policy authority,
prove separate humans, or strengthen the local runner sandbox.
