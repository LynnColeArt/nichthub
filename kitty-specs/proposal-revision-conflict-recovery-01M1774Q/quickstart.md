# Quickstart: Recover a Proposal with an Immutable Revision

This scenario assumes `nh init` and project policy are already configured.

## 1. Observe the conflict safely

On the target branch with a clean worktree:

```sh
nh merge <proposal-id>
```

If Git reports a conflict, Nichthub aborts the merge and restores the clean
worktree. The error identifies `<proposal-id>` as the predecessor for recovery.

## 2. Resolve the code with Git

Create a normal branch/worktree, update it onto the desired target base, resolve
the conflicts, test the result, and commit it. Nichthub does not replace Git's
conflict-resolution tools.

```sh
git switch -c resolved-change <proposal-head>
git merge <new-target-base>
# resolve files, test, and commit
```

## 3. Publish a fresh revision

```sh
nh proposal revise <proposal-id> \
  --base <new-target-base> \
  --head resolved-change \
  --body "Resolved the conflict against the new target base."
nh sync
```

The original proposal, its code ref, reviews, CI results, and decisions remain
unchanged. The new revision has a different proposal ID and no inherited
evidence.

## 4. Inspect exact lineage and evidence

```sh
nh proposal show <proposal-id>
nh proposal show <revision-id>
nh proposal status <revision-id>
```

The predecessor reports that it is superseded. The revision reports its direct
predecessor and the evidence required under the policy from its new exact base.

## 5. Review, run, decide, and merge the exact revision

```sh
nh review <revision-id> --approve --body "Resolved code looks good."
nh run request <revision-id> test
nh sync

# A trusted runner executes the exact request, then maintainers synchronize.
nh proposal status <revision-id>
nh decide <revision-id> --accept
nh merge <revision-id>
```

Using `<proposal-id>` instead of `<revision-id>` for a new acceptance or merge
fails and names its known successor.

## 6. Preserve siblings without a winner

An author may publish another valid revision naming the same predecessor:

```sh
nh proposal revise <proposal-id> --base <other-base> --head <other-head> \
  --body "Alternative resolution."
```

Both revisions are sibling candidates. Listing and inspection show both exact
IDs; Nichthub does not designate either as latest. Publication still uses the
author's single linear actor chain—two disconnected clones must not append
concurrently with the same private identity.

If disconnected repositories later reveal that different siblings were already
merged, inspection preserves both signed merge facts and reports a lineage
merge conflict. It never deletes history or picks a winner.

## Expected Verification

```sh
gofmt -w *.go
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

No Docker daemon or Nichthub server is involved.
