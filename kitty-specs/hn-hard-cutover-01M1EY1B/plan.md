# Implementation Plan: `hn` Hard Cutover

**Branch**: `feat/hn-hard-cutover` | **Date**: 2026-09-01 | **Spec**: [spec.md](spec.md)
**Input**: Breaking rename of Hubnot's complete active short namespace from
`nh` to `hn`, with no compatibility behavior.

## Summary

Cut the executable, Git ref root, private state root, repository configuration,
wire identifiers, environment prefix, runner labels, tests, current docs, and
active charter over to `hn` as one atomic release. Preserve legacy signed facts
and audit artifacts as inert history. Use a pre-cutover binary to govern the
exact final commit, then bootstrap new Ed25519 actors and publish independent
`refs/hn/*` facts.

## Technical Context

**Language/Version**: Go 1.26  
**Primary Dependencies**: Go standard library, Git command-line client;
Bubblewrap remains an optional hardened Linux CI backend  
**Storage**: Git objects/refs, tracked `.hn/*.json`, private `.git/hn/*.json`  
**Testing**: `go test`, race detector, real temporary repositories/remotes,
live two-party public smoke  
**Target Platform**: Portable Git hosts; Linux for sandboxed CI  
**Project Type**: Single Go CLI  
**Performance Goals**: Preserve current bounded reads, 1 MiB logs, configured
replication budgets, and ordinary CLI response behavior  
**Constraints**: No third-party Go dependency, server, Docker requirement,
compatibility shim, history rewrite, or force-push  
**Scale/Scope**: Roughly 50 Go source/test files, 9 active docs/config surfaces,
four wire identifiers, one ref hierarchy, and one private-state hierarchy

## Charter Check

- **Repository-native state**: Pass. The change keeps ordinary Git objects,
  refs, and transports and changes only the namespace root.
- **Evidence integrity**: Pass. Exact historical bytes are excluded from edits;
  the final old governance event names the exact cutover commit.
- **Hostile input / least authority**: Pass. Validation and sandbox boundaries
  remain behaviorally unchanged and receive full race/acceptance coverage.
- **Immutable recovery**: Pass. No history or event is rewritten and no actor
  continuity is fabricated across the namespace boundary.
- **Protocol/docs lockstep**: Pass. Current protocol documents, tests, and active
  charter are required outputs of the same mission.
- **Quality gates**: Pass by plan. Build, vet, formatting, race, acceptance,
  namespace scans, fresh-state isolation, and public Git smoke are mandatory.

The active charter itself is a governed output: `nh` becomes `hn` and project
directive IDs `NH-001..005` become `HN-001..005`. Completed mission references
remain historical exceptions.

## Project Structure

### Documentation (this mission)

```text
kitty-specs/hn-hard-cutover-01M1EY1B/
├── spec.md
├── research.md
├── data-model.md
├── occurrence_map.yaml
├── plan.md
├── quickstart.md
├── contracts/namespace-cutover-v0.md
├── tasks.md
└── tasks/WP*.md
```

### Source and active configuration

```text
.
├── *.go                         # CLI, protocols, storage, governance, CI, memory
├── *_test.go                    # unit, integration, security, acceptance
├── .hn/
│   ├── policy.json              # new actor fingerprints; hn.policy/0
│   └── pipelines/test.json      # hn.pipeline/0
├── .nh/                         # frozen transition evidence; ignored by hn
├── README.md
├── docs/                        # current contracts plus marked history
└── .kittify/charter/            # active HN directives and human companion
```

**Structure Decision**: Keep the single-package Go design. Namespace constants
remain close to the security-critical modules they govern; this mission avoids
an unrelated abstraction refactor. Add one focused isolation test file if
needed to make the non-compatibility contract explicit.

## Namespace Change Matrix

| Concern | Old active value | New active value | Treatment |
|---|---|---|---|
| CLI / binary | `nh` | `hn` | hard replace |
| collaboration wire | `nh/0` | `hn/0` | reject old |
| memory wire | `nh-memory/0` | `hn-memory/0` | reject old |
| policy / pipeline | `nh.policy/0`, `nh.pipeline/0` | `hn.policy/0`, `hn.pipeline/0` | reject old |
| Git refs | `refs/nh/*` | `refs/hn/*` | new root; old ignored |
| private state | `.git/nh/*` | `.git/hn/*` | new root; no migration |
| tracked config | `.nh/*` | `.hn/*` | new active copy; old frozen |
| environment | `NH_*` | `HN_*` | hard replace |
| runner / temp labels | `nh/*`, `nh-*` | `hn/*`, `hn-*` | hard replace |
| charter IDs | `NH-001..005` | `HN-001..005` | active charter only |

