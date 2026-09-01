# Research: `hn` Hard Cutover

## Decision 1: Treat the rename as a protocol reset

**Decision**: Replace every active `nh` surface with `hn` and intentionally
reject the old wire versions, refs, private paths, config paths, commands, and
environment variables.

**Rationale**: The project has no users, so compatibility creates cost and
ambiguity without protecting deployed consumers. A hard boundary also prevents
legacy facts from accidentally participating in new trust decisions.

**Evidence**: The pre-edit inventory found the same namespace embedded in CLI
usage, Git refs, private filesystem paths, four serialized version strings,
runner labels, environment variables, current docs, and tests. Changing only
the executable would leave a misleading mixed protocol.

## Decision 2: Preserve history without making it executable

**Decision**: Do not rewrite completed `kitty-specs/**` missions, immutable
event journals, historical self-hosting reports, or existing `refs/nh/*`.
Keep the checked-in `.nh/` policy and pipeline unchanged because they are the
inputs to the final old-protocol governance event. Add parallel active `.hn/`
configuration; new code never reads `.nh/`.

**Rationale**: Signed event IDs bind exact bytes, and the project audit trail
must continue to describe what actually happened. Preserving a path is not a
compatibility promise when the new runtime has no code path to read it.

## Decision 3: Bootstrap independent trust roots

**Decision**: Generate fresh maintainer and reviewer Ed25519 identities below
`.git/hn/`, then place only their fingerprints in `.hn/policy.json`.

**Rationale**: Copying old key material or actor fingerprints would blur the
namespace boundary and could create the unsupported multiple-writer identity
condition. New keys make the cutover cryptographically explicit.

## Decision 4: Use the old protocol exactly once to attest the cutover

**Decision**: Build a pre-cutover `nh` executable from the public base and use
the frozen `.nh/pipelines/test.json` plus existing actors to propose, execute,
review, decide, and merge the exact final candidate. Afterward, publish new
facts only with `hn`.

**Rationale**: This retains an unbroken public governance story without adding
legacy-reading code to the new binary. The final old merge fact becomes a
historical boundary marker.

## Decision 5: Rename active charter identifiers, not historical references

**Decision**: The authoritative charter uses `HN-001` through `HN-005` and names
the `hn` CLI. Older mission plans that cited `NH-###` remain unchanged.

**Rationale**: The charter is an active project contract, while completed
missions are provenance. Applying one policy to both would either preserve a
stale active name or falsify history.

## Decision 6: Add explicit isolation tests

**Decision**: Beyond mechanically updating existing tests, add tests which seed
legacy `.git/nh/`, `.nh/`, `refs/nh/*`, and `NH_*` values and prove the new
runtime ignores them.

**Rationale**: A zero-occurrence scan proves source spelling, but not runtime
non-interference. Isolation tests directly verify the no-compatibility contract.

## Risks and mitigations

- **Mechanical replacement damages unrelated text**: classify all eight bulk
  edit categories and use token-aware replacement plus a post-edit scan.
- **Final old governance cannot execute the candidate**: retain the frozen old
  pipeline at its exact path; the pre-cutover binary evaluates it from the
  candidate while the candidate binary ignores it.
- **Fresh policy fingerprints create a circular dependency**: build the
  candidate binary first, generate private identities outside tracked source,
  then write their public fingerprints into `.hn/policy.json` and rerun all
  verification.
- **Protocol security regresses during broad edits**: keep behavior changes
  limited to namespace constants/paths and isolation tests, then run the race
  suite and live two-party governance/replication scenarios.
- **Public history is rewritten accidentally**: use append-only commits and
  ordinary ref pushes; never force-push or alter old signed objects.

## Open questions

None requiring user direction. The user explicitly selected a breaking cutover
with no compatibility commitment.
