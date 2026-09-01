---
work_package_id: WP01
title: Cut over runtime and executable tests
dependencies: []
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-004
- FR-005
- FR-006
- FR-007
- FR-009
- FR-013
- NFR-001
- NFR-002
- NFR-003
- NFR-005
- C-001
- C-002
- C-003
- C-005
planning_base_branch: feat/hn-hard-cutover
merge_target_branch: feat/hn-hard-cutover
branch_strategy: Planning artifacts for this mission were generated on feat/hn-hard-cutover. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hn-hard-cutover unless the human explicitly redirects the landing branch.
subtasks:
- T001
- T002
- T003
- T004
- T005
- T006
- T007
history: []
agent_profile: implementer-ivan
authoritative_surface: .
create_intent: []
execution_mode: code_change
model: ''
owned_files:
- '*.go'
- .gitignore
role: implementer
tags: []
tracker_refs: []
---

# WP01: Cut over runtime and executable tests

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Replace Hubnot's entire active runtime namespace from `nh` to `hn` across the
Go command, wire contracts, Git refs, private state, repository configuration,
environment controls, runner labels, and executable tests. This is a hard
protocol cutover: the resulting program must not alias, read, migrate, accept,
write, or suggest the legacy namespace.

## Context

Hubnot has no users whose installations require compatibility, so preserving a
second executable or a fallback reader would create ambiguity without product
value. The mission therefore treats the rename as a new protocol root, while
Git history, existing signed facts, completed mission records, and frozen
tracked `.nh/` transition inputs remain immutable evidence outside this WP.

WP01 owns all root Go sources and executable tests plus `.gitignore`. It changes
runtime behavior and proves that old state is inert. WP02 depends on this
behavior before it can add active `.hn/` configuration and update current docs
and the charter. WP03 then performs aggregate QA, the final old-governance
attestation, fresh trust bootstrap, and public smoke verification.

Apply `occurrence_map.yaml` semantically. Do not use a blind repository-wide
replacement: unrelated substrings are out of scope, and files outside the
owned surface contain deliberate history. Within production Go code there
must be no fallback root, compatibility parser, dual write, legacy environment
read, or command alias. Preserve the existing cryptographic and operational
invariants: exact-byte Ed25519 signing, per-actor append CAS, fail-closed policy
evaluation, quarantine-before-promote replication, bounded reads and logs,
keyring permission checks, and sandbox isolation.

The exact new namespace is `hn` for the executable and diagnostics, `hn/0` for
collaboration, `hn-memory/0` for memory, and `hn.pipeline/0`/`hn.policy/0` for
configuration schemas. Storage uses `refs/hn/*`, `.git/hn/*`, and `.hn/*`;
environment controls use `HN_*`; runner and temporary labels use
`hn/<version>` and `hn-*`.

Run the implementation with:

```bash
spec-kitty agent action implement WP01 --agent codex
```

### Subtask T001: Establish failing namespace-constant and CLI tests

**Purpose**

Pin the desired breaking contract with focused red tests before production
edits. The tests must distinguish a true hard cutover from a self-consistent
cosmetic rename and must fail against the pre-cutover runtime for the expected
old namespace behavior.

**Steps**

1. Inspect `main.go`, `event.go`, `memory_event.go`, `policy.go`, `ci.go`,
   `store.go`, `identity.go`, `identity_keyring.go`, and the current test helpers
   to identify existing constants and observable entry points.
2. Add a focused root test file such as `namespace_cutover_test.go`; keep it in
   package `main` so it can assert package constants and invoke existing command
   helpers without introducing exported test hooks.
3. Add a table-driven test that pins the four exact active wire identifiers:
   `hn/0`, `hn-memory/0`, `hn.pipeline/0`, and `hn.policy/0`.
4. Add assertions for representative ref constructors and path helpers so actor,
   proposal, memory, remote-tracking, quarantine, identity, and replication
   state resolve only below `refs/hn/` and `.git/hn/`.