## Implementation Concern Map

### IC-01 — Runtime namespace constants and paths

- **Purpose**: Change every reader/writer boundary used by collaboration,
  governance, CI, identity, memory, replication, shallow recovery, and keyring.
- **Relevant requirements**: FR-001 through FR-006, FR-013.
- **Affected surfaces**: `*.go`, generated binary ignore rules.
- **Sequencing/depends-on**: none.
- **Risks**: Overlooking a recovery diagnostic or one exact-ref refspec; blind
  substring replacement could affect unrelated identifiers.

### IC-02 — Executable tests and legacy isolation

- **Purpose**: Update existing expectations and prove old refs, files, private
  state, and environment values have no effect.
- **Relevant requirements**: FR-007, FR-009, FR-013; NFR-001 through NFR-005.
- **Affected surfaces**: `*_test.go`, temporary Git repositories and remotes.
- **Sequencing/depends-on**: IC-01.
- **Risks**: Tests may pass after self-consistent renaming without proving the
  absence of fallback behavior; explicit adversarial fixtures close that gap.

### IC-03 — Active policy, pipeline, docs, and charter

- **Purpose**: Make every current contract executable as written while clearly
  separating historical records.
- **Relevant requirements**: FR-004, FR-008 through FR-010, FR-012.
- **Affected surfaces**: `.hn/**`, README, current `docs/*.md`,
  `.kittify/charter/**`, occurrence audit.
- **Sequencing/depends-on**: IC-01 for exact behavior; fresh fingerprints from
  IC-04 before policy finalization.
- **Risks**: Dated evidence blocks in host-compatibility docs must not be made to
  claim the new program produced old refs.

### IC-04 — Fresh identity and trust bootstrap

- **Purpose**: Create independent maintainer/reviewer keys, populate active
  policy, and verify two-party `hn` governance and replication.
- **Relevant requirements**: FR-009 through FR-011; NFR-006.
- **Affected surfaces**: untracked `.git/hn/**`, reviewer clone private state,
  tracked `.hn/policy.json`, public `refs/hn/*`.
- **Sequencing/depends-on**: IC-01 and a buildable candidate.
- **Risks**: Private keys must never enter diffs; fingerprints must be finalized
  before the exact candidate is governed.

### IC-05 — Final old-governance transition and public cutover

- **Purpose**: Attest the exact candidate using the last `nh` proposal loop,
  publish it, then begin the independent `hn` history.
- **Relevant requirements**: FR-009 through FR-011; SC-003 through SC-005.
- **Affected surfaces**: old local/public `refs/nh/*`, new `refs/hn/*`, `main`,
  remote smoke clones, mission verification report.
- **Sequencing/depends-on**: IC-01 through IC-04 and final QA.
- **Risks**: Any post-attestation source change invalidates evidence and requires
  a new proposal revision; ref publication must remain append-only.

## Implementation Sequence

1. Record the occurrence baseline and `.nh` file digests; implement runtime and
   test namespace changes plus explicit legacy-isolation tests.
2. Build `hn`, generate fresh private identities, and add `.hn` policy/pipeline
   using only their public fingerprints.
3. Update current docs and active charter through their governed workflows;
   audit every residual old token against `occurrence_map.yaml`.
4. Run build, vet, formatting, race, acceptance, keyring, sandbox, selective
   replication, shallow recovery, and fresh-repository smoke gates.
5. Freeze the candidate. Use a separately built pre-cutover `nh` executable and
   the old actors to run the final public proposal/review/CI/decision/merge.
6. Publish `main`, establish and synchronize new `hn` facts, and verify a fresh
   clone can validate them while ignoring old facts.

## Verification Strategy

- Unit and integration suite: `go test -count=1 ./...`.
- Security/concurrency suite: `go test -race -count=1 ./...`.
- Static gates: `go build ./...`, `go vet ./...`, `gofmt -l .`,
  `git diff --check`.
- Active namespace scan: zero old spellings outside occurrence-map exceptions.
- Frozen evidence check: `.nh/policy.json` and `.nh/pipelines/test.json` retain
  baseline SHA-256 digests; old refs retain append-only ancestry.
- Isolation fixture: seed only legacy state, run new commands, and assert no old
  fact discovery or file mutation.
- Public smoke: fresh clones, new identities, issue/proposal/CI/review/decision
  path as practical, explicit `git ls-remote` inspection.
- Dependency check: `go.mod` remains `module hubnot` with no `require` entries.

## Complexity Tracking

No charter violation or new architectural layer is introduced. The additional
complexity is operational: two namespace eras coexist in Git history, but only
one is executable at a time.
