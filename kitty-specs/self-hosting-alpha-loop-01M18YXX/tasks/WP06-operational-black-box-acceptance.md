---
work_package_id: WP06
title: Operational Black-Box Acceptance
dependencies:
- WP01
- WP02
- WP03
- WP04
- WP05
requirement_refs:
- FR-003
- FR-005
- FR-006
- FR-007
- FR-008
- FR-009
- FR-010
- FR-011
- FR-012
- FR-013
- FR-014
- FR-015
- FR-016
- FR-017
- FR-018
- FR-019
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T028
- T029
- T030
- T031
- T032
- T033
phase: Phase 4 - Operational Proof
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: reviewer-renata
agent: codex
authoritative_surface: main.go
create_intent:
- operational_acceptance_test.go
execution_mode: code_change
model: ''
owned_files:
- main.go
- operational_acceptance_test.go
role: reviewer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP06 – Operational Black-Box Acceptance

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Integrate every new command at the executable boundary and prove the complete
Operational Self-Hosting Alpha behavior without internal knowledge. The test
must drive a compiled `nh` binary in separate processes across multiple real
Git repositories and one disposable bare remote.

The scenario must amend policy under its old rules, establish distinct actor
continuity, isolate hostile/unselected/over-budget replication, recover a
shallow dependency explicitly, require non-author evidence for a later
candidate, merge it, and reconstruct the result from an identity-free clone.
It passes three consecutive times within 120 seconds per run.

## Context & Constraints

Read all acceptance scenarios in [spec.md](../spec.md), the target operator
flow in [quickstart.md](../quickstart.md), and the verification matrix in
[plan.md](../plan.md). This is a black-box test: assertions use CLI inputs,
outputs, exit statuses, Git refs/objects, and repository files, not internal Go
functions or implementation layout.

Automated tests are offline and may never mutate the public remote or require
credentials. Do not invoke this repository's `go test ./...` from a pipeline
inside the acceptance scenario; use a tiny synthetic action to prevent
recursive test execution. No Docker requirement is allowed.

## Branch Strategy

- **Strategy**: integration lane based on completed WP01-WP05; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Use the integration lane from `lanes.json`, which contains all dependencies.
`main.go` is owned here specifically so top-level routing and usage are changed
once after every command contract exists.

## Subtasks & Detailed Guidance

### T028 – Build a subprocess-level multi-repository harness

- Begin `operational_acceptance_test.go` with helpers that build the current
  package once into a temporary binary and execute it with explicit working
  directory, environment, stdin, timeout, and captured stdout/stderr.
- Create isolated author, reviewer/runner, and verifier repositories plus one
  ordinary bare remote. Configure Git locally in each repository.
- Scrub credential helpers, tokens, global config surprises, and inherited
  Nichthub private paths from subprocess environments.
- Provide helpers for exact ref/object inspection and full identifier parsing
  from stable output labels; do not rely on column spacing or short IDs.
- Use per-process working directories rather than `os.Chdir`, allowing tests to
  run safely under the race detector and alongside other packages.
- Fail every process within a bounded timeout and print concise captured output
  on failure without leaking synthetic private key values.

### T029 – Exercise policy amendment under the old policy

- Initialize a one-maintainer repository and commit the baseline policy and a
  tiny pipeline definition/action.
- Prepare a proposed policy adding the second actor, requiring role-distinct
  evidence, and disallowing author approval where specified.
- Run policy show/check over base and head; assert exact digests, sorted actor
  changes, and base-governs language.
- Open the amendment candidate and attempt evidence from the newly added actor
  before merge. It must remain a valid signed fact but non-qualifying under the
  old base policy.
- Satisfy the amendment only with actors authorized by the old policy, decide,
  merge, and publish collaboration refs/primary branch as separate observable
  actions to the disposable remote.
- Prove the amended policy takes effect only for a later candidate whose base
  contains it.

### T030 – Exercise distinct actors and planned rotation

- Initialize the second clone independently and compare full actor/public-key
  material; assert neither `.git/nh` tree nor private key is copied.
- Authorize and accept a device relationship through synchronized exact facts;
  both clones must display the same accepted relationship while policy remains
  unchanged.
- Perform a disposable planned successor rotation and verify two actor chains,
  both signatures, predecessor history retention, and successor sequence one.
- Interrupt/retry at least one rotation boundary through a supported test seam
  or controlled process failure and prove the old signer stays active until
  completion.
- Add a competing/cyclic fixture through the remote and assert visible
  ambiguity with no inferred authority winner.
- Confirm only exact policy actor lists determine which second-actor claims
  qualify.

### T031 – Exercise selected replication, hostile input, and budgets

- Publish selected valid actor/candidate refs plus unselected, malformed,
  mismatched, missing-dependency, and oversized refs to the disposable remote.
- Save full selections and positive budgets in each participant clone; assert
  local selection files are untracked and never pushed.
- Synchronize mixed valid/invalid selections and inspect accepted remote refs:
  independent valid refs promote; invalid, over-budget, dependency-missing,
  and unselected refs do not.
- Exercise each event/object/object-byte/attachment-byte/total-byte boundary at
  one below, exactly at, and one above. Focused lower-level tests may supply the
  exhaustive matrix, but the black-box scenario must cross every budget type.
- Verify missing dependencies report full referencing/missing IDs and an exact
  additional selection/recovery action.
- Verify identity-free sync imports accepted public facts but publishes no
  actor ref of its own.

