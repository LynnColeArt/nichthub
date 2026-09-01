---
work_package_id: WP03
title: Prove compatibility and public cutover
dependencies:
- WP01
- WP02
requirement_refs:
- FR-004
- FR-007
- FR-008
- NFR-001
- NFR-002
- NFR-003
- NFR-004
- C-001
- C-002
- C-003
planning_base_branch: feat/hubnot-product-rename
merge_target_branch: feat/hubnot-product-rename
branch_strategy: Planning artifacts for this mission were generated on feat/hubnot-product-rename. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hubnot-product-rename unless the human explicitly redirects the landing branch.
subtasks:
- T009
- T010
- T011
- T012
history: []
agent_profile: researcher-robbie
authoritative_surface: kitty-specs/hubnot-product-rename-01M1ERYJ/
create_intent: []
execution_mode: planning_artifact
model: ''
owned_files:
- kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md
role: researcher
tags: []
tracker_refs: []
---

# WP03: Prove compatibility and public cutover

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `researcher-robbie`
- **Role**: `researcher`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Produce the auditable proof that the integrated Hubnot rename preserves every
established compatibility and historical-integrity boundary, passes the full
release gate, and works from the explicitly renamed public Git remote. Record
all commands, exact revisions, and outcomes in the mission-owned verification
artifact without modifying product code, compatibility namespaces, or history.

## Context

WP01 and WP02 perform the active code, module, documentation, project-metadata,
and repository-URL changes. This package begins only after both dependencies
are integrated into the working branch. It is the convergence point for plan
concerns IC-02 and IC-03: it proves that the visible rename did not become a
protocol migration and that the public cutover is real rather than a redirect
or credential-assisted local success.

The rename deliberately preserves `nh`, `.nh`, `refs/nh/*`, `nh/0`,
`nh.pipeline/0`, `nh.policy/0`, and `nh-memory/0`. It also preserves all
pre-mission signed facts and append-only event journals byte-for-byte. The
source term remains legitimate only where `occurrence_map.yaml` or an immutable
historical surface explicitly classifies it as an exception. Never normalize,
rewrite, regenerate, or recommit historical evidence merely to make a search
clean.

The sole owned tracked file is:

- `kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md`

Operational commands may inspect the repository, Git object database, local
configuration, and public remote. Do not edit files outside the owned surface.
If a verification failure identifies a defect in WP01 or WP02, stop, preserve
the failing output, and route the defect back to the owning package instead of
repairing it here.

Run this work package with:

```bash
spec-kitty agent action implement WP03 --agent <name>
```

Before recording evidence, capture the exact working revision, branch, Go
version, Git version, and UTC timestamp. Commands in the report must be
reproducible from the repository root, and every pass/fail statement must be
paired with its command and exit result rather than inferred from earlier QA.

### Subtask T009: Audit occurrence classification and compatibility namespaces

**Purpose**: Demonstrate that every remaining tracked use of the former product
name is an intentional classified exception and that the stable `nh`
compatibility namespace was not renamed or semantically altered.

**Steps**:

1. Read `occurrence_map.yaml` and treat it as the authoritative classification
   policy for the bulk edit:
   - Confirm its target is the former product term and its replacement is
     `Hubnot`.
   - Confirm the source-term occurrence inside the map itself is preserved as
     governance metadata.
   - Enumerate every exception pattern and its required action before running
     repository-wide searches.

2. Search tracked files case-insensitively for the former product term:

   ```bash
   git grep -n -i 'nichthub' -- .
   ```

   Capture the complete result in a temporary log or directly in the evidence
   report. Classify every hit as one of:
   - the occurrence map's governed target declaration;
   - an append-only canonical or mission event journal;
   - completed historical mission evidence explicitly excluded by the map;
   - a literal historical slug or filesystem path required for truthfulness;
   - a defect that must be returned to WP01 or WP02.