5. Add CLI assertions for the root help/usage surface and representative error
   paths. Require `hn` command suggestions and reject any shipped `nh` alias or
   legacy recovery instruction.
6. Pin the build-artifact rule by changing `.gitignore` from the root `/nh`
   entry to `/hn`; do not add a second legacy binary rule as compatibility.
7. Where constants are private, prefer asserting observable serialized bytes,
   ref names, paths, or diagnostics over exposing new production APIs.
8. Run only the new tests and capture their expected pre-implementation
   failures. Failures should cite old values, not unrelated fixture breakage.
9. Do not weaken existing goldens to get an early green run; their coordinated
   update belongs to T006 after production behavior changes.

**Files**

- Create `namespace_cutover_test.go` (approximately 120–220 lines initially).
- Modify `.gitignore` (one exact root binary rule).

**Validation**

```bash
go test -run 'TestNamespaceCutover|TestCLI.*Namespace' -count=1 ./...
git diff -- .gitignore namespace_cutover_test.go
```

Record the intentional red result before production edits, then run
`spec-kitty agent tasks mark-status T001 --status done`.

### Subtask T002: Cut collaboration refs, event protocol, and diagnostics to `hn`

**Purpose**

Move the collaboration event store, signed envelope version, CLI spelling, and
all recovery guidance to the new namespace without changing event semantics or
the security properties of append and validation.

**Steps**

1. Change the collaboration protocol constant and validation expectations in
   `event.go` from `nh/0` to exactly `hn/0`; preserve JSON field order, canonical
   encoding, limits, actor derivation, signature coverage, and timestamp rules.
2. Update `store.go` ref constructors and every enumeration prefix from
   `refs/nh/` to `refs/hn/`, including actors, proposals, remote projections,
   proposal-code publication, and compare-and-swap update targets.
3. Rename Git reflog/commit messages and synthetic email domains where they are
   Hubnot-owned labels (`nh:` and `@nh.invalid`) to `hn:` and `@hn.invalid`.
   Preserve message structure and event ID computation inputs.
4. Update the complete command dispatcher and root help in `main.go` and
   `commands.go` so help, arity errors, subcommand routing, and examples say
   `hn`; do not retain an `nh` executable branch or alias.
5. Update collaboration diagnostics and recovery commands in `proposal.go`,
   `governance.go`, `lineage.go`, identity-facing command sources, and any other
   production Go file under the ownership glob.
6. Ensure proposal refs, decision evidence, merge recovery, issue comments,
   review commands, and status hints all use the same new command/ref root.
7. Search production `.go` files after edits for exact `nh/0`, `refs/nh`,
   `usage: nh`, quoted `nh ` commands, and Hubnot-owned `@nh.invalid` labels.
8. Do not alter event kind names, JSON property names, SHA-256 formatting,
   sequence rules, full-ID enforcement, merge-repair behavior, or QA fixes F-1
   through F-5 except where an error string contains the namespace itself.

**Files**

- Modify `event.go`, `store.go`, `main.go`, `commands.go`, `proposal.go`,
  `governance.go`, `lineage.go`, and adjacent root production `.go` files whose
  collaboration diagnostics contain the old active token.
- Expected change size: many exact constants/strings, no structural refactor.

**Validation**

```bash
go test -run 'TestNamespaceCutover|TestEvent|TestStore|TestProposal|TestGovernance|TestLineage' -count=1 ./...
rg -n 'nh/0|refs/nh|usage: nh|@nh\.invalid' --glob '*.go' --glob '!*_test.go'
go run . help
```

The production scan must be empty unless an occurrence is proven unrelated to
the namespace and recorded for T007 review. Then run
`spec-kitty agent tasks mark-status T002 --status done`.

### Subtask T003: Cut identity, keyring, and private paths to `.git/hn`

**Purpose**

Make fresh `hn` identity state cryptographically and physically independent
from legacy `.git/nh/` state while retaining permission, symlink, bounded-read,
atomic-write, rotation, and continuity safeguards.

