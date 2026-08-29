---
affected_files: []
cycle_number: 1
mission_slug: proposal-revision-conflict-recovery-01M1774Q
reproduction_command:
reviewed_at: '2026-08-29T17:45:45Z'
reviewer_agent: codex
wp_id: WP01
---

# WP01 review cycle 1 — changes requested

Reviewer: Reviewer Renata (loaded manually because `spec-kitty agent action review`
rolled back on its empty-status-commit defect).

## Blocking findings

1. `event.go:143` accepts `Title: "   "` for `proposal.revise` because it checks
   `strings.TrimSpace(event.Title) != ""`. The signed-wire model says the title
   is not copied into the revision event; any non-empty JSON title value is a
   second authoritative value. Reject `event.Title != ""` and add the
   whitespace-only case to `TestProposalRevisionContentValidation`.

2. `store.go` now admits revision subjects for `run.request`,
   `proposal.decision`, and `proposal.merged`, but the new tests exercise only
   `review.submit`. Add observable relationship tests proving the three newly
   generalized paths accept a valid revision while retaining their exact
   commit/head/policy/evidence checks. This closes T004's changed behavior and
   FR-015's compatibility boundary.

3. `TestProposalRevisionRelationshipRejections` only checks `err != nil`.
   Strengthen hostile cases to assert the stable reason and involved short
   revision/event identity required by T004 and FR-013. The checks should fail
   if validation regresses to an unhelpful generic error.

## Review evidence

- PASS — owned-file/locality boundary: product diff is only `event.go`,
  `store.go`, and `revision_event_test.go`; no dependency or ref-namespace
  change.
- PASS — architecture: narrow event classifier plus complete-set relationship
  validation; no timestamp, arrival-order, mutable-latest, or actor-fork rule.
- PASS — deletion test: restoring pre-WP `event.go` and `store.go` makes every
  `TestProposalRevision*` group fail for the intended missing behaviors.
- PASS — `go test -count=1 -race ./...`, `go test -count=1 -cover ./...`
  (47.5%), `go vet ./...`, `go build ./...`, and diff check.
- FAIL — frozen wire surface: whitespace title probe fails because the invalid
  serialized value is accepted.
- FAIL — FR/changed-path coverage: revision run/decision/merge relationships
  and actionable error content are not contract-pinned.
- N/A — dependency/supply-chain review; no dependency files changed.
- N/A — dead public API, silent empty return, and production raise/fragility;
  this WP adds no public API, dependency, exception mechanism, or silent-return
  path.

Re-review after the focused tests are red-first, the fix is applied, and all
repository gates pass again.
