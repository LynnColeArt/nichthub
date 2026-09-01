---
work_package_id: WP01
title: Rename runtime and module identity
dependencies: []
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-007
- FR-010
- NFR-001
- NFR-002
- C-002
- C-003
planning_base_branch: feat/hubnot-product-rename
merge_target_branch: feat/hubnot-product-rename
branch_strategy: Planning artifacts for this mission were generated on feat/hubnot-product-rename. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hubnot-product-rename unless the human explicitly redirects the landing branch.
subtasks:
- T001
- T002
- T003
- T004
history: []
agent_profile: implementer-ivan
authoritative_surface: .
create_intent: []
execution_mode: code_change
model: ''
owned_files:
- .gitignore
- go.mod
- main.go
- ci.go
- replication.go
- memory_commands_test.go
- policy_commands_test.go
role: implementer
tags: []
tracker_refs: []
---

# WP01: Rename runtime and module identity

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Rename the active Go module, generated binary identity, and user-visible runtime
branding from the former product name to Hubnot while leaving every established
`nh` protocol, storage, environment, ref, and CLI compatibility namespace
unchanged. Update only the owned source and test surfaces, and prove that this
cosmetic rename does not alter signed bytes or runtime behavior.

## Context

The public Git remote has already moved to `github.com/LynnColeArt/hubnot`, but
the code at the mission baseline still introduces the product using its former
name. This work package handles the root-level executable surfaces that can be
changed independently of the documentation and active project metadata owned by
WP02.

This mission is a classified bulk edit, not a protocol migration. Apply
`occurrence_map.yaml` semantically: change current branding and test-only brand
sentinels, but do not perform a blind substring replacement. In particular,
the following remain authoritative compatibility contracts:

- the executable command spelling `nh`;
- the protocol/version strings `nh/0`, `nh.pipeline/0`, `nh.policy/0`, and
  `nh-memory/0`;
- `.nh` and private `.git/nh/` storage paths;
- `refs/nh/*` collaboration refs;
- established `NH_*` or other genuine compatibility environment names;
- serialized field names, event payload formats, and signed fact bytes.

No dependencies are added, removed, or upgraded. No source paths move. Runtime
behavior, performance, event construction, validation, signing, governance,
replication, CI execution, and memory semantics must stay unchanged except for
human-readable product branding.

WP02 can proceed independently on README, guides, and active Spec Kitty
metadata. WP03 depends on both WPs and will perform the aggregate occurrence,
history, full-suite, and public-cutover proof. This WP must therefore leave its
owned surfaces internally consistent and provide clear test evidence for the
later aggregate verification.

### Subtask T001: Rename module identity and generated artifact hygiene

**Purpose**

Change the current Go package identity to `hubnot` and make the ignored local
binary basename match the new product identity. These changes define what Go
tooling and ordinary local builds call the product without changing the stable
`nh` command interface.

**Steps**

1. Inspect `go.mod` and confirm that it is a zero-dependency, single-module
   declaration whose first line is the former lowercase module name.
2. Change only the module declaration to:

   ```go
   module hubnot
   ```

3. Do not introduce an import-path migration beyond that declaration. The
   codebase is a single `main` package and should not need mechanical import
   rewrites or a compatibility shim.
4. Inspect `.gitignore` and replace the root-level ignored generated binary
   basename for the former product with `/hubnot`.
5. Preserve all other ignore rules byte-for-byte unless formatting requires
   otherwise. Do not add broad binary globs that could hide unrelated files.
6. Do not rename the `nh` command in help text, examples, executable behavior,
   protocol strings, private paths, or refs. `nh` is the stable automation
   surface and is explicitly outside this product-brand rename.
7. If a local ignored binary with the former basename exists, treat it as
   generated state rather than source. Do not add it to Git and do not modify
   tracked files outside this WP in order to manage it.
8. Confirm with `go list -m` that Go reports `hubnot` as the module identity.
9. Confirm with `git check-ignore -v hubnot` that the new local build basename
   is ignored by the exact root-level rule.
10. Confirm that the former generated basename is no longer presented as the
    current product artifact by `.gitignore`.

**Files**

- Modify `go.mod`: one module declaration line; no dependency block changes.
- Modify `.gitignore`: one root-level generated artifact rule.
- Expected change size: approximately two replacement lines.

**Validation**

Run:

```bash
go list -m
go build -o hubnot .
git check-ignore -v hubnot
git status --short
```

`go list -m` must print `hubnot`; the build must succeed; `hubnot` must remain
ignored and absent from `git status`. Record T001 completion with:

```bash
spec-kitty agent tasks mark-status T001 --status done
```

### Subtask T002: Rename user-visible runtime branding

**Purpose**

Ensure every new human-facing introduction, CI log header, truncation notice,
and replication selection description emitted by the owned production sources
uses Hubnot, while compatibility tokens remain exact.

**Steps**

1. In `main.go`, change the introductory help sentence from the former product
   brand to `Hubnot`.