**Steps**

1. Trace all identity and keyring path construction through `identity.go`,
   `identity_keyring.go`, and `identity_continuity.go` before changing literals.
2. Replace the active private root with `.git/hn/` for `identity.json`, the
   `identities/` keyring, rotation state, and any continuity metadata.
3. Update identity init/show/list/public/authorize/accept/rotate diagnostics to
   suggest `hn`, including the no-identity recovery message.
4. Preserve the existing 0700 directory and 0600 file checks, symlink refusal,
   maximum file sizes, strict JSON decoding, atomic rename, and fsync behavior.
5. Do not probe `.git/nh/`, copy a legacy key, infer continuity from old actor
   refs, or create a migration marker. A repo containing only the old private
   root must behave exactly like a fresh repo with no active identity.
6. Update interruption-control environment names used by identity rotation from
   `NH_*` to `HN_*`, with no fallback reads of the old names.
7. Ensure newly generated actors still derive fingerprints solely from their
   Ed25519 public keys and begin independent `hn/0` sequences.
8. Add or extend focused tests for private-root creation, absent legacy import,
   permissions, symlink rejection, interrupted rotation, and diagnostics.

**Files**

- Modify `identity.go`, `identity_keyring.go`, `identity_continuity.go` and their
  corresponding `*_test.go` files.
- Extend `namespace_cutover_test.go` if a small shared isolation fixture helps.

**Validation**

```bash
go test -run 'TestIdentity|TestKeyring|TestRotation|TestNamespaceCutover' -count=1 ./...
rg -n '\.git[/", ]+nh|filepath\.Join\([^\n]*"nh"|NH_TEST_ROTATION' identity*.go
```

The scan may match only explicit old-state isolation fixture input in tests,
never production readers. Then run
`spec-kitty agent tasks mark-status T003 --status done`.

### Subtask T004: Cut governance, CI, config, environment, and runner surfaces to `hn`

**Purpose**

Move every active trust-policy, pipeline, execution, and runner boundary to the
new config/schema/environment namespace while preserving fail-closed evidence
evaluation and sandbox hardening.

**Steps**

1. In `policy.go` and `policy_commands.go`, change the active policy schema to
   `hn.policy/0` and every lookup, diff, and diagnostic path to
   `.hn/policy.json`.
2. In `ci.go`, change the pipeline schema to `hn.pipeline/0` and pipeline/action
   lookups to `.hn/pipelines/<name>.json` and `.hn/actions/<name>`.
3. Update run request/result construction, trusted-runner evidence, platform
   labels, log headers, temp prefixes, and executable diagnostics from `nh` to
   `hn` without changing signed payload fields or their order.
4. Change all Hubnot-owned test/interruption/sandbox environment controls from
   `NH_*` to `HN_*`. Search calls to `os.Getenv`, `LookupEnv`, `Setenv`, and
   command `Env`; do not read both spellings.
5. Update `runner.go` command usage, stderr prefixes, watch/once behavior, and
   runner-version labels to `hn`.
6. Preserve bubblewrap resolution hardening, namespace isolation, capability
   dropping, clear environment, read-only system binds, no-network behavior,
   restricted PATH, timeout and log caps, tar extraction defenses, and the host
   backend's double opt-in.
7. Preserve policy digest binding, required approvals/results/accepts,
   maintainer checks, exact proposal/commit binding, rerun semantics, and the
   retroactive merge-event repair path.
8. Prove old `.nh/` config and `NH_*` controls have no production read path.
   Do not delete or modify the checked-in `.nh/` tree; it is frozen transition
   evidence outside this WP.

**Files**

- Modify `policy.go`, `policy_commands.go`, `ci.go`, `backend.go`, `runner.go`,
  `governance.go`, and their root tests.
- No tracked `.hn/` files are created here; WP02 owns that surface.

**Validation**

