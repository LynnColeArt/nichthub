---
work_package_id: WP07
title: Living Protocol and Public Proof
dependencies:
- WP06
requirement_refs:
- FR-016
- FR-017
- FR-020
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T034
- T035
- T036
- T037
- T038
- T039
phase: Phase 5 - Living Protocol and Public Evidence
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: curator-carla
agent: codex
authoritative_surface: docs/
create_intent:
- docs/identity-v0.md
- docs/replication-v0.md
- docs/self-hosting-alpha.md
execution_mode: code_change
model: ''
owned_files:
- README.md
- docs/protocol-v0.md
- docs/governance-v0.md
- docs/host-compatibility.md
- docs/identity-v0.md
- docs/replication-v0.md
- docs/self-hosting-alpha.md
role: curator
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP07 – Living Protocol and Public Proof

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `curator-carla`
- **Role**: `curator`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Make the shipped operational model understandable and independently
reproducible, then use it to govern Nichthub's own staged public landing. The
documentation must be an accurate security and protocol boundary—not an
aspirational description—and the public verification record must contain full
immutable actor, event, policy, ref, and Git object identifiers.

Success means a fresh participant can follow repository documentation from a
depth-limited clone with no private identity, explicitly select the recorded
facts, perform bounded recovery, and reconstruct every identifier in the
inaugural public record without a Nichthub service, GitHub collaboration API,
Docker, or copied private key.

## Context & Constraints

Read all mission contracts and [quickstart](../quickstart.md), then compare them
line by line with the behavior and tests delivered by WP01–WP06. NH-005 and
DIRECTIVE_037 require protocol, threat model, compatibility observations,
examples, and code behavior to move together.

The live proof is authorized project work, but it must remain staged and
recoverable. Collaboration-ref publication and primary-branch publication are
separate actions. Never publish private keyring/selection data, temporary
paths, tokens, or credential-bearing URLs. Use ordinary Git transport; hosting
provider UI/API state is not evidence.

Be precise about limitations: budgets are hard before promotion/retention but
not portable hard pre-download network quotas; actor-key separation does not
prove separate humans; rotation is planned and does not recover a lost key;
published immutable facts cannot promise global erasure.

## Branch Strategy

- **Strategy**: final documentation/evidence lane based on completed WP06; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Use the final lane workspace from `lanes.json`. Before any public mutation,
rehearse the exact workflow against a disposable remote and verify the branch,
actor, candidate, and policy IDs. Publish only the intended explicit refs.

## Subtasks & Detailed Guidance

### T034 – Update canonical protocol and governance documentation

- Update `docs/protocol-v0.md` with exact optional event fields and validation
  rules for `identity.authorize` and `identity.accept`; state that existing
  event payloads/IDs are unchanged.
- Document deterministic continuity projection, pending/missing/conflicting
  states, duplicate facts, cycles, and the absence of timestamp/order winners.
- Update `docs/governance-v0.md` with policy show/check semantics and the rule
  that an amendment is an ordinary candidate governed solely by base bytes.
- Explain continuity and replication selection as non-authoritative inputs;
  explicit base policy actor lists remain the only role source.
- Align examples and terminology with full actor/event IDs and “candidate,” not
  provider-specific PR vocabulary.
- Cross-check all statements against automated tests and CLI help output.

### T035 – Document identity continuity and keyring safety

- Create `docs/identity-v0.md` describing actor/key invariants, mutual claims,
  local keyring layout, active pointer, legacy migration, and planned rotation.
- Separate public signed facts from private untracked keyring/transaction
  records visually and textually.
- State file permission and atomic-recovery expectations without exposing a
  private key example.
- Provide operator flows for separate device initialization, public-material
  exchange, authorization, acceptance, inspection, and planned rotation.
- Explain incomplete, competing, cyclic, and fetched-later conflict behavior.
- Explicitly defer lost-key, compromise, social/organizational recovery,
  shared-key concurrency, and implicit policy migration.

### T036 – Document replication and threat boundaries

- Create `docs/replication-v0.md` covering saved per-remote selections, exact
  refspecs, quarantine lifecycle, measurements, validation, object copy,
  atomic accepted-ref promotion, and accepted projection roots.
- Document every budget and exact one-below/equal/one-above semantics reflected
  by tests.
- Explain per-selection failure isolation and missing-dependency diagnostics
  with full IDs and exact next actions.
- Describe compatibility-all as bounded and quarantined; do not imply the old
  direct wildcard import remains.