3. Reject the verification if any hit occurs in active Go branding, `go.mod`,
   README prose, maintained guide prose, active project configuration, active
   charter prose, or the current glossary unless the map explicitly preserves
   that exact occurrence.

4. Audit the frozen namespace inventory using exact-string searches. At a
   minimum, prove continued use of:
   - command name `nh`;
   - private directory `.nh` where applicable to repository data;
   - collaboration refs under `refs/nh/`;
   - event protocol `nh/0`;
   - pipeline protocol `nh.pipeline/0`;
   - policy protocol `nh.policy/0`;
   - memory protocol `nh-memory/0`.

5. Compare the relevant constants, fixtures, and golden assertions to the
   baseline commit named by the mission plan (`7dcfbc1`). Use read-only Git
   inspection such as `git show`, `git diff`, and `git cat-file`; do not check
   out or rewrite baseline files. Explain any diff touching a compatibility
   token and prove it changes only surrounding product prose.

6. Verify that the Go module reports `hubnot`, while test and automation command
   examples still invoke `nh`. This distinction is intentional: module identity
   is current branding, while the CLI spelling is a compatibility contract.

7. Add an occurrence-audit table to `rename-verification.md` with columns for
   surface, command or query, result count, classification, and disposition.
   Include a separate frozen-namespace table with the expected literal and the
   evidence that it remains authoritative.

**Files**:

- Modify `kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md`
  (approximately 50–90 lines for occurrence and namespace evidence).
- Read `occurrence_map.yaml`, active product files, historical exceptions, and
  baseline Git objects; do not modify them.

**Validation**:

- Every tracked source-term occurrence has a one-to-one classification backed
  by `occurrence_map.yaml`; there are zero unclassified active occurrences.
- All frozen namespace literals remain present and semantically authoritative.
- `go list -m` reports `hubnot`.
- The report distinguishes current branding from protocol compatibility rather
  than presenting a mechanically empty search as success.

### Subtask T010: Prove immutable journals, signed facts, and ref lineage

**Purpose**: Establish that the rename did not rewrite append-only Spec Kitty
journals, signed collaboration facts, or their Git object identities, and that
any governance ref movement is append-only.

**Steps**:

1. Use the baseline commit `7dcfbc1` as the pre-mission comparison point. Resolve
   it to a full commit ID and record that ID. Also record the full candidate
   commit being verified; abbreviations alone are insufficient evidence.

2. Enumerate protected tracked journals at both revisions:
   - `.kittify/canonical-events.jsonl`;
   - every tracked `kitty-specs/**/status.events.jsonl` that existed at the
     baseline;
   - any additional journal path explicitly excluded by
     `occurrence_map.yaml`.

3. For each protected baseline journal, compare the baseline blob ID with the
   candidate blob ID. A path that existed at the baseline must retain the exact
   blob ID. New mission journals may exist, but existing journals may not be
   edited, reformatted, regenerated, or reordered.

4. Enumerate all signed collaboration events reachable from baseline
   `refs/nh/*` actor and governance refs. Record enough deterministic data to
   reproduce the set, such as sorted full event IDs and a digest of that sorted
   manifest. Avoid copying private keyring material into the report.

5. Enumerate the same facts from the candidate repository and prove every
   baseline event object remains reachable and byte-identical:
   - compare Git object IDs directly;
   - run the existing event and chain validators through supported read-only
     `nh` commands or tests;
   - verify signatures rather than assuming object reachability implies trust.

6. For each relevant `refs/nh/*` ref that moved since the baseline, verify its
   old tip is an ancestor of its new tip. Classify new tips as ordinary governed
   mission facts. Reject deletions, rewinds, force-updates, forks, or replacement
   objects.

7. Run the compatibility-focused tests that cover exact event bytes, signatures,
   policy/event versions, identity continuity, replication re-projection, and
   memory formats. Use `-count=1` and isolate host Git configuration where the
   test harness permits it.

