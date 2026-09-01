---
work_package_id: WP02
title: Bootstrap active policy, documentation, and charter
dependencies:
- WP01
requirement_refs:
- FR-004
- FR-008
- FR-009
- FR-010
- FR-012
- FR-013
- NFR-003
- NFR-004
- NFR-005
- C-002
- C-003
- C-004
- C-005
planning_base_branch: feat/hn-hard-cutover
merge_target_branch: feat/hn-hard-cutover
branch_strategy: Planning artifacts for this mission were generated on feat/hn-hard-cutover. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hn-hard-cutover unless the human explicitly redirects the landing branch.
subtasks:
- T008
- T009
- T010
- T011
- T012
- T013
- T014
history: []
agent_profile: curator-carla
authoritative_surface: .
create_intent:
- .hn/policy.json
- .hn/pipelines/test.json
execution_mode: code_change
model: ''
owned_files:
- .hn/policy.json
- .hn/pipelines/test.json
- README.md
- docs/**
- .kittify/charter/**
role: curator
tags: []
tracker_refs: []
---

# WP02: Bootstrap active policy, documentation, and charter

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `curator-carla`
- **Role**: `curator`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Bootstrap Hubnot's active `.hn/` policy and pipeline with fresh public actor
fingerprints, then bring every current human-facing contract and active charter
surface into exact agreement with the new `hn` runtime. Preserve legacy
configuration and dated evidence byte-for-byte while making their historical,
non-executable status unmistakable.

## Context

WP01 cuts the executable, wire versions, refs, private paths, environment
variables, and executable tests from `nh` to `hn`. This package begins only
after WP01 is available, because it must build and execute the candidate `hn`
binary to create fresh trust roots and must document the behavior that the
candidate actually implements. WP03 depends on this package and will run the
aggregate verification, attest the exact candidate through the final legacy
governance loop, merge it, publish it, and establish public `refs/hn/*` facts.

This mission is a hard protocol reset. There is no alias, fallback read,
migration, dual write, copied private key, or actor-continuity claim. Existing
`.git/nh/**`, `.nh/**`, `refs/nh/*`, completed mission artifacts, immutable
journals, and dated self-hosting evidence remain historical records only. The
new program must ignore them. In particular, `.nh/policy.json` and
`.nh/pipelines/test.json` are frozen inputs required for WP03's last
old-protocol attestation and are outside this work package's ownership.

Treat `occurrence_map.yaml` as the authoritative bulk-edit classification.
Current documentation must use `hn`, `.hn`, `.git/hn`, `refs/hn/*`, `HN_*`,
`hn/0`, `hn-memory/0`, `hn.pipeline/0`, and `hn.policy/0`. Historical blocks
must retain the tokens that were true when their evidence was produced. Do not
blindly replace substrings, rewrite signed bytes, alter the Go module `hubnot`,
or modify files outside the declared `owned_files` without recording a concise
ownership exception for the reviewer.

The fresh maintainer and reviewer private keys are operational state, not
source artifacts. Generate them only below `.git/hn/` in their intended local
repositories, verify restrictive permissions, and inspect the Git diff before
and after policy creation. Only actor fingerprints and other intentionally
public authorization material may enter `.hn/policy.json`.

Implement this work package with:

```bash
spec-kitty agent action implement WP02 --agent codex
```

### Subtask T008: Build candidate and generate fresh maintainer/reviewer `hn` identities

**Purpose**

Create independent active trust roots using the WP01 candidate without
tracking, copying, or exposing any private key material from the legacy
namespace.

**Steps**

1. Confirm WP01 is integrated and the runtime advertises `hn` rather than
   `nh`; do not proceed against a partially renamed executable.
2. Build the candidate binary as `hn` using the repository's documented Go
   toolchain. Keep the binary untracked and confirm `.gitignore` covers the
   generated artifact.
3. Snapshot `git status --short` and the tracked-file list before creating any
   identity so an accidental private-key addition is immediately visible.
4. Initialize a fresh maintainer identity in the primary repository with the
   candidate. Its private state must be rooted under `.git/hn/`, never
   `.git/nh/`, a worktree source directory, or a tracked path.
5. Prepare or reuse a dedicated reviewer clone whose lifetime extends through
   WP03. Initialize a different fresh reviewer identity there with the same
   candidate runtime.
6. Record the two full public actor fingerprints in a secure working note for
   T009. Do not record private keys, seed bytes, full identity JSON, or other
   secret material in mission artifacts or terminal summaries.
7. Confirm the fingerprints differ from each other and from every legacy actor
   named by `.nh/policy.json`.
8. Inspect the primary repository and reviewer clone directly: active keyring
   files must be under `.git/hn/`, and legacy `.git/nh/` contents must not have
   been read, copied, or changed.
9. Verify directory and file permissions retain the existing keyring contract
   (0700 directories and 0600 sensitive files where applicable).
10. Re-run `git status --short` and `git ls-files` checks to prove no private
    identity material is staged or tracked.

**Files**

- No tracked source file is created by this subtask.
- Untracked local maintainer state: `.git/hn/**` in the primary repository.
- Untracked local reviewer state: `.git/hn/**` in the dedicated reviewer clone.
- Generated candidate binary: repository-root `hn`, ignored by Git.

**Validation**

- The candidate reports the `hn` command surface and initializes both actors.
- Two distinct full fingerprints are available for T009.
- `git status --short` contains no keyring or identity file.
- `git ls-files` reports nothing below `.git/hn/` or any reviewer private path.
- Legacy identity material retains its pre-subtask metadata and content.
- Record completion with `spec-kitty agent tasks mark-status T008 --status done`.

### Subtask T009: Create active `.hn` policy and pipeline with fresh fingerprints

**Purpose**

Add the repository configuration consumed by the new runtime while retaining
the exact legacy policy and pipeline bytes needed to govern the transition.

**Steps**

1. Compute and retain SHA-256 digests for `.nh/policy.json` and
   `.nh/pipelines/test.json` before touching active configuration.
2. Create `.hn/policy.json` using schema version `hn.policy/0`. Preserve the
   established policy semantics and thresholds unless the mission plan
   explicitly requires a namespace-derived change.
3. Populate every maintainer, reviewer, runner, and other actor authorization
   field with the fresh public fingerprints from T008. Do not inherit a legacy
   actor fingerprint merely because its role was equivalent.
4. Create `.hn/pipelines/test.json` using schema version `hn.pipeline/0` and
   update runner labels, commands, paths, or environment names to their active
   `hn` forms.
5. If active actions are referenced, keep them under `.hn/actions/` and verify
   that no path points into `.nh/`.
6. Compare active policy and pipeline structure with the runtime validators
   and with WP01 tests; serialized keys unrelated to the namespace must remain
   stable.
7. Exercise policy and pipeline loading through the candidate binary so the
   files are validated by production parsing rather than JSON syntax alone.
8. Recompute the two `.nh/**` SHA-256 digests and require exact equality with
   the baseline values.
9. Search `.hn/**` for `nh`, `NH_`, `.nh`, `refs/nh`, or old wire versions;
   investigate tokens semantically rather than accepting a broad false
   positive or deleting unrelated substrings.
10. Inspect the diff for any private key, identity document, or accidental
    legacy configuration edit before staging public configuration.

**Files**

- Create `.hn/policy.json` (small JSON policy, approximately 20–60 lines).
- Create `.hn/pipelines/test.json` (small JSON pipeline, approximately 10–40
  lines).
- Create `.hn/actions/**` only if the active pipeline already requires a
  repository action; do not invent a new execution mechanism.
- Do not modify `.nh/**`.

**Validation**

- The candidate accepts both active JSON files and rejects old schema versions.
- Active policy contains exactly the new intended actors.
- No active config path or value relies on `.nh/**` or a legacy actor.
- Baseline and final SHA-256 digests for frozen `.nh` files match exactly.
- No secret or private identity material appears in `git diff --cached`.
- Record completion with `spec-kitty agent tasks mark-status T009 --status done`.

### Subtask T010: Update README active command and configuration examples

**Purpose**

Make the primary onboarding document executable verbatim with the shipped `hn`
binary and active repository namespace.

**Steps**

1. Read the README end-to-end before editing; classify each occurrence as
   current guidance, historical evidence, unrelated text, or an explicit
   discussion of the cutover.
2. Change build instructions to produce `hn`, and update every active command,
   usage snippet, recovery instruction, and shell example to invoke `hn`.
3. Update active paths and refs to `.hn/**`, `.git/hn/**`, and `refs/hn/**`.
4. Update wire and configuration versions to `hn/0`, `hn-memory/0`,
   `hn.pipeline/0`, and `hn.policy/0` wherever the current contract is
   described.
5. Update Hubnot-owned environment variables and runner labels to `HN_*` and
   `hn/<version>` forms.
6. Explain concisely that legacy `.nh/**` and `refs/nh/**` may remain as frozen
   historical evidence but are ignored by `hn`; do not imply migration or
   backward compatibility.
7. Preserve the product name `Hubnot`, module name `hubnot`, ordinary Git
   transport model, optional Bubblewrap boundary, and absence of a Docker or
   service requirement.
8. Keep previously resolved QA claims accurate, including full trust-bearing
   IDs, merge-event recovery, restricted Bubblewrap resolution, `--rerun`, and
   current shallow-recovery semantics.
9. Follow the revised README commands in a clean temporary repository where
   practical; record discrepancies as defects instead of documenting around
   them.

**Files**

- Modify `README.md`; expected changes are distributed namespace corrections
  and a short historical-boundary note, not a wholesale rewrite.

**Validation**

- Every active README command begins with `hn`.
- Build instructions create only the documented `hn` artifact.
- Current paths, refs, schemas, variables, and runner labels match WP01.
- No passage promises an alias, automatic import, server, or Docker runtime.
- Record completion with `spec-kitty agent tasks mark-status T010 --status done`.

### Subtask T011: Update current protocol, governance, CI, identity, memory, and replication docs

**Purpose**

Keep each maintained technical contract in lockstep with the hard-cutover
runtime across all security-critical subsystems.

**Steps**

1. Inventory `docs/**` and distinguish current normative guidance from dated
   transcripts or historical ledgers before editing any occurrence.
2. Update current collaboration protocol contracts to `hn/0`, `refs/hn/**`,
   and the `hn` command surface, including exact-ID and recovery examples.
3. Update governance and policy documentation to `.hn/policy.json`,
   `hn.policy/0`, fresh trust-root language, and the absence of actor continuity
   across the namespace boundary.
4. Update CI documentation to `.hn/pipelines/**`, `.hn/actions/**`,
   `hn.pipeline/0`, `HN_*`, active runner labels, and current sandbox behavior.
5. Update identity and keyring guidance to `.git/hn/**`, preserving the strict
   permissions, symlink refusal, bounded-read, atomic-write, and fsync
   invariants.
6. Update memory documentation to `hn-memory/0`, active memory refs, and the
   `.git/hn/memory/**` index without changing the trust semantics.
7. Update replication and shallow-recovery documentation to the exact
   `refs/hn/**` fetch, quarantine, remote-tracking, selection, transaction, and
   anchor paths implemented by WP01.
8. Retain ordinary Git transport and host portability; do not introduce a
   provider API, Hubnot service, Docker requirement, or compatibility daemon.
9. Cross-check every normative path and identifier against source constants
   and focused tests rather than relying solely on textual replacement.

**Files**

- Modify current normative documents under `docs/**`.
- Exclude `docs/self-hosting-alpha.md` from edits; it is immutable historical
  evidence.
- Handle `docs/host-compatibility.md` only according to T012.

**Validation**

- Current docs name only active namespace values except when explicitly
  explaining rejected legacy input.
- Security, budget, sandbox, and recovery guarantees remain substantively
  unchanged.
- Command and path examples match executable behavior.
- Historical files excluded by the occurrence map remain byte-identical.
- Record completion with `spec-kitty agent tasks mark-status T011 --status done`.

### Subtask T012: Preserve dated history and add `hn` host-compatibility evidence

**Purpose**

Extend compatibility evidence for the new namespace without retroactively
rewriting what earlier Git hosts and pre-cutover binaries actually produced.

**Steps**

1. Record a baseline digest for `docs/self-hosting-alpha.md` and do not edit the
   file.
2. Read `docs/host-compatibility.md` as an evidence ledger. Identify dated
   blocks, exact commands, commit IDs, actor IDs, ref listings, and conclusions
   belonging to the legacy era.
3. Preserve every dated legacy block verbatim, including `nh` commands and
   `refs/nh/**`; they are facts, not stale instructions.
4. Update only undated active guidance that is meant to describe the current
   program.
5. Add a clearly dated `hn` compatibility section based on evidence actually
   gathered with the candidate or by WP03. If public publication has not yet
   occurred, label pre-public results accurately and leave final public fields
   for WP03's verification artifact rather than inventing outcomes.
6. Record the binary/version, host/transport, tested command sequence, observed
   `refs/hn/**`, and explicit absence of fallback to `refs/nh/**`.
7. Distinguish coexistence on a Git host from protocol compatibility: a host
   may advertise both roots, while the active runtime selects only `refs/hn/**`.
8. Recompute the self-hosting-alpha digest and require equality.

**Files**

- Modify `docs/host-compatibility.md` with a bounded current section while
  retaining dated evidence.
- Do not modify `docs/self-hosting-alpha.md`.

**Validation**

- Historical command transcripts and exact identifiers remain unchanged.
- New statements are tied to reproducible `hn` evidence and a date/context.
- No pre-public test is misrepresented as final public verification.
- `docs/self-hosting-alpha.md` has identical baseline and final digest.
- Record completion with `spec-kitty agent tasks mark-status T012 --status done`.

### Subtask T013: Update active charter through charter doctrine

**Purpose**

Rename the active CLI prose and project directive identifiers from `NH-###` to
`HN-###` using the repository's governed charter workflow, while preserving
historical mission references and external directive provenance.

**Steps**

1. Before editing charter files, load and follow the `/spk-doctrine-charter`
   skill in full. Use its canonical interview, generation, context, and sync
   workflow rather than manually editing generated surfaces in isolation.
2. Inventory `.kittify/charter/**` and identify the authoritative source,
   generated human-readable companion, interview answers, and any provenance
   fields governed by the doctrine.
3. Change active references to the Hubnot CLI from `nh` to `hn`.
4. Rename project-owned active directive IDs `NH-001` through `NH-005` to
   `HN-001` through `HN-005`, preserving numbering, meaning, severity, and
   enforcement language.
5. Do not rename unrelated external `DIRECTIVE_*` identifiers, source
   citations, historical mission artifacts, or append-only event bytes.
6. Regenerate or synchronize all charter representations required by doctrine
   so the machine-readable source and readable companion agree.
7. Inspect the resulting diff for semantic drift beyond the approved namespace
   change; revert incidental formatter or generator churn if doctrine permits.
8. Run the doctrine-prescribed charter validation and context synchronization
   checks.

**Files**

- Modify only the active files required by the charter workflow below
  `.kittify/charter/**`.
- Do not modify completed mission specs merely because they cite `NH-###`.

**Validation**

- Active charter prose names `hn` and project directives are `HN-001..005`.
- Directive semantics and non-project provenance remain intact.
- Machine-readable and readable charter surfaces validate and agree.
- No completed historical artifact was rewritten.
- Record completion with `spec-kitty agent tasks mark-status T013 --status done`.

### Subtask T014: Run occurrence-map audit and documentation/config validation

**Purpose**

Prove the package completed a meaning-aware cutover and that every surviving
legacy token belongs to a declared historical exception.

**Steps**

1. Run scoped, case-aware searches for the exact active-era forms: `nh`,
   `NH_`, `NH-`, `.nh`, `refs/nh`, `.git/nh`, `nh/0`, `nh-memory/0`,
   `nh.pipeline/0`, and `nh.policy/0`.
2. Classify each result against `occurrence_map.yaml`. Do not treat an expected
   match in the occurrence map itself as a defect, and do not whitelist an
   active-document occurrence merely to make the scan pass.
3. Require zero legacy occurrences in `.hn/**`, current README guidance,
   current normative docs, and active charter content except explicit prose
   describing rejected or frozen legacy state.
4. Verify the inverse: active config and docs contain the expected new forms
   and do not accidentally omit an entire subsystem.
5. Validate `.hn` JSON through both a strict JSON parser and the candidate's
   production policy/pipeline loaders.
6. Re-run frozen-file digests for `.nh/**`, `docs/self-hosting-alpha.md`, and
   any other baseline evidence touched by the audit boundary.
7. Run `git diff --check`, inspect `git diff --stat`, and review the complete
   diff for out-of-scope files, private material, generated binaries, or broad
   replacement damage.
8. Confirm `go.mod` remains `module hubnot`, contains no third-party
   requirements, and no document adds Docker or a service prerequisite.
9. Run focused tests covering policy, pipeline, keyring, governance, memory,
   replication, and legacy isolation. Leave the full race and public smoke
   gates to WP03, but fix any package-owned documentation/config mismatch found
   here.
10. Preserve a concise audit summary for WP03 showing commands, counts, allowed
    exception classes, fresh actor fingerprints, and frozen-file digests;
    never include private key material.

**Files**

- Correct findings only within `.hn/**`, `README.md`, `docs/**`, or
  `.kittify/charter/**`.
- The audit summary may be supplied through task evidence; WP03 owns the final
  tracked `namespace-verification.md` report.

**Validation**

- Every remaining old-token occurrence has a specific occurrence-map reason.
- Active config validates under the candidate and contains no legacy actor.
- Frozen evidence digests match their baselines.
- `git diff --check` is clean and no secrets or binaries are tracked.
- Module/dependency and portability constraints remain satisfied.
- Record completion with `spec-kitty agent tasks mark-status T014 --status done`.

## Definition of Done

- T008 has event-sourced completion evidence and two distinct fresh public
  fingerprints; their private keys exist only in intended `.git/hn/**` roots.
- T009 has event-sourced completion evidence; active `.hn` config validates,
  names only fresh actors, and frozen `.nh` digests are unchanged.
- T010 has event-sourced completion evidence; README instructions can be
  followed with `hn` and state the no-compatibility boundary accurately.
- T011 has event-sourced completion evidence; current subsystem contracts match
  runtime constants and retain all documented security invariants.
- T012 has event-sourced completion evidence; historical compatibility records
  remain truthful and current evidence is clearly dated and scoped.
- T013 has event-sourced completion evidence from a charter-doctrine-compliant
  update, with active `HN-001..005` identifiers and synchronized surfaces.
- T014 has event-sourced completion evidence and a zero-unclassified-occurrence
  audit suitable for WP03's final verification report.
- The complete diff is confined to declared ownership, contains no key
  material, and does not modify `.nh/**`, immutable journals, completed mission
  records, `docs/self-hosting-alpha.md`, or `go.mod`.
- The package introduces no compatibility layer, Docker dependency, hosted
  service requirement, third-party Go module, or security-invariant change.

## Risks

- **Private key leakage**: identity creation can produce sensitive local files.
  Mitigate with `.git/hn/**` placement, permission checks, tracked-file scans,
  and explicit diff review before every commit.
- **Circular policy bootstrap**: the policy needs fingerprints generated by
  the candidate it will govern. Mitigate by building WP01 first, creating keys
  outside tracked source, then finalizing public policy and rerunning parsing.
- **Historical falsification**: mechanical replacement can make old evidence
  claim it used `hn`. Mitigate with baseline digests, dated-block review, and
  strict occurrence-map exceptions.
- **Mixed active namespace**: self-consistent docs can still miss one subsystem.
  Mitigate with both old-token and expected-new-token scans plus production
  loader checks.
- **Charter divergence**: editing only a generated charter file can desynchronize
  its source. Mitigate by using charter doctrine and its validation workflow.
- **Ownership collision**: WP03 owns final verification evidence and public
  operations. Keep interim audit evidence event-sourced or untracked and avoid
  editing WP03's report.

## Reviewer Guidance

Review the public/private boundary first: only full actor fingerprints may be
tracked, no keyring data may appear in any diff, and the new actors must not be
legacy actors under new labels. Recompute frozen `.nh` and historical-document
digests rather than trusting prose assertions.

Then compare every `.hn` schema, README command, normative doc path, and charter
identifier with the WP01 runtime. Pay special attention to negative behavior:
the documentation must never promise a fallback read, migration, alias, dual
write, or identity continuity. Check that `docs/host-compatibility.md` preserves
dated legacy evidence instead of rewriting it, and that final-public claims are
left to WP03 unless already reproducible.

Finally, verify the charter update followed doctrine, the complete occurrence
audit has no unexplained active-era legacy token, `go.mod` remains `hubnot`
without third-party dependencies, and all T008–T014 completion claims are backed
by `spec-kitty agent tasks mark-status` events rather than checked boxes.
