---
affected_files: []
cycle_number: 1
mission_slug: hn-hard-cutover-01M1EY1B
reproduction_command:
reviewed_at: '2026-09-01T17:28:47Z'
reviewer_agent: user
wp_id: WP01
---

# WP01 Review Feedback 1

Verdict: changes requested. Commit `17eb2a9` is mechanically disciplined and
passes its automated gates, but two hard-cutover requirements are not yet met.

## Issue 1 — Active merge commits still carry the legacy `NH` label (blocking)

`governance.go:191` still constructs merge subjects as
`Merge NH proposal ...`. This is active, newly produced user-visible protocol
metadata, not frozen history. It contradicts FR-001, SC-001, the occurrence
map's user-facing/log treatment, and WP01's objective that the runtime neither
write nor suggest the legacy namespace.

Fix the active subject to `Merge HN proposal ...` and add a focused assertion
covering the generated merge subject (or an equivalent production-label
test). Expand the final active-source audit to include case-insensitive exact
token forms such as `\bNH\b`; the current prescribed lowercase scan does not
catch this occurrence.

## Issue 2 — Legacy isolation stops before the required sync/storage boundary (blocking)

`TestReplicationNamespaceIgnoresLegacyRemoteAdvertisements` exercises
`resolveReplicationRequests` only. It does not run `hn sync` or the production
quarantine/promote/publish transaction, and therefore does not prove T007 step
4's contract that a mixed remote cannot cause legacy refs to be fetched,
projected, promoted, or pushed. Likewise, the old-only fixture seeds a single
`.git/nh/identity.json`, but not the keyring, replication records, and memory
index required by T007 steps 2–3, nor does it exercise the identity, memory
index/recall, and replication command paths against snapshots of that state.

Keep the existing constructor-level tests—they are useful—and extend them with
a real-Git mixed-remote test that runs the sync transaction with valid active
selection/state, then asserts only `refs/hn/*` local/remote projections and
publications changed while colliding `refs/nh/*` stayed byte/ref identical.
Seed representative legacy keyring, replication, and memory-index files under
`.git/nh/`, snapshot them before the relevant command paths, and assert the
entire legacy tree remains unchanged afterward. This must exercise production
commands/transactions rather than only path helpers.

## Verification evidence

- `go test -count=1 ./...`: PASS (`ok hubnot 68.691s`)
- focused namespace/security suite: PASS
- `go vet ./...`, `go build -o hn .`, formatting, and `git diff --check`: PASS
- `.gitignore` manual review: PASS; `hn` is ignored while `nh` and `hubnot`
  are visible, matching the one-executable cutover
- production legacy scan: lowercase ref/path/version/env/command forms are
  absent; case-insensitive scan found the issue above
- frozen `.nh/policy.json` and `.nh/pipelines/test.json`: byte-identical SHA-256
  hashes to the WP base
- changed paths: PASS; only owned root `*.go` files and `.gitignore`
- mutation check: reverting the identity root, policy root, or actor-ref
  enumeration in isolated temporary copies independently caused the matching
  legacy-isolation test to fail, so those tests are not synthetic fixtures
- QA F-1 through F-5 paths retain their base logic; this WP introduces only
  namespace-literal changes on those paths, aside from the stale label above

## Anti-pattern checklist

1. Dead code: **PASS** — no new production public function, type, or module.
2. Synthetic-fixture test: **PASS** for implemented coverage — tests invoke
   production readers/validators and the mutation check failed as expected;
   missing boundary breadth is separately blocking under Issue 2.
3. Silent empty return: **N/A** — no new semantic return path was introduced.
4. FR coverage: **FAIL** — the explicit mixed-remote sync and complete legacy
   private-state isolation behaviors are not exercised end to end.
5. Frozen surface: **PASS** — frozen paths are absent from the diff and the
   tracked `.nh/` hashes match the base.
6. Locked decision: **FAIL** — `Merge NH proposal` contradicts the hard-cutover
   MUST/MUST NOT clauses.
7. Shared-file ownership: **PASS** — all changes are inside WP01 ownership; no
   WP02/WP03 surface is modified.
8. Production fragility: **N/A** — no new control-flow/error path was added.
