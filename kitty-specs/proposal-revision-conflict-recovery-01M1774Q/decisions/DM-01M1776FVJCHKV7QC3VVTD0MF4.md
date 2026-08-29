# Decision Moment `01M1776FVJCHKV7QC3VVTD0MF4`

- **Mission:** `proposal-revision-conflict-recovery-01M1774Q`
- **Origin flow:** `specify`
- **Slot key:** `specify.revision.concurrent-successors`
- **Input key:** `concurrent_successors`
- **Status:** `resolved`
- **Created:** `2026-08-29T16:56:10.866535+00:00`
- **Resolved:** `2026-08-29T16:58:43.067639+00:00`
- **Opened by:** `cli`
- **Other answer:** `false`

## Question

If the original author publishes two proposal revisions from disconnected clones, how should Nichthub represent the competing successors?

## Options

- Keep both as valid sibling revisions and require explicit proposal IDs
- Choose one deterministic current revision and mark the other obsolete
- Reject the second revision once either successor is observed

## Final answer

Keep both as valid sibling proposal revisions. Preserve the fork in the signed lineage, expose every sibling, and require commands to name an explicit proposal ID; do not infer a global latest revision.

## Rationale

_(none)_

## Change log

- `2026-08-29T16:56:10.866535+00:00` — opened
- `2026-08-29T16:58:43.067639+00:00` — resolved (final_answer="Keep both as valid sibling proposal revisions. Preserve the fork in the signed lineage, expose every sibling, and require commands to name an explicit proposal ID; do not infer a global latest revision.")
