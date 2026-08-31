# Mission Review Report: self-hosting-alpha-loop-01M18YXX

**Reviewer**: Codex post-merge mission reviewer  
**Date**: 2026-08-30  
**Mission**: `self-hosting-alpha-loop-01M18YXX` — Operational Self-Hosting Alpha  
**Baseline commit**: `2353a077c38353334a86739d6e71d394585539f4`  
**HEAD at review**: `5273820cbedd16a2206f66580b21b6c256210386`  
**WPs reviewed**: WP01–WP07

## Gate Results

### Gate 1 — Contract tests

- Command: `go test ./...`
- Exit code: 0
- Result: PASS
- Notes: The repository is a single-package Go CLI, so its executable contract,
  integration, compatibility, and black-box tests live in the root `*_test.go`
  files. The merged target passed the complete suite.

### Gate 2 — Architectural tests

- Commands: `go vet ./...`; `go build ./...`; byte comparison against approved
  WP06 commit `bd99802`
- Exit codes: 0, 0, identical
- Result: PASS
- Notes: The Python-only Spec Kitty `tests/architectural/` gate is not a
  Nichthub repository surface. The applicable architecture is the planned flat
  Go package with deep boundary files; vet and build pass, all new production
  modules have live CLI callers, and the merged Go surface is byte-identical to
  the independently approved WP06 lane.

### Gate 3 — Cross-repository operational E2E

- Command: `go test ./... -run '^TestOperationalSelfHostingAlpha' -count=3 -v`
- Exit code: 0
- Result: PASS
- Notes: Three consecutive compiled-CLI runs over disposable repositories and
  bare remotes passed in 45.98 seconds total. The principal full-loop cases took
  8.97–9.03 seconds each. The suite includes policy amendment, distinct actors,
  rotation, hostile selected replication, every budget, shallow recovery,
  role-distinct governance, crash recovery, hostile private-state paths, and
  identity-free reconstruction. This project does not modify the Spec Kitty
  runtime, so the Spec Kitty cross-repository scenario suite is not applicable.

### Gate 4 — Issue matrix

- File: not scaffolded because the mission references no external issues
- Rows: 0
- Empty or unknown verdicts: 0
- Result: PASS
- Notes: The merge gate reported: “No issue references discovered — nothing to
  enforce.”

### Additional release gates

- `go test -race ./...`: PASS
- `git diff --check 2353a077..HEAD`: PASS
- Tracked `.git/nh` private state: none
- Private-key material in shipped docs/artifacts: none
- Acceptance matrix: 20/20 criteria PASS
- Post-merge stale-assertion scan: no findings

## FR Coverage Matrix

| FR | Brief promise | WP owner | Principal proof | Adequacy | Finding |
|---|---|---|---|---|---|
| FR-001 | Inspect exact policy | WP03 | `TestPolicyShowAndCheckContract` | ADEQUATE | — |
| FR-002 | Validate and diff amendment | WP03 | `TestPolicyShowAndCheckContract`, `TestPolicyCheckRejectsInvalidSidesWithoutMutation` | ADEQUATE | — |
| FR-003 | Old policy governs amendment | WP03/WP06 | `TestPolicyEvaluationUsesExactBaseAcrossAmendment`, operational E2E | ADEQUATE | — |
| FR-004 | Reject lockout | WP03 | `TestPolicyCheckRejectsInvalidSidesWithoutMutation` | ADEQUATE | — |
| FR-005 | Distinct device actors | WP01/WP02/WP06 | identity continuity command tests, operational E2E | ADEQUATE | — |
| FR-006 | Planned rotation | WP01/WP02/WP06 | rotation crash/retry matrix, operational E2E | ADEQUATE | — |
| FR-007 | Explicit deterministic continuity | WP02/WP06 | deterministic, cycle, competing-successor, remote-ambiguity tests | ADEQUATE | — |
| FR-008 | No implicit project authority | WP02/WP03/WP06 | `TestIdentityContinuityDoesNotGrantPolicyAuthority`, base-policy evaluation | ADEQUATE | — |
| FR-009 | Persist exact local selection | WP04/WP06 | `TestReplicationSelectionRoundTripAndRejectsAmbiguity`, E2E | ADEQUATE | — |
| FR-010 | Quarantine before projection | WP04/WP06 | independent promotion and hostile-layer tests, crash denial tests | ADEQUATE | — |
| FR-011 | Isolate invalid selections | WP04/WP06 | selected replication and hostile E2E | ADEQUATE | — |
| FR-012 | Enforce positive budgets | WP04/WP06 | real transaction boundary matrix and E2E five-budget matrix | ADEQUATE | — |
| FR-013 | Exact missing-dependency guidance | WP04/WP05/WP06 | real missing event/object/policy/pipeline tests | ADEQUATE | — |
| FR-014 | Detect actual shallow boundary | WP05/WP06 | wrong-type/malformed/disjoint preservation and exact-gap tests | ADEQUATE | — |
| FR-015 | Bounded selected recovery | WP05/WP06 | selection-preserving recovery and budget tests | ADEQUATE | — |
| FR-016 | Role-distinct public loop | WP03/WP06/WP07 | operational E2E and public evidence record | ADEQUATE | — |
| FR-017 | Fresh shallow reconstruction | WP04/WP05/WP06/WP07 | operational shallow verifier and credential-disabled public reconstruction | ADEQUATE | — |
| FR-018 | Recoverable fail-closed errors | WP01–WP06 | keyring, rotation, replication, shallow, receipt and pending-anchor matrices | ADEQUATE | — |
| FR-019 | Automated offline scenario | WP04–WP06 | compiled-CLI operational suite, three consecutive runs | ADEQUATE | — |
| FR-020 | Durable public evidence | WP07 | `docs/self-hosting-alpha.md` and fresh public reconstruction | ADEQUATE | — |

