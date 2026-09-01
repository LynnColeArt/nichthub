# Data Model: Hubnot Product Rename

This mission adds no runtime data schema. Its relevant conceptual entities are
classification and compatibility boundaries.

## Product Brand

- **Canonical value**: `Hubnot`
- **Lowercase/module value**: `hubnot`
- **Applies to**: maintained public prose, runtime human-readable text, active
  project metadata, module identity, and hosted repository URLs.

## Compatibility Namespace

- **Canonical values**: `nh`, `.nh`, `refs/nh/*`, `nh/0`, `nh.pipeline/0`,
  `nh.policy/0`, `nh-memory/0`, and existing `NH_*` environment variables.
- **Rule**: byte-preserved by this mission.
- **Reason**: existing repositories, scripts, policies, and signed facts bind
  these exact strings.

## Historical Record

- **Kinds**: append-only Spec Kitty events, signed collaboration/memory facts,
  literal old project slugs and paths, and completed mission evidence.
- **Rule**: never rewritten by the rename.
- **Verification**: tracked historical blobs remain identical; every existing
  signed event stays byte-identical and reachable; governance refs only advance
  by fast-forward.

## Occurrence Classification

- **Inputs**: path, semantic category, exact occurrence, historical status.
- **Actions**: rename, rename-if-user-visible, manual review, or do-not-change.
- **Authority**: `occurrence_map.yaml` plus the mission spec and plan.

## Relationships

Each former-name occurrence belongs to exactly one semantic classification.
Product-brand occurrences map to the new brand. Compatibility namespaces are
not former-name occurrences and remain stable. Historical records may contain
the former name but are retained as evidence rather than treated as current
branding.