8. Add an immutable-evidence section to `rename-verification.md` containing:
   - full baseline and candidate revisions;
   - the journal path/blob comparison;
   - baseline signed-event count and deterministic manifest digest;
   - reachability and signature-validation results;
   - old/new governance ref tips and fast-forward verdicts.

9. If any protected blob or signed event differs, stop immediately. Do not
   attempt to repair the mismatch by recreating an object, moving a ref, or
   rewriting a journal. Preserve exact diagnostic output and mark the package
   blocked pending an integrity review.

**Files**:

- Modify only `kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md`
  (approximately 60–100 lines for immutable-history evidence).
- Inspect Git objects, refs, event journals, and tests read-only.

**Validation**:

- Every protected pre-mission journal has the same full blob ID.
- Every pre-mission signed event has the same object ID and payload bytes and is
  still reachable through an append-only valid chain.
- Every moved collaboration ref is a proven fast-forward.
- Compatibility tests pass without rewriting golden payloads or protocol
  fixtures.

### Subtask T011: Run the complete local release and acceptance gate

**Purpose**: Verify the integrated rename at the same reliability level as the
alpha release, including uncached tests, race detection, static checks, build,
formatting, diff hygiene, and compiled end-to-end scenarios.

**Steps**:

1. Start from a clean integrated worktree after WP01 and WP02. Record
   `git status --short`, the full `HEAD`, and the exact Go and Git versions.
   Account for the expected WP03 evidence-file modification when reporting
   cleanliness.

2. Create a temporary, task-specific directory for test caches and outputs.
   Do not write generated binaries into the repository and do not reuse the
   obsolete former-brand binary basename.

3. Disable ambient Git configuration for commands intended to prove portable
   behavior:

   ```bash
   GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -count=1 ./...
   ```

   Record elapsed time and exit status. If the environment requires an explicit
   temporary home or cache, name it task-specifically and document it.

4. Run the race suite uncached with the same Git isolation:

   ```bash
   GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -race -count=1 ./...
   ```

   Do not substitute a cached earlier run or a run against one dependency branch.

5. Run all non-test gates:

   ```bash
   gofmt -l .
   go vet ./...
   go build ./...
   git diff --check
   ```

   `gofmt -l .` must emit no tracked Go source. Build output must stay outside
   the repository or use Go's package-only build behavior.

6. Run the compiled end-to-end acceptance scenarios that exercise real Git
   repositories, signing, governance, replication, CI, and memory. Identify the
   exact test names or supported acceptance command in the report rather than
   describing them generically.

7. Where Bubblewrap is available, run the repository-defined sandbox pipeline
   against the exact candidate revision. Record backend, request/result IDs,
   exit status, duration, and whether the runner had network access. If the host
   cannot supply Bubblewrap, mark that check unavailable with evidence; do not
   silently replace it with the host backend.

8. Scan the repository after all gates for unexpected generated files. Confirm
   that the obsolete former-brand build artifact is absent from tracked and
   untracked status and that no test contaminated Git refs or private keyring
   state in the working repository.

9. Add a release-gate matrix to `rename-verification.md`. For each command,
   record exact invocation, environment isolation, revision, exit code, elapsed
   time, and concise outcome. Preserve meaningful failure output even if a later
   rerun passes.

**Files**:

- Modify only `kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md`
  (approximately 50–90 lines for release and acceptance results).
- Use temporary locations for binaries, logs, caches, and isolated repositories.

**Validation**:

- Uncached full and race suites pass at the exact candidate revision.
- `go vet`, `go build`, formatting, and diff checks pass.
- Compiled end-to-end acceptance scenarios pass with host Git configuration
  disabled.
- Sandbox evidence is explicit and no generated artifact pollutes the worktree.

### Subtask T012: Verify the renamed public remote with a credential-free clone

