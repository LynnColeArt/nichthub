# Research: Operational Self-Hosting Alpha

## Research question

How should Nichthub extend its tested repository-native event model so the
project can safely evolve policy and device keys, replicate only bounded
selected collaboration state, operate from shallow clones, and prove those
capabilities through its own public development workflow?

## Baseline findings

### Existing integrity boundaries are reusable

Every actor is currently the SHA-256 fingerprint of one Ed25519 public key.
Each event repeats that key, is signed over exact payload bytes, and belongs to
one monotonically sequenced actor chain. Git ref compare-and-swap prevents two
local commands from silently appending to the same observed head, while global
projection rejects forks, gaps, broken predecessors, invalid relationships,
and unsupported event kinds. [S002] [S003]

Policy is already immutable in the correct place: every candidate is evaluated
against exact `.nh/policy.json` bytes from its signed base commit. A candidate
may change policy for later candidates but cannot use its head policy to
authorize itself. Policy validation already rejects empty maintainer sets,
duplicate or malformed actors, and unsatisfiable thresholds. The missing layer
is operator tooling to inspect and compare these facts before publication.
[S004]

Synchronization is the weakest public boundary. It currently fetches wildcard
actor and proposal refs directly into the accepted remote-tracking namespace,
pushes all local proposal refs, and then validates the union globally. One bad
fetched history can therefore make every command that collects events fail,
and participants cannot persist a per-remote actor/candidate selection. [S005]

### The public transport prerequisite has passed

On 2026-08-30, the public `LynnColeArt/nichthub` remote advertised the active
actor ref and an existing proposal ref unchanged after `nh sync`:

```text
refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd
refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290
```

A fresh clone with no private identity synchronized seven events and displayed
the new operational-alpha issue. Its full ID is:

```text
sha256:7204f1b6c2eb16b1d25f1f6269f0cce1eed354636c432baf981e44e1cf83a79a
```

This extends the previous compatibility record from private actor refs to a
public repository carrying both actor and proposal namespaces. [S006] [S007]

### Existing tests prove pieces, not the new operating model

The test suite already proves distributed pipeline execution, explicit unsafe
host opt-in, optional Bubblewrap isolation, proposal code-ref transfer,
base-policy governance, merge guards, and immutable conflict revision. It does
not cover policy-change preview, actor continuity, selected/quarantined sync,
shallow history recovery, or one role-distinct black-box CLI workflow. [S008]

## Decisions

### D-001 — Policy amendments remain ordinary candidates

**Decision**: Do not introduce a `policy.amend` event. Add operator commands
that load, validate, compare, and explain policies from exact files or commits;
the amendment itself remains a normal Git change proposed through
`proposal.open` or `proposal.revise`.

**Rationale**: The signed base/head candidate already binds the old and new
policy bytes. The old policy's digest already governs review, CI, decisions,
and merge. A second event type would duplicate authority and create ambiguity
about which artifact actually changes policy. [S004] [S009]

**Rejected alternatives**:

- Let the head policy govern a policy-changing candidate: self-authorization.
- Add an out-of-band administrator command: violates repository-native
  governance.
- Automatically rewrite policy: obscures a security-sensitive Git diff.

### D-002 — Identity continuity uses two-sided signed actor claims

**Decision**: Add an authorization event signed by an existing actor naming a
target actor, target public key, and relationship (`device` or `successor`). Add
an acceptance event signed by the target actor referencing that exact
authorization. A relationship becomes complete only when both events verify.

**Rationale**: The old signature proves intent by the established actor; the
new signature proves control of the target key. Including the target public key
allows independent recomputation of its actor fingerprint. Separate actor
chains preserve single-writer append and attribution. [S002] [S003] [S009]

**Projection rules**:

- A `device` relationship leaves both actors active.
- An accepted `successor` relationship marks the predecessor retired in the
  identity projection but never deletes its history.
- Incomplete, cyclic, replayed, or competing accepted successors remain
  visible as conflicts; clients do not select a global winner.
- A continuity edge carries no maintainer, reviewer, runner, or decision
  authority. Project policy must name the target actor independently.

**Rejected alternatives**:

- Copy one private key to every device: creates actor-chain forks and destroys
  device attribution.
- Treat actor display names as identity: unsigned and non-unique.
- Let rotation implicitly rewrite policy: bypasses project governance.
- Claim lost-key recovery: impossible to prove from predecessor signatures
  when the predecessor key is unavailable.

