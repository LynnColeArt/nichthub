---
affected_files: []
cycle_number: 4
mission_slug: hn-hard-cutover-01M1EY1B
reproduction_command:
reviewed_at: '2026-09-01T18:11:11Z'
reviewer_agent: user
wp_id: WP01
---

# WP01 Review Feedback 4

Verdict: changes requested. Commit `51405b0` closes the core QA findings in
ordinary repositories, and its new tests are mutation-sensitive, but two
integration paths still violate the intended full-ID/recovery contract.

## Issue 1 — Decision recovery guidance emits an unusable short ID (blocking)

`governance.go:66` and `governance.go:88` tell the operator to run
`hn proposal status <shortID>`. `cmdProposalStatus` now correctly requires a
full `sha256:<64-hex>` event ID, so following either diagnostic immediately
fails with the new full-ID guard. This leaves the F-1 cutover internally
inconsistent even though direct trust-bearing inputs are guarded.

Independent reproduction created a not-ready proposal and called `cmdDecide`
with its full ID. The returned command contained only the 12-character display
prefix and failed an assertion that the suggested command include the exact
proposal ID.

Use `proposal.ID` rather than `shortID(proposal.ID)` in both executable recovery
commands. Add a command-path test that extracts or checks the diagnostic and
proves the suggested `hn proposal status` invocation uses the full ID accepted
by `cmdProposalStatus`. Continue using short IDs for non-command display text.

## Issue 2 — Shallow verification blocks merge-event repair before it starts (blocking)

In a shallow repository, `cmdMerge` first calls `prepareShallowVerification`.
The `proposal merge` branch in `verifyShallowOperation` still returns
`proposal head is already contained in current branch` when containment is
true. Containment is now the trigger for F-2 reconciliation, so this preflight
prevents the new repair branch in `cmdMerge` from ever running in a shallow
checkout, even when all exact objects and evidence are locally available.

Independent reproduction used the real crash-after-code-merge fixture, marked
the repository shallow at the proposal base, and retried the full-ID merge.
The retry failed with the preflight error above and recorded no merge event.

Update shallow merge verification so a fully verified contained head is an
allowed precondition for the main command's repair path, while missing exact
objects still produce the existing fail-closed shallow dependency flow. Add a
shallow crash/retry test proving the repair reaches the same exact-policy,
approval, CI, accepting-decision, merge-commit, idempotency, and append-CAS
checks as the ordinary path.

## Verified good

- F-1 direct call-path guards cover issue comment, proposal revise/status,
  review, run request/execute, decision, merge, and identity acceptance; guards
  run before shallow verification, so locally unambiguous prefixes fail closed.
- F-2 ordinary repair re-resolves the exact proposal, reloads the immutable
  base policy, recomputes current approvals/results/decisions, locates a unique
  first-parent merge directly naming the signed head, and records that exact
  commit without rerunning `git merge`.
- F-2 ordinary idempotency and stale approval/CI rejection pass. A reviewer-only
  test injected a second append lock during reconciliation; the failed repair
  remained retryable, preserved `HEAD`, and ultimately produced exactly one
  merge event.
- F-3 resolves `bwrap` only through the fixed absolute directories returned by
  `sandboxSystemDirectories`, and `sandboxPath` uses that identical ordered
  set. Reverting discovery to ambient `exec.LookPath` made the hostile-PATH
  test fail.
- F-5 has current `errShallowRecoverySelectionRequired` semantics with no WP05
  text or stale production reference. Restoring the old message made its test
  fail.
- Focused QA tests, `go vet ./...`, `go build -o hn .`, formatting, and diff
  checks: PASS.
- `go test -race` for the QA-focused suite: PASS (`ok hubnot 4.399s`).
- Full `go test -count=1 ./...`: PASS (`ok hubnot 81.078s`).
- Namespace scan, WP ownership, module boundary, and frozen `.nh/` hashes:
  PASS.

## Mutation/deletion evidence

In isolated copies, each of these changes independently made its corresponding
test fail: removing the issue-comment full-ID guard, removing F-2's current
evidence readiness check, restoring ambient `LookPath`, and restoring the stale
WP05 sentinel. The new tests therefore invoke production behavior rather than
asserting synthetic fixtures.

## Anti-pattern checklist

1. Dead code: **PASS** — each new production helper has live production callers.
2. Synthetic-fixture test: **PASS** — focused mutation checks killed all four
   intended regressions; the two missing integration cases are separately
   blocking under Issues 1–2.
3. Silent empty return: **N/A** — no silent empty production return was added.
4. FR coverage: **FAIL** — executable full-ID recovery guidance and shallow
   merge-event reconciliation are untested and currently fail.
5. Frozen surface: **PASS** — frozen paths are absent from the diff and both
   checked-in `.nh/` SHA-256 hashes match the WP base.
6. Locked decision: **FAIL** — the short-ID command contradicts F-1, and the
   shallow preflight contradicts F-2's retry repair requirement.
7. Shared-file ownership: **PASS** — all changes remain in WP01-owned Go files.
8. Production fragility: **N/A** — no bare exception/raise analogue was added;
   fail-loud branches have explicit recovery context.
