package main

import (
	"os"
	"path/filepath"
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
	mustGit(t, "config", "user.email", "test@nh.invalid")
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
	mustGit(t, "-C", seed, "config", "user.email", "seed@nh.invalid")
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "base")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", remote, aliceDirectory)
	mustGit(t, "clone", "-q", remote, bobDirectory)

	if err := os.Chdir(aliceDirectory); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "config", "user.name", "Alice")
	mustGit(t, "config", "user.email", "alice@nh.invalid")
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
