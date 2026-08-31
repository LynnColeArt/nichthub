---
affected_files: []
cycle_number: 2
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T05:48:06Z'
reviewer_agent: user
wp_id: WP04
---

# WP04 Review Feedback — Cycle 2

Verdict: **changes requested**.

Cycle 1's exact-before-lexical, production persistence/recovery, corruption-matrix, permission, ambient-secret, and hostile-content gaps are substantially remediated. Two narrower contract failures remain.

## Blocking issue 1: only the commit is bound, not the full applicability context

The new runtime-only `projectionCommit` (`memory_index.go:38-40`) is assigned after rebuild/verification (`memory_index.go:306`, `memory_index.go:597`) and checked by query (`memory_index.go:800-805`). That correctly prevents reuse across commits.

However, WP03's projection also derives applicability from `MemoryProjectionContext.Subject` and `MemoryProjectionContext.Path` (`memory_projection.go:353`, `memory_projection.go:357-360`). Neither value is retained in the verified runtime binding. `MemoryIndexQuery.Subject` and `.Path` are only applied as post-projection row filters (`memory_index.go:918-930`). Consequently, an index verified at commit C with subject/path context X can be queried at the same commit with context Y and silently reuse stale applicability classifications.

Concrete false-negative case: rebuild at commit C with a nonmatching path B, which projects a record anchored at path A as `inapplicable`; then query the same verified index at C for path A plus `Applicabilities: [applicable]`. The row's exact path matches, but the cached classification was computed for B, so the valid A result is omitted. The equivalent subject case follows from subject applicability.

Required remediation:

- Bind the complete applicability context used by WP03—at least exact `(AtCommit, Subject, Path)`—to the verified/rebuilt runtime index without adding it to canonical persisted schema.
- Reject a candidate query whose applicability-bearing context differs from the verified projection context, or redesign the internal API so such a mismatch cannot be expressed.
- Add matching/mismatching subject and path tests analogous to the commit test, including an assertion that recomputation for the second context produces the correct opposite applicability classification.

## Blocking issue 2: the 10,000-record corpus is still homogeneous

`TestMemoryIndexTenThousandRecordsPerformanceAndRecovery` now exercises the real rebuild, encoding, atomic persistence, strict load, verification, query, deletion, and two byte-identical rebuilds. That resolves the main cycle-1 defect.

But corpus construction (`memory_index_test.go:832-856`) uses one actor, one `decision` kind, the same topics, no paths, exact/applicable records, one trusted policy classification, and the default active lifecycle for all 10,000 rows. T020 explicitly requires a corpus representative of all record kinds, actors, lifecycle states, paths, topics, trust classes, and Unicode lexical terms. The current timing and recovery proof therefore does not cover the required mixed projection/query shapes.

Required remediation:

- Deterministically distribute the 10,000 records across all six record kinds, multiple actors, all lifecycle states, path/no-path anchors, varied normalized topics, and all trust classifications while retaining Unicode lexical coverage.
- Make the production-path exact-plus-lexical queries select across that mixed corpus and prove stable membership/order before deletion and after both rebuilds.
- Keep the existing complete rebuild and p95 assertions; they pass comfortably and should not be weakened.

## Anti-pattern checklist

1. **Dead code — PASS**: the planned WP04 prerequisite surface has live internal production call chains and is ready for WP05 integration.
2. **Synthetic-fixture test — PASS**: the revised 10k test invokes real rebuild, persistence, strict load, verify, query, delete, and rebuild paths through the documented injectable verified-source seam.
3. **Silent empty return — PASS**: no undocumented swallowed failure was found; invalid UTF-8 is rejected at query/record boundaries.
4. **FR coverage — FAIL**: full commit/subject/path applicability-context binding and the required mixed 10k corpus lack acceptance coverage.
5. **Frozen surface — PASS**: commit `bc4e8ef` changes only `memory_index.go` and `memory_index_test.go`.
6. **Locked decision — FAIL**: the exact requested applicability context can still diverge between verification and query at the same commit.
7. **Shared-file ownership — PASS**: only WP04-owned files changed; no overlap was found.
8. **Production fragility — N/A**: no exception-style raise path or dependency change exists; fail-loud diagnostics remain bounded.

## Independent verification evidence

- `go test ./... -run TestMemoryIndex -count=1 -v` — PASS; 10k production rebuild/persist `292.812273ms`, exact-plus-lexical p95 `46.649286ms`.
- Focused permutation/context/strict-load tests at `-count=20` — PASS.
- `go test ./... -count=1` — PASS (`46.249s`).
- `go test -race ./... -count=1` — PASS (`100.865s`).
- `go vet ./...` — PASS.
- `go build ./...` — PASS.
- Mission-base and commit-local `git diff --check` — PASS.
