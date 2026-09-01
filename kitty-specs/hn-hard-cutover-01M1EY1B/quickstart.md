# Quickstart: Verify the `hn` Cutover

## Build and inspect a fresh repository

```sh
go build -trimpath -o hn .

work=$(mktemp -d)
git -C "$work" init
git -C "$work" config user.name "Hubnot smoke"
git -C "$work" config user.email "smoke@hn.invalid"
cp hn "$work/hn"
(
  cd "$work"
  ./hn init --name "Smoke device"
  ./hn issue open --body "Fresh namespace" "Verify hn"
  ./hn log
  git for-each-ref --format='%(refname)' refs/hn
)
```

Expected: private state appears below `$work/.git/hn/`, signed facts appear only
below `refs/hn/*`, and emitted events use `hn/0`.

## Prove legacy isolation

In a disposable repository, create sentinel files below `.git/nh/` and a
sentinel `refs/nh/actors/*` ref before running `hn`. Record their hashes/ref
targets. After `hn init`, issue creation, log, and sync attempts:

- the sentinels are byte-identical;
- the old ref is unchanged and absent from `hn log`;
- all new state is below `.git/hn/` and `refs/hn/*`;
- setting only `NH_*` test variables changes no behavior.

Use the automated isolation tests for the exact fixture construction; malformed
legacy refs are not a valid reason to weaken normal Git ref validation.

## Run the quality gates

```sh
go build ./...
go vet ./...
test -z "$(gofmt -l .)"
git diff --check
go test -count=1 ./...
go test -race -count=1 ./...
```

Then run the occurrence audit documented in
`kitty-specs/hn-hard-cutover-01M1EY1B/occurrence_map.yaml`. Every residual old
token must be in a declared historical/manual-review path.

## Public transition

Before publication, build the legacy executable from the public base into a
temporary path. Use that binary and the existing actors to govern the exact
candidate through proposal, trusted sandbox result, independent review,
maintainer decision, and merge. Do not rebuild or amend the candidate afterward.

After the merged commit is public, use the new `hn` executable and fresh actors
to publish `refs/hn/*`. A fresh clone should select/synchronize those exact refs,
validate the new facts, and show no old facts in active projections.
