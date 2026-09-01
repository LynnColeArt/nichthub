# Implementation Plan: Hubnot Product Rename

**Branch**: `feat/hubnot-product-rename` | **Date**: 2026-09-01 | **Spec**: [spec.md](spec.md)
**Input**: Rename the active public product identity to Hubnot without changing
the established `nh` compatibility namespaces or rewriting historical facts.

## Summary

Apply the validated `occurrence_map.yaml` as a semantic rename. Current Go
module metadata, user-visible runtime text, maintained documentation, active
Spec Kitty configuration/charter prose, tests, and hosted-repository URLs move
to Hubnot. Existing protocol/storage/CLI namespaces and append-only historical
records remain byte-identical. Verification combines occurrence auditing,
protocol compatibility tests, full Go gates, and a credential-free clone from
the renamed remote.

## Technical Context

**Language/Version**: Go 1.26.4  
**Primary Dependencies**: Go standard library and Git 2.x; zero third-party Go modules  
**Storage**: Git objects/refs plus private `.git/nh/` state  
**Testing**: Go unit, integration, hostile-input, race, and compiled end-to-end acceptance tests  
**Target Platform**: Linux alpha, with portable non-runner commands where documented  
**Project Type**: Single-package CLI  
**Performance Goals**: No runtime behavior or performance change  
**Constraints**: Preserve signed bytes, Git refs, protocol strings, command spelling, and event journals  
**Scale/Scope**: Approximately 30 classified tracked files plus one local remote URL and one ignored generated binary

## Charter Check

- **Bulk-edit guardrail**: PASS. `change_mode` is `bulk_edit`; all eight
  occurrence categories and explicit historical/configuration exceptions are
  present and schema-valid.
- **Protocol integrity**: PASS. `nh/0`, `nh-memory/0`, `nh.pipeline/0`,
  `nh.policy/0`, `.nh`, `refs/nh/*`, environment prefixes, and the `nh` command
  are frozen for this mission.
- **Append-only history**: PASS. Signed collaboration refs,
  `.kittify/canonical-events.jsonl`, mission `status.events.jsonl` files, and
  completed historical mission evidence are excluded from edits.
- **Dependency safety**: PASS. No dependency is added, removed, or upgraded.
- **Test-first evidence**: PASS. Pre-rename baseline is the QA-reviewed public
  commit `7dcfbc1`; post-rename gates exercise the same byte-compatibility and
  end-to-end suites.
- **Scope discipline**: PASS. The separate QA hardening findings are not folded
  into the rename.

## Occurrence Classification and Edit Policy

| Surface | Action | Representative paths | Rationale |
| --- | --- | --- | --- |
| Public brand prose | Rename | `README.md`, `docs/*.md`, active charter | Current user-facing identity |
| Runtime brand strings | Rename | `main.go`, `ci.go`, `replication.go` | New output must say Hubnot |
| Module identity | Rename lowercase | `go.mod` | Current package identity |
| Test-only branded sentinels | Rename | `memory_commands_test.go`, diagnostics | Fixtures follow current brand without touching protocol goldens |
| Hosted repository URLs | Rename | host compatibility/self-hosting docs, local `origin` | Remote is now `LynnColeArt/hubnot` |
| Protocol/storage/CLI namespaces | Preserve | `nh/0`, `.nh`, `refs/nh/*`, command examples | Existing repositories and signed facts depend on them |
| Append-only journals and signed facts | Preserve | canonical/status event JSONL, `refs/nh/*` | Historical integrity and exact bytes |
| Completed mission evidence | Preserve | `kitty-specs/proposal-revision-conflict-recovery-*` | Truthful historical product/path context |
| Rename mission classification | Preserve source term | `occurrence_map.yaml` | The governed map must identify its target |

## Implementation Strategy

1. Capture baseline blob IDs for excluded tracked journals and the complete set
   of reachable pre-mission signed event IDs.
2. Rename active source, module, docs, charter/configuration, test fixtures,
   and current repository URLs according to the occurrence map.
