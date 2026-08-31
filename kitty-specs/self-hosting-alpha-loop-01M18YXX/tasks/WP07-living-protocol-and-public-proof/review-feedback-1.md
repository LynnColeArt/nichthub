# WP07 review feedback — cycle 1

## Verdict: REJECTED

### Issue 1 — gap-free shallow recovery wording does not match the shipped CLI

`docs/self-hosting-alpha.md` says that, after the documented selected sync has
already supplied every dependency, explicitly running
`nh sync origin --recover-shallow` "would report that no shallow recovery is
required." The exact documented public verification flow does not do that.

Independent reproduction from a fresh credential-disabled HTTPS depth-1 clone
of `https://github.com/LynnColeArt/nichthub.git` selected the recorded two
actors and three proposals with the recorded budgets, promoted 24 verified
events, and left the repository shallow. Immediately afterward:

```text
$ nh sync origin --recover-shallow
$ echo $?
0
```

The command emitted no message. This agrees with
`recoverSelectedShallow`: when there is no durable gap and closure is already
complete, it returns the result of fresh accepted-projection verification as
an idempotent no-op. The separate statement in `docs/replication-v0.md` that a
*complete repository* is rejected is accurate; the reproduced repository was
shallow but gap-free.

Because WP07 treats documentation as an executable protocol/security surface,
the gap-free shallow case must describe the actual silent successful
verification, or the CLI must implement the promised diagnostic and cover it
with an observable-contract test. Re-run that exact post-sync command before
resubmission.

## Verified evidence that does not require rework

- Fresh public depth-1 reconstruction matched the recorded `main`, two actor,
  and three proposal advertisements; exact selection and budgets promoted 24
  events and five histories.
- The release reconstructed as merged with one qualifying distinct review,
  one trusted `sandbox` result, and one acceptance decision. The accepted
  device relation reconstructed, the clone remained shallow, and it contained
  no keyring, legacy identity record, or local actor ref.
- Independent hashing matched all public event IDs, both actor fingerprints,
  both policy digests, the pipeline digest, every recorded event commit,
  candidate ref/head binding, and merge commit/tree binding. Both actor chains
  had exact sequence, `previous`, and Git-parent continuity.
- The release candidate commit changed only product/test/documentation paths;
  its diff introduced no Spec Kitty metadata or private state. The final public
  merge tree equals the governed candidate head tree.
- Collaboration/main histories are forward-extending and the remote advertises
  only the recorded explicit refs. Documentation correctly records that all
  local public actor refs are published and that collaboration refs and `main`
  are separate publication outcomes.
- Placeholder, credential/token/private-key, credential-URL, host-private-path,
  and tracked `.git/nh` scans were clean for the proof surface.
- `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`,
  `go build ./...`, `go test ./... -run TestOperationalSelfHostingAlpha
  -count=3 -v`, and `git diff --check` all passed.

## WP anti-pattern checklist

1. Dead code — N/A (documentation-only WP; no production symbols added).
2. Synthetic-fixture test — PASS (FR-016/FR-017/FR-020 were exercised through
   production CLI/Git paths and independently through the public remote).
3. Silent empty return — N/A for the WP diff; the observed silent no-op is an
   existing production behavior misdescribed by the new documentation.
4. FR coverage — PASS.
5. Frozen surface — PASS (no frozen surface was modified).
6. Locked decision — PASS.
7. Shared-file ownership — PASS (the two WP07 commits touch only its declared
   owned files).
8. Production fragility — N/A (no production code or new raise/panic path).

No dependency files changed, so the supply-chain evidence check is N/A.
