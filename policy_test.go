package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyEvidenceDecisionAndMerge(t *testing.T) {
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

	alice := testIdentity(t, "Alice")
	reviewer := testIdentity(t, "Reviewer")
	runner := testIdentity(t, "Runner")
	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@nh.invalid")
	writeTestPolicy(t, root, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{alice.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals:   1,
			RequiredAccepts:     1,
			TrustedReviewers:    []string{reviewer.Actor},
			AllowAuthorApproval: false,
		},
		Pipelines: map[string]PipelinePolicy{
			"test": {RequiredResults: 1, TrustedRunners: []string{runner.Actor}},
		},
	})
	if err := os.MkdirAll(filepath.Join(root, ".nh", "pipelines"), 0o755); err != nil {
		t.Fatal(err)
	}
	pipeline := []byte("{\"version\":\"nh.pipeline/0\",\"steps\":[{\"name\":\"Test\",\"command\":\"true\"}]}\n")
	if err := os.WriteFile(filepath.Join(root, ".nh", "pipelines", "test.json"), pipeline, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".nh")
	mustGit(t, "commit", "-q", "-m", "base policy")
	base := mustGitText(t, "rev-parse", "HEAD")

	// The proposed head changes policy, but this proposal must continue using
	// the policy from its signed base commit.
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
	mustGit(t, "add", ".nh/policy.json")
	mustGit(t, "commit", "-q", "-m", "proposed policy change")
	head := mustGitText(t, "rev-parse", "HEAD")

	proposalEvent := newEvent(alice, "proposal.open", 1, "")
	proposalEvent.Title = "Governed change"
	proposalEvent.Base = base
	proposalEvent.Head = head
	proposal, err := appendEvent(proposalEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Ready || evaluation.Policy.Proposals.RequiredApprovals != 1 {
		t.Fatal("proposal did not use the immutable base policy")
	}

	reviewEvent := newEvent(reviewer, "review.submit", 1, "")
	reviewEvent.Subject = proposal.ID
	reviewEvent.Verdict = "approve"
	if _, err := appendEvent(reviewEvent, reviewer); err != nil {
		t.Fatal(err)
	}
	requestEvent, err := nextEvent(alice, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	_, _, definition, err := loadPipeline(head, "test")
	if err != nil {
		t.Fatal(err)
	}
	requestEvent.Subject = proposal.ID
	requestEvent.Pipeline = "test"
	requestEvent.Definition = definition
	requestEvent.Commit = head
	request, err := appendEvent(requestEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	log := []byte("tests passed\n")
	resultEvent := newEvent(runner, "run.result", 1, "")
	resultEvent.Subject = request.ID
	resultEvent.Pipeline = "test"
	resultEvent.Definition = definition
	resultEvent.Commit = head
	resultEvent.Outcome = "passed"
	resultEvent.DurationMS = 1
	resultEvent.Log = eventID(log)
	resultEvent.Backend = "sandbox"
	resultEvent.Platform = "test/test"
	resultEvent.Runner = "nh/test"
	if _, err := appendEventWithAttachments(resultEvent, runner, map[string][]byte{"log.txt": log}); err != nil {
		t.Fatal(err)
	}

	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	proposal, err = resolveEvent(events, proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err = evaluateProposal(proposal, events)
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Ready || len(evaluation.Evidence) != 2 {
		t.Fatalf("evaluation ready=%v evidence=%d, want true and 2", evaluation.Ready, len(evaluation.Evidence))
	}

	decisionEvent, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	decisionEvent.Subject = proposal.ID
	decisionEvent.Policy = evaluation.PolicyDigest
	decisionEvent.Verdict = "accept"
	decisionEvent.Evidence = evaluation.Evidence
	decision, err := appendEvent(decisionEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	proposal, _ = resolveEvent(events, proposal.ID)
	evaluation, err = evaluateProposal(proposal, events)
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Accepted {
		t.Fatal("valid signed decision did not accept proposal")
	}

	mergeEvent, err := nextEvent(alice, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	mergeEvent.Subject = proposal.ID
	mergeEvent.Policy = evaluation.PolicyDigest
	mergeEvent.Head = head
	mergeEvent.Commit = head
	mergeEvent.Evidence = []string{decision.ID}
	if _, err := appendEvent(mergeEvent, alice); err != nil {
		t.Fatal(err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	proposal, _ = resolveEvent(events, proposal.ID)
	evaluation, err = evaluateProposal(proposal, events)
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Merged || proposalStatus(evaluation) != "merged" {
		t.Fatal("valid merge evidence did not produce merged status")
	}
}

func writeTestPolicy(t *testing.T, root string, policy PolicyDocument) {
	t.Helper()
	encoded, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Join(root, ".nh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".nh", "policy.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}