- State the residual standard-Git pre-download resource risk prominently and
  distinguish it from the hard before-promotion boundary.
- Document shallow recovery as explicit selected fetch through the same
  transaction, never global unshallow or implicit trust expansion.

### T037 – Update README, host compatibility, and public operator flow

- Replace the README's now-implemented omission claims with an accurate alpha
  capability/scope table and retain explicit deferrals.
- Add concise command examples for identity, policy, replication, sync
  recovery, and the role-distinct workflow. Make the no-service/no-Docker
  baseline unmistakable while preserving Bubblewrap's optional/default live
  runner description.
- Update `docs/host-compatibility.md` with the already observed public actor and
  proposal ref support, selected/shallow test results, and host limitations.
- Put the final public operator flow in `docs/self-hosting-alpha.md`; use the
  mission quickstart as a planning input and ensure every trust-bearing
  placeholder in the shipped guide requires a full ID.
- Separate `nh sync` collaboration-ref publication from ordinary `git push`
  primary-branch publication.
- Ensure no command instructs users to copy `.git/nh` or a private key.

### T038 – Record deterministic offline verification evidence

- Create the structure of `docs/self-hosting-alpha.md` from the successful
  black-box acceptance run before inserting public IDs.
- Record the acceptance command, three-run result, elapsed bounds, synthetic
  remote topology, policy digests, actor separation checks, selection budgets,
  promoted/rejected categories, shallow gap/recovery, and compatibility result.
- Use full immutable identifiers when a deterministic fixture supplies them;
  otherwise distinguish reproducible assertions from run-specific IDs.
- Record zero public network mutation, zero required provider API, zero Docker,
  and zero copied private identity for automated proof.
- Do not record temporary absolute paths, environment values, raw keys beyond
  explicitly public keys, credentials, or logs containing secrets.
- Link each evidence section to the relevant protocol/operator document and
  acceptance requirement.

### T039 – Perform and record the staged public self-hosting proof

- Rehearse the complete sequence against a fresh disposable remote using the
  release-candidate binary and final docs. Resolve all discrepancies before
  touching `origin`.
- Stage 1: establish the second distinct actor/continuity facts and govern the
  policy amendment using only the old base policy. Record base/head policy
  digests, candidate, exact evidence, decision, merge, commits, and refs.
- Verify the newly added actor supplied no qualifying authority to the
  amendment that first names it.
- Stage 2: open a later real candidate based on the amended policy, prove the
  author alone is not ready, obtain required exact evidence from the distinct
  actor, decide, merge, and publish collaboration refs and `main` separately.
- If the candidate conflicts, use immutable proposal revision and require new
  exact evidence; never rewrite or silently retarget signed facts.
- From a fresh public depth-limited clone with no `.git/nh` private identity,
  save exact selections/budgets, synchronize/recover, and reconstruct every
  recorded actor, event, policy, candidate code ref, merge, and Git commit ID.
- Confirm remote advertised refs and primary branch targets by ordinary Git
  commands, not provider API/UI assertions.
- Fill `docs/self-hosting-alpha.md` with the full public record, observation
  date, commands sufficient to verify, and honest same-human/transport limits.

## Verification Strategy

Before public execution:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
go test ./... -run TestOperationalSelfHostingAlpha -count=3
```

After public execution, run the documented fresh-clone verification exactly as
written and compare every full ID. Search tracked changes for private-key
fields, tokens, credential URLs, and host-private paths. Verify docs contain no
placeholder IDs before review.

## Risks & Mitigations

- **Documentation drift**: derive claims from final CLI/tests and execute every
  example against the release candidate.
- **Irreversible public error**: rehearse offline, resolve exact refs/IDs, and
  publish staged explicit targets; corrections add facts rather than rewrite.
- **Secret leakage**: inspect diffs and generated record before push; record
  only public keys and immutable public facts.
- **Overclaiming**: state transport quota, same-human, planned-rotation, and
  immutable-publication limitations directly.
- **Split publication**: record collaboration refs and primary branch as
  separate outcomes with retry guidance.

## Review Guidance

The reviewer should treat documentation as a test surface. Follow it from a
fresh depth-limited clone, verify no private identity exists, and compare all
full IDs with the record. Cross-check every protocol field and error claim with
the implementation and black-box tests.

Reject placeholder identifiers, provider UI/API evidence, leaked local state,
claims of lost-key recovery or hard network quota, implication that continuity
grants roles, or a public record that cannot be reconstructed with ordinary Git
plus `nh`.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.
