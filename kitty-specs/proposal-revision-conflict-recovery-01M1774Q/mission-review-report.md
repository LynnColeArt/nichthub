---
verdict: pass_with_notes
mode: post-merge
reviewed_at: 2026-08-29T18:31:00-05:00
reviewer: codex
baseline_commit: 3f4164705694f869ec2959fe6066833a9b84d5be
reviewed_head: 99b21dc
findings: 2
gates_recorded:
  - id: gate_1
    name: go_contract_tests
    command: go test -count=1 -race ./...
    exit_code: 0
    result: pass
  - id: gate_2
    name: go_architecture_and_build
    command: go vet ./... && go build ./...
    exit_code: 0
    result: pass
  - id: gate_3
    name: cross_repo_e2e
    command: not_applicable
    exit_code: 0
    result: not_applicable
  - id: gate_4
    name: issue_matrix
    command: spec-kitty review --mission 01M1774Q --mode post-merge
    exit_code: 0
    result: not_applicable
issue_matrix_present: not_applicable
mission_exception_present: false
---

# Mission Review Report: Proposal Revision and Conflict Recovery

**Mission**: `proposal-revision-conflict-recovery-01M1774Q`

**Baseline commit**: `3f4164705694f869ec2959fe6066833a9b84d5be`

**HEAD reviewed**: `99b21dc`

**WPs reviewed**: WP01–WP05

## Gate Results

### Gate 1 — Contract tests

- Command: `go test -count=1 -race ./...`
- Exit code: 0
- Result: PASS
- Notes: All root-package unit, CLI, real-Git, governance, and bare-remote
  integration tests passed. `go test -count=1 -cover ./...` also passed with
  63.8% statement coverage.

### Gate 2 — Architectural and build checks

- Command: `go vet ./... && go build ./...`
- Exit code: 0
- Result: PASS
- Notes: The repository is a flat Go package, so the Python architectural-test
  path named by the generic Spec Kitty review profile does not exist. Manual
  live-caller tracing confirms `proposal.revise` enters through `main.go` and
  `proposal.go`; `lineage.go` is called by proposal display, policy evaluation,
  acceptance, and merge paths. `git diff --check` also passed.

### Gate 3 — Cross-repository E2E

- Command: not applicable
- Exit code: 0
- Result: NOT APPLICABLE
- Notes: This mission changes Nichthub's repository-local Go protocol and does
  not claim behavior in an external Spec Kitty E2E repository. Its distributed
  boundary is instead exercised through two real clones and a bare Git remote
  by `TestProposalRevisionSyncAndConvergence`.

### Gate 4 — Issue matrix

- File: not applicable; the mission declares zero canonical issue references.
- Rows: 0
- Result: PASS / NOT APPLICABLE

### Tooling compatibility note

`spec-kitty review --mission 01M1774Q --mode post-merge` reported
`MISSION_REVIEW_DEAD_CODE_UNDETERMINABLE` because its scanner supports changed
Python files only. This is not a dead-code finding. The Go-specific caller,
build, vet, and test checks above replace that undeterminable scan.

## FR Coverage Matrix

