# `hn` Namespace Cutover Contract

## Active contract

The first Hubnot release after this mission exposes exactly these roots:

```text
command:            hn
collaboration wire: hn/0
memory wire:        hn-memory/0
policy wire:        hn.policy/0
pipeline wire:      hn.pipeline/0
Git refs:           refs/hn/*
private state:      <git-dir>/hn/*
repository config:  .hn/*
environment:        HN_*
runner label:       hn/<version>
```

These values form one version boundary. Implementations MUST NOT accept a mixed
combination.

## Non-compatibility rules

The `hn` program MUST NOT:

- dispatch under the executable name `nh` through a shipped alias or wrapper;
- enumerate, fetch, push, project, or validate `refs/nh/*`;
- read, migrate, copy, or delete `.git/nh/*`;
- use `.nh/policy.json`, `.nh/pipelines/*`, or `.nh/actions/*`;
- accept `nh/0`, `nh-memory/0`, `nh.policy/0`, or `nh.pipeline/0` as current;
- read `NH_*` as an alias for `HN_*`;
- emit both old and new refs, fields, variables, or paths.

Unsupported old wire values fail through the same strict-version validation used
for every unknown future version.

## Historical evidence

Existing `refs/nh/*`, completed mission documents, append-only journals,
historical compatibility/self-hosting reports, and the checked-in `.nh/`
transition inputs may remain in a repository. Their presence MUST NOT affect
`hn` output or trust projections. Preservation is provenance, not support.

## Isolation examples

1. With only `refs/nh/actors/<actor>` present, `hn log` returns no event from
   that ref.
2. With only `.git/nh/identity.json` present, `hn identity show` reports that
   `hn init` is required and does not create `.git/hn/` as a read side effect.
3. With both `refs/nh/*` and `refs/hn/*` advertised, `hn sync` constructs
   refspecs exclusively for `refs/hn/*`.
4. With `NH_INTERNAL_TESTING=1` and no corresponding `HN_*` variable, production
   behavior is unchanged.

## Transition attestation

The last old-protocol merge event may name the exact first `hn` commit. This is
an external historical relationship; the new runtime neither requires nor
parses the old event. Any source change after that attestation requires a new
legacy proposal revision before publication.