**Purpose**: Complete the public cutover proof by showing that the repository is
advertised and consumable at `github.com/LynnColeArt/hubnot`, without relying on
an old-URL redirect, cached credentials, or the current checkout.

**Steps**:

1. Inspect the current checkout's remote configuration and record the complete
   fetch and push URLs for `origin`. The canonical URL must name
   `github.com/LynnColeArt/hubnot`; a former-name URL that happens to redirect is
   a failure.

2. If the local `origin` still names the old repository, update only the local
   Git remote configuration to the canonical renamed URL, then re-read it. This
   changes local operational configuration, not a tracked file or protocol
   namespace. Do not rewrite URLs inside immutable journals or signed facts.

3. Query the explicit public HTTPS URL with credentials and interactive prompts
   disabled. Record the advertised default branch and its full object ID. Do
   not derive the URL from `origin`, because that could conceal stale local
   configuration.

4. Create a fresh temporary directory and clone the explicit URL using a clean
   Git configuration and no credential helper. Suggested controls include:
   - `GIT_CONFIG_GLOBAL=/dev/null`;
   - `GIT_CONFIG_NOSYSTEM=1`;
   - `GIT_TERMINAL_PROMPT=0`;
   - command-scoped `credential.helper=`;
   - an empty, task-specific home directory where needed.

5. Confirm the clone resolves directly to the renamed repository and checks out
   the intended public revision. Do not accept a redirect-only result without
   also proving the configured origin in the fresh clone is the canonical
   Hubnot URL.

6. In the fresh clone, run:
   - `go list -m` and verify `hubnot`;
   - `go build ./...`;
   - the compiled end-to-end acceptance scenarios with host Git configuration
     disabled;
   - a CLI help or smoke command proving visible branding is Hubnot while the
     executable command surface remains `nh`.

7. Inspect public collaboration refs needed to audit the governed cutover. The
   public transport must advertise the expected fast-forwarded `refs/nh/*`
   state without requiring a private keyring. Never copy maintainer or reviewer
   private identity files into the fresh clone.

8. Add a public-cutover section to `rename-verification.md` containing the exact
   URL, public branch OID, fresh-clone OID, credential-isolation controls,
   acceptance commands, collaboration-ref observation, and final verdict.

9. Remove the temporary clone after evidence is captured using an explicit
   validated path. Do not remove the current checkout, a workspace root, or a
   path selected through an unresolved variable or glob.

10. Conclude the report with a requirement matrix mapping FR-004, FR-007,
    FR-008, NFR-001 through NFR-004, and C-001 through C-003 to concrete evidence
    sections. Note that the QA-hardening findings remain outside this mission.

**Files**:

- Modify `kitty-specs/hubnot-product-rename-01M1ERYJ/rename-verification.md`
  (approximately 50–90 lines for public transport evidence and traceability).
- Local `.git/config` may be updated only to set the canonical `origin` URL;
  this is operational configuration, not a tracked deliverable.

**Validation**:

- `origin` uses the explicit renamed Hubnot URL for fetch and push.
- A credential-free fresh clone from that URL succeeds without depending on a
  former-name redirect.
- The fresh clone builds, identifies module `hubnot`, presents Hubnot branding,
  and passes the required acceptance scenarios.
- The evidence report maps every owned requirement and constraint to a
  reproducible result.

## Definition of Done

- `rename-verification.md` identifies the full baseline and candidate commit IDs,
  exact toolchain versions, environment controls, and UTC execution time.
- Every tracked former-name occurrence is classified, with zero unclassified
  active product occurrences.
- The `nh` CLI, storage, ref, event, policy, pipeline, and memory compatibility
  namespaces are explicitly proven unchanged.
- Every protected pre-mission event journal retains its exact baseline blob ID.
- Every pre-mission signed event remains byte-identical, signature-valid, and
  reachable; every moved collaboration ref is a verified fast-forward.
- Uncached tests, race tests, vet, build, formatting, and diff hygiene pass on
  the exact integrated candidate.