### T032 – Exercise shallow detection and bounded recovery

- Create a genuine depth-limited verifier clone whose selected candidate has a
  required base/policy/ancestry object outside its initial branch depth.
- First run without recovery and assert a non-zero trust-sensitive result, full
  missing identifier, no signed output event, no merge commit, and unchanged
  accepted refs.
- Run with explicit `--recover-shallow` using an already selected supplying ref
  and sufficient budgets. Verify recovery uses accepted quarantine promotion
  and the restarted operation succeeds.
- Assert the repository remains shallow where unrelated history is concerned;
  no global unshallow or unrelated actor/candidate ref may appear.
- Repeat with a missing unselected dependency and assert the CLI instructs the
  exact selection rather than adding it.
- Compare reconstructed facts with the authoritative author/reviewer clones.

### T033 – Exercise role-distinct governance and repeatability

- Open a real later candidate based on the amended policy. The author alone
  must not make it ready.
- Issue the exact pipeline request and have a distinct trusted runner execute
  the tiny synthetic action with the default isolated backend where available;
  use explicit controlled host fallback only in portable automated fixtures.
- Have the distinct trusted reviewer publish approval, then have an explicit
  maintainer decide and merge from a clean base.
- Verify every request/result/review/decision/merge binds the exact candidate,
  head, pipeline definition, policy digest, and evidence IDs.
- From the fresh shallow identity-free clone, reconstruct all actors, candidate
  code refs, evidence, decision, merge fact, and resulting commit.
- Preserve the existing public seven-event compatibility baseline and all
  current proposal-revision conflict tests.
- Run the complete operational scenario three consecutive times; each run must
  finish within 120 seconds and produce equivalent relationships (timestamps
  and generated IDs may differ unless fixtures fix their inputs).

## Top-Level CLI Integration

Update `main.go` once to route and document:

```text
nh identity show|list|public|authorize|accept|rotate
nh policy show|check
nh replication select|show
nh sync [REMOTE] [--recover-shallow]
```

Retain every existing command. Usage text must describe full-ID trust inputs
and must not promise a service, Docker, lost-key recovery, or hard pre-download
byte quotas.

## Test Strategy

Required gates:

```bash
go test ./... -run TestOperationalSelfHostingAlpha -count=1
go test ./... -run TestOperationalSelfHostingAlpha -count=3
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

If Bubblewrap is unavailable in a test environment, isolate that platform fact
from protocol acceptance and retain explicit coverage of default sandbox
behavior on supported Linux. Do not silently turn host execution into the
default.

## Risks & Mitigations

- **Recursive execution**: the synthetic pipeline action must not run this
  repository's test suite.
- **Global-state flakes**: explicit cwd/env per subprocess; no process-wide
  directory or environment mutation.
- **Brittle output**: assert stable labels, full IDs, exit codes, and Git state.
- **Credential/public leakage**: offline bare remote only; scrub environment.
- **False role distinction**: prove distinct actor keys, while docs later state
  this does not prove independent humans/organizations.
- **Slow suite**: reuse the built binary, bound subprocesses, and keep fixture
  histories minimal while preserving real Git behavior.

## Review Guidance

The reviewer should run the three-count scenario and inspect that it invokes
only the compiled CLI and Git. Verify the new head policy cannot authorize its
own amendment, an author cannot satisfy the later role-distinct candidate
alone, and a fresh shallow clone has no identity private state.

Reject mocked signatures/Git, public network use, recursive testing, hidden
host execution, incomplete budget coverage, short trust identifiers, or tests
that pass without inspecting accepted refs and exact evidence relationships.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.
- 2026-08-30T18:32:22Z – codex – Accepted incoming handoff from WP03: main.go may already route the nh policy command. Preserve or deliberately supersede that route while implementing operational acceptance; include it in black-box coverage.
- 2026-08-30T21:53:58Z – codex – Accepted mandatory ARBITER handoff from WP05: before WP06 may pass, fix the post-promotion denial-clear crash seam centrally. collectEvents/loadStoredEvent and every accepted trust path must reject transaction-pending denied object/event commits even after shallow markers are gone. Add black-box crash/restart reconciliation, missing/corrupt denial-record fail-closed behavior, and serialized/per-transaction union or ref-count semantics so one transaction cannot clear another's denial and pre-existing accepted objects are never falsely denied. Reproduce WP05 review-cycle-3 failure. Mission acceptance is prohibited until this passes independent review.
- 2026-08-30T22:33:41Z – codex – Implemented the authorized arbiter repair across store.go, shallow.go, and quarantine.go: durable per-transaction pending/accepted object receipts now gate every accepted trust loader independently of shallow markers, fail closed on incomplete/corrupt state, preserve concurrent denial unions, spare pre-existing accepted objects, and reconcile only after completion is durable. Added a compiled-process crash/restart reproduction of the WP05 review-cycle-3 post-promotion/marker-release seam.
- 2026-08-30T22:33:41Z – codex – Completed T028–T033 operational proof in commit 8eca178 with compiled nh subprocesses, isolated repositories and bare remotes, old-policy amendment authority, device continuity and interrupted/retried successor rotation, remote cycle ambiguity without inferred authority, hostile exact replication across all five budgets, bounded shallow recovery including refusal of an unselected supplier, role-distinct governance, and identity-free reconstruction. Operational x1 and x3, full, race, vet, build, formatting, and diff checks passed; no dependencies changed and the diff-scoped ruff check had no Python files.