2. Read the complete help block around the changed sentence before editing.
   Preserve command names, command hierarchy, spacing, and all `nh` examples.
3. In `ci.go`, change the pipeline log header to begin with `Hubnot pipeline`.
4. In `ci.go`, change the log-cap marker to read
   `[log truncated by Hubnot]`.
5. Preserve log ordering, newlines, pipeline name, commit display, backend
   display, log-capping thresholds, and truncation behavior exactly.
6. In `replication.go`, change the `--all` flag description so it says that all
   advertised `Hubnot` refs are selected.
7. Preserve the actual ref namespace, selection record schema, object kinds,
   budget logic, quarantine behavior, and every `refs/nh/*` value.
8. Search the three production sources case-insensitively for the former brand
   after the targeted edits. Investigate each result rather than replacing it
   mechanically.
9. Search those sources for compatibility tokens and inspect the diff to ensure
   they have not changed.
10. Do not address QA hardening findings, error wording unrelated to branding,
    or new features in this work package.

**Files**

- Modify `main.go`: one public introduction string.
- Modify `ci.go`: the pipeline header and log-truncation marker.
- Modify `replication.go`: the user-visible replication `--all` description.
- Expected change size: approximately four targeted string replacements.

**Validation**

Run focused source audits and a CLI smoke check:

```bash
rg -n -i 'former-product-name' main.go ci.go replication.go
rg -n 'nh/0|nh\.pipeline/0|nh\.policy/0|nh-memory/0|refs/nh|\.git/nh|usage: nh' main.go ci.go replication.go
go run . help
go test -count=1 ./...
```

When executing the first audit, substitute the literal former product name
identified by `occurrence_map.yaml`; zero active production matches are
expected. The CLI must introduce Hubnot but continue to show `nh` commands.
Record T002 completion with:

```bash
spec-kitty agent tasks mark-status T002 --status done
```

### Subtask T003: Align non-protocol test sentinels and diagnostics

**Purpose**

Update test-only brand labels that represent current product diagnostics or
hostile-input sentinels, without rewriting protocol goldens, serialized keys,
signed event fixtures, or compatibility assertions.

**Steps**

1. In `memory_commands_test.go`, inspect the hostile prompt/environment test
   around the branded temporary path and environment variable before editing.
2. Rename the temporary sentinel path to use the `hubnot` product basename.
   Preserve its absolute temporary-directory location and the assertion that
   the hostile command must not create the file.
3. Rename the test-only environment sentinel from the former branded spelling
   to `HUBNOT_MEMORY_SENTINEL`.
4. Verify that this variable is only adversarial test input and is not an
   established production environment namespace. Do not rename real runtime
   `NH_*` compatibility variables if any are encountered elsewhere.
5. Preserve the hostile string contents, escaping, prompt-injection coverage,
   memory capture rules, and failure condition except for the current brand.
6. In `policy_commands_test.go`, change the human-readable failure diagnostic
   from the former product brand to `Hubnot refs changed`.
7. Preserve the ref values and ref namespace being compared. The diagnostic
   label changes; the compatibility assertion does not.
8. Inspect all diffs in the two tests. Reject any change to JSON payloads,
   protocol versions, event IDs, expected signatures, memory canonical bytes,
   policy digests, or `refs/nh/*` literals.
9. Run the closest focused tests for the modified functions where their names
   are clear, then run the package suite uncached.
10. Confirm that no `/tmp` sentinel with either old or new brand remains after
    tests complete.

**Files**

- Modify `memory_commands_test.go`: test-only hostile path and environment
  sentinel occurrences.
- Modify `policy_commands_test.go`: one current diagnostic label.
- Expected change size: approximately four localized replacements.

**Validation**

Run:

```bash
go test -count=1 ./...
git diff -- memory_commands_test.go policy_commands_test.go
rg -n 'nh/0|nh-memory/0|refs/nh|NH_' memory_commands_test.go policy_commands_test.go
```

The tests must retain their security and compatibility assertions. Review the
diff manually to ensure no expected protocol byte string changed. Record T003
completion with:

```bash
spec-kitty agent tasks mark-status T003 --status done
```

### Subtask T004: Audit owned surfaces and run compatibility gates

**Purpose**

Prove that WP01 is a narrow semantic rename: all owned active brand occurrences
are updated, all compatibility namespaces remain intact, and the project still
builds, vets, formats, and passes uncached tests.

**Steps**

1. Review `occurrence_map.yaml` again and compare each owned-file edit against
   its category and exception action.
2. Audit all seven owned files for every case variant and possessive form of
   the former brand. Any remaining occurrence must be either removed by this WP
   or explicitly justified as historical or classification evidence.
3. Do not edit `occurrence_map.yaml` from this WP; WP01 does not own it and its
   target term must remain literal so the audit stays meaningful.
4. Audit for all frozen compatibility namespaces named in the mission spec.
   Compare the diff rather than relying only on test success.