| FR | WP | Executable evidence | Adequacy |
|---|---|---|---|
| FR-001 | WP01/WP03 | `TestProposalRevisionSignedRoundTrip`; `TestProposalRevisionCommandAndLineageInspection` | ADEQUATE |
| FR-002 | WP01 | `TestProposalRevisionRelationships`; `TestProposalRevisionRelationshipRejections` | ADEQUATE |
| FR-003 | WP01/WP03/WP05 | `TestProposalRevisionCommandAndLineageInspection`; `TestProposalRevisionSyncAndConvergence` | ADEQUATE |
| FR-004 | WP02/WP04 | `TestLineageStatesPreserveMergeFacts`; `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-005 | WP04 | `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-006 | WP04 | `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-007 | WP02/WP03/WP04 | `TestProposalRevisionCommandAndLineageInspection`; `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-008 | WP02/WP03/WP05 | `TestLineageTopologyConvergesAcrossInputOrders`; `TestProposalRevisionSyncAndConvergence` | ADEQUATE |
| FR-009 | WP03/WP04 | `TestProposalRevisionGovernanceRelationshipsAndLaterMergeFact`; `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-010 | WP01/WP02/WP03 | `TestLineageTopologyConvergesAcrossInputOrders`; `TestProposalRevisionCommandAndLineageInspection` | ADEQUATE |
| FR-011 | WP02/WP04 | `TestLineageStatesPreserveMergeFacts`; `TestRevisionEvidenceAndLineageGovernance` | ADEQUATE |
| FR-012 | WP02/WP04/WP05 | `TestLineageStatesPreserveMergeFacts`; `TestRevisionEvidenceAndLineageGovernance`; `TestProposalRevisionSyncAndConvergence` | ADEQUATE |
| FR-013 | WP01/WP03 | `TestProposalRevisionRelationshipRejections`; `TestProposalRevisionCommandAndLineageInspection` | ADEQUATE |
| FR-014 | WP03/WP04 | `TestProposalMergeConflictAbortsWithRevisionGuidance` | ADEQUATE |
| FR-015 | WP04/WP05 | Existing unchanged suite; `TestProposalAndReviewSync`; `TestProposalRevisionSyncAndConvergence` | ADEQUATE |
| FR-016 | WP05 | `TestProposalRevisionSyncAndConvergence` | ADEQUATE |

The tests exercise production command and storage paths rather than replacing
them with expected-output fixtures. The lineage unit fixtures are appropriate
only for the pure projection module and are complemented by public CLI and
real-remote tests.

## Spec-to-code fidelity

- The signed wire kind and strict inherited-title rule are implemented in
  `event.go`; author, predecessor, exact-head, and cycle relationships are
  validated in `store.go`.
- `proposal.go` publishes a new event and a new immutable proposal ref without
  rewriting the predecessor, prints exact IDs, refuses merged predecessors,
  and exposes lineage through list/show.
- `lineage.go` derives sorted predecessors, successors, siblings, members,
  merge facts, supersession, closure, and conflicts from verified facts without
  a timestamp winner or mutable latest pointer.
- `policy.go`, `ci.go`, and `governance.go` keep evidence scoped to the exact
  candidate ID/base/head and apply superseded, closed, and competing-merge
  guards before acceptance or Git mutation.
- `commands.go`, `go.mod`, and `go.sum` did not change. Real transport tests
  prove the existing actor/proposal wildcard refs carry revisions and code.
- README and protocol/governance documents match the shipped CLI and state that
  no Docker daemon, service, new ref namespace, or same-key multiwriter support
  was introduced.

No non-goal invasion, locked-decision violation, dead production module,
silent default-on-error path, new network call, dynamic shell execution, or
dependency change was found.

## Drift Findings

### DRIFT-1: Public 10,000-event latency threshold is not measured end to end

**Type**: NFR-MISS

**Severity**: MEDIUM

**Spec reference**: NFR-003 and SC-007

`BenchmarkLineageRepresentativeIndex` constructs exactly 10,000 collaboration
events, 1,000 proposals, and 100 revision links and measures index construction
plus one state query at approximately 0.50 ms/op. This strongly validates the
new pure projection, but it does not invoke `nh proposal list`, `show`, or
`status`, so it excludes Git-object collection, signature verification, policy
loading, and output formatting named by NFR-003. The implementation is not
known to miss the two-second threshold; the public-boundary threshold remains
unproven.

## Risk Findings

### RISK-1: Partial-publication retry outcome is inferred, not contract-pinned

**Type**: ERROR-PATH TEST GAP

**Severity**: LOW

**Spec reference**: NFR-004

`TestProposalRevisionCommandAndLineageInspection` injects proposal-ref
publication failure and proves that the signed revision event remains, its
exact ID is reported, no code ref exists for it, and the predecessor ref is
unchanged. It does not perform the subsequent retry and assert that the retry
produces a separate fully valid revision. The code path supports that outcome,
but the final clause of NFR-004 is not directly constrained. The injected
permission failure is also skipped on Windows.

## Silent Failure Candidates

None found in the changed production paths. Errors from event collection,
validation, ref publication, policy evaluation, and Git operations propagate
to the CLI.

## Security Notes

- Git is invoked with argument vectors through `exec.Command`, never through a
  shell. Revision strings beginning with `-` are rejected and revisions resolve
  with `--end-of-options` before use.
- Fetched events are signature- and relationship-validated before lineage
  projection; fetched proposal refs must agree with the signed exact head.
- Acceptance reloads and reevaluates current verified events before emission;
  merge applies lineage gates before changing Git state.
- No credential, HTTP, service, database, or new dependency surface was added.

## Final Verdict

**PASS WITH NOTES**

All sixteen functional requirements have adequate executable coverage through
the actual Go/Git boundaries, all merged product gates pass, the original WP01
review blockers are regression-pinned, and no blocking security or correctness
finding remains. The two notes are test-evidence gaps around a medium-priority
public performance threshold and the post-failure retry clause; neither is
evidence of a shipped functional defect.

### Open items (non-blocking)

1. Add a bounded end-to-end benchmark for list/show/status over Git-backed
   10,000-event history to close NFR-003 and SC-007 literally.
2. Extend the partial-publication test through a successful retry, with a
   platform-neutral failure seam if Windows is a supported target.
3. Extend Spec Kitty's mission-review dead-code scan to understand Go projects,
   or classify unsupported languages as `undeterminable` without forcing a
   product-failure verdict.

## Retrospective Reminder

The runtime authored
`kitty-specs/proposal-revision-conflict-recovery-01M1774Q/retrospective.yaml`
at mission terminus. Review it with `spec-kitty retrospect summary` and inspect
staged proposals with `spec-kitty agent retrospect synthesize --mission
01M1774Q` (dry-run by default); use `--apply` only after approving a proposal.
