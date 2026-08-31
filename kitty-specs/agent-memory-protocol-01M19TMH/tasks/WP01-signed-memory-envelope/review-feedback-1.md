# WP01 Review Feedback — Round 1

**Verdict:** Changes requested.

**Issue 1 — Leading traversal paths are accepted (blocking correctness/security defect).**

`validMemoryPath` relies on `path.Clean(value) == value` to reject traversal. That catches collapsible internal forms such as `docs/../README.md`, but it does not reject an already-clean leading traversal such as `..` or `../secret`: `path.Clean("../secret")` is still `../secret`. Those values satisfy every other current predicate and are accepted. This violates T003 step 8 and the hostile-input boundary requiring repository-relative paths with no traversal.

Fix by explicitly rejecting `..` as a complete path and `../` as the leading path segment (or equivalently rejecting any segment equal to `..`) in addition to the normalization check. Add table-driven production-path tests for at least `..`, `../secret`, `a/../../secret`, and valid names that merely contain two dots such as `docs/file..md`.

**Issue 2 — The aggregate handoff-byte regression test does not reach the aggregate limit branch (blocking test-fidelity defect).**

`TestMemoryPathAndHandoffBoundaries` distributes the aggregate handoff totals across exactly 16 entries. At 65,537 bytes, at least one entry becomes 4,097 bytes, so validation fails on `maxMemoryHandoffEntryBytes` before it can test `maxMemoryHandoffBytes`. Removing the aggregate total check would therefore leave the purported aggregate one-above test green. This fails the WP's exact one-below/exact/one-above requirement and the synthetic-fixture anti-pattern check.

Fix the fixture so the aggregate cases use enough entries (for example 17 or more, within the 64-entry limit) that every entry remains at or below 4,096 bytes for totals 65,535, 65,536, and 65,537. Assert that the one-above rejection is caused solely by the aggregate bound. While correcting boundary fidelity, make the multibyte content cases hit exact byte totals 65,535, 65,536, and 65,537 by mixing multibyte and single-byte characters; the current pure two-byte-rune cases exercise 65,534/65,536/65,538 rather than the specified adjacent byte boundaries.

## Review evidence

- `go test ./... -count=1 -run 'TestMemory(StrictDecode|EventValidation|OperationShapes|Kinds|ContentTopicEvidenceBoundaries|PathAndHandoffBoundaries|SignVerify|EverySignedFieldTamper|IDAndDefaultStream)|TestIdentityFieldsPreserve'` — passed.
- `go test -race ./...` — passed.
- `go vet ./...` — passed.
- `go build ./...` — passed.
- `git diff --check` — passed.
- Independent default-stream derivation over the required NUL-delimited bytes produced `d5c42b67775cd9520bb0089b495230aad66693ef5385fe3ab766a2d5afec2076`, matching the golden.

## WP anti-pattern checklist

1. Dead code — **PASS**: this file is a planned package-level foundation, creates no separate Go module, and exposes no new public functions; its internal call graph is live under the package and intended for dependent WPs.
2. Synthetic-fixture test — **FAIL**: the aggregate handoff one-above case is masked by the per-entry bound (Issue 2).
3. Silent empty return — **PASS**: no silent empty-return failure paths were found.
4. FR coverage — **PASS for WP01's bounded wire slice**: tests invoke production encoding, validation, signing, verification, IDs, compatibility, and inert-data paths; later projection/recall behavior remains assigned to dependent WPs.
5. Frozen surface — **PASS**: implementation commit `5b01cbf` adds only `memory_event.go` and `memory_event_test.go`; legacy event fixtures are untouched.
6. Locked decision — **FAIL**: leading traversal contradicts the required no-traversal anchor rule (Issue 1).
7. Shared-file ownership — **PASS**: the implementation commit changes only WP01-owned files.
8. Production fragility — **PASS**: new fail-loud validation errors are bounded and justified at the hostile wire boundary; no bare panic/raise path was added.