```bash
go test -run 'TestPolicy|TestRun|TestCI|TestRunner|TestBackend|TestMerge' -count=1 ./...
rg -n 'nh\.policy/0|nh\.pipeline/0|\.nh/|NH_[A-Z0-9_]+|\bnh runner\b' --glob '*.go' --glob '!*_test.go'
```

Any match in production is a blocker unless it is demonstrably unrelated and
classified. Then run `spec-kitty agent tasks mark-status T004 --status done`.

### Subtask T005: Cut memory, replication, quarantine, and shallow surfaces to `hn`

**Purpose**

Move the remaining distributed-state subsystems to independent `hn` refs,
wire values, local records, temporary names, and recovery commands while
retaining validation budgets and atomic promotion semantics.

**Steps**

1. In `memory_event.go`, change the signed envelope version to
   `hn-memory/0`. Rename Hubnot-owned domain-separation strings such as stream,
   policy-missing, recall-query, and recall-cursor tags from `nh-memory-*` to
   `hn-memory-*`; update exact-byte goldens later in T006.
2. Change memory refs and private index paths throughout `memory_store.go`,
   `memory_index.go`, `memory_projection.go`, `memory_commands.go`, and related
   helpers to `refs/hn/memory/*`, `.git/hn/memory/*`, and `.hn/policy.json`.
3. In `replication.go` and `quarantine.go`, change fetch/push refspecs, allowlist
   checks, private quarantine roots, remote projections, pending transaction
   records, and atomic promotion destinations to the `hn` root.
4. In `shallow.go`, change dependency ref construction, recovery anchors,
   selection records, exact tree lookups, temporary prefixes, and every recovery
   instruction to `hn`; preserve the existing fail-closed gap classification.
5. Update replication interruption variables from `NH_*` to `HN_*` and ensure
   the legacy spelling cannot trigger test-only crash seams in production.
6. Audit prefix checks carefully: every `HasPrefix`, `TrimPrefix`, ref allowlist,
   `for-each-ref`, fetch destination, push source, and `update-ref` transaction
   must agree on `refs/hn/`; partial replacement could promote into the wrong
   namespace.
7. Preserve post-download budget accounting, object/tree validation, linear
   parents, signature and identity-continuity reprojection, crash-safe pending
   anchors, and a single atomic ref transaction.
8. Preserve memory applicability, trust filtering, bounded recall, cursor
   integrity, index recovery, typed evidence, and inert-content guarantees.

**Files**

- Modify `memory_*.go`, `replication.go`, `quarantine.go`, `shallow.go`, and
  their root test files.
- Expected change size: broad literal/path edits plus focused isolation checks,
  not a subsystem redesign.

**Validation**

```bash
go test -run 'TestMemory|TestReplication|TestQuarantine|TestShallow|TestNamespaceCutover' -count=1 ./...
rg -n 'nh-memory|refs/nh|\.git[/", ]+nh|NH_TEST_REPLICATION|\bnh sync\b|\bnh replication\b' --glob '*.go' --glob '!*_test.go'
```

Production results must contain no legacy namespace input. Then run
`spec-kitty agent tasks mark-status T005 --status done`.

### Subtask T006: Update all existing executable tests and fixtures

**Purpose**

Bring the complete existing test suite into lockstep with the new active
protocol while preserving the intent and adversarial coverage of each test.

**Steps**

1. Update every root `*_test.go` active expectation from the old command, wire,
   ref, private path, config path, environment, runner, temp, and synthetic
   email namespace to its exact `hn` form.
2. Recompute exact serialized-event goldens where protocol or domain-separation
   bytes intentionally change. Never edit only the expected digest: inspect the
   complete payload and confirm field order/signature coverage remain correct.
3. Update temporary repository setup to write `.hn/policy.json` and
   `.hn/pipelines/*.json`, add those paths to Git, and inspect the corresponding
   `HEAD:.hn/...` objects.