### D-003 — Local identity storage becomes an explicit keyring

**Decision**: Migrate the single local identity file to a private keyring with
one active actor pointer and separately protected actor records. Planned
rotation creates or accepts a new actor, completes both signed claims, then
switches the active pointer; the predecessor key is retained only for explicit
historical or recovery inspection and is not the default signer.

**Rationale**: Rotation crosses two actor refs and local secret state, so one
flat identity file cannot represent a recoverable intermediate state. An
explicit active pointer makes crash recovery and accidental retired-key use
observable. Existing repositories must migrate without changing actor IDs or
event payloads. [S002] [S003]

**Safety boundary**: No command exports or transmits private keys. A separately
initialized device shares only its public actor/key material before mutual
authorization.

### D-004 — Replication selection and policy trust remain separate

**Decision**: Persist local, untracked per-remote selections of full actor and
candidate IDs plus positive resource budgets. Selection decides which remote
facts may enter the local projection. Existing project policy independently
decides which verified claims qualify for governance.

**Rationale**: Conflating fetch selection with trust would make it impossible
to inspect an untrusted contribution without granting it authority, or to
retain a valid non-authoritative fact needed as evidence context. [S005] [S009]

**Compatibility**: Explicit selection is the safe public workflow. Legacy
wildcard synchronization remains available only as an explicit all-ref mode
subject to the same quarantine and budgets; it is not used implicitly once a
remote selection exists.

### D-005 — Fetch selected refs into an isolated quarantine repository

**Decision**: Resolve explicit remote ref names, fetch them into a generated
bare quarantine repository, measure and validate the selected object graph,
then promote accepted refs into the main repository through an atomic ref
transaction. Failed imports are discarded without accepted refs pointing at
their objects.

**Rationale**: Fetching directly into `refs/nh/remotes/*` makes invalid data
visible before verification. A separate Git object database provides a real
validation boundary while reusing Git's transport and object verification.
Atomic promotion prevents a partially updated selected set. [S003] [S005]

**Failure isolation**: Actor-chain structural validation occurs per actor.
Cross-actor relationship validation uses the union of already accepted facts
and the selected quarantine candidates. An invalid actor or candidate remains
quarantined; independently valid selections can still promote. Missing
cross-selection dependencies are reported by full ID rather than treated as
invalid signatures.

### D-006 — Budgets govern promotion; transport exhaustion remains explicit

**Decision**: Enforce configured event-count, object-count, attachment-size,
individual-object-size, and total reachable-byte limits in quarantine before
promotion. Use direct selected refspecs so unselected object graphs are not
requested.

**Rationale**: Git can advertise many refs without transferring their objects,
and direct refspecs limit requested reachability. Once a selected pack arrives,
quarantine measurement can deterministically prevent expensive projection and
retained invalid state. [S005] [S010]

**Residual risk**: Standard `git fetch` does not expose a portable hard
pre-download pack-size ceiling through the current wrapper. A malicious
selected remote may consume network and temporary disk before post-fetch
budget rejection. The mission must document this honestly; OS-level transport
quotas or a lower-level pack protocol are separate hardening work. This does
not weaken the before-promotion requirement.

### D-007 — Shallow recovery follows selected dependencies, never global trust

**Decision**: Detect shallow repositories before integrity-sensitive
operations. First verify whether the exact selected actor, proposal, base,
policy, pipeline, and ancestry objects are already available. If not, report
the missing boundary and offer an explicit bounded fetch of the selected refs
through the same quarantine path; never silently unshallow the repository.

**Rationale**: A shallow marker is not itself failure—custom refs may already
carry every required object. Integrity depends on exact object availability,
not whether the branch clone was deep. A global `--unshallow` fetch would
discard the benefits of selection and may retrieve unrelated history. [S003]
[S005] [S010]

### D-008 — The live proof has two governed stages

**Decision**: The public landing proof occurs in stages:

1. Land the operational capabilities and a policy change under the current
   base policy.
2. Use the amended policy for a subsequent real candidate that cannot become
   ready without qualifying evidence from a second actor.
3. Verify both stages from a fresh depth-limited clone using explicit
   selection.

**Rationale**: A new trusted actor cannot authorize the proposal that first
adds it. The staged sequence is the direct observable proof that policy is
base-bound and not self-amending. [S004] [S009]

### D-009 — Automated acceptance stays offline; public proof is recorded

