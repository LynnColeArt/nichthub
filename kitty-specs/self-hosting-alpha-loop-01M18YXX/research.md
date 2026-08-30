# Research: Self-Hosting Alpha Loop

## Research question

Can the currently implemented Nichthub commands compose into one real,
publicly reconstructable development loop for the Nichthub repository itself,
and what is the smallest durable change needed to prove that claim repeatedly?

## Findings

### The protocol pieces already compose

The current CLI exposes every state transition required by the happy path:
issue creation, immutable proposal creation, pipeline request and execution,
review, policy evaluation, acceptance, merge, and synchronization. Proposal
code is transported through a content-bearing ref, while actor histories carry
the signed evidence. Merge changes the checked-out target locally; publishing
that target branch remains an ordinary explicit Git push. [S002] [S003]

Existing tests prove the individual distributed seams, including two-clone
pipeline execution, code-ref transfer, policy evaluation, and immutable
revision recovery. They do not yet exercise the entire operator-visible CLI
workflow as one black-box scenario. Most tests call command or domain functions
inside the test process. [S005]

### The public remote supports the required namespaces

On 2026-08-30, `nh sync` published the active actor history to the public
`LynnColeArt/nichthub` remote. The remote advertised both the actor ref and the
existing proposal code ref unchanged:

```text
refs/nh/actors/36944394addccd027292abc8183f332af8b4590925291ee8f1e8d8f09446b7dd
refs/nh/proposals/4fd8813d5849241173eb48bbebd6858535545fa514a80b88d3fdae0e59377290
```

This closes two gaps in the previous compatibility record: the repository is
public, and `refs/nh/proposals/*` has now been observed on GitHub.com rather
than only on a local bare remote. [S004] [S006]

### Fresh verification does not require an identity

A fresh public clone with no `.git/nh/identity.json` ran `nh sync`, verified
seven events, and displayed the new self-hosting issue. The full issue event ID
is:

```text
sha256:7204f1b6c2eb16b1d25f1f6269f0cce1eed354636c432baf981e44e1cf83a79a
```

The issue is actor sequence 7 and follows the earlier complete governance
proposal chain. Read-only synchronization intentionally tolerates the absence
of a local private identity; it fetches and validates before conditionally
publishing local actor state. [S002] [S006] [S007]

### The live loop is a bootstrap landing gate

Spec Kitty assembles and validates this mission on
`feat/self-hosting-alpha-loop`. The complete candidate can only be proposed
after that feature branch exists. Therefore the mission has two explicit
boundaries:

1. Build and accept the repeatable scenario and documentation on the feature
   branch under Spec Kitty.
2. Use Nichthub to propose that exact feature head against `main`, earn live
   evidence, merge locally, publish `main`, and verify from a fresh clone.

This is not circular: Spec Kitty decides whether the candidate is internally
ready; Nichthub supplies the public landing gate that the mission is proving.
[S001]

### The inaugural run proves transport and integrity, not role independence

The committed base policy trusts the current maintainer actor as maintainer,
reviewer, and runner and explicitly permits author approval. One identity can
therefore complete the inaugural loop without changing policy. This validates
composition, exact evidence binding, public transport, and reconstruction. It
does not demonstrate independent organizational checks or multi-device actor
append, both of which remain later roadmap work. [S001] [S003]

## Decisions

### D-001 — Add one black-box acceptance scenario

**Decision**: Exercise the full offline workflow through the CLI boundary in
isolated subprocesses against a disposable ordinary Git remote.

**Rationale**: A single scenario detects composition gaps that function-level
tests cannot, while subprocesses isolate working directory, standard streams,
and private identity state. The scenario must use a tiny repository-owned
action rather than the Nichthub repository's own `go test ./...` pipeline, so
the acceptance test cannot recursively invoke itself. [S005]

**Consequences**: The test will be slower than a unit test but remains bounded
to 90 seconds. The synthetic repository action is controlled test input, so the
explicit unsafe-host backend may be used for deterministic offline coverage;
the live public run must still use the default isolated backend. [S001] [S003]

### D-002 — Keep the public host out of automated tests

**Decision**: Automated acceptance uses a local bare Git remote. The real
public remote is exercised once as release evidence and documented by exact
identifiers.

**Rationale**: Automated tests must be deterministic, offline-capable, and
unable to mutate public state or consume credentials. A bare remote exposes the
same Git ref behavior the protocol depends on. [S001] [S004]

### D-003 — Correct only observed blockers

**Decision**: Do not add a new event kind, issue-to-proposal relation, service,
or provider integration unless the black-box or live loop fails because the
existing behavior cannot satisfy the specification.

**Rationale**: The mission is an integration proof. Inventing adjacent product
behavior would hide whether the current protocol already works and expand the
review surface without evidence. The issue and candidate are correlated in the
verification record for this alpha; a signed protocol relationship is a future
feature. [S001] [S008]

### D-004 — Record exact public identifiers

**Decision**: Add a self-hosting operations document containing the operator
flow, role assumptions, exact inaugural identifiers, fresh-clone observations,
and limitations. Update the hosted compatibility matrix with the new public and
proposal-ref result.

**Rationale**: The durable proof is the signed repository data. Exact IDs make
the narrative independently checkable and avoid screenshots or provider UI as
the authority. [S004] [S006] [S007]

### D-005 — Treat partial publication as an explicit operator boundary

**Decision**: Documentation must distinguish `nh sync`, which publishes
collaboration refs, from the ordinary Git push that publishes the merged target
branch. It must verify both after the live merge.

**Rationale**: The CLI deliberately does not push code branches as a side
effect of synchronization or merge. Concealing this boundary would make the
happy path appear to succeed while the public primary branch remained stale.
[S002] [S003]

## Candidate implementation seams

- Add a black-box end-to-end test file that invokes the existing CLI entry
  point in subprocesses and creates its repositories under the test temporary
  directory.
- Reuse existing test Git helpers where that does not cross the subprocess
  boundary; keep scenario-specific orchestration local to the acceptance test.
- Add a public self-hosting operations/evidence document.
- Update `docs/host-compatibility.md` for the observed public actor and proposal
  namespaces.
- Change production behavior only if a failing acceptance step reproduces an
  actual blocker through the current CLI.

## Risks and mitigations

| Risk | Consequence | Mitigation |
| --- | --- | --- |
| Acceptance test recursively runs itself | Unbounded execution | Use a synthetic repository pipeline with a tiny repository-owned executable, not `go test ./...`. |
| Tests mutate the public repository | Nondeterminism and unwanted public history | Use only a temporary bare remote in automation; keep live actions manual and identifier-recorded. |
| One actor fills every trusted role | Proof could be overstated | State that the alpha proves composition and integrity, not independent approval. |
| Collaboration refs publish but `main` does not | Public code and evidence disagree operationally | Verify both custom refs and `refs/heads/main` after separate publish actions. |
| Existing actor history contains earlier experiments | Fresh-clone counts are not mission-specific | Address all inaugural facts by full ID rather than assuming an empty history. |
| Default isolated runner is unavailable | Live acceptance cannot satisfy C-004 | Preflight backend availability before requesting live evidence; do not silently substitute host execution. |
| A post-proposal blocker changes the candidate | Evidence becomes stale | Publish an immutable revision and re-earn review, CI, and acceptance. |

## Open questions

There are no unresolved scope decisions. Exact proposal, run, review, decision,
merge, and resulting commit IDs will be filled into the verification record as
the live landing gate advances. Any observed CLI blocker becomes a test-first
work item; absence of a blocker is a valid and expected outcome.
