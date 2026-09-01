package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestProposalRevisionSyncAndConvergence(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "project.git")
	aliceDirectory := filepath.Join(root, "alice")
	bobDirectory := filepath.Join(root, "bob")
	if err := os.Mkdir(seed, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "-C", seed, "init", "-q", "-b", "main")
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "seed")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", remote, aliceDirectory)
	mustGit(t, "clone", "-q", remote, bobDirectory)

	if err := os.Chdir(aliceDirectory); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "config", "user.name", "Alice")
	mustGit(t, "config", "user.email", "alice@hn.invalid")
	alice, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}
	writeTestPolicy(t, aliceDirectory, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{alice.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals:   0,
			RequiredAccepts:     1,
			AllowAuthorApproval: true,
		},
		Pipelines: map[string]PipelinePolicy{},
	})
	mustGit(t, "add", ".hn/policy.json")
	mustGit(t, "commit", "-q", "-m", "policy")
	mustGit(t, "push", "-q", "origin", "main")
	base := mustGitText(t, "rev-parse", "main")

	rootHead := revisionSyncCommit(t, "root-candidate", "root candidate")
	rootProposal := revisionSyncOpen(t, base, rootHead, "Distributed recovery")
	legacyHead := revisionSyncCommit(t, "legacy-candidate", "legacy candidate")
	legacyProposal := revisionSyncOpen(t, base, legacyHead, "Legacy proposal")
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}

	firstHead := revisionSyncCommit(t, "first-revision", "first revision")
	firstRevision := revisionSyncRevise(t, rootProposal.ID, base, firstHead)
	secondHead := revisionSyncCommit(t, "second-revision", "second revision")
	secondRevision := revisionSyncRevise(t, rootProposal.ID, base, secondHead)
	_, _, policyDigest, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	firstDecision := appendRevisionDecision(t, alice, firstRevision.ID, policyDigest, nil)
	firstMerge := appendRevisionMerge(t, alice, firstRevision, policyDigest, firstDecision)
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(bobDirectory); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "config", "user.name", "Bob")
	mustGit(t, "config", "user.email", "bob@hn.invalid")
	if _, _, err := createIdentity("Bob"); err != nil {
		t.Fatal(err)
	}
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	for _, revision := range []*StoredEvent{firstRevision, secondRevision} {
		fetched, err := resolveEvent(events, revision.ID)
		if err != nil {
			t.Fatal(err)
		}
		verified, verifiedID, err := verifyEvent(fetched.Payload, fetched.Signature)
		if err != nil {
			t.Fatalf("verify fetched revision %s: %v", revision.ID, err)
		}
		if verifiedID != revision.ID || verified.Kind != "proposal.revise" || verified.Subject != rootProposal.ID || verified.Actor != alice.Actor || verified.Base != base || verified.Head != revision.Event.Head {
			t.Fatalf("fetched revision changed signed relationship: %#v", verified)
		}
		fetchedHead, exists, err := proposalHead(revision.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !exists || fetchedHead != revision.Event.Head {
			t.Fatalf("fetched revision code = %q, exists=%t; want %s", fetchedHead, exists, revision.Event.Head)
		}
		mustGit(t, "cat-file", "-e", revision.Event.Head+"^{commit}")
	}

	forward, err := buildLineageIndex(events)
	if err != nil {
		t.Fatal(err)
	}
	reversedEvents := slices.Clone(events)
	slices.Reverse(reversedEvents)
	reversed, err := buildLineageIndex(reversedEvents)
	if err != nil {
		t.Fatal(err)
	}
	for _, proposalID := range []string{rootProposal.ID, firstRevision.ID, secondRevision.ID} {
		forwardState, err := forward.state(proposalID)
		if err != nil {
			t.Fatal(err)
		}
		reversedState, err := reversed.state(proposalID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(forwardState, reversedState) {
			t.Fatalf("lineage state depends on event presentation order:\nforward: %#v\nreverse: %#v", forwardState, reversedState)
		}
	}
	rootState := mustLineageState(t, forward, rootProposal.ID)
	firstState := mustLineageState(t, forward, firstRevision.ID)
	secondState := mustLineageState(t, forward, secondRevision.ID)
	wantSuccessors := []string{firstRevision.ID, secondRevision.ID}
	slices.Sort(wantSuccessors)
	if !rootState.Superseded || !reflect.DeepEqual(rootState.SuccessorIDs, wantSuccessors) {
		t.Fatalf("root lineage state = %#v", rootState)
	}
	if !firstState.Merged || !reflect.DeepEqual(firstState.SiblingIDs, []string{secondRevision.ID}) || !reflect.DeepEqual(firstState.MergeEventIDs, []string{firstMerge.ID}) {
		t.Fatalf("merged revision state = %#v", firstState)
	}
	if !secondState.LineageClosed || !reflect.DeepEqual(secondState.MergedCandidateIDs, []string{firstRevision.ID}) || !reflect.DeepEqual(secondState.SiblingIDs, []string{firstRevision.ID}) {
		t.Fatalf("closed sibling state = %#v", secondState)
	}

	mustGit(t, "update-ref", proposalRef(secondRevision.ID), base)
	if err := cmdReview([]string{secondRevision.ID, "--approve"}); err == nil || !strings.Contains(err.Error(), "conflicting code refs") {
		t.Fatalf("review with conflicting fetched code ref returned %v", err)
	}
	mustGit(t, "update-ref", "-d", proposalRef(secondRevision.ID))
	if err := cmdReview([]string{secondRevision.ID, "--approve", "--body", "Exact fetched revision reviewed"}); err != nil {
		t.Fatal(err)
	}
	legacyOutput, err := captureTestOutput(t, func() error { return cmdProposalShow(legacyProposal.ID) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(legacyOutput, "Legacy proposal") || !strings.Contains(legacyOutput, "Code: available and matched") || strings.Contains(legacyOutput, "Predecessor:") || strings.Contains(legacyOutput, "Successors:") || strings.Contains(legacyOutput, "State:") {
		t.Fatalf("legacy proposal presentation changed:\n%s", legacyOutput)
	}
	if err := cmdReview([]string{legacyProposal.ID, "--approve"}); err != nil {
		t.Fatalf("review legacy proposal: %v", err)
	}
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(aliceDirectory); err != nil {
		t.Fatal(err)
	}
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}
	synchronizedEvents, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if reviews := currentReviews(synchronizedEvents, secondRevision.ID); len(reviews) != 1 || reviews[0].Event.Body != "Exact fetched revision reviewed" {
		t.Fatalf("synchronized exact revision reviews = %#v", reviews)
	}
	if reviews := currentReviews(synchronizedEvents, legacyProposal.ID); len(reviews) != 1 {
		t.Fatalf("synchronized legacy reviews = %#v", reviews)
	}
	if reviews := currentReviews(synchronizedEvents, rootProposal.ID); len(reviews) != 0 {
		t.Fatalf("revision review leaked to predecessor: %#v", reviews)
	}

	remoteRefs := strings.Fields(mustGitText(t, "--git-dir="+remote, "for-each-ref", "--format=%(refname)", "refs/hn"))
	proposalRefCount := 0
	for _, ref := range remoteRefs {
		switch {
		case strings.HasPrefix(ref, "refs/hn/actors/"):
		case strings.HasPrefix(ref, "refs/hn/proposals/"):
			proposalRefCount++
		default:
			t.Fatalf("sync introduced unexpected remote ref namespace %s", ref)
		}
	}
	if proposalRefCount != 4 {
		t.Fatalf("remote proposal refs = %d, want root, legacy, and two revisions", proposalRefCount)
	}
	trackingRefs := strings.Fields(mustGitText(t, "for-each-ref", "--format=%(refname)", "refs/hn/remotes/origin"))
	for _, ref := range trackingRefs {
		if !strings.HasPrefix(ref, "refs/hn/remotes/origin/actors/") && !strings.HasPrefix(ref, "refs/hn/remotes/origin/proposals/") {
			t.Fatalf("sync introduced unexpected tracking ref namespace %s", ref)
		}
	}
}

func revisionSyncCommit(t *testing.T, branch, message string) string {
	t.Helper()
	mustGit(t, "switch", "-q", "-C", branch, "main")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", message)
	return mustGitText(t, "rev-parse", "HEAD")
}

func mustLineageState(t *testing.T, index *lineageIndex, proposalID string) proposalLineageState {
	t.Helper()
	state, err := index.state(proposalID)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func revisionSyncOpen(t *testing.T, base, head, title string) *StoredEvent {
	t.Helper()
	if _, err := captureTestOutput(t, func() error {
		return cmdProposalOpen([]string{"--base", base, "--head", head, title})
	}); err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	for index := len(events) - 1; index >= 0; index-- {
		if events[index].Event.Kind == "proposal.open" && events[index].Event.Head == head {
			return &events[index]
		}
	}
	t.Fatalf("proposal for head %s not found", head)
	return nil
}

func revisionSyncRevise(t *testing.T, predecessorID, base, head string) *StoredEvent {
	t.Helper()
	if _, err := captureTestOutput(t, func() error {
		return cmdProposalRevise([]string{predecessorID, "--base", base, "--head", head})
	}); err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	for index := len(events) - 1; index >= 0; index-- {
		if events[index].Event.Kind == "proposal.revise" && events[index].Event.Subject == predecessorID && events[index].Event.Head == head {
			return &events[index]
		}
	}
	t.Fatalf("revision for head %s not found", head)
	return nil
}
