# Mission Specification: Hubnot Product Rename

**Mission Branch**: `hubnot-product-rename-01M1ERYJ`  
**Created**: 2026-09-01  
**Status**: Draft  
**Input**: Rename the public product and repository identity from the former
name to Hubnot after learning that the former name has a derogatory meaning in
Dutch.

This is a cross-cutting bulk edit. It renames the former product name to
`Hubnot` across the active codebase and public documentation, with per-category
rules captured in `occurrence_map.yaml`.

## User Scenarios & Testing

### User Story 1 - Present one unambiguous public identity (Priority: P1)

As a user encountering the project, I see Hubnot consistently in the README,
runtime messages, current documentation, module identity, and repository URLs,
without encountering the former product name in active product surfaces.

**Why this priority**: Removing the accidental derogatory branding is the
reason for the mission.

**Independent Test**: Search the classified active product surfaces and build
the CLI; every public brand occurrence is Hubnot and runtime output introduces
the product as Hubnot.

**Acceptance Scenarios**:

1. **Given** a fresh checkout, **when** a user reads the README and current
   guides, **then** the product and public repository are named Hubnot.
2. **Given** a built CLI, **when** help, CI logs, truncation diagnostics, or
   replication help are emitted, **then** public branding says Hubnot.
3. **Given** the Go module, **when** Go tooling reports the package identity,
   **then** it reports `hubnot`.

### User Story 2 - Preserve repository and protocol compatibility (Priority: P1)

As an existing operator, I can use repositories and signed histories created
before the rename without migrating or invalidating their protocol namespaces.

**Why this priority**: Existing signed facts bind exact bytes and Git refs; a
cosmetic rename must not silently fork the protocol.

**Independent Test**: Run the complete compatibility, signing, governance,
replication, CI, and memory suites against existing fixtures with byte-identity
assertions unchanged.

**Acceptance Scenarios**:

1. **Given** existing `nh/0` collaboration facts, **when** the renamed binary
   loads them, **then** their bytes, IDs, signatures, and projections are
   unchanged.
2. **Given** existing `.nh`, `refs/nh/*`, `nh.pipeline/0`, and `nh-memory/0`
   storage, **when** normal commands run, **then** those compatibility
   namespaces remain authoritative.
3. **Given** the established `nh` command spelling, **when** users follow
   existing automation, **then** the command surface remains compatible.

### User Story 3 - Preserve truthful historical records (Priority: P2)

As a maintainer or auditor, I can distinguish current branding from immutable
historical records rather than seeing event journals rewritten after the fact.

**Why this priority**: Append-only and exact-byte history is a core project
claim.

**Independent Test**: Compare canonical event journals and signed collaboration
refs before and after the rename; their object IDs and bytes are identical.

**Acceptance Scenarios**:

1. **Given** append-only Spec Kitty event journals and signed `refs/nh/*`
   facts, **when** the rename lands, **then** those historical bytes are not
   edited or replaced.
2. **Given** current prose derived from older work, **when** it describes the
   product rather than a literal historical value, **then** it uses Hubnot.
3. **Given** a literal historical path, slug, event payload, or immutable
   digest input, **when** occurrences are classified, **then** it remains
   unchanged and is recorded as an explicit exception.

### Edge Cases

- Case variants and possessive forms of the former name must be classified.
- Hosted-repository URLs must use `LynnColeArt/hubnot.git` after the remote
  rename.
- `nh` substrings in protocol versions, refs, directories, environment
  variables, CLI examples, and test fixtures are not matches for the product
  name and must not be mechanically changed.
- An ignored local build artifact bearing the former name is generated state,
  not a tracked compatibility contract.
- Absolute paths and historical project slugs in event journals remain
  historical even if the local checkout directory is later renamed.

## Requirements

### Functional Requirements

| ID | Title | Requirement | Priority | Status |
|----|-------|-------------|----------|--------|
| FR-001 | Public brand | Rename active user-facing product text from the former name to Hubnot. | High | Open |
| FR-002 | Runtime brand | Rename CLI help, CI log headers, truncation diagnostics, and replication help text to Hubnot. | High | Open |
| FR-003 | Module identity | Change the Go module identity from the former lowercase name to `hubnot`. | High | Open |
| FR-004 | Repository URL | Update current hosted-repository URLs and the local `origin` URL to `github.com/LynnColeArt/hubnot`. | High | Open |
| FR-005 | Active metadata | Change active project configuration, charter prose, and glossary terminology to Hubnot. | High | Open |
| FR-006 | Current docs | Rename product references in README, protocol guides, operating guides, and current prose mission artifacts. | High | Open |
| FR-007 | Compatibility namespaces | Preserve `nh`, `.nh`, `refs/nh/*`, `nh/0`, `nh.pipeline/0`, `nh.policy/0`, and `nh-memory/0`. | High | Open |
| FR-008 | Historical integrity | Preserve canonical event journals, signed collaboration histories, literal historical slugs, and exact historical filesystem paths. | High | Open |
| FR-009 | Accurate runner example | Replace the obsolete branded runner example with the implementation's `nh/<version>` form. | Medium | Open |
| FR-010 | Generated artifact hygiene | Stop treating a build artifact named after the former product as the current product artifact. | Medium | Open |

### Non-Functional Requirements

| ID | Title | Requirement | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| NFR-001 | Regression safety | Uncached tests, race tests, vet, build, formatting, and diff checks pass after the rename. | Reliability | High | Open |
| NFR-002 | Byte compatibility | Existing collaboration and memory byte/ID compatibility tests pass without fixture rewrites to protocol payloads. | Compatibility | High | Open |
| NFR-003 | Occurrence completeness | Every tracked occurrence is classified as rename, preserve, historical exception, or generated artifact before implementation review. | Quality | High | Open |
| NFR-004 | Fresh-clone usability | A credential-free clone from the renamed public URL builds and runs the end-to-end acceptance scenarios. | Operability | High | Open |

### Constraints

| ID | Title | Constraint | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| C-001 | No history rewrite | Do not rewrite Git history, append-only event journals, or signed collaboration facts to erase the former name. | Integrity | High | Open |
| C-002 | No namespace migration | Do not rename the established protocol, storage, ref, environment, or CLI compatibility namespaces in this mission. | Compatibility | High | Open |
| C-003 | No feature expansion | Do not combine the rename with the separate QA hardening findings or post-alpha features. | Scope | High | Open |

### Key Entities

- **Product brand**: The current public name `Hubnot` shown to users.
- **Compatibility namespace**: Stable technical identifiers beginning with
  `nh` that bind existing repositories, facts, scripts, and wire formats.
- **Historical record**: An append-only event payload, signed fact, literal
  former slug/path, or other exact evidence whose bytes must remain truthful.
- **Current prose artifact**: Maintained explanatory text that should use the
  current product name even when its original mission predates the rename.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Active code, README, maintained guides, Go module identity, and
  current project metadata contain zero unclassified former-name occurrences.
- **SC-002**: All classified compatibility namespace strings are byte-identical
  before and after the rename.
- **SC-003**: All canonical event journals and signed collaboration refs retain
  their pre-mission bytes and object IDs.
- **SC-004**: Full test, race, vet, build, formatting, and diff gates pass.
- **SC-005**: The public `origin` advertises the renamed URL and a
  credential-free fresh clone can build and execute acceptance tests.
