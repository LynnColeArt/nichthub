---
work_package_id: WP03
title: Prove, govern, and publish the hard cutover
dependencies:
- WP01
- WP02
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-004
- FR-005
- FR-006
- FR-007
- FR-008
- FR-009
- FR-010
- FR-011
- FR-012
- FR-013
- NFR-001
- NFR-002
- NFR-003
- NFR-004
- NFR-005
- NFR-006
- C-001
- C-002
- C-003
- C-004
- C-005
planning_base_branch: feat/hn-hard-cutover
merge_target_branch: feat/hn-hard-cutover
branch_strategy: Planning artifacts for this mission were generated on feat/hn-hard-cutover. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hn-hard-cutover unless the human explicitly redirects the landing branch.
subtasks:
- T015
- T016
- T017
- T018
- T019
- T020
- T021
history: []
agent_profile: reviewer-renata
authoritative_surface: kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md
create_intent:
- kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md
execution_mode: planning_artifact
model: ''
owned_files:
- kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md
role: reviewer
tags: []
tracker_refs: []
---

# WP03: Prove, govern, and publish the hard cutover

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `reviewer-renata`
- **Role**: `reviewer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Prove the complete `nh` to `hn` hard cutover, land the exact frozen candidate
through the final legacy governance loop, publish independently rooted `hn`
facts, and leave a precise recovery-grade verification ledger. No source or
active configuration may change after legacy attestation; later edits are
limited to this mission evidence file and must be identified as post-attestation
reporting rather than part of the governed candidate.

Implement this work package with:

```sh
spec-kitty agent action implement WP03 --agent codex
```

## Context

WP01 changes the executable and all runtime/test namespace boundaries. WP02
adds active `.hn/` configuration with fresh actor fingerprints and updates the
current documentation and charter. This package is the final dependency gate:
it verifies those outputs, exercises disposable real-Git scenarios, freezes one
candidate commit, uses a separately built pre-cutover `nh` binary and legacy
actors to attest that commit, and only then establishes public `refs/hn/*`.

The two namespace eras have different authority. Legacy `.nh/`, `.git/nh/`,
and `refs/nh/*` exist only to attest the transition and preserve provenance.
The new binary must never consult them. Conversely, the final legacy governance
loop must use the frozen `.nh/pipelines/test.json` and pre-cutover executable;
do not accidentally use the candidate `hn` binary to manufacture old facts.

External mutations in this package are append-only and deliberate. Never force
push, rewrite an actor ref, amend an attested candidate, copy a private key into
the repository, or expose keyring contents in logs. Perform destructive or
remote-changing steps only after resolving exact commits, actors, remotes, and
ref targets with read-only commands. A failed gate stops publication.

### Subtask T015: Run build, vet, format, diff, and unit gates

**Purpose**: Establish a clean, reproducible static and unit-test baseline for
the exact candidate before any governance or public ref mutation occurs.

**Steps**:

1. Confirm the worktree and branch are the expected cutover candidate:
   - Record `git status --short`, `git branch --show-current`,
     `git rev-parse HEAD`, and `git remote -v`.
   - Require the source/config/doc changes from WP01 and WP02 to be committed.
   - Treat an unexpected staged file, unrelated edit, or wrong remote as a stop
     condition; do not clean or discard user work.
2. Record the toolchain with `go version` and `git --version`, then run:

   ```sh
   go build ./...
   go vet ./...
   test -z "$(gofmt -l .)"
   git diff --check
   go test -count=1 ./...
   ```

3. Build the named release candidate as an ignored local artifact with
   `go build -trimpath -o hn .`. Run `./hn help` and at least one invalid-command
   path, confirming diagnostics and recovery instructions say `hn`, not `nh`.
4. Capture each command, exit status, elapsed time where useful, and concise
   output in working notes for T021. Do not paste private paths, environment
   secrets, key JSON, or full repetitive test output into the report.
5. If formatting is dirty, return the work to its owning package instead of
   using this evidence-only WP to edit `*.go`, docs, charter, or `.hn/**`.

**Files**: No tracked source files. Record evidence later in
`kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md` (~30 lines).

**Validation**: Every command exits zero, `gofmt -l .` emits nothing,
`git diff --check` emits nothing, the built executable is `hn`, and the initial
candidate commit and toolchain are recorded.

### Subtask T016: Run full race, security, and acceptance gates

**Purpose**: Reconfirm the security-critical behavior under concurrency and
real Git operations after the namespace reset, not merely after a textual rename.

**Steps**:

1. Run the complete suite under the race detector without using cached results:

   ```sh
   go test -race -count=1 ./...
   ```

2. Confirm the output includes the real-repository acceptance scenarios and
   the tests covering exact-byte Ed25519 signing, actor-chain CAS behavior,
   policy/decision validation, keyring permissions, sandbox execution,
   quarantine-before-promote replication, selective replication, shallow
   recovery, and agent memory. Use `go test -list .` or focused `-run` patterns
   only to demonstrate coverage; do not substitute focused runs for the full
   race suite.
3. Explicitly inspect the QA regression tests associated with findings F-1
   through F-5:
   - full IDs remain mandatory on trust-bearing commands;
   - a merged-code/missing-event state has a repair path;
   - Bubblewrap resolution is restricted to trusted system directories;
   - `run execute --rerun` remains documented and accepted;
   - stale shallow-recovery semantics do not reappear.
4. Confirm tests do not silently select a host execution backend or depend on
   Docker. Any host-backend acceptance must retain its explicit double opt-in.
5. Preserve the final package summary and duration for T021. A race report,
   timeout, flaky acceptance result, or skipped required backend check blocks
   candidate freezing until diagnosed and rerun cleanly.

**Files**: No tracked source files. Add a summarized gate table to
`namespace-verification.md` (~25 lines) after the public transition.

**Validation**: `go test -race -count=1 ./...` exits zero with no race report,
all security/acceptance domains are demonstrably represented, and QA F-1 through
F-5 remain resolved.

### Subtask T017: Verify namespace completeness, frozen evidence, and dependencies

**Purpose**: Prove that every surviving legacy token is an intentional historical
exception and that the cutover adds no service, container, or library dependency.

**Steps**:

1. Run a token-aware repository scan, excluding `.git/**`, for at least:
   `refs/nh/`, `.git/nh/`, `.nh/`, `nh/0`, `nh-memory/0`, `nh.pipeline/0`,
   `nh.policy/0`, `NH_*`, standalone command `nh`, and `NH-[0-9]{3}`.
2. Classify every result against
   `kitty-specs/hn-hard-cutover-01M1EY1B/occurrence_map.yaml`:
   - active Go, current docs, active charter, `.hn/**`, and `.gitignore` may not
     contain an obsolete active spelling;
   - `.nh/**`, completed mission records, immutable journals, and historical
     self-hosting evidence retain their original bytes;
   - manual-review paths must distinguish current guidance from dated evidence.
3. Search separately for each new boundary (`refs/hn/`, `.git/hn/`, `.hn/`,
   `hn/0`, `hn-memory/0`, `hn.pipeline/0`, `hn.policy/0`, `HN_*`, and `HN-###`)
   and confirm it appears in the intended implementation/test/document surface.
4. Compare `.nh/policy.json` and `.nh/pipelines/test.json` against the public
   base using both Git object identity and SHA-256 digests. Require no diff in
   `.nh/**`; record exact digests for later recovery checks.
5. Inspect all legacy public actor/proposal refs read-only before transition
   with `git for-each-ref` and `git ls-remote`. Record their tips so T019 can
   demonstrate append-only advancement rather than replacement.
6. Verify `go.mod` still says `module hubnot`, has no third-party `require`
   entries, and `go list -m -mod=readonly all` lists no external module. Confirm
   there is no new Docker, daemon, or hosted-service prerequisite in build/test
   instructions.

**Files**: No tracked source files. Add the occurrence classification, frozen
digests, dependency result, and pre-transition ref tips to
`namespace-verification.md` (~45 lines).

**Validation**: The active-tree scan has zero unclassified legacy occurrences,
all declared frozen files are byte-identical, old ref tips are recorded, and the
module/deployment boundary remains standard-library Go plus ordinary Git.

### Subtask T018: Execute fresh local and two-party `hn` smoke tests

**Purpose**: Exercise the new namespace in real disposable repositories before
publishing it, including explicit proof that legacy state is inert.

**Steps**:

1. Create disposable directories with `mktemp -d`; record the resolved paths
   only in temporary notes and arrange cleanup after evidence is captured. Copy
   the candidate `hn` binary into the test area rather than installing it over
   any existing executable.
2. In a fresh local Git repository, configure a synthetic name/email and run:
   `hn init`, `hn issue open`, `hn log`, `hn show`, and representative invalid
   input. Inspect event objects and `git for-each-ref refs/hn` to confirm the
   collaboration wire is `hn/0`, private state is below `.git/hn/`, and no
   `refs/nh/*` is created.
3. In a second disposable repository, seed sentinel `.git/nh/` files and a
   valid legacy ref before invoking `hn`. Record sentinel hashes and ref tips.
   Run identity/log/sync operations with an `NH_*` test variable set, then prove:
   - old files and refs are byte-for-byte unchanged;
   - old facts are absent from active projections;
   - only `.git/hn/`, `refs/hn/*`, and `HN_*` affect Hubnot behavior.
4. Create a disposable bare remote and two working clones. Initialize distinct
   `hn` actors, place their public fingerprints into a disposable `.hn/policy.json`,
   and commit a disposable policy/pipeline using the new wire identifiers.
5. Exercise the documented two-party path as far as the policy supports:
   proposal, `run request`, sync, trusted execution, independent review, status,
   maintainer decision, merge, and sync. Use full event IDs on every trust-bearing
   command and inspect both clones' `refs/hn/*` projections.
6. Keep this smoke remote separate from the configured public `origin`. Do not
   publish any new public `refs/hn/*` yet; FR-011 requires the legacy transition
   to land first. Never copy test or production private key files into logs.

**Files**: Disposable repositories and binaries only; summarize actors with
public fingerprints and facts with full IDs in `namespace-verification.md`
(~35 lines).

**Validation**: Both smoke scenarios succeed using only `hn`; legacy sentinels
remain unchanged; two independent actors exchange and validate only
`refs/hn/*`; the disposable remote is demonstrably not the public remote.

### Subtask T019: Freeze and govern the exact candidate with legacy `nh`

**Purpose**: Create the final old-protocol attestation that names the exact
first `hn` candidate without introducing compatibility behavior into that candidate.

**Steps**:

1. Finish all candidate-affecting work before freezing. Require a clean status,
   rerun the cheap T015 gates, resolve `candidate=$(git rev-parse HEAD)`, and
   record its tree ID. From this point forward, do not amend, rebase, or modify
   source, tests, `.hn/**`, `.nh/**`, README, current docs, or active charter.
2. Build a pre-cutover executable from the exact public base in a separate
   temporary worktree or clean clone. Name it `nh`, record the base commit and
   binary SHA-256, and verify its help reports the legacy namespace. Never build
   the legacy attestor from candidate sources.
3. Reconfirm the old maintainer/reviewer actor fingerprints, private key
   availability, frozen `.nh/policy.json`, frozen `.nh/pipelines/test.json`,
   public remote URL, target branch, and current public main tip before writing
   any event.
4. With the legacy binary, open or revise a proposal whose head is the exact
   candidate commit and whose base is the current public `main`. Record the full
   proposal event ID and synchronize append-only `refs/nh/*`.
5. Request the frozen legacy `test` pipeline for that full proposal ID. On the
   independent legacy actor, sync, execute the request in the trusted sandbox,
   review the exact proposal, and publish the full run-result and review IDs.
6. Back on the maintainer actor, sync and inspect status. Accept only if policy,
   approval, and exact CI evidence validate against the candidate. Record the
   decision ID, run `nh merge` for the full proposal ID, and record both the
   merge event ID and resulting Git merge commit.
7. If the candidate changes or any evidence binds another commit/policy digest,
   stop. Create a proposal revision and repeat CI/review/decision; never reuse
   stale evidence. If code merges but event recording fails, use the implemented
   reconciliation path and record the recovery instead of fabricating a fact.

**Files**: External append-only `refs/nh/*` and a public-branch merge; no source
file edits. Exact identifiers are reported later in `namespace-verification.md`.

**Validation**: The legacy proposal names the frozen candidate, all trust facts
use full IDs and old actors, the merge event validates, old ref histories only
advance by ancestry, and candidate source/tree content remains unchanged.

### Subtask T020: Publish main, establish `refs/hn/*`, and validate a fresh clone

**Purpose**: Complete the public namespace boundary only after the old protocol
has attested and merged the candidate, then demonstrate ordinary-Git discovery.

**Steps**:

1. Inspect the post-merge local branch, merge commit, configured upstream, and
   remote main tip. Push the legacy-governed merge and legacy collaboration refs
   through the normal commands. Never use `--force`, a leading `+` refspec, ref
   deletion, or a non-fast-forward update.
2. Verify with `git ls-remote` that public `main` is the expected merge commit
   and that the recorded final `refs/nh/*` tips are present. Confirm the merged
   tree contains the exact candidate tree at the expected lineage and that no
   ungoverned source/config change intervened.
3. Using the new `hn` binary and the fresh maintainer/reviewer private identities
   established for `.hn/policy.json`, create the minimal documented signed facts
   needed to establish both new actors. Synchronize only `refs/hn/*` to the
   public remote and record full event IDs and resulting ref tips.
4. Have the second actor sync from the public remote and validate the new facts.
   Confirm active policy recognizes only the fresh actors and all emitted wire,
   runner, and ref values use `hn` forms.
5. Clone the public repository into a new temporary directory with no copied
   private state. Build `hn`, run the documented read/sync/log flow, and inspect
   `git for-each-ref refs/hn` plus `git ls-remote origin 'refs/hn/*'`.
6. Prove the fresh clone can see and validate new facts while `hn log` and trust
   projections ignore the still-advertised historical `refs/nh/*`. Verify the
   clone creates no `.git/nh/` state and requires neither Docker nor a Hubnot
   server.
7. If publication is partial or a remote tip is unexpected, stop further writes,
   capture local/remote tips, and diagnose. Retry idempotent sync where safe;
   never repair an append-only history by force-pushing or deleting evidence.

**Files**: Public `main`, append-only public `refs/nh/*` final evidence, new
append-only `refs/hn/*`, and disposable fresh-clone state. No tracked source edits.

**Validation**: Public main resolves to the expected governed merge, both ref
eras are enumerable, only fresh `refs/hn/*` affect the new runtime, two new
actors validate, and a clean clone reproduces the documented flow over Git alone.

### Subtask T021: Author the namespace verification and recovery ledger

**Purpose**: Leave one concise, auditable record connecting requirements,
candidate identity, legacy governance, public publication, new trust roots, and
failure recovery without exposing private key material.

**Steps**:

1. Create
   `kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md`. Clearly mark
   which evidence was captured before candidate freeze and which external IDs
   were added after publication. State that post-attestation reporting did not
   change the governed source/config tree.
2. Record exact Git identities:
   - public base, frozen candidate commit and tree, legacy merge commit, and
     final public main commit;
   - old and new actor fingerprints;
   - before/after tips for every changed `refs/nh/*` and `refs/hn/*` ref.
3. Record full `sha256:<64-hex>` IDs for the legacy proposal, run request,
   trusted run result, independent review, maintainer decision, and merge event,
   plus the initial new `hn` facts. Do not use display prefixes in the ledger.
4. Include command/result tables for build, vet, formatting, diff, unit, race,
   security/acceptance, occurrence audit, dependency inspection, disposable
   local smoke, two-party smoke, public sync, and fresh-clone validation.
5. Include frozen `.nh/**` SHA-256 digests, the legacy binary base and digest,
   active `.hn/**` wire values, the public remote URL, toolchain versions, and
   relevant timestamps/timezone. Link conclusions to FR/NFR/constraint IDs.
6. Document recovery notes for candidate drift, stale evidence, code-merged but
   event-missing reconciliation, interrupted sync, unexpected remote tips,
   unavailable keys, and fresh-clone disagreement. Recovery must preserve refs
   append-only and must never recommend force push, history rewrite, dual writes,
   fallback reads, or private-key copying.
7. Review the report for secret leakage and internal inconsistency. Private keys,
   raw keyring JSON, authorization tokens, browser codes, and unnecessary local
   filesystem details must not appear. Every claim should be reproducible from
   an exact commit, full event ID, ref tip, digest, or command result.

**Files**:
`kitty-specs/hn-hard-cutover-01M1EY1B/namespace-verification.md` (new, approximately
180–300 lines). This is the only tracked file owned by WP03.

**Validation**: The report names every exact commit/event/ref/digest required to
audit the transition, distinguishes governed candidate bytes from later evidence
reporting, contains no secrets, and gives safe recovery instructions for every
identified publication failure mode.

## Definition of Done

- T015 has an event-sourced completion record after all static and unit gates
  pass for the recorded candidate baseline.
- T016 has an event-sourced completion record after the full uncached race suite
  and security/acceptance checks pass.
- T017 has an event-sourced completion record after zero unclassified active
  legacy occurrences, frozen evidence, and zero third-party dependencies are proven.
- T018 has an event-sourced completion record after both disposable local and
  two-party smoke tests prove new behavior and old-state isolation.
- T019 has an event-sourced completion record after full-ID legacy evidence
  attests and merges the exact frozen candidate without rewriting refs.
- T020 has an event-sourced completion record after public main, new `refs/hn/*`,
  two fresh actors, and a clean public clone all validate.
- T021 has an event-sourced completion record after `namespace-verification.md`
  contains exact evidence and recovery guidance with no secret material.
- Record each completion with
  `spec-kitty agent tasks mark-status Txxx --status done`; Markdown checkboxes
  are not completion evidence.
- No source, test, active configuration, README/current docs, or charter file was
  changed after the candidate was frozen and legacy-attested.
- No force push, ref deletion, history rewrite, compatibility alias, fallback
  read, dual write, private-key copy, Docker requirement, or service dependency
  was introduced.

## Risks

- **Self-invalidating evidence**: Editing candidate-owned source/config after the
  proposal makes prior CI and review stale. Freeze a commit and repeat the whole
  legacy evidence chain after any candidate change.
- **Attestor confusion**: Running candidate `hn` for the legacy loop can create
  facts in the wrong namespace. Build and hash the old binary from the recorded
  public base and keep the two binaries at explicit absolute temporary paths.
- **Remote ref damage**: A force update would destroy the append-only audit
  property. Resolve before/after tips and allow only fast-forward ref updates.
- **Premature new trust root**: Publishing `refs/hn/*` before the legacy merge
  reverses the intended attestation order. Keep all pre-merge `hn` smoke remotes
  disposable and verify the public remote URL before each sync.
- **Secret exposure**: Verbose diagnostics can reveal private state. Record
  fingerprints, event IDs, hashes, and modes only; never record private key bytes.
- **False-positive namespace scans**: Blind substring scans can rewrite history
  or unrelated words. Apply the exact token families and exception paths in
  `occurrence_map.yaml`, then manually review every residual.
- **Evidence-report timing**: Final external IDs exist only after publication.
  Treat this WP-owned report as post-attestation mission evidence and explicitly
  state that it is not part of the frozen source/config candidate.

## Reviewer Guidance

Reviewers should independently resolve the candidate, merge, event, actor, and
ref identifiers rather than trusting prose. Confirm that the old proposal and
every old trust fact bind the exact frozen candidate and frozen policy digest;
then confirm public `refs/hn/*` begin only after the legacy merge.

Re-run the namespace scan against `occurrence_map.yaml` and inspect the explicit
legacy-isolation tests. Pay special attention to exact refspecs, private-state
paths, strict wire versions, environment reads, runner labels, command recovery
text, and the active charter IDs. Historical exceptions must remain inert and
byte-preserving, not become an undocumented compatibility layer.

Finally, audit external mutation safety: all changed actor refs must advance by
ancestry, public main must be a normal fast-forward publication of the governed
merge, no force/delete refspec may appear, and report contents must contain no
private key or credential material. Reject the WP if any reported identifier is
abbreviated, any gate was run against a different commit, or evidence cannot be
reproduced from ordinary Git commands.
