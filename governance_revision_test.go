package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func appendRevisionRunResult(t *testing.T, runner *Identity, request *StoredEvent) *StoredEvent {
	t.Helper()
	log := []byte("revision pipeline passed\n")
	event, err := nextEvent(runner, "run.result")
	if err != nil {
		t.Fatal(err)
	}
	event.Subject = request.ID
	event.Pipeline = request.Event.Pipeline
	event.Definition = request.Event.Definition
	event.Commit = request.Event.Commit
	event.Outcome = "passed"
	event.DurationMS = 1
	event.Log = eventID(log)
	event.Backend = "host"
	event.Platform = "test/test"
	event.Runner = "hn/test"
	stored, err := appendEventWithAttachments(event, runner, map[string][]byte{"log.txt": log})
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func appendRevisionReview(t *testing.T, reviewer *Identity, proposalID string) *StoredEvent {
	t.Helper()
	event, err := nextEvent(reviewer, "review.submit")
	if err != nil {
		t.Fatal(err)
	}
	event.Subject = proposalID
	event.Verdict = "approve"
	stored, err := appendEvent(event, reviewer)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func appendRevisionDecision(t *testing.T, maintainer *Identity, proposalID, policy string, evidence []string) *StoredEvent {
	t.Helper()
	event, err := nextEvent(maintainer, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	event.Subject = proposalID
	event.Policy = policy
	event.Verdict = "accept"
	event.Evidence = append([]string(nil), evidence...)
	stored, err := appendEvent(event, maintainer)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func appendRevisionMerge(t *testing.T, maintainer *Identity, proposal *StoredEvent, policy string, decision *StoredEvent) *StoredEvent {
	t.Helper()
	event, err := nextEvent(maintainer, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	event.Subject = proposal.ID
	event.Policy = policy
	event.Head = proposal.Event.Head
	event.Commit = proposal.Event.Head
	event.Evidence = []string{decision.ID}
	stored, err := appendEvent(event, maintainer)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func TestRevisionEvidenceAndLineageGovernance(t *testing.T) {
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
	bob := testIdentity(t, "Bob")
	writeTestPolicy(t, root, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{alice.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals:   1,
			RequiredAccepts:     1,
			TrustedReviewers:    []string{bob.Actor},
			AllowAuthorApproval: false,
		},
		Pipelines: map[string]PipelinePolicy{
			"check": {RequiredResults: 1, TrustedRunners: []string{bob.Actor}},
		},
	})
	if err := os.MkdirAll(filepath.Join(root, ".hn", "pipelines"), 0o755); err != nil {
		t.Fatal(err)
	}
	pipeline := []byte("{\"version\":\"hn.pipeline/0\",\"steps\":[{\"name\":\"Check\",\"command\":\"true\"}]}\n")
	if err := os.WriteFile(filepath.Join(root, ".hn", "pipelines", "check.json"), pipeline, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".hn")
	mustGit(t, "commit", "-q", "-m", "base policy and pipeline")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "predecessor head")
	predecessorHead := mustGitText(t, "rev-parse", "HEAD")
	revisedPipeline := []byte("{\"version\":\"hn.pipeline/0\",\"steps\":[{\"name\":\"Revised check\",\"command\":\"true\"}]}\n")
	if err := os.WriteFile(filepath.Join(root, ".hn", "pipelines", "check.json"), revisedPipeline, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".hn/pipelines/check.json")
	mustGit(t, "commit", "-q", "-m", "revision head pipeline")
	revisionHead := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "sibling head")
	siblingHead := mustGitText(t, "rev-parse", "HEAD")

	predecessorEvent, err := nextEvent(alice, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	predecessorEvent.Title = "Exact evidence"
	predecessorEvent.Base = base
	predecessorEvent.Head = predecessorHead
	predecessor, err := appendEvent(predecessorEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(predecessor.ID, predecessorHead); err != nil {
		t.Fatal(err)
	}
	predecessorReview := appendRevisionReview(t, bob, predecessor.ID)
	_, _, definition, err := loadPipeline(predecessorHead, "check")
	if err != nil {
		t.Fatal(err)
	}
	predecessorRequestEvent, err := nextEvent(alice, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	predecessorRequestEvent.Subject = predecessor.ID
	predecessorRequestEvent.Pipeline = "check"
	predecessorRequestEvent.Definition = definition
	predecessorRequestEvent.Commit = predecessorHead
	predecessorRequest, err := appendEvent(predecessorRequestEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	predecessorResult := appendRevisionRunResult(t, bob, predecessorRequest)
	_, _, predecessorPolicy, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	predecessorDecision := appendRevisionDecision(t, alice, predecessor.ID, predecessorPolicy, []string{predecessorReview.ID, predecessorResult.ID})

	revisionEvent, err := nextEvent(alice, "proposal.revise")
	if err != nil {
		t.Fatal(err)
	}
	revisionEvent.Subject = predecessor.ID
	revisionEvent.Base = predecessorHead
	revisionEvent.Head = revisionHead
	revision, err := appendEvent(revisionEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(revision.ID, revisionHead); err != nil {
		t.Fatal(err)
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := evaluateProposal(revision, events)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Ready || evaluation.Accepted || len(evaluation.Approvals) != 0 || len(evaluation.AcceptDecisions) != 0 || len(evaluation.Pipelines) != 1 || len(evaluation.Pipelines[0].Passed) != 0 {
		t.Fatalf("revision inherited predecessor evidence: %#v", evaluation)
	}
	mustGit(t, "update-ref", "-d", proposalRef(revision.ID))
	missingCode, err := evaluateProposal(revision, events)
	if err != nil || missingCode.CodeAvailable || missingCode.CodeMatches || missingCode.Ready {
		t.Fatalf("missing revision code evaluation = %#v, %v", missingCode, err)
	}
	if err := createProposalRef(revision.ID, revisionHead); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "update-ref", proposalRef(revision.ID), predecessorHead, revisionHead)
	mismatchedCode, err := evaluateProposal(revision, events)
	if err != nil || !mismatchedCode.CodeAvailable || mismatchedCode.CodeMatches || mismatchedCode.Ready {
		t.Fatalf("mismatched revision code evaluation = %#v, %v", mismatchedCode, err)
	}
	mustGit(t, "update-ref", proposalRef(revision.ID), revisionHead, predecessorHead)
	_, _, revisionPolicy, err := loadPolicy(revision.Event.Base)
	if err != nil {
		t.Fatal(err)
	}
	hostileDecision := lineageTestEvent("hostile-revision-decision", "proposal.decision", revision.ID)
	hostileDecision.Event.Actor = alice.Actor
	hostileDecision.Event.Policy = revisionPolicy
	hostileDecision.Event.Verdict = "accept"
	hostileDecision.Event.Evidence = []string{predecessorReview.ID, predecessorResult.ID}
	if err := validateEventRelationships(append(append([]StoredEvent(nil), events...), hostileDecision)); err == nil || !strings.Contains(err.Error(), shortID(hostileDecision.ID)) {
		t.Fatalf("predecessor evidence decision error = %v, want hostile decision ID", err)
	}
	hostileMerge := lineageTestEvent("hostile-revision-merge", "proposal.merged", revision.ID)
	hostileMerge.Event.Actor = alice.Actor
	hostileMerge.Event.Policy = revisionPolicy
	hostileMerge.Event.Head = revision.Event.Head
	hostileMerge.Event.Evidence = []string{predecessorDecision.ID}
	if err := validateEventRelationships(append(append([]StoredEvent(nil), events...), hostileMerge)); err == nil || !strings.Contains(err.Error(), shortID(hostileMerge.ID)) {
		t.Fatalf("predecessor acceptance merge error = %v, want hostile merge ID", err)
	}
	hostileRequest := lineageTestEvent("hostile-revision-request", "run.request", revision.ID)
	hostileRequest.Event.Actor = alice.Actor
	hostileRequest.Event.Pipeline = "check"
	hostileRequest.Event.Definition = definition
	hostileRequest.Event.Commit = predecessorHead
	if err := validateEventRelationships(append(append([]StoredEvent(nil), events...), hostileRequest)); err == nil || !strings.Contains(err.Error(), shortID(hostileRequest.ID)) {
		t.Fatalf("predecessor run request reuse error = %v, want hostile request ID", err)
	}

	revisionReview := appendRevisionReview(t, bob, revision.ID)
	if err := cmdRunRequest([]string{revision.ID, "check"}); err != nil {
		t.Fatalf("request revision run: %v", err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	requests := runRequests(events)
	var revisionRequest *StoredEvent
	for index := range requests {
		if requests[index].Event.Subject == revision.ID {
			revisionRequest = &requests[index]
		}
	}
	if revisionRequest == nil {
		t.Fatal("revision run request not found")
	}
	writeActiveTestIdentity(t, bob)
	if _, err := captureTestOutput(t, func() error {
		return cmdRunExecute([]string{revisionRequest.ID, "--backend", "host", "--allow-unsafe-host-execution"})
	}); err != nil {
		t.Fatalf("execute revision run: %v", err)
	}
	writeActiveTestIdentity(t, alice)
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	revisionResults := currentRunResults(events, revisionRequest.ID)
	if len(revisionResults) != 1 {
		t.Fatalf("revision run results = %d, want 1", len(revisionResults))
	}
	revisionResult := revisionResults[0]
	evaluation, err = evaluateProposal(revision, events)
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Ready || len(evaluation.Approvals) != 1 || len(evaluation.Pipelines[0].Passed) != 1 || len(evaluation.AcceptDecisions) != 0 {
		t.Fatalf("fresh revision evidence did not qualify exactly: %#v", evaluation)
	}
	if evaluation.Approvals[0].ID != revisionReview.ID || evaluation.Pipelines[0].Passed[0].ID != revisionResult.ID {
		t.Fatal("revision evaluation selected predecessor evidence")
	}

	err = cmdDecide([]string{predecessor.ID, "--accept"})
	if err == nil || !strings.Contains(err.Error(), revision.ID) || !strings.Contains(err.Error(), "superseded") {
		t.Fatalf("superseded acceptance error = %v, want exact successor %s", err, revision.ID)
	}
	if err := cmdDecide([]string{revision.ID, "--accept"}); err != nil {
		t.Fatalf("accept ready revision: %v", err)
	}
	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	revision, err = resolveEvent(events, revision.ID)
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err = evaluateProposal(revision, events)
	if err != nil || len(evaluation.AcceptDecisions) != 1 {
		t.Fatalf("revision acceptance evaluation = %#v, %v", evaluation, err)
	}
	revisionDecision := evaluation.AcceptDecisions[0]

	siblingEvent, err := nextEvent(alice, "proposal.revise")
	if err != nil {
		t.Fatal(err)
	}
	siblingEvent.Subject = predecessor.ID
	siblingEvent.Base = predecessorHead
	siblingEvent.Head = siblingHead
	sibling, err := appendEvent(siblingEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(sibling.ID, siblingHead); err != nil {
		t.Fatal(err)
	}
	appendRevisionMerge(t, alice, revision, evaluation.PolicyDigest, &revisionDecision)

	err = cmdDecide([]string{sibling.ID, "--accept"})
	if err == nil || !strings.Contains(err.Error(), revision.ID) || !strings.Contains(err.Error(), "lineage") {
		t.Fatalf("lineage-closed acceptance error = %v, want merged revision %s", err, revision.ID)
	}
	if err := cmdDecide([]string{sibling.ID, "--reject", "--body", "historical rejection remains allowed"}); err != nil {
		t.Fatalf("reject lineage-closed sibling: %v", err)
	}
	err = cmdMerge([]string{predecessor.ID})
	if err == nil || !strings.Contains(err.Error(), "superseded") || !strings.Contains(err.Error(), revision.ID) || !strings.Contains(err.Error(), sibling.ID) {
		t.Fatalf("superseded merge error = %v, want both successors", err)
	}

	siblingReview := appendRevisionReview(t, bob, sibling.ID)
	siblingRequestEvent, err := nextEvent(alice, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	siblingRequestEvent.Subject = sibling.ID
	siblingRequestEvent.Pipeline = "check"
	siblingRequestEvent.Definition = revisionRequest.Event.Definition
	siblingRequestEvent.Commit = siblingHead
	siblingRequest, err := appendEvent(siblingRequestEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	siblingResult := appendRevisionRunResult(t, bob, siblingRequest)
	_, _, siblingPolicy, err := loadPolicy(sibling.Event.Base)
	if err != nil {
		t.Fatal(err)
	}
	siblingDecision := appendRevisionDecision(t, alice, sibling.ID, siblingPolicy, []string{siblingReview.ID, siblingResult.ID})
	appendRevisionMerge(t, alice, sibling, siblingPolicy, siblingDecision)

	events, err = collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	predecessor, err = resolveEvent(events, predecessor.ID)
	if err != nil {
		t.Fatal(err)
	}
	conflict, err := evaluateProposal(predecessor, events)
	if err != nil {
		t.Fatal(err)
	}
	if !conflict.Lineage.MergeConflict || len(conflict.Lineage.MergedCandidateIDs) != 2 {
		t.Fatalf("competing merges not preserved: %#v", conflict.Lineage)
	}
	statusOutput, err := captureTestOutput(t, func() error { return cmdProposalStatus(predecessor.ID) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statusOutput, "merge conflict") || !strings.Contains(statusOutput, revision.ID) || !strings.Contains(statusOutput, sibling.ID) {
		t.Fatalf("conflict status lacks exact merged candidates:\n%s", statusOutput)
	}
	for operation, run := range map[string]func() error{
		"accept": func() error { return cmdDecide([]string{revision.ID, "--accept"}) },
		"merge":  func() error { return cmdMerge([]string{revision.ID}) },
	} {
		err := run()
		if err == nil || !strings.Contains(err.Error(), revision.ID) || !strings.Contains(err.Error(), sibling.ID) {
			t.Fatalf("conflicting %s error = %v, want both merged candidate IDs", operation, err)
		}
	}
}

func TestProposalMergeConflictAbortsWithRevisionGuidance(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(root, "conflict.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".hn/policy.json", "conflict.txt")
	mustGit(t, "commit", "-q", "-m", "base")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "switch", "-q", "-c", "proposal")
	if err := os.WriteFile(filepath.Join(root, "conflict.txt"), []byte("proposal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", "conflict.txt")
	mustGit(t, "commit", "-q", "-m", "proposal change")
	head := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "switch", "-q", "main")
	if err := os.WriteFile(filepath.Join(root, "conflict.txt"), []byte("target\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", "conflict.txt")
	mustGit(t, "commit", "-q", "-m", "target change")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")

	proposalEvent, err := nextEvent(alice, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	proposalEvent.Title = "Conflicting proposal"
	proposalEvent.Base = base
	proposalEvent.Head = head
	proposal, err := appendEvent(proposalEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	_, _, policy, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	appendRevisionDecision(t, alice, proposal.ID, policy, nil)

	err = cmdMerge([]string{proposal.ID})
	if err == nil || !strings.Contains(err.Error(), proposal.ID) || !strings.Contains(err.Error(), "hn proposal revise "+proposal.ID) {
		t.Fatalf("merge conflict error = %v, want exact proposal and revision guidance", err)
	}
	if afterHead := mustGitText(t, "rev-parse", "HEAD"); afterHead != beforeHead {
		t.Fatalf("HEAD after conflict = %s, want %s", afterHead, beforeHead)
	}
	if status := mustGitText(t, "status", "--porcelain"); status != "" {
		t.Fatalf("worktree after conflict is dirty:\n%s", status)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	for _, stored := range events {
		if stored.Event.Kind == "proposal.merged" && stored.Event.Subject == proposal.ID {
			t.Fatalf("conflicting merge emitted merge event %s", stored.ID)
		}
	}
}
