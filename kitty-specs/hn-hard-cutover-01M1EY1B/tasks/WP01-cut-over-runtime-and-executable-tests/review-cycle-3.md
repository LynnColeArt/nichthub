---
affected_files: []
cycle_number: 3
mission_slug: hn-hard-cutover-01M1EY1B
reproduction_command:
reviewed_at: '2026-09-01T17:51:01Z'
reviewer_agent: user
wp_id: WP01
---

# WP01 reopened: unresolved QA baseline findings

WP02 correctly discovered that the 2026-09-01 QA findings were reported but not
implemented before this mission. FR-013 therefore cannot be satisfied by
preservation alone. Close these product-code gaps in WP01-owned Go files and
tests, keeping the `hn` namespace throughout:

1. **F-1 trust-bearing event IDs**: require a full `sha256:<64-hex>` ID before
   resolving every trust-bearing command/dependency. Short IDs remain display
   conveniences only. Add positive/negative command-path tests, including a
   shallow-selection setting where a locally unambiguous prefix must still fail.
2. **F-2 merge-event repair**: when the proposal head is already contained in
   the target branch but no valid `proposal.merged` event exists, revalidate the
   exact base-policy, approvals, CI results, and accepting decisions, then append
   the missing merge event without rerunning `git merge`. Make retries
   idempotent and fail closed if evidence/policy no longer validates. Add the
   crash/retry acceptance test described by QA.
3. **F-3 bubblewrap resolution**: locate `bwrap` only through the same restricted
   system-directory set used for sandbox PATH, never ambient PATH. Test a hostile
   PATH shadow binary cannot become the sandbox guard.
4. **F-5 stale shallow sentinel**: remove or rename the obsolete WP05-era
   `errShallowRecoveryUnavailable` semantics so current recovery errors are
   accurate. Keep existing shallow recovery behavior.

Do not modify README/docs (WP02 owns F-4 and accurate documentation). Run
focused mutation-style tests, full unit suite, vet, build, gofmt, diff, and
namespace scans. The reviewer must verify these behaviors independently before
WP02 rebases and resumes.
