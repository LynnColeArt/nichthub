---
affected_files: []
cycle_number: 1
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T16:44:57Z'
reviewer_agent: user
wp_id: WP05
---

# WP05 Review Feedback — Changes Requested

Reviewed commit `8a91061` independently against the WP05 prompt, `spec.md`,
`data-model.md`, and `contracts/memory-cli-v0.md`.

## Blocking issues

### 1. `show` and `recall` cannot inspect valid memory when the repository has no policy file

`cmdMemoryShow` and `cmdMemoryRecall` both call
`memoryProjectionContextAt`, which unconditionally calls
`LoadMemoryProjectionPolicy` (`memory_commands.go:711-715`,
`memory_commands.go:943-953`). A freshly initialized repository can record a
valid signed memory, but `nh memory show <full-id> --json` then aborts with:

```text
nh: no .nh/policy.json at commit <full-commit-id>
```

That prevents exact inspection of an already verified local record and makes
the frozen `policy-missing` trust classification unreachable at this public
boundary. `show` must remain available without fetching or mutating anything;
explicit untrusted recall must likewise classify rather than abort when the
policy context is absent. Add black-box coverage beginning with `nh init`, an
empty commit, a deliberate record, exact `show`, and explicit untrusted recall.

### 2. Strict machine recall rejects the documented default bounds

`RecallRequestV0` uses non-pointer integer fields and
`normalizeRecallRequestV0` rejects zero (`memory_commands.go:52-67`,
`memory_commands.go:810-823`). Therefore this valid version-0 request shape:

```json
{"version":0,"atCommit":"<full-commit-id>","includeUntrusted":true}
```

fails with `recall bounds must be positive` instead of applying the
data-model/CLI-contract defaults of 20 records and 65,536 encoded content
bytes. Preserve rejection of explicit zero/negative values while making omitted
machine fields normalize to the frozen defaults. Add strict JSON tests that
distinguish omitted fields from explicit invalid values and exercise the public
`--input FILE|- --json` path.

### 3. `show` is incomplete for lifecycle facts

`memoryShowEnvelopeMetadataV0` has no signature-status field, and
`cmdMemoryShow` only attaches `Projection` when the selected ID is a
record-producing envelope (`memory_commands.go:95-122`,
`memory_commands.go:716-744`). Showing a retraction or challenge returns the
relationship metadata but omits the required signature status and the target's
anchor, digest, applicability, lifecycle, evidence, and trust context. This is
observable for the full lifecycle-fact ID returned by the CLI itself.

Make `show` complete and unambiguous for `record`, `supersede`, `retract`, and
`challenge`: preserve exact fact metadata and signature validity, and expose
the relevant projected record/target context with full IDs and missing
dependencies. Cover both JSON and human output for every operation.

### 4. First-item content overflow does not produce the contracted bounded page

When the first deterministic match exceeds `maxContentBytes`,
`buildMemoryRecallEnvelope` returns an error (`memory_commands.go:1000-1012`)
instead of a response with consistent `matched`, `returned`, `truncated`, and
continuation metadata. The current test at
`memory_commands_test.go:476-490` freezes that error even though T024 requires
truncation metadata and says a cursor is omitted only when no deterministic
match remains.

Reconcile the implementation with the frozen paging contract and add
black-box boundary cases for valid maximum-size records, escape-heavy content,
multibyte UTF-8, a first oversized encoded item, and an oversized item after a
returned item. Pagination must either make deterministic progress without
gaps/duplicates or reject the request before emitting a partial page under a
contracted, tested rule.

### 5. The required neutral-adapter and compatibility proof is incomplete

`TestMemoryAdapterNeutralRequestsProduceIdenticalCanonicalPayloads` compares a
Go struct and a map after direct normalization; it does not drive two
independently shaped adapters through the public JSON byte interface, append
through signing/CAS, or consume a mixed recall envelope. No test asserts that
both consumers preserve every provenance/lifecycle/applicability/evidence/trust
field, and the sentinel test checks only the stored payload rather than index,
stdout/stderr, errors, and recall output.

Complete T025 with two genuinely independent test adapters using only the
version-0 JSON bytes. Exercise all six kinds and every public operation, scan
all observable artifacts for ambient sentinels, and add explicit collaboration
compatibility assertions for unchanged event payload bytes/public IDs and
collaboration-only behavior with no memory refs/index.

## Gate evidence

- `go test ./... -run 'TestMemory(Command|Recall|Adapter)' -count=3` — pass
- `go test ./...` — pass
- `go test -race ./...` — pass (`129.015s`)
- `go vet ./...` — pass
- `go build ./...` — pass
- `git diff --check kitty/mission-agent-memory-protocol-01M19TMH..HEAD` — pass
- Independent compiled-binary black-box record — pass
- Independent compiled-binary `show` without a policy file — fail as described
- Independent compiled-binary machine recall with omitted documented bounds — fail as described
- Independent compiled-binary lifecycle-fact `show` — incomplete as described

## WP anti-pattern checklist

1. Dead code — **PASS**: `main.go` calls `cmdMemory`; the new command module's
   production paths are reachable from that route.
2. Synthetic-fixture test — **FAIL**: the claimed two-adapter acceptance proof
   stops at direct DTO normalization/canonical encoding and does not exercise
   the production CLI append/recall path.
3. Silent empty return — **PASS**: reviewed empty/nil returns are normal
   successful helper exits or explicit collection initializers, not swallowed
   failures.
4. FR coverage — **FAIL**: FR-013/FR-021 public behavior fails the no-policy and
   machine-default probes; T022/T025 operation/adapter coverage is incomplete.
5. Frozen surface — **PASS**: commit `8a91061` changes only `main.go`,
   `memory_commands.go`, and `memory_commands_test.go`.
6. Locked decision — **PASS**: no network, process execution, automatic capture,
   policy mutation, service, Docker, or vendor dependency was added.
7. Shared-file ownership — **PASS**: the WP implementation commit stays within
   its declared files and the `main.go` change is narrowly additive.
8. Production fragility — **FAIL**: unconditional policy loading turns a valid
   read-only exact-inspection path into a fail-closed abort for an otherwise
   valid signed record.
