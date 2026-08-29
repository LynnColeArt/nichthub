# Hosted Git compatibility

Nichthub needs Git remotes to accept, advertise, and transfer objects reachable
through `refs/nh/*`. This document records observed behavior from live tests.
It does not infer support for providers that have not been tested.

| Remote | Result | Tested |
| --- | --- | --- |
| Local bare Git repository | Passed | 2026-08-29 |
| GitHub.com private repository over HTTPS | Passed | 2026-08-29 |
| GitLab | Not tested | — |
| Other hosted providers | Not tested | — |

## GitHub.com

The test used a temporary private repository under an authenticated personal
account. GitHub accepted actor refs outside the branch and tag namespaces:

```text
refs/nh/actors/2f2534ef...
refs/nh/actors/e675ea17...
```

This live test preceded the addition of `refs/nh/proposals/*`; that specific
namespace has only been exercised with a local bare Git remote so far.

The following behaviors were directly observed:

1. Pushing a new `refs/nh/actors/<actor>` ref succeeded.
2. `git ls-remote origin 'refs/nh/actors/*'` advertised the ref unchanged.
3. A fresh clone fetched the actor histories with Nichthub's explicit wildcard
   refspec.
4. A second actor published a separate history containing a comment.
5. The first clone fetched that history and reconstructed the signed issue and
   comment.
6. All referenced event commits and their parent histories were transferred.

An ordinary `git clone` and an ordinary `git fetch origin` did **not** fetch the
custom refs. This is expected because a standard clone configures a branch
refspec, not a wildcard for `refs/nh/*`. After `nh sync`, the same fresh clone
contained both remote-tracking actor refs and projected both verified events.

Therefore GitHub.com can act as an unaware Nichthub transport with the current
layout, but users still need `nh sync` (or an equivalent configured refspec) to
receive collaboration data.

## Not yet covered

The test did not establish behavior for:

- public repositories or organization policy variations;
- GitHub Enterprise Server;
- very large actor counts or histories;
- provider garbage-collection and long-term retention behavior;
- deletion, force-push, and rollback handling;
- per-actor authorization for collaborators;
- partial clones or shallow clones.

These are distinct from the basic transport result and need separate tests.