All tests above invoke production storage, signing, Git, projection, and CLI
paths. The synthetic-continuity and synthetic-budget false positives identified
in early review cycles were replaced by real signed repository transactions.

## Drift Findings

### DRIFT-1: Draft replication contract retains the superseded active-only publication rule

**Type**: LOCKED-DOCUMENT DRIFT  
**Severity**: MEDIUM  
**Reference**: selected replication contract step 9; FR-005, FR-006, FR-017

The draft contract says synchronization publishes only the active local actor
ref. Shipped `publishLocalFacts` and the canonical operator documentation
publish every local public actor ref. The shipped behavior is necessary: a
rotation's predecessor history must remain available so a fresh verifier can
validate continuity. Private keyring, selection, receipt, and transaction state
remain local. This intentional design correction is documented in WP07 history,
but the mission contract itself was not synchronized.

Evidence:

- `contracts/selected-replication-v0.md:56`
- `commands.go:272`
- `docs/replication-v0.md:155`

Follow-up: update the contract's step 9 to match the approved public behavior.

### DRIFT-2: Spec Kitty governance artifacts contain workstation-absolute paths

**Type**: NFR-MISS  
**Severity**: LOW  
**Reference**: NFR-002

The shipped Nichthub product, signed facts, README, public protocol docs, and
public verification record contain no host-private path. Spec Kitty's tracked
analysis/status/retrospective metadata does contain absolute workspace paths.
Most instances predate the implementation baseline; the terminus event added
one absolute retrospective path. These paths disclose no credential or private
key and do not enter Nichthub's signed/public protocol, but the broad wording of
NFR-002 says “tracked files,” so the governance metadata is a literal exception.

Follow-up: have Spec Kitty serialize repository-relative evidence and feedback
paths in public project metadata.

## Risk Findings

### RISK-1: WP05's terminal UI retains a stale rejected-artifact warning

**Type**: PROCESS / REVIEW-PROVENANCE  
**Severity**: LOW

WP05 reached the three-rejection cap and was conditionally advanced so WP06
could own the required crash/restart repair. WP06 cycle 2 independently proved
that repair, and the acceptance matrix plus merged tests pass. The canonical
status is `done`, but `spec-kitty agent tasks status` still warns that WP05's
latest ordinary review artifact was rejected. The arbiter handoff and later
approval are durable, so this is not a product defect; it can confuse future
automated readers that ignore the handoff evidence.

## Silent Failure Candidates

No blocking silent-failure candidate was found. Empty results on unchanged
policy comparisons and gap-free shallow recovery are explicit success/no-op
contracts. Best-effort directory-sync cleanup paths retain durable retry state
and were covered by interruption tests.

## Security Notes

| Finding | Location | Risk class | Disposition |
|---|---|---|---|
| Git commands use argument vectors, not shell strings | `git.go:48` | SHELL-INJECTION | PASS |
| Identity records and pointers reject symlinks, unsafe modes, malformed and oversized state | `identity_keyring.go`, identity tests | PRIVATE-STATE | PASS |
| Replication receipts and pre-copy pending anchors fail closed across loss, corruption, symlinks, unsafe modes, and concurrency | `shallow.go:424`, operational tests | LOCK/CRASH-TOCTOU | PASS |
| Transport errors redact URLs, credentials, quarantine paths, and local transaction paths | `replication_test.go:528` | CREDENTIAL-DISCLOSURE | PASS |
| Spec Kitty metadata retains absolute workspace paths | mission analysis/status/retrospective | PATH-DISCLOSURE | Non-blocking DRIFT-2 |

## Final Verdict

**PASS WITH NOTES**

### Verdict rationale

All 20 functional requirements have adequate production-path tests, the
release-gating NFR behavior passes, the three-run operational scenario is well
inside its time limit, and the approved crash/restart repair survived the
post-merge suite. No high or critical code, authority, transport, or secret
handling defect remains. The two notes are documentation/process hygiene: one
stale draft contract sentence and Spec Kitty's absolute-path metadata.

### Open items (non-blocking)

1. Align selected-replication contract step 9 with all-local-public-actor
   publication.
2. Track upstream Spec Kitty work to emit repository-relative metadata paths.
3. Preserve the WP05 arbiter-to-WP06 closure evidence when summarizing mission
   history so the stale rejected-artifact warning is not mistaken for an open
   product defect.

## Retrospective Reminder

The terminus runtime authored `retrospective.yaml` for mission
`01M18YXXWM2WBE3DCHD6DS2RDK`. The canonical sequence after this report is to
review `spec-kitty retrospect summary` and run
`spec-kitty agent retrospect synthesize --mission self-hosting-alpha-loop-01M18YXX`
in dry-run mode before applying any staged proposal.
