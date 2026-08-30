# Quickstart: Operational Self-Hosting Alpha

This is the target operator flow for the mission. Command contracts are draft
until implementation and acceptance complete.

## 1. Inspect the current policy

```sh
nh policy show main
```

Record the full base policy digest and current trusted actors.

## 2. Prepare a second device actor

In a separate clone:

```sh
nh init --name "Review device"
nh identity public
```

Move only the printed public actor/key material to the existing device. Never
copy `.git/nh` or a private key.

On the existing actor's clone:

```sh
nh identity authorize \
  --relationship device \
  --actor <full-new-actor> \
  --public-key <new-public-key>
nh sync
```

After the new clone selects and synchronizes the existing actor, accept the
full authorization event:

```sh
nh identity accept <full-authorization-event-id>
nh sync
```

Both clones should now display the same accepted device relationship. Neither
policy nor project authority changed.

## 3. Amend policy under the old policy

On a candidate branch, add the new actor as a trusted reviewer or runner and
make author-only evidence insufficient. Validate before committing:

```sh
nh policy check --base main --file .nh/policy.json
```

After committing:

```sh
nh policy check --base main --head HEAD
nh proposal open --base main --head HEAD \
  --body "Add the second trusted actor without self-authorization." \
  "Amend operational policy"
```

Complete CI, review, acceptance, and merge using only actors authorized by the
old base policy. Publish collaboration refs with `nh sync` and publish the
merged branch with ordinary Git.

## 4. Configure selected replication

Each participant records full IDs and positive local budgets:

```sh
nh replication select origin \
  --actor <full-maintainer-actor> \
  --actor <full-review-actor> \
  --proposal <full-candidate-id> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456

nh replication show origin
nh sync origin
```

Selection controls what is imported. The newly amended policy separately
controls whose verified claims count.

## 5. Land a role-distinct candidate

Open a later candidate whose signed base contains the amended policy:

```sh
nh proposal open --base main --head <feature-revision> \
  --body "Exercise operational self-hosting." \
  "Use the amended collaboration policy"
nh run request <full-proposal-id> test
nh sync
```

The second actor synchronizes the exact selection, runs or reviews as required,
and publishes its evidence. The author then verifies readiness, records the
maintainer decision, merges from a clean `main`, publishes `main`, and
synchronizes the merge fact.

## 6. Verify from a fresh shallow clone

```sh
git clone --depth 1 https://github.com/LynnColeArt/nichthub.git verify
cd verify
nh replication select origin \
  --actor <full-maintainer-actor> \
  --actor <full-review-actor> \
  --proposal <full-policy-candidate> \
  --proposal <full-role-distinct-candidate> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh sync origin --recover-shallow
nh proposal status <full-role-distinct-candidate>
nh proposal show <full-role-distinct-candidate>
nh log
```

Verification must succeed without `.git/nh` private identity material. Compare
every full event, policy, ref, and commit ID with `docs/self-hosting-alpha.md`.

## 7. Planned local rotation

Use a disposable actor for the inaugural proof:

```sh
nh identity rotate --name "Rotated device"
nh identity show
nh sync
```

The predecessor and successor histories remain visible. The successor does not
inherit any project role; amend policy separately if that is desired. Lost-key
or compromise recovery is not claimed by this command.
