---
affected_files: []
cycle_number: 1
mission_slug: self-hosting-alpha-loop-01M18YXX
reproduction_command:
reviewed_at: '2026-08-30T18:06:59Z'
reviewer_agent: user
wp_id: WP01
---

## WP01 review feedback — cycle 1

### 1. Completed rotation state is never retired

`switchActiveIdentity` atomically replaces the active pointer but leaves
`.git/nh/rotation.json` in place, including on idempotent retries where the
target actor is already active. This contradicts
`contracts/identity-continuity-v0.md`, which defines `rotation.json` as present
only while a local rotation is incomplete, and the data model's optional
in-progress transaction state.

Make completion recoverable across the pointer-switch/cleanup boundary: after
the target becomes active, durably remove the matching completed transaction;
on retry, recognize an already-switched target and converge even if cleanup
was interrupted. Add behavioral tests for completion cleanup and interruption
at that boundary. The previous signer must remain active before the switch,
and a retry after the switch must finish cleanly without selecting or creating
another actor.

### 2. The tracked-secret scan is vacuous

`TestIdentityInspectionAndTrackedFilesDoNotExposePrivateKey` calls
`inIdentityTestRepository`, whose repository has an empty tracked tree, then
iterates `git ls-files`. The loop therefore examines zero files and cannot
detect a private-key encoding written into a tracked fixture or artifact.

Keep the real command-output assertion, but make the tracked-artifact check
non-vacuous: assert the scan has meaningful tracked inputs and inspect the
actual WP/project tracked changes or a deliberately tracked disposable
fixture produced through the relevant production path. The test should fail
if the implementation writes the generated private encoding into tracked
state.

## Verification already passing

- `go test ./... -run 'Test.*(Identity|Keyring|Migration|RotationState)'`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Independent pre-WP/current-binary migration preserved the actor, legacy
  bytes, actor ref before the next event, and `0700`/`0600` modes.

The out-of-map `proposal_test.go` helper adjustment is justified and scoped:
it now selects test actors through the canonical keyring instead of writing
the former legacy authority.

## WP anti-pattern checklist

1. Dead code: **PASS** — the keyring module is reached by production
   `createIdentity`/`loadIdentity` and all existing signer call sites.
2. Synthetic/vacuous fixture: **FAIL** — the tracked-secret scan iterates an
   empty tracked tree.
3. Silent empty return: **PASS** — empty results are explicit absence/list
   semantics; no swallowed-error path was found.
4. FR coverage: **FAIL** — recoverable completion does not cover the required
   pointer-switched cleanup boundary.
5. Frozen surface: **PASS** — no frozen or untouchable file was modified.
6. Locked decision: **FAIL** — completed transaction retention contradicts
   the canonical in-progress-only rotation-state contract.
7. Shared-file ownership: **PASS** — no other WP owns these changed files;
   the `proposal_test.go` exception is explicitly justified in the commit.
8. Production fragility: **N/A** — no exception-style/bare-raise production
   path was added.