- Compiled end-to-end and available Bubblewrap acceptance evidence is recorded
  without substituting unsafe execution for unavailable sandboxing.
- The local `origin` and explicit public HTTPS URL name
  `github.com/LynnColeArt/hubnot`.
- A credential-free fresh clone builds and passes the required operational
  scenarios while exposing Hubnot branding and preserving the `nh` command.
- No private key, credential, generated binary, or temporary repository is
  added to the worktree or evidence artifact.
- The report contains a requirement-to-evidence matrix covering FR-004, FR-007,
  FR-008, NFR-001, NFR-002, NFR-003, NFR-004, C-001, C-002, and C-003.
- Completion of each subtask is recorded through the event-sourced task surface:

  ```bash
  spec-kitty agent tasks mark-status T009 --status done
  spec-kitty agent tasks mark-status T010 --status done
  spec-kitty agent tasks mark-status T011 --status done
  spec-kitty agent tasks mark-status T012 --status done
  ```

- No ticked checkbox is treated as completion evidence; the task-status events
  and committed verification artifact are authoritative.

## Risks

- **A redirect hides stale configuration**: An old GitHub URL may continue to
  clone successfully. Mitigate by querying and cloning the explicit Hubnot URL,
  then checking the fresh clone's configured origin.
- **A clean search rewards history rewriting**: Historical occurrences are
  expected in classified immutable surfaces. Mitigate by requiring a complete
  classification table instead of demanding zero repository-wide matches.
- **Verification changes what it measures**: Regenerating journals, moving
  refs, or running mutating commands can invalidate the baseline comparison.
  Mitigate with read-only Git plumbing and isolated temporary repositories.
- **Local Git configuration masks portability defects**: A user-level default
  branch or credential helper can make tests pass accidentally. Mitigate with
  disabled system/global config and credential helpers for portability checks.
- **The evidence package repairs dependency defects**: Cross-package edits would
  bypass ownership and blur review. Mitigate by failing with captured evidence
  and returning defects to WP01 or WP02.
- **Private material leaks into proof**: Keyrings or credentials are never
  needed for public-clone verification. Record public fingerprints and event IDs
  only, and inspect reports before commit.
- **Race and sandbox gates are reported against different revisions**: Record a
  full revision for every gate and reject mixed-revision evidence.
- **Scope expands into QA hardening**: Full-ID enforcement, merge recovery,
  sandbox executable resolution, and other post-alpha changes remain separate.

## Reviewer Guidance

Review this package as an evidence and integrity audit, not as a prose-only
documentation change. Re-run representative commands and verify that every
claim names the exact commit and environment in which it was observed.

Focus especially on these questions:

1. Does every remaining former-name occurrence map to an explicit historical or
   classification exception, with no active branding accidentally retained?
2. Does the report preserve and verify `nh` compatibility tokens rather than
   treating them as missed rename occurrences?
3. Are protected journals compared by exact Git blob ID, and are signed facts
   checked for byte identity, signature validity, reachability, and append-only
   ref movement?
4. Were the uncached, race, static, build, formatting, and acceptance gates run
   against one exact integrated revision with host Git configuration isolated?
5. Was the public clone made from the explicit Hubnot URL with credentials and
   prompts disabled, rather than through a redirect or authenticated checkout?
6. Does the fresh clone prove both sides of the compatibility decision: current
   Hubnot product identity and the unchanged `nh` command/protocol namespace?
7. Does the report avoid private keys, credentials, enormous raw logs, and
   unsupported claims while retaining enough exact output to reproduce results?
8. Are any failures routed back to their owning WP instead of being repaired
   through out-of-scope edits in this planning-artifact package?

The reviewer should reject the package if immutable-history evidence is based
only on textual diffs, if public transport relies on the old URL, if tests are
cached or run against mixed revisions, or if the report edits or replaces any
historical journal to eliminate a former-name occurrence.