4. Update real-Git acceptance helpers to build a binary named `hn`, report `hn`
   command failures, inspect `.git/hn`, and enumerate `refs/hn`.
5. Update interruption tests to set `HN_INTERNAL_TESTING`,
   `HN_TEST_REPLICATION_INTERRUPT_AFTER`, and
   `HN_TEST_ROTATION_INTERRUPT_AFTER` only for active behavior.
6. Preserve every test's original security assertion: ambiguity rejection,
   exact full IDs on trust-bearing commands, merge-event repair, restricted
   bubblewrap resolution, keyring permission checks, hostile tar rejection,
   quarantine atomicity, shallow recovery, and memory trust/applicability.
7. Keep legacy spellings only in the explicit isolation tests from T007. Do not
   leave ordinary fixtures using old values merely because they still compile.
8. Run the entire non-race suite once. Diagnose failures as namespace-boundary
   inconsistencies before changing assertions; do not delete coverage or make
   comparisons less precise.
9. Format all modified Go files and inspect the aggregate diff for accidental
   edits outside root Go files and `.gitignore`.

**Files**

- Modify all affected root `*_test.go` files, including event, policy,
  governance, CI, identity, memory, replication, shallow, and operational
  acceptance tests.
- Extend but do not duplicate shared helpers where a single namespace update
  serves multiple suites.

**Validation**

```bash
gofmt -w *.go
go test -count=1 ./...
gofmt -l .
git diff --check
```

The suite must be green with its original behavioral strength. Then run
`spec-kitty agent tasks mark-status T006 --status done`.

### Subtask T007: Add adversarial legacy-isolation tests and run focused verification

**Purpose**

Prove absence of compatibility behavior rather than relying only on renamed
happy-path fixtures, then perform a scoped active-source audit before WP02
builds policy and documentation on the runtime contract.

**Steps**

1. Extend `namespace_cutover_test.go` or the most appropriate real-Git test file
   with a repository that contains valid-looking legacy `refs/nh/actors/*`,
   `refs/nh/proposals/*`, and `refs/nh/memory/*` but no active `refs/hn/*`.
2. Seed legacy `.git/nh/identity.json`, keyring/replication/index files, and a
   tracked `.nh/policy.json`/pipeline. Hash or snapshot them before invoking the
   new program and assert they remain byte-for-byte and ref-for-ref unchanged.
3. Verify `hn log`, identity inspection, policy lookup, memory recall/index, and
   replication do not discover or import seeded legacy state. Expected errors
   must describe missing active `hn` state, not malformed legacy input.
4. Create a bare remote advertising both ref roots. Run `hn sync` under a valid
   active selection and prove only `refs/hn/*` is fetched, projected, promoted,
   and pushed; old refs must remain invisible even when identifiers collide.
5. Set old `NH_*` interruption variables while running active identity and
   replication operations. Assert those operations complete normally and do not
   create legacy transaction state.
6. Feed otherwise-valid envelopes/policies/pipelines carrying `nh/0`,
   `nh-memory/0`, `nh.policy/0`, and `nh.pipeline/0` to new validators and assert
   explicit unsupported-version failures.
7. Scan production Go code for semantic legacy forms, including split path
   construction and diagnostic fragments. Test-only old matches must be confined
   to named isolation cases with comments explaining why they are intentional.
8. Run focused namespace/security suites, `go vet`, `go build`, formatting, and
   diff checks. Leave the full race suite for WP03's aggregate gate unless time
   permits an early run.
9. Confirm `go.mod` remains `module hubnot` with no `require` block, no Docker or
   service was added, and no file outside this WP's ownership was modified.

**Files**

- Finalize `namespace_cutover_test.go` (target 220–420 lines total).
- Modify an existing acceptance/helper test only if required for realistic
  mixed-remote coverage; keep all changes under root `*.go`.

**Validation**

