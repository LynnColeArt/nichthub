---
work_package_id: WP02
title: Rename documentation and active project metadata
dependencies: []
requirement_refs:
- FR-001
- FR-004
- FR-005
- FR-006
- FR-007
- FR-008
- FR-009
- NFR-003
- C-001
- C-002
- C-003
planning_base_branch: feat/hubnot-product-rename
merge_target_branch: feat/hubnot-product-rename
branch_strategy: Planning artifacts for this mission were generated on feat/hubnot-product-rename. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/hubnot-product-rename unless the human explicitly redirects the landing branch.
base_branch: kitty/mission-hubnot-product-rename-01M1ERYJ
base_commit: f59e94461eee61766c8e7de7bd9c53c6d7439b98
created_at: '2026-09-01T15:46:24.219126+00:00'
subtasks:
- T005
- T006
- T007
- T008
history: []
agent_profile: curator-carla
authoritative_surface: .
create_intent: []
execution_mode: code_change
model: ''
owned_files:
- README.md
- docs/**
- .kittify/config.yaml
- .kittify/charter/charter.md
- .kittify/charter/interview/answers.yaml
- .kittify/glossaries/team_domain.yaml
role: curator
tags: []
tracker_refs: []
---

# WP02: Rename documentation and active project metadata

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `curator-carla`
- **Role**: `curator`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Rename the current public documentation and active project metadata from the
former product name to Hubnot while preserving every established `nh` compatibility
namespace and every immutable historical record. Update maintained repository
URLs and the protocol runner example without broadening this rename into a
protocol migration, history rewrite, or QA-hardening mission.

## Context

The former product name has a derogatory meaning in Dutch, and the public
repository has already been renamed to `github.com/LynnColeArt/hubnot`.
This work package owns the human-facing and active Spec Kitty surfaces of the
rename. WP01 independently owns runtime strings, the Go module name, tests,
and generated-artifact hygiene. WP03 depends on both packages and will prove
byte compatibility, run the aggregate quality gates, update the local remote,
and verify the public cutover from a credential-free clone.

This is a classified bulk edit, not an unrestricted search-and-replace. Treat
`kitty-specs/hubnot-product-rename-01M1ERYJ/occurrence_map.yaml` as the
authoritative occurrence policy, but do not edit it because it is outside this
work package's ownership. Current prose uses Hubnot. Stable technical tokens
remain exactly as they are: the `nh` executable and command examples, `.nh`,
`refs/nh/*`, `nh/0`, `nh.pipeline/0`, `nh.policy/0`, `nh-memory/0`, and related
environment or storage namespaces. The source term in append-only journals,
signed facts, completed historical evidence, literal old slugs, and literal
historical filesystem paths remains truthful history and must not be rewritten.

The active project configuration is current state rather than protocol state,
so its project slug moves to `hubnot`. The readable charter, its interview
source, and the team glossary are likewise maintained project context. Do not
edit `.kittify/charter/charter.yaml`, canonical event journals, mission status
journals, or any file not listed in `owned_files`.

### Subtask T005: Rename the README's public product identity

**Purpose**: Make the repository's primary entry point consistently present
Hubnot as the product while leaving all compatibility-facing commands,
protocol names, paths, and examples intact.

**Steps**:

1. Read `README.md` from top to bottom before editing and classify each
   case-insensitive occurrence of `Nichthub` by meaning.
   - Rename product-title, explanatory prose, status prose, and other current
     human-facing brand references to `Hubnot`.
   - Preserve every standalone `nh` command, `.nh` path, `refs/nh/*` ref,
     version string, memory namespace, and environment-compatible spelling.
   - Do not treat `nh` embedded in a compatibility token as an abbreviation
     that should be expanded or renamed.

2. Change the H1 title and introductory identity paragraphs so a first-time
   reader encounters only the new public name.
   - Keep the architectural claim unchanged: Git remains the storage and
     transport, and Hubnot supplies portable signed collaboration state.
   - Keep the operational-alpha and deferred-work claims semantically
     identical; this mission changes identity, not maturity or scope.
   - Retain the explicit statement that no mandatory service or database is
     required.

3. Rename current prose in the operational status, replication, memory,
   conflict-recovery, and completed-alpha sections.
   - Preserve command invocations such as `nh sync`, `nh proposal revise`,
     and `nh memory recall` byte-for-byte except for unrelated surrounding
     prose.
   - Preserve protocol literals including `nh-memory/0` and `.nh/policy.json`.
   - Preserve existing security-boundary and deferred-capability wording.

4. Review links and examples in `README.md` for repository identity.
   - If a current hosted URL naming `LynnColeArt/nichthub` is present, replace
     it with the exact current repository identity `LynnColeArt/hubnot`.
   - Do not add redirects, aliases, migration machinery, or claims that the
     former remote will continue to resolve.
   - Keep link targets into `docs/` unchanged because documentation filenames
     and directory layout do not move in this mission.

5. Keep the README command-surface contract scoped to the rename.
   - Do not implement or document the separate QA hardening findings as part
     of this edit.
   - Do not rename the CLI from `nh` to `hubnot`.
   - Do not change trust, policy, runner, replication, or merge semantics.

**Files**:

- `README.md` (modify in place; brand and any current repository URL only,
  with no structural rewrite expected)

**Validation**:

1. Run `rg -n -i 'nichthub' README.md` and manually classify every remaining
   match. The expected result is no active former-brand prose.
2. Run targeted namespace checks for `nh/0`, `nh-memory/0`, `refs/nh/`, `.nh`,
   and lines beginning with or containing `nh `; compare them with the parent
   version to ensure only surrounding brand prose changed.
3. Render or inspect all Markdown links and fenced examples for accidental
   path or command changes.
4. Confirm the README still says Docker is not required and preserves the
   operational-alpha boundary without adding a new feature claim.
5. Record T005 completion with
   `spec-kitty agent tasks mark-status T005 --status done` only after the
   checks above succeed.

### Subtask T006: Rename maintained guides and current hosted-repository URLs

**Purpose**: Apply the new product identity throughout maintained guides,
change current GitHub URLs to the renamed remote, and correct the documented
runner example to match the already-implemented compatibility namespace.

**Steps**:

1. Enumerate current documentation under `docs/` and inspect every
   case-insensitive former-name occurrence before modifying it.
   - Expected maintained guides include protocol, CI, governance, identity,
     replication, memory, memory safety, host compatibility, and self-hosting.
   - Rename current prose references to `Hubnot`, including possessive or
     sentence-position variants if present.
   - Do not mechanically replace lowercase `nh` substrings or filenames.

2. Update current hosted-repository URLs in the operating evidence.
   - Replace `https://github.com/LynnColeArt/nichthub.git` with
     `https://github.com/LynnColeArt/hubnot.git` in maintained host and
     self-hosting documentation.
   - Update current clone commands and remote-advertisement descriptions to
     the explicit new URL.
   - Preserve commit IDs, event IDs, timings, actor IDs, command output, and
     other historical evidence unless the text is plainly maintained prose or
     an executable current clone command rather than a literal signed fact.
   - Do not rely on or document a GitHub redirect from the former URL.

3. Manually review `docs/protocol-v0.md` as required by the occurrence map.
   - Rename the document title and current prose from Nichthub to Hubnot.
   - Preserve the exact event protocol field value `"protocol": "nh/0"`.
   - Preserve `refs/nh/actors/*`, `refs/nh/proposals/*`, accepted tracking
     refs, `.git/nh/`, and every event-kind spelling.
   - Change the obsolete illustrative runner value `"runner": "nichthub/0"`
     to the implementation's existing `"runner": "nh/<version>"` form. This
     is documentation correction FR-009, not a wire-format migration.
   - Do not modify serialized event fixtures or claim that historical
     `run.result` payload bytes were rewritten.

4. Review the self-hosting and host-compatibility evidence carefully.
   - Distinguish narrative product labels and current URL instructions from
     immutable IDs and truthful historical observations.
   - Rename prose such as “Nichthub service” to “Hubnot service” where it is a
     current architectural description.
   - Leave ordinary Git and `refs/nh/*` transport semantics unchanged.
   - Preserve exact shell commands except for the hosted repository URL when
     that URL is the command's current target.

5. Review the remaining maintained guides for semantic rather than mechanical
   edits.
   - `docs/ci-v0.md` must retain sandbox/backend terminology and pipeline
     namespace literals.
   - `docs/replication-v0.md` must retain selection, quarantine, accepted-ref,
     and recovery contracts.
   - `docs/governance-v0.md` must retain proposal and merge semantics.
   - `docs/identity-v0.md` must retain actor/key continuity terminology.
   - `docs/memory-v0.md` and `docs/memory-safety.md` must retain
     `nh-memory/0`, memory ref names, and the untrusted-content boundary.

6. Keep documentation scope closed around the rename.
   - Do not fold F-1 full-ID enforcement, F-2 merge reconciliation, F-3
     Bubblewrap resolution, or any other QA hardening into these guides.
   - Do not revise technical claims merely to improve wording unless a small
     local grammar adjustment is required by the new name.
   - Do not rename documents or create new redirect documents.

**Files**:

- `docs/protocol-v0.md` (brand prose plus the runner example)
- `docs/host-compatibility.md` (brand prose and current hosted URL)
- `docs/self-hosting-alpha.md` (brand prose and current clone URLs)
- Other `docs/*.md` files containing active former-brand prose (targeted
  in-place edits only)

**Validation**:

1. Run `rg -n -i 'nichthub' docs` and explain every remaining match; no
   unclassified active product occurrence may remain.
2. Run `rg -n 'github\.com/LynnColeArt/(nichthub|hubnot)' docs README.md` and
   verify every current repository URL uses `hubnot` and no stale old URL is
   presented as the current clone target.
3. Verify `docs/protocol-v0.md` shows `"runner": "nh/<version>"` while retaining
   `"protocol": "nh/0"` exactly.
4. Diff all compatibility tokens in `docs/` against the parent version; reject
   accidental changes to `nh`, `.nh`, `refs/nh/*`, `nh/0`, `nh.pipeline/0`,
   `nh.policy/0`, or `nh-memory/0`.
5. Run a Markdown-link/path inspection so unchanged relative documentation
   links still resolve.
6. Record T006 completion with
   `spec-kitty agent tasks mark-status T006 --status done` only after all
   current-guide occurrences and URLs are accounted for.

### Subtask T007: Rename active Spec Kitty project metadata

**Purpose**: Make the live project configuration, readable charter sources,
and canonical team terminology identify the project as Hubnot without editing
append-only Spec Kitty history or runtime-generated charter state.

**Steps**:

1. Update `.kittify/config.yaml` as current configuration.
   - Change only `project.slug` from `nichthub` to `hubnot` for the product
     rename unless another owned current-brand field is discovered.
   - Preserve `project.uuid`, `node_id`, `build_id`, mission activations, VCS
     type, agent availability, and all unrelated configuration values.
   - Do not regenerate project identity or initialize Spec Kitty again.

2. Update `.kittify/charter/charter.md` as maintained human-readable doctrine.
   - Rename the title and current explanatory references to Hubnot.
   - Preserve the engineering constraints, quality gates, self-hosting
     qualification, and product boundary semantically.
   - Keep the command spelling `nh` and all technical compatibility language.
   - Do not edit `.kittify/charter/charter.yaml`; it is outside ownership and
     is not listed as a user-facing rename target in this package.

3. Update `.kittify/charter/interview/answers.yaml` as the maintained source
   prose behind the active charter.
   - Rename product references in `project_intent` and `review_policy` to
     Hubnot.
   - Preserve the selected mission/profile, language/toolchain decisions,
     testing requirements, quality gates, deployment constraints, paradigms,
     and directives.
   - Preserve YAML scalar structure and avoid formatting churn unrelated to
     the renamed word.

4. Review `.kittify/glossaries/team_domain.yaml` for active product terminology.
   - If the glossary already contains a product-name entry, rename its active
     surface and current definition to Hubnot without changing unrelated
     domain terms.
   - If there is no product-name entry, do not invent a broad glossary
     reorganization merely to create churn; confirm the existing proposal,
     evidence, and recovery terminology remains applicable under Hubnot.
   - Do not rename `nh` compatibility terms if any are present.

5. Immediately verify that changing the active slug does not make the current
   mission undiscoverable.
   - Use the immutable mission identifier or the explicit mission handle
     `hubnot-product-rename-01M1ERYJ` for Spec Kitty commands.
   - Run a read-only mission status or context-resolution command by explicit
     mission identifier after the edit.
   - If resolution fails, investigate configuration compatibility without
     editing journals, mission IDs, or historical slugs.

6. Protect Spec Kitty history during the metadata rename.
   - Do not edit `.kittify/canonical-events.jsonl`.
   - Do not edit any `status.events.jsonl`.
   - Do not edit completed mission evidence, including
     `kitty-specs/proposal-revision-conflict-recovery-01M1774Q/**`.
   - Do not rewrite a literal former checkout path or slug in a historical
     record merely because it contains the former name.

**Files**:

- `.kittify/config.yaml` (one active slug value)
- `.kittify/charter/charter.md` (maintained charter brand prose)
- `.kittify/charter/interview/answers.yaml` (maintained source answers)
- `.kittify/glossaries/team_domain.yaml` (only an existing active product term,
  if present; otherwise inspect without forced changes)

**Validation**:

1. Parse both edited YAML files with the repository's normal Spec Kitty
   commands; do not use a formatter that rewrites unrelated keys.
2. Confirm `.kittify/config.yaml` contains `slug: hubnot` and retains the exact
   UUID, node ID, and build ID from the parent version.
3. Run a read-only `spec-kitty` context/status lookup using the explicit
   mission identifier and confirm it still resolves.
4. Run `rg -n -i 'nichthub'` across the five owned metadata paths and classify
   any remaining match as a deliberate literal; active prose should have none.
5. Compare the canonical event journal and all mission status journals with
   the parent tree and confirm this subtask did not modify them.
6. Record T007 completion with
   `spec-kitty agent tasks mark-status T007 --status done` after resolution and
   immutability checks pass.

### Subtask T008: Audit classified occurrences and protect compatibility/history

**Purpose**: Close WP02 with a reproducible occurrence audit proving that its
owned active surfaces say Hubnot and that similarly shaped compatibility and
historical tokens were not swept into the rename.

**Steps**:

1. Re-read `occurrence_map.yaml` without editing it and build an audit list for
   the WP02-owned paths.
   - `user_facing_strings` and current prose are rename targets.
   - Current hosted repository URLs are rename targets.
   - `serialized_keys` and `cli_commands` are preservation targets.
   - `docs/protocol-v0.md` is a manual-review target.
   - Active config and charter prose follow their explicit exceptions.

2. Search every owned path case-insensitively for the former product term.
   - Include title case, lowercase, uppercase, possessive, and URL forms.
   - Classify each residual occurrence as an approved historical literal or
     report it as an implementation defect.
   - Expect zero residual active product-brand occurrences in these owned
     current surfaces.
   - Do not “fix” a residual by broad replacement until its category is known.

3. Audit protected namespaces inside the changed files.
   - Enumerate `nh`, `.nh`, `.git/nh`, `refs/nh/`, `nh/0`, `nh.pipeline/0`,
     `nh.policy/0`, and `nh-memory/0` before and after the edit.
   - Review every changed line containing one of those tokens.
   - Require exact preservation unless the line's surrounding human prose has
     changed while the token itself remains byte-identical.
   - Confirm executable examples still invoke `nh` rather than `hubnot`.

4. Audit the hosted URL transition separately.
   - All maintained current clone/fetch/advertisement URLs must use
     `https://github.com/LynnColeArt/hubnot.git`.
   - This package does not change `.git/config`; WP03 owns the local `origin`
     cutover and public fresh-clone proof.
   - A GitHub redirect or successful clone of the former URL is not acceptable
     evidence that current docs are correct.

5. Verify protected history was not edited.
   - Use `git diff --name-only` and `git status --short` to confirm every edit
     made by this package falls under its `owned_files` list.
   - Explicitly assert no canonical/status event JSONL file appears in the
     diff.
   - Explicitly assert no completed historical mission artifact appears in
     the diff.
   - Do not amend, reserialize, or normalize historical content even if it
     contains the former product name.

6. Review the aggregate documentation diff for scope and truthfulness.
   - Reject wording changes unrelated to the rename and FR-009 runner example.
   - Reject any accidental claim that protocol compatibility is stable,
     production hardening is complete, or deferred capabilities now exist.
   - Reject changes that incorporate the separate post-alpha QA findings.
   - Ensure statements using Hubnot remain grammatical and clear.

7. Run package-level formatting and whitespace checks.
   - Run `git diff --check`.
   - Inspect YAML indentation and Markdown fences manually.
   - Verify relative links still target existing files.
   - Leave full build, test, race, public remote, and fresh-clone acceptance
     evidence to dependent WP03, while reporting any obvious blocker now.

**Files**:

- No new file is required. T008 audits the aggregate changes already made to
  `README.md`, `docs/**`, and the four active `.kittify` metadata files.
- Do not write the audit into `occurrence_map.yaml`, mission journals, or the
  WP03-owned `rename-verification.md`.

**Validation**:

1. The former-term search has no unexplained result in WP02-owned files.
2. The protected-namespace comparison shows no compatibility-token rename.
3. The current-URL search finds the Hubnot URL and no stale current clone URL.
4. `git diff --name-only` is contained by this package's ownership map.
5. `git diff --check` succeeds.
6. The diff contains no journal, signed-fact, historical mission, protocol
   migration, feature expansion, or unrelated QA-hardening edit.
7. Record T008 completion with
   `spec-kitty agent tasks mark-status T008 --status done` after the complete
   audit is reviewable.

## Definition of Done

- `README.md` introduces and describes Hubnot consistently, with no active
  former-brand occurrence and no change to the `nh` CLI or compatibility
  namespaces.
- Every maintained guide under `docs/` uses Hubnot in current prose.
- Every maintained current GitHub URL in owned documentation points to
  `github.com/LynnColeArt/hubnot.git`.
- `docs/protocol-v0.md` preserves `nh/0` and other protocol literals while its
  runner inventory example uses the implementation-accurate
  `"nh/<version>"` form.
- `.kittify/config.yaml` uses the active project slug `hubnot` without changing
  the project's UUID, node ID, or build ID.
- The readable charter and interview source use Hubnot, retain their doctrine,
  and remain valid Markdown/YAML.
- The team glossary contains no stale active former-brand terminology; no
  unrelated glossary reorganization was introduced.
- Canonical event journals, mission status journals, signed facts, completed
  historical mission artifacts, literal old paths, and literal historical
  slugs are untouched.
- All `nh` command, ref, directory, protocol, policy, pipeline, memory, and
  environment compatibility tokens remain byte-identical.
- Every residual former-term occurrence in the owned surfaces is explicitly
  classified; no active occurrence remains unexplained.
- The aggregate diff is restricted to `owned_files`, is whitespace-clean, and
  contains no QA hardening or feature expansion.
- T005, T006, T007, and T008 each have an event-sourced completion record from
  `spec-kitty agent tasks mark-status <Txxx> --status done`; completion is not
  represented by editing checkboxes in this prompt.

## Risks

- **Compatibility-token collateral damage**: A blind replacement could change
  `nh/0`, `nh-memory/0`, `.nh`, refs, or command examples. Mitigate with
  semantic classification and before/after token audits.
- **Historical evidence rewrite**: Self-hosting prose and Spec Kitty state mix
  maintained explanations with exact evidence. Mitigate by limiting edits to
  owned current prose and excluding every journal, signed fact, historical
  slug, and completed mission artifact.
- **Misleading URL evidence**: GitHub may redirect the old repository URL,
  masking stale documentation. Mitigate by requiring the explicit Hubnot URL
  in every current command and leaving credential-free verification to WP03.
- **Mission resolution after slug change**: Active configuration changes can
  affect tool discovery. Mitigate by making the change only in implementation,
  using the explicit immutable mission identifier, and checking resolution
  immediately.
- **Protocol example confusion**: Changing `"nichthub/0"` to
  `"nh/<version>"` could be mistaken for changing historic payloads. Mitigate
  by limiting it to the documented field inventory and explicitly preserving
  all signed event bytes.
- **Scope creep from QA feedback**: Nearby README or guide defects may tempt
  unrelated corrections. Mitigate by restricting this package to the rename
  and FR-009; the hardening findings belong to a separate mission.
- **Cross-package overlap**: WP01 owns source/runtime/test branding and WP03
  owns verification evidence and remote cutover. Mitigate by making no edits
  outside this package's frontmatter ownership map.

## Reviewer Guidance

Review the diff by semantic category rather than by counting replacements.
First confirm all active public prose and project metadata say Hubnot. Then
inspect every changed line near `nh`, `.nh`, `refs/nh/*`, version strings,
pipeline identifiers, memory identifiers, command examples, and exact IDs to
ensure the compatibility boundary has not moved.

Pay special attention to `docs/protocol-v0.md`: the document brand should be
Hubnot, `nh/0` must remain the event protocol, and only the illustrative runner
value should move to `nh/<version>`. In host compatibility and self-hosting
guides, distinguish current clone URLs—which must use the new public remote—
from immutable IDs and truthful historical evidence—which must not change.

For active Spec Kitty metadata, verify the slug change is narrow and that the
project UUID/build/node identities are preserved. Confirm no canonical event
journal, status journal, generated charter state, or completed historical
mission artifact is present in the diff. Finally, reject any bundled QA
hardening or feature work, even if individually desirable.

## Implementation Command

```sh
spec-kitty agent action implement WP02 --agent <name>
```
