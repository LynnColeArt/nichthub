---
affected_files: []
cycle_number: 1
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T05:24:57Z'
reviewer_agent: user
wp_id: WP04
---

# WP04 Review Feedback — Cycle 1

Verdict: **changes requested**.

## Blocking issue 1: query context is accepted but not enforced, and lexical matching runs before exact filters

`MemoryIndexQuery.AtCommit` is validated in `memory_index.go:822-825`, but it is never consumed by `queryMemoryIndexV0` or `memoryIndexRecordMatches` (`memory_index.go:785-917`). A caller can therefore present a different valid commit than the commit used to project/verify the index and receive candidates without a context-mismatch error. The contract requires recall applicability to be evaluated for the exact requested commit; an inert trust-bearing field is unsafe.

The same query path computes the global lexical posting intersection at `memory_index.go:792-800` and only then applies exact filters at `memory_index.go:802`. This reverses the frozen recall order in `plan.md` and the WP prompt: exact filters must precede lexical matching.

Required remediation:

- Bind the candidate query to the exact `MemoryProjectionContext` used by rebuild/verify, or remove `AtCommit` from this internal query surface and make a verified context-bound index/candidate API impossible to misuse.
- Apply exact filters before lexical intersection, with deterministic results and no second public recall envelope.
- Add tests with identical refs/policy under two requested commits that prove a context mismatch cannot silently reuse candidates; cover the `Applicabilities` dimension and exact-before-lexical staging.

## Blocking issue 2: the 10,000-record acceptance test bypasses the real rebuild/recovery path

`TestMemoryIndexTenThousandRecordsPerformanceAndRecovery` (`memory_index_test.go:472-524`) calls `ProjectMemories` and `buildMemoryIndexV0` directly. It never calls `rebuildMemoryIndexV0`, persists the index, strictly reloads it, calls `verifyMemoryIndexV0`, deletes it, or performs either required clean rebuild. Its “rebuild” timing therefore excludes source collection, deterministic encoding, atomic persistence, permission handling, strict load, and verification. The test name claims recovery that it does not exercise.

Required remediation:

- Drive the 10,000-record corpus through the real rebuild entry point and private file.
- Measure the complete clean rebuild against 30 seconds and indexed exact-plus-lexical query against the 1-second p95 target.
- Strictly load/verify the persisted index, delete it, rebuild twice, and assert byte-identical bytes plus identical membership/order after each rebuild.

## Blocking issue 3: required isolation and failure-recovery evidence is absent

The suite verifies normal `0600`/`0700` modes and one injected writer error, but it does not cover the T020 acceptance matrix: permission-denied directory/file/atomic-replace cases, absence of readable partial output, ambient secret/token/environment/transcript sentinel exclusion, or hostile instruction-like content with no execution/network/ref/policy/authorization side effect. The corruption matrix also mutates only four shapes (`memory_index_test.go:371-399`), not every persisted field family required by T018 (membership, anchor/path, lifecycle edges, evidence details, dependencies, trust, inert data, and tokens).

Required remediation:

- Add deterministic failure fixtures for directory creation/validation and atomic replacement; assert bounded diagnostics, no partial valid-looking cache, and unchanged canonical refs/policy.
- Add ambient sentinel and hostile-content fixtures and inspect both index bytes and returned diagnostics/observable repository state.
- Expand strict verification mutations across every persisted field family and assert the cache always loses disagreements with verified sources.

## Anti-pattern checklist

1. **Dead code — PASS**: WP04 is an explicitly planned prerequisite surface for WP05; its internal rebuild, verify, build, validation, and query paths have production call chains within the owned module.
2. **Synthetic-fixture test — FAIL**: the 10k acceptance fixture bypasses the production rebuild/persist/load/verify/recovery path.
3. **Silent empty return — PASS**: the invalid-UTF-8 tokenizer empty result is guarded by query and record validation; no undocumented swallowed error was found.
4. **FR coverage — FAIL**: requested-commit/applicability behavior and T020 recovery/isolation requirements lack behavioral assertions.
5. **Frozen surface — PASS**: commit `ee9dc07` changes only `memory_index.go` and `memory_index_test.go`; no frozen mission artifact was modified.
6. **Locked decision — FAIL**: lexical matching precedes exact filtering, contrary to the frozen recall flow; `AtCommit` is accepted without enforcement.
7. **Shared-file ownership — PASS**: the implementation commit changes only the two WP04-owned files; no ownership overlap was found.
8. **Production fragility — N/A**: no exception-style `raise` path exists in this Go change; fail-loud errors are bounded and no new dependency was introduced.

## Independent verification evidence

- `go test ./... -run TestMemoryIndex -count=1 -v` — PASS; reported 10k build ~85 ms and query p100 ~40 ms, subject to the coverage limitation above.
- `go test ./... -run 'TestMemoryIndex(BuildIsByteIdenticalAcrossInputPermutations|ExactAndUnicodeLexicalQuery|StrictLoad)' -count=20` — PASS.
- `go test ./... -count=1` — PASS (`45.424s`).
- `go test -race ./... -count=1` — PASS (`72.264s`).
- `go vet ./...` — PASS.
- `go build ./...` — PASS.
- `git diff --check kitty/mission-agent-memory-protocol-01M19TMH..HEAD` and `git diff --check ee9dc07^..ee9dc07` — PASS.