3. Replace `.gitignore`'s obsolete generated binary basename and remove only
   the ignored local generated binary with that basename.
4. Audit every remaining case-insensitive source-term occurrence and classify
   it as an approved historical/map exception; reject any active occurrence.
5. Verify excluded journal blobs are byte-identical and every pre-mission signed
   event is still byte-identical and reachable; any actor-ref movement must be
   an append-only fast-forward caused by governed mission facts.
6. Run formatting, diff, build, vet, uncached tests, race tests, and compiled
   operational acceptance scenarios with host Git configuration disabled.
7. Review the aggregate diff, then land through the existing `nh` governance
   protocol and verify a credential-free clone from the renamed URL.

## Project Structure

```text
go.mod                         # active module identity
main.go                        # CLI introduction
ci.go                          # runner log branding
replication.go                 # replication help branding
README.md                      # primary public identity and commands
docs/                          # maintained protocol and operator guides
.kittify/config.yaml           # active project slug
.kittify/charter/              # active project doctrine prose
.kittify/glossaries/           # canonical Hubnot product term
*_test.go                      # renamed non-protocol fixtures and assertions
kitty-specs/hubnot-product-rename-01M1ERYJ/
├── spec.md
├── plan.md
├── occurrence_map.yaml
└── tasks/
```

**Structure Decision**: No source paths move. This is an in-place semantic
rename governed by per-surface classification; compatibility and history
exceptions stay visible to reviewers.

## Data and Compatibility Flow

```text
classified occurrence
        |
        +-- current brand/module/URL ------> rename to Hubnot
        |
        +-- nh compatibility namespace ----> preserve exact bytes
        |
        `-- immutable historical record ----> preserve and verify OID
```

No persisted protocol migration exists or is needed. Newly emitted human log
text uses Hubnot; event envelopes, ref names, policy formats, memory formats,
and CLI command spelling remain unchanged.

## Risks and Mitigations

| Risk | Mitigation |
| --- | --- |
| Blind replacement mutates exact historical data | Occurrence map exceptions plus pre/post blob and ref OID checks |
| Renaming `nh` breaks existing repositories | Explicit frozen-namespace audit and byte-compatibility suites |
| Active Spec Kitty slug changes during the mission | Change active config only after planning/task artifacts are finalized; immediately verify mission resolution |
| Old hosted URL survives in current docs | Case-insensitive tracked-file audit plus credential-free clone from the new URL |
| Generated local binary reappears as untracked | Update ignore rule before recoverably removing the exact generated artifact |

## Implementation Concern Map

### IC-01 — Brand and module identity

- **Purpose**: Rename active public and runtime identity to Hubnot.
- **Relevant requirements**: FR-001 through FR-006, FR-009, FR-010.
- **Affected surfaces**: Go sources, `go.mod`, README, docs, active charter/config,
  glossary, tests, `.gitignore`, remote URL.
- **Sequencing/depends-on**: none.
- **Risks**: Missing case variants or updating a historical literal.

### IC-02 — Compatibility and history preservation

- **Purpose**: Prove the rename does not alter established technical namespaces
  or immutable evidence.
- **Relevant requirements**: FR-007, FR-008; NFR-002, NFR-003.
- **Affected surfaces**: Protocol constants, storage paths, CLI command text,
  event journals, collaboration refs, historical mission artifacts.
- **Sequencing/depends-on**: IC-01 classification, but checks are captured before edits.
- **Risks**: A mechanically similar `nh` token or generated JSONL value is
  mistaken for branding.

### IC-03 — Verification and public cutover

- **Purpose**: Demonstrate the renamed product from local build through public
  Git transport.
- **Relevant requirements**: NFR-001, NFR-004; SC-004, SC-005.
- **Affected surfaces**: Full Go suite, race detector, sandbox acceptance,
  public Git remote, README commands.
- **Sequencing/depends-on**: IC-01 and IC-02.
- **Risks**: GitHub redirect masks a stale URL; fresh-clone verification uses
  the explicit new URL with credentials disabled.

## Complexity Tracking

No charter violations or new architectural abstractions are required.
