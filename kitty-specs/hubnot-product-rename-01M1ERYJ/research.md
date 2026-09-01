# Research: Hubnot Product Rename

## Question

How can the project adopt Hubnot everywhere users currently encounter the
product identity without invalidating existing repositories, signed facts,
automation, or historical evidence?

## Findings and Decisions

### D-01 — Treat the change as a product-brand rename

The user's rename of the public repository to `LynnColeArt/hubnot` is the
authoritative product decision. Current public prose, runtime branding, module
identity, active project configuration, and hosted URLs should use Hubnot.
No external language research is needed to justify or restate the user's reason.

### D-02 — Keep `nh` as the compatibility namespace

The current implementation binds `nh` into signed protocol versions, Git ref
names, private storage paths, policy/pipeline versions, memory protocol names,
environment variables, and the established CLI spelling. Those tokens do not
contain the former product name. Renaming them would be a protocol and storage
migration rather than a brand rename, so this mission preserves them exactly.

### D-03 — Separate maintained prose from immutable history

README, maintained guides, current runtime strings, the active charter, and the
active project slug are living product surfaces and should change. Canonical
event journals, mission status event journals, signed collaboration objects,
literal historical filesystem paths, and completed mission evidence remain
unchanged. The occurrence map makes these exceptions explicit.

### D-04 — Change module and hosted repository identity

`go.mod` has no external imports and the project is a single `main` package, so
changing the module declaration to `hubnot` has no dependency migration cost.
Current documentation and the local `origin` should use the renamed public URL.
Verification must use the explicit new URL with credential helpers disabled so
a provider redirect cannot hide a stale reference.

### D-05 — Correct the runner example without changing the runner namespace

The implementation emits `nh/<version>` for new CI results. The maintained
protocol guide's obsolete branded sample should become an accurate
`nh/0.0.1-dev` example. Existing signed run results remain unchanged.

## Assumptions

- The public repository rename has already occurred and the new URL advertises
  the same Git object graph.
- The established `nh` CLI spelling remains acceptable as a compatibility name.
- Git history is not rewritten to erase old commits; current-tree branding and
  forward behavior are the mission boundary.

## Open Questions

None blocking. A future explicitly versioned migration may reconsider the `nh`
namespace, but it is out of scope here.