5. Run `gofmt` only on changed Go files. A pure string/module rename should not
   cause broad formatting churn.
6. Run `git diff --check` and confirm there is no whitespace damage.
7. Run `go vet ./...` and `go build ./...` with the renamed module declaration.
8. Run the full suite uncached with host Git configuration disabled where the
   test harness supports it:

   ```bash
   GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -count=1 ./...
   ```

9. If a failure occurs, distinguish a rename regression from a host/tooling
   issue. Fix only failures caused by owned-file changes and do not expand into
   the separate post-alpha QA hardening work.
10. Inspect `git diff --stat`, `git diff --name-only`, and the full diff. Only
    the seven owned files may be modified by this work package.
11. Capture the commands and outcomes in the WP implementation handoff so WP03
    can reuse them for aggregate verification.
12. Do not run the mission-wide race suite or public fresh-clone cutover as a
    substitute for WP03; those remain its aggregate responsibilities unless the
    runtime explicitly asks for them here.

**Files**

- Reinspect all owned files; no additional file should be created.
- Formatting may touch only changed Go files and should produce no semantic
  changes beyond the classified replacements.
- Expected aggregate WP diff: seven files and a small number of lines.

**Validation**

Run the complete WP01 gate:

```bash
gofmt -w main.go ci.go replication.go memory_commands_test.go policy_commands_test.go
git diff --check
go vet ./...
go build ./...
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -count=1 ./...
git diff --name-only
git diff --stat
```

Then record T004 completion with:

```bash
spec-kitty agent tasks mark-status T004 --status done
```

## Definition of Done

- `go.mod` declares `module hubnot` and contains no unrelated dependency or
  toolchain changes.
- `.gitignore` ignores the root `hubnot` build artifact and no longer presents
  the former product basename as the current generated binary.
- CLI introduction text, pipeline log header, truncation marker, and replication
  help use `Hubnot`.
- Test-only current-brand sentinels and human-readable diagnostics use Hubnot.
- The `nh` command spelling and all protocol, storage, ref, serialized, and
  environment compatibility namespaces remain unchanged.
- The WP diff is limited to `.gitignore`, `go.mod`, `main.go`, `ci.go`,
  `replication.go`, `memory_commands_test.go`, and `policy_commands_test.go`.
- `gofmt`, `git diff --check`, `go vet`, `go build`, and uncached tests pass.
- No protocol golden, signed payload, event ID, fixture digest, or historical
  record has been rewritten.
- T001, T002, T003, and T004 each have an event-sourced completion record from
  `spec-kitty agent tasks mark-status <Txxx> --status done`; completion is not
  represented by editing checkboxes in this file.
- The implementation handoff includes the exact validation commands and their
  outcomes for review and downstream WP03 verification.

## Risks

- **Blind replacement breaks compatibility**: The former brand resembles the
  stable `nh` namespace conceptually, but `nh` tokens are not brand matches.
  Mitigate with targeted edits and explicit frozen-namespace diff review.
- **Test fixture rewrite hides a protocol change**: Some test strings may be
  exact signed or serialized bytes. Change only the identified hostile-input
  sentinels and diagnostic label; inspect every test diff manually.
- **Module rename creates unnecessary import churn**: This is a single-package
  CLI. Change only the module declaration unless Go tooling proves another
  owned-file edit is required.
- **Generated binary pollutes the worktree**: Build to the ignored `hubnot`
  basename or a temporary destination, and confirm `git status` remains clean
  apart from intended source edits.
- **Scope expands into QA hardening**: F-1 through F-5 are a separate mission.
  Preserve behavior and report unrelated findings without fixing them here.
- **An active brand occurrence is missed due to case variation**: Perform a
  case-insensitive owned-surface audit using the literal target term from the
  occurrence map and inspect possessive forms.

## Reviewer Guidance

Review this as a compatibility-sensitive rename, not as ordinary prose churn.
Start with the full diff and verify that every change belongs to one of the
seven owned files and one of the classified current-brand surfaces. Confirm
that `nh`, `.nh`, `.git/nh`, `refs/nh/*`, `nh/0`, `nh.pipeline/0`,
`nh.policy/0`, and `nh-memory/0` are untouched wherever present.

Pay special attention to `memory_commands_test.go`: the changed environment
variable and temporary path must be test-only hostile sentinels, not runtime
compatibility names. In `policy_commands_test.go`, only the failure message may
change; the ref comparison must remain exact. In `ci.go`, ensure the header and
truncation text changed without altering cap accounting or emitted newlines.

Re-run the uncached suite with host Git configuration disabled, then verify the
module name and ignored build artifact independently. Reject unrelated cleanup,
dependency movement, protocol edits, repository URL edits, documentation edits,
or QA hardening changes from this WP. Those either belong to WP02/WP03 or to a
separate mission.

## Implementation Command

After loading the assigned profile, start implementation with:

```bash
spec-kitty agent action implement WP01 --agent <name>
```
