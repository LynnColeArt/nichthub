# Memory CLI and Machine Contract v0

## Command surface

```text
nh memory record --kind KIND --at REV --applies MODE [record fields]
nh memory record --input FILE|- --json
nh memory handoff --at REV --applies MODE --input FILE|- --json
nh memory supersede MEMORY [record fields]
nh memory retract MEMORY --reason REASON
nh memory challenge MEMORY --reason REASON [--evidence TYPED-ID]...
nh memory show MEMORY [--json]
nh memory recall [filters] [bounds] --json
nh memory index rebuild|verify
```

Full actor, stream, memory, evidence, and subject IDs are mandatory on
trust-bearing inputs. Short IDs may appear only in human display and unambiguous
read-only lookup.

## Deliberate input

Record commands consume only explicit flags, positional content, or the exact
versioned JSON input. They do not capture prompts, responses, terminal history,
environment variables, process arguments beyond the command, clipboard data,
or arbitrary working-tree contents.

Machine input uses strict decoding with unknown fields rejected. Equivalent
normalized human and JSON requests produce the same internal record shape; the
timestamp remains an explicitly observable source of distinct signed IDs.

## Recall defaults and filters

Default `atCommit` is current `HEAD`; default lifecycle is active; default trust
is policy-qualified; default bounds are 20 records and 65,536 encoded content
bytes. Filters cover exact subject, path, topic, kind, actor, lifecycle, trust,
and deterministic lexical query.

Explicit `--include-untrusted` reveals non-qualifying valid claims with their
actual classification. It never changes policy or labels them trusted.

## Output safety

JSON recall always contains:

- a version number and constant inert-data warning;
- normalized-query/source digest;
- matched, returned, truncated, and continuation metadata;
- full provenance and separate signature, applicability, lifecycle, evidence,
  and trust fields for every item;
- author prose only beneath `memories[].data.content` and structured handoff
  data only beneath `memories[].data.handoff`;
- sorted exact missing dependencies and recovery guidance.

Recall must not execute commands, invoke tools, fetch data, append events, update
refs, amend policy, or grant authorization. JSON encoding handles all control
characters. Human output uses existing control-safe helpers and includes the
same warning/classifications.

Continuation cursors bind the normalized request and accepted-source
fingerprint. A cursor used after sources, policy, filters, or bounds change is
rejected rather than continuing a different result set.

## Errors

Errors name the invalid field and safe full public IDs when actionable. They do
not echo private key bytes, tokens, environment values, remote credentials, or
unbounded hostile content. Missing exact dependencies are distinct from invalid
or malformed dependencies.