```bash
go test -run 'TestNamespaceCutover|TestLegacy.*Ignored|Test.*RejectsLegacy|TestReplication.*Namespace' -count=1 ./...
go test -count=1 ./...
go vet ./...
go build -o hn .
test -z "$(gofmt -l .)"
git diff --check
git check-ignore -v hn
```

Also run a reviewed active-source scan equivalent to:

```bash
rg -n 'refs/nh|\.git/nh|\.nh/|nh/0|nh-memory/0|nh\.pipeline/0|nh\.policy/0|NH_[A-Z0-9_]+|usage: nh|\bnh (init|sync|log|show|issue|proposal|review|decide|merge|run|runner|identity|memory|policy|replication)\b' --glob '*.go' --glob '!*_test.go'
```

The production scan must be empty. Review each test match manually, retain only
adversarial legacy inputs, then run
`spec-kitty agent tasks mark-status T007 --status done`.

## Definition of Done

- T001–T007 each have an event-sourced completion record created with
  `spec-kitty agent tasks mark-status <Txxx> --status done`.
- The only shipped command and ignored generated executable is `hn`; help,
  usage errors, and recovery diagnostics never suggest `nh`.
- Production code recognizes only the four `hn` wire versions, `refs/hn/*`,
  `.git/hn/*`, `.hn/*`, `HN_*`, and `hn` runner/temp labels.
- No alias, migration, fallback read, dual write, or legacy protocol acceptance
  exists in production code.
- Existing signed-event, governance, CI, identity, memory, replication,
  quarantine, shallow, and operational tests pass with the new contract.
- Dedicated adversarial tests prove legacy refs, config, private state,
  environment variables, and wire versions are ignored or rejected and never
  mutated.
- `go test -count=1 ./...`, `go vet ./...`, `go build -o hn .`, `gofmt -l .`,
  and `git diff --check` complete cleanly.
- `go.mod` remains the zero-third-party-dependency `hubnot` module; no Docker,
  service, or provider-specific dependency is introduced.
- The checked-in `.nh/` transition files and all files outside this WP's owned
  surface remain unchanged.

## Risks

- **Self-consistent false green**: Mechanically renaming production and tests
  can conceal fallback behavior. Mitigate with seeded old-only and mixed-remote
  isolation tests that snapshot legacy state before and after execution.
- **Partial ref cutover**: One stale prefix in fetch, projection, validation, or
  atomic promotion could cross trust namespaces. Audit every constructor,
  prefix check, refspec, and transaction destination as a connected flow.
- **Signed-byte drift beyond intent**: Protocol/domain tag changes alter event
  IDs by design, but field order or validation must not. Review full goldens and
  preserve exact-byte signing tests.
- **Private-key exposure**: Tests use temporary identities only. Never copy a
  real `.git/nh` or `.git/hn` keyring into source, fixtures, logs, or diffs.
- **Historical evidence damage**: The broad edit is limited to root Go files
  and `.gitignore`. Do not touch `.nh/`, existing refs, journals, reports, or
  completed mission artifacts.
- **Security regression hidden by namespace churn**: Preserve test bodies and
  assertions wherever possible, rerun focused security suites, and treat any
  weakening or deletion as a review blocker.

## Reviewer Guidance

Review the change as a security boundary, not a spelling patch. Trace one actor
event, proposal, CI result, memory event, selective replication transaction,
and shallow recovery path from construction through validation and storage; all
must remain entirely inside the `hn` namespace. Confirm production code has no
legacy read or dual-write branch and that old values appear in tests only as
hostile isolation input.

Pay special attention to exact refspecs and `HasPrefix`/`TrimPrefix` pairs,
private-path permission checks, environment-gated interruption seams, policy
and pipeline tree lookups, memory domain-separation strings, and recovery
diagnostics. Re-run the focused isolation tests and inspect the snapshots that
prove legacy files and refs were not mutated. Finally, verify the prior QA fixes
remain covered: full trust-bearing IDs, merge-event repair, restricted
`bwrap` resolution, documented rerun behavior, and current shallow recovery
semantics.