**Decision**: Exercise the full behavior through CLI subprocesses against a
disposable ordinary Git remote. Exercise the public remote deliberately as the
live landing gate and record exact IDs.

**Rationale**: Automated tests must not require credentials or mutate public
state. The synthetic repository action used by the test must not run this
repository's own complete test pipeline, avoiding recursive test invocation.
[S008] [S009]

## Component boundaries

### Identity keyring and continuity projection

- Owns private local key records, active-actor selection, migration, planned
  rotation state, public handoff material, and deterministic relationship
  projection.
- Does not decide project governance roles.
- Provides a small signer/inspection interface to existing event commands.

### Policy inspection

- Owns exact policy loading, validation, digest display, deterministic policy
  comparison, and amendment warnings.
- Reuses the same validator as proposal evaluation; there is one policy truth.
- Does not create a parallel policy event or mutate Git automatically.

### Replication transaction

- Owns per-remote selection, quarantine lifecycle, budget accounting,
  dependency classification, validation, and atomic promotion.
- Does not interpret policy authority.
- Supplies accepted refs to the existing event projection.

### Shallow dependency resolver

- Classifies missing Git objects versus invalid facts.
- Maps missing objects back to selected actor/candidate refs and explicit
  recovery actions.
- Uses replication quarantine rather than a second fetch implementation.

### Existing domain projection and governance

- Extends event relationship validation for identity claims.
- Continues to own proposal/revision/evidence/policy/merge semantics.
- Never treats an identity-continuity relation as policy authorization.

## Test strategy

- Test-first unit coverage for policy diffs, identity relationship validation,
  keyring migration/active selection, budget accounting, and shallow-gap
  classification.
- Integration coverage with multiple real Git repositories for quarantine,
  atomic promotion, invalid-selection isolation, and shallow recovery.
- Black-box subprocess acceptance covering policy amendment, two actor chains,
  selected synchronization, role-distinct evidence, merge, and fresh-clone
  reconstruction.
- Golden compatibility fixtures proving the seven existing public event IDs
  remain byte-for-byte readable.
- Live public verification recorded by full actor, event, policy, ref, and Git
  object IDs; no screenshots or provider UI claims.

## Risks and mitigations

| Risk | Consequence | Mitigation |
| --- | --- | --- |
| Rotation is mistaken for authority transfer | New key bypasses policy | Keep identity projection and policy evaluation separate; test valid-but-nonqualifying successor claims. |
| Crash occurs between old and new actor events | Incomplete rotation | Persist explicit local rotation state, keep old signer active until mutual acceptance, and make retry idempotent. |
| Competing successors are observed | Clients choose different active keys | Preserve all claims, derive conflict, and refuse an unambiguous successor projection. |
| Invalid remote data enters accepted refs | Global projection denial | Fetch into separate quarantine, validate first, and atomically promote only valid selected refs. |
| Selected pack is huge before measurement | Temporary disk/network exhaustion | Request direct refs, use a generated quarantine location, enforce post-fetch promotion budgets, cleanly report residual transport limit. |
| Missing dependency is classified as corruption | Valid partial collaboration becomes unusable | Return a distinct missing-dependency result with full ID and explicit additional selection. |
| Shallow recovery fetches unrelated history | Selection and resource promises break | Recover only selected refs under the same budgets; never silently unshallow. |
| Existing identity migration changes signing facts | Published actor history forks | Preserve exact key bytes and actor fingerprint; migration tests compare pre/post signatures and actor refs. |
| Acceptance test invokes itself recursively | Unbounded test process tree | Use a tiny synthetic action in the scenario repository rather than `go test ./...`. |
| Same human operates both alpha actors | Role independence is overstated | Claim cryptographic actor separation, not independent human organizations; document the limitation. |

## Explicitly deferred

- Lost-key, compromise, social, organizational, or threshold recovery.
- Shared private-key concurrent writers.
- Secrets and runner network/resource hardening.
- Portable and container backends.
- Shared moderation and redaction semantics.
- Merge queues, notifications, web UI, search, and discovery.
- Hard pre-download transport byte quotas beyond standard Git capabilities.

## Open questions

There are no unresolved product-scope decisions. Exact CLI names and the local
keyring/selection file layouts are implementation-plan concerns; they must
remain private under `.git/nh`, use atomic file replacement, and expose the
behaviors and invariants defined above. Exact public verification IDs will be
filled as the two-stage landing gate advances.
