---
affected_files: []
cycle_number: 3
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T17:25:23Z'
reviewer_agent: user
wp_id: WP05
---

# WP05 downstream compatibility feedback

WP07 acceptance exposed a frozen CLI-contract mismatch. The documented handoff
form is:

`nh memory handoff --at HEAD --applies descendants --input handoff.json --json`

The landed WP05 parser rejects `--input` when `--at` or `--applies` is also
present. Permit the documented handoff composition with explicit, deterministic
precedence and strict conflict rejection, add compiled black-box coverage, and
re-run the full WP05 gates. Do not change the frozen mission contracts.
