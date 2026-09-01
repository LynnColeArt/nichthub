package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProposalAndReviewStorage(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@hn.invalid")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "base")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "feature")
	head := mustGitText(t, "rev-parse", "HEAD")

	alice := testIdentity(t, "Alice")
	proposal, err := nextEvent(alice, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	proposal.Title = "Test proposal"
	proposal.Base = base
	proposal.Head = head
	storedProposal, err := appendEvent(proposal, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(storedProposal.ID, head); err != nil {
		t.Fatal(err)
	}

	availableHead, exists, err := proposalHead(storedProposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || availableHead != head {
		t.Fatalf("proposal head = %q, exists = %v; want %q", availableHead, exists, head)
	}

	bob := testIdentity(t, "Bob")
	review, err := nextEvent(bob, "review.submit")
	if err != nil {
		t.Fatal(err)
	}
	review.Subject = storedProposal.ID
	review.Verdict = "approve"
	storedReview, err := appendEvent(review, bob)
	if err != nil {
		t.Fatal(err)
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("collected %d events, want 2", len(events))
	}
	reviews := currentReviews(events, storedProposal.ID)
	if len(reviews) != 1 || reviews[0].ID != storedReview.ID {
		t.Fatalf("current reviews = %#v, want %s", reviews, storedReview.ID)
	}
	approvals, changes := reviewCounts(reviews)
	if approvals != 1 || changes != 0 {
		t.Fatalf("review counts = +%d/-%d, want +1/-0", approvals, changes)
	}
}

func mustGit(t *testing.T, args ...string) {
	t.Helper()
	if _, err := gitOutput(args...); err != nil {
		t.Fatal(err)
	}
}

func mustGitText(t *testing.T, args ...string) string {
	t.Helper()
	value, err := gitText(args...)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func writeActiveTestIdentity(t *testing.T, identity *Identity) {
	t.Helper()
	if _, err := storeIdentityRecord(identity, identityLifecycleAvailable); err != nil {
		t.Fatal(err)
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		t.Fatal(err)
	}
	if err := writePrivateFileAtomic(paths.active, []byte(identity.Actor+"\n")); err != nil {
		t.Fatal(err)
	}
}

func captureTestOutput(t *testing.T, run func() error) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = previous }()
	runErr := run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(output), runErr
}

func TestProposalRevisionCommandAndLineageInspection(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@hn.invalid")
	alice, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}
	writeTestPolicy(t, root, PolicyDocument{
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
	mustGit(t, "commit", "-q", "-m", "base policy")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "proposal head")
	proposalHeadOID := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "revision one")
	revisionHeadOne := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "revision two")
	revisionHeadTwo := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "revision three")
	revisionHeadThree := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "revision blocked")
	blockedHead := mustGitText(t, "rev-parse", "HEAD")

	if _, err := captureTestOutput(t, func() error {
		return cmdProposalOpen([]string{"--base", base, "--head", proposalHeadOID, "Immutable recovery"})
	}); err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	proposals := proposalEvents(events)
	if len(proposals) != 1 {
		t.Fatalf("proposal count = %d, want 1", len(proposals))
	}
	predecessor := proposals[0]
	predecessorPayload := append([]byte(nil), predecessor.Payload...)
	predecessorRef, exists, err := refValue(proposalRef(predecessor.ID))
	if err != nil || !exists {
		t.Fatalf("predecessor ref = %q, %t, %v", predecessorRef, exists, err)
	}

	bob := testIdentity(t, "Bob")
	writeActiveTestIdentity(t, bob)
	eventCount := len(events)
	err = cmdProposal([]string{"revise", predecessor.ID, "--base", proposalHeadOID, "--head", revisionHeadOne})
	if err == nil || !strings.Contains(err.Error(), predecessor.ID) || !strings.Contains(err.Error(), "author") {
		t.Fatalf("non-author error = %v, want predecessor ID and author reason", err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != eventCount {
		t.Fatalf("non-author attempt changed event count from %d to %d", eventCount, len(events))
	}

	writeActiveTestIdentity(t, alice)
	revisionOutput, err := captureTestOutput(t, func() error {
		return cmdProposal([]string{"revise", predecessor.ID, "--base", proposalHeadOID, "--head", revisionHeadOne, "--body", "Resolved conflicts"})
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, policyDigest, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	rejection, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	rejection.Subject = predecessor.ID
	rejection.Policy = policyDigest
	rejection.Verdict = "reject"
	rejection.Body = "Resolve the conflict"
	if _, err := appendEvent(rejection, alice); err != nil {
		t.Fatal(err)
	}
	if err := cmdProposalRevise([]string{predecessor.ID, "--base", proposalHeadOID, "--head", revisionHeadTwo}); err != nil {
		t.Fatalf("revise rejected predecessor: %v", err)
	}

	acceptance, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	acceptance.Subject = predecessor.ID
	acceptance.Policy = policyDigest
	acceptance.Verdict = "accept"
	accepted, err := appendEvent(acceptance, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmdProposalRevise([]string{predecessor.ID, "--base", proposalHeadOID, "--head", revisionHeadThree}); err != nil {
		t.Fatalf("revise accepted-but-unmerged predecessor: %v", err)
	}

	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	revisions := make([]StoredEvent, 0)
	for _, stored := range proposalEvents(events) {
		if stored.Event.Kind == "proposal.revise" {
			revisions = append(revisions, stored)
		}
	}
	if len(revisions) != 3 {
		t.Fatalf("revision count = %d, want 3", len(revisions))
	}
	for _, revision := range revisions {
		if revision.Event.Subject != predecessor.ID {
			t.Fatalf("revision predecessor = %s, want %s", revision.Event.Subject, predecessor.ID)
		}
		head, exists, err := proposalHead(revision.ID)
		if err != nil || !exists || head != revision.Event.Head {
			t.Fatalf("revision ref = %q, %t, %v; want %q", head, exists, err, revision.Event.Head)
		}
	}
	createdRevision := revisions[0]
	if !strings.Contains(revisionOutput, "proposal revision") || !strings.Contains(revisionOutput, predecessor.ID) {
		t.Fatalf("revision output = %q, want revision wording and exact predecessor ID", revisionOutput)
	}
	storedPredecessor, err := resolveEvent(events, predecessor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(storedPredecessor.Payload, predecessorPayload) {
		t.Fatal("predecessor payload changed after revision creation")
	}
	afterRef, exists, err := refValue(proposalRef(predecessor.ID))
	if err != nil || !exists || afterRef != predecessorRef {
		t.Fatalf("predecessor ref after revisions = %q, %t, %v; want %q", afterRef, exists, err, predecessorRef)
	}

	if _, err := captureTestOutput(t, func() error {
		return cmdReview([]string{predecessor.ID, "--approve"})
	}); err != nil {
		t.Fatal(err)
	}
	writeActiveTestIdentity(t, bob)
	if _, err := captureTestOutput(t, func() error {
		return cmdReview([]string{createdRevision.ID, "--approve"})
	}); err != nil {
		t.Fatalf("review revision: %v", err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if got := currentReviews(events, createdRevision.ID); len(got) != 1 || got[0].Event.Actor != bob.Actor {
		t.Fatalf("revision reviews = %#v, want only Bob's exact revision review", got)
	}
	if got := currentReviews(events, predecessor.ID); len(got) != 1 || got[0].Event.Actor != alice.Actor {
		t.Fatalf("predecessor reviews = %#v, want only Alice's predecessor review", got)
	}

	showOutput, err := captureTestOutput(t, func() error { return cmdProposalShow(createdRevision.ID) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(showOutput, "Immutable recovery") || !strings.Contains(showOutput, "Predecessor: "+predecessor.ID) || !strings.Contains(showOutput, "Siblings:") {
		t.Fatalf("revision show output lacks inherited title or lineage:\n%s", showOutput)
	}
	for _, sibling := range revisions {
		if sibling.ID != createdRevision.ID && !strings.Contains(showOutput, sibling.ID) {
			t.Fatalf("revision show output lacks sibling %s:\n%s", sibling.ID, showOutput)
		}
	}
	listOutput, err := captureTestOutput(t, cmdProposalList)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listOutput, "State: superseded") || !strings.Contains(listOutput, "Successors:") {
		t.Fatalf("proposal list output lacks superseded lineage:\n%s", listOutput)
	}

	writeActiveTestIdentity(t, alice)
	merge, err := nextEvent(alice, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	merge.Subject = predecessor.ID
	merge.Policy = policyDigest
	merge.Head = predecessor.Event.Head
	merge.Commit = predecessor.Event.Head
	merge.Evidence = []string{accepted.ID}
	if _, err := appendEvent(merge, alice); err != nil {
		t.Fatal(err)
	}
	eventsBeforeBlocked, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	err = cmdProposalRevise([]string{predecessor.ID, "--base", proposalHeadOID, "--head", blockedHead})
	if err == nil || !strings.Contains(err.Error(), predecessor.ID) || !strings.Contains(err.Error(), "independent proposal") {
		t.Fatalf("merged predecessor error = %v, want exact ID and independent-proposal guidance", err)
	}
	eventsAfterBlocked, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsAfterBlocked) != len(eventsBeforeBlocked) {
		t.Fatalf("merged-predecessor refusal changed event count from %d to %d", len(eventsBeforeBlocked), len(eventsAfterBlocked))
	}

	_, _, revisionPolicyDigest, err := loadPolicy(createdRevision.Event.Base)
	if err != nil {
		t.Fatal(err)
	}
	revisionAcceptance, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	revisionAcceptance.Subject = createdRevision.ID
	revisionAcceptance.Policy = revisionPolicyDigest
	revisionAcceptance.Verdict = "accept"
	acceptedRevision, err := appendEvent(revisionAcceptance, alice)
	if err != nil {
		t.Fatal(err)
	}
	revisionMerge, err := nextEvent(alice, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	revisionMerge.Subject = createdRevision.ID
	revisionMerge.Policy = revisionPolicyDigest
	revisionMerge.Head = createdRevision.Event.Head
	revisionMerge.Commit = createdRevision.Event.Head
	revisionMerge.Evidence = []string{acceptedRevision.ID}
	if _, err := appendEvent(revisionMerge, alice); err != nil {
		t.Fatal(err)
	}
	conflictOutput, err := captureTestOutput(t, func() error { return cmdProposalShow(createdRevision.ID) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(conflictOutput, "State: merged, lineage closed, merge conflict") ||
		!strings.Contains(conflictOutput, "Merged lineage members:") ||
		!strings.Contains(conflictOutput, predecessor.ID) || !strings.Contains(conflictOutput, createdRevision.ID) {
		t.Fatalf("conflicting lineage output lacks exact merged members:\n%s", conflictOutput)
	}

	if runtime.GOOS != "windows" {
		gitDirectory, err := requireGitRepository()
		if err != nil {
			t.Fatal(err)
		}
		proposalRefDirectory := filepath.Join(gitDirectory, "refs", "hn", "proposals")
		info, err := os.Stat(proposalRefDirectory)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(proposalRefDirectory, 0o500); err != nil {
			t.Fatal(err)
		}
		publicationErr := cmdProposalRevise([]string{revisions[1].ID, "--base", proposalHeadOID, "--head", blockedHead})
		if err := os.Chmod(proposalRefDirectory, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
		if publicationErr == nil || !strings.Contains(publicationErr.Error(), "was created but its code ref failed") {
			t.Fatalf("partial-publication error = %v", publicationErr)
		}
		events, err := collectEvents()
		if err != nil {
			t.Fatal(err)
		}
		var unpublished *StoredEvent
		for index := range events {
			if events[index].Event.Kind == "proposal.revise" && events[index].Event.Subject == revisions[1].ID {
				unpublished = &events[index]
			}
		}
		if unpublished == nil || !strings.Contains(publicationErr.Error(), unpublished.ID) {
			t.Fatalf("partial-publication error %q lacks created exact revision ID", publicationErr)
		}
		if _, exists, err := proposalHead(unpublished.ID); err != nil || exists {
			t.Fatalf("unpublished revision code ref exists=%t, err=%v", exists, err)
		}
		if ref, exists, err := refValue(proposalRef(predecessor.ID)); err != nil || !exists || ref != predecessorRef {
			t.Fatalf("partial publication changed predecessor ref to %q, exists=%t, err=%v", ref, exists, err)
		}
	}
}

func TestProposalAndReviewSync(t *testing.T) {
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
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "base")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", remote, aliceDirectory)
	mustGit(t, "clone", "-q", remote, bobDirectory)

	if err := os.Chdir(aliceDirectory); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "config", "user.name", "Alice")
	mustGit(t, "config", "user.email", "alice@hn.invalid")
	base := mustGitText(t, "rev-parse", "main")
	mustGit(t, "switch", "-q", "-c", "feature")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "feature")
	head := mustGitText(t, "rev-parse", "HEAD")
	alice, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := nextEvent(alice, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	proposal.Title = "Synced proposal"
	proposal.Base = base
	proposal.Head = head
	storedProposal, err := appendEvent(proposal, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(storedProposal.ID, head); err != nil {
		t.Fatal(err)
	}
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(bobDirectory); err != nil {
		t.Fatal(err)
	}
	bob, _, err := createIdentity("Bob")
	if err != nil {
		t.Fatal(err)
	}
	if err := cmdSync(nil); err != nil {
		t.Fatal(err)
	}
	fetchedHead, exists, err := proposalHead(storedProposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || fetchedHead != head {
		t.Fatalf("fetched proposal head = %q, exists = %v; want %q", fetchedHead, exists, head)
	}
	mustGit(t, "cat-file", "-e", head+"^{commit}")
	review, err := nextEvent(bob, "review.submit")
	if err != nil {
		t.Fatal(err)
	}
	review.Subject = storedProposal.ID
	review.Verdict = "approve"
	storedReview, err := appendEvent(review, bob)
	if err != nil {
		t.Fatal(err)
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
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	reviews := currentReviews(events, storedProposal.ID)
	if len(reviews) != 1 || reviews[0].ID != storedReview.ID {
		t.Fatalf("synced reviews = %#v, want %s", reviews, storedReview.ID)
	}
}
