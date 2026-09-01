package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShallowRepositoryAllowsPresentExactDependency(t *testing.T) {
	_, clone, _, head := depthOneRepository(t)
	withTestDirectory(t, clone)

	if err := requireExactDependency(exactDependency{
		Operation: "proposal inspection",
		Kind:      shallowBaseCommit,
		MissingID: head,
	}); err != nil {
		t.Fatalf("present dependency was rejected merely because the repository is shallow: %v", err)
	}
}

func TestShallowDependencyGapNamesExactMissingObjectAndRecovery(t *testing.T) {
	root, clone, base, _ := depthOneRepository(t)
	withTestDirectory(t, clone)
	mustGit(t, "remote", "set-url", "origin", "file://"+filepath.Join(root, "remote.git"))

	err := requireExactDependency(exactDependency{
		Operation:   "proposal decision",
		Kind:        shallowBaseCommit,
		MissingID:   base,
		OwnerKind:   replicationProposal,
		OwnerID:     "sha256:" + strings.Repeat("a", 64),
		Remote:      "origin",
		RequiredRef: "refs/hn/proposals/" + strings.Repeat("a", 64),
	})
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) {
		t.Fatalf("error = %v, want ShallowDependencyGap", err)
	}
	text := err.Error()
	for _, exact := range []string{base, "base commit", "origin", "refs/hn/proposals/" + strings.Repeat("a", 64), "hn sync origin --recover-shallow"} {
		if !strings.Contains(text, exact) {
			t.Fatalf("gap diagnostic omitted %q: %s", exact, text)
		}
	}
	if strings.Contains(text, root) || strings.Contains(text, clone) {
		t.Fatalf("gap diagnostic leaked a local path: %s", text)
	}
}

func TestCompleteRepositoryPreservesOrdinaryMissingObjectError(t *testing.T) {
	root := inReplicationTestRepository(t)
	missing := strings.Repeat("f", 40)
	err := requireExactDependency(exactDependency{
		Operation: "proposal inspection",
		Kind:      shallowBaseCommit,
		MissingID: missing,
	})
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) {
		t.Fatalf("complete repository error = %v, want ordinary missing-object error", err)
	}
	if strings.Contains(err.Error(), root) {
		t.Fatalf("ordinary error leaked repository path: %v", err)
	}
}

func TestShallowClassifierPreservesPresentWrongType(t *testing.T) {
	_, clone, _, _ := depthOneRepository(t)
	withTestDirectory(t, clone)
	blobOutput, err := gitInput([]byte("present blob"), nil, "hash-object", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	blob := strings.TrimSpace(string(blobOutput))
	err = requireExactDependency(exactDependency{
		Operation: "proposal merge", Kind: shallowMergeAncestor,
		MissingID: blob, ObjectType: "commit",
	})
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "type blob, want commit") {
		t.Fatalf("wrong-type error = %v, want ordinary type mismatch", err)
	}
	err = cmdPolicyShow([]string{blob})
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "type blob, want commit") {
		t.Fatalf("policy command wrong-type error = %v, want ordinary type mismatch", err)
	}
}

func TestShallowClassifierPreservesUnrelatedProbeFailure(t *testing.T) {
	_, clone, base, _ := depthOneRepository(t)
	withTestDirectory(t, clone)
	previous := shallowObjectProbe
	shallowObjectProbe = func(string) (gitObjectProbe, error) {
		return gitObjectProbe{}, errors.New("injected repository permission failure")
	}
	t.Cleanup(func() { shallowObjectProbe = previous })
	err := requireExactDependency(exactDependency{
		Operation: "policy show", Kind: shallowBaseCommit,
		MissingID: base, ObjectType: "commit",
	})
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "permission failure") {
		t.Fatalf("probe failure = %v, want original unrelated error", err)
	}
}

func TestShallowAcceptedFactsPreservePresentMalformedSignedEvent(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "visible")
	withTestDirectory(t, publisher)
	actor := testIdentity(t, "Malformed Event Actor")
	event := newEvent(actor, "issue.open", 1, "")
	event.Title = "signed before corruption"
	stored, err := appendEvent(event, actor)
	if err != nil {
		t.Fatal(err)
	}
	eventBlob := mustGitText(t, "rev-parse", stored.Commit+":event.json")
	badSignature, err := gitInput([]byte("not-a-valid-signature"), nil, "hash-object", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	treeInput := "100644 blob " + eventBlob + "\tevent.json\n100644 blob " + strings.TrimSpace(string(badSignature)) + "\tsignature\n"
	badTree, err := gitInput([]byte(treeInput), nil, "mktree")
	if err != nil {
		t.Fatal(err)
	}
	badCommit, err := gitInput(nil, nil, "commit-tree", strings.TrimSpace(string(badTree)), "-m", "malformed signed event")
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "update-ref", actorRef(actor.Actor), strings.TrimSpace(string(badCommit)))
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	mustGit(t, "push", "-q", "origin", actorRef(actor.Actor)+":"+actorRef(actor.Actor))
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(actor.Actor)+":"+acceptedActorRef("origin", actor.Actor))
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	beforeShallow := readShallowBytes(t)
	err = guardShallowEventClosure("proposal status")
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("malformed present signed event = %v, want ordinary verification error", err)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
		t.Fatalf("malformed event changed refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
	if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
		t.Fatalf("malformed event changed shallow bytes:\nbefore=%q\nafter=%q", beforeShallow, got)
	}
}

func TestPolicyCommandsClassifyMissingExactCommitsAndPreserveMalformedPolicy(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", seed)
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	if err := os.MkdirAll(filepath.Join(seed, ".hn"), 0o755); err != nil {
		t.Fatal(err)
	}
	policy := `{"version":"hn.policy/0","maintainers":["` + strings.Repeat("a", 64) + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`
	if err := os.WriteFile(filepath.Join(seed, ".hn", "policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "-C", seed, "add", ".hn/policy.json")
	mustGit(t, "-C", seed, "commit", "-q", "-m", "policy base")
	missingBase := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	missingPolicyBlob := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", missingBase+":.hn/policy.json")))
	visiblePolicy := strings.Replace(policy, strings.Repeat("a", 64), strings.Repeat("c", 64), 1)
	if err := os.WriteFile(filepath.Join(seed, ".hn", "policy.json"), []byte(visiblePolicy), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "-C", seed, "add", ".hn/policy.json")
	mustGit(t, "-C", seed, "commit", "-q", "-m", "visible policy")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)

	for _, test := range []struct {
		name string
		kind ShallowDependencyKind
		run  func() error
	}{
		{"show", shallowBaseCommit, func() error { return cmdPolicyShow([]string{missingBase}) }},
		{"check-head", shallowProposalCodeRef, func() error {
			return cmdPolicyCheck([]string{"--base", "HEAD", "--head", missingBase})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			var gap *ShallowDependencyGap
			if !errors.As(err, &gap) || gap.Kind != test.kind || gap.MissingID != missingBase {
				t.Fatalf("error = %v, want %s gap %s", err, test.kind, missingBase)
			}
		})
	}

	copyCommitAndTreesWithoutBlobs(t, seed, missingBase)
	err := cmdPolicyShow([]string{missingBase})
	var policyGap *ShallowDependencyGap
	if !errors.As(err, &policyGap) || policyGap.Kind != shallowPolicyBlob || policyGap.MissingID != missingPolicyBlob {
		t.Fatalf("missing policy blob = %v, want policy gap %s", err, missingPolicyBlob)
	}

	malformed := filepath.Join(root, "malformed")
	mustGit(t, "clone", "-q", clone, malformed)
	if err := os.MkdirAll(filepath.Join(malformed, ".hn"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(malformed, ".hn", "policy.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "-C", malformed, "config", "user.name", "Malformed")
	mustGit(t, "-C", malformed, "config", "user.email", "malformed@hn.invalid")
	mustGit(t, "-C", malformed, "add", ".hn/policy.json")
	mustGit(t, "-C", malformed, "commit", "-q", "-m", "malformed policy")
	withTestDirectory(t, malformed)
	err = cmdPolicyShow([]string{"HEAD"})
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "parse policy") {
		t.Fatalf("malformed present policy = %v, want ordinary parse error", err)
	}
}

func TestProposalCommandsClassifyRealMissingBaseCodeAndMergeAncestry(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", seed)
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	if err := os.MkdirAll(filepath.Join(seed, ".hn"), 0o755); err != nil {
		t.Fatal(err)
	}
	policy := `{"version":"hn.policy/0","maintainers":["` + strings.Repeat("b", 64) + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`
	if err := os.WriteFile(filepath.Join(seed, ".hn", "policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "-C", seed, "add", ".hn/policy.json")
	mustGit(t, "-C", seed, "commit", "-q", "-m", "missing historical base")
	missingCommit := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "visible candidate head")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	visible := mustGitText(t, "rev-parse", "HEAD")
	actor := testIdentity(t, "Local Candidate Author")
	missingBaseEvent := newEvent(actor, "proposal.open", 1, "")
	missingBaseEvent.Title = "missing base"
	missingBaseEvent.Base = missingCommit
	missingBaseEvent.Head = visible
	missingBase, err := appendEvent(missingBaseEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(missingBase.ID, visible); err != nil {
		t.Fatal(err)
	}
	missingCodeEvent, err := nextEvent(actor, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	missingCodeEvent.Title = "missing code"
	missingCodeEvent.Base = visible
	missingCodeEvent.Head = missingCommit
	missingCode, err := appendEvent(missingCodeEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")
	for _, test := range []struct {
		name string
		kind ShallowDependencyKind
		run  func() error
	}{
		{"status-base", shallowBaseCommit, func() error { return cmdProposalStatus(missingBase.ID) }},
		{"review-code", shallowProposalCodeRef, func() error { return cmdReview([]string{missingCode.ID, "--approve"}) }},
		{"merge-ancestry", shallowMergeAncestor, func() error { return cmdMerge([]string{missingBase.ID}) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			var gap *ShallowDependencyGap
			if !errors.As(err, &gap) || gap.Kind != test.kind || gap.MissingID != missingCommit {
				t.Fatalf("error = %v, want %s gap %s", err, test.kind, missingCommit)
			}
			if gap.OwnerKind != replicationProposal || gap.OwnerID == "" || !strings.Contains(gap.RequiredRef, "refs/hn/proposals/") {
				t.Fatalf("object gap lacks candidate supplier: %#v", gap)
			}
			if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
				t.Fatalf("command advanced refs:\nbefore=%s\nafter=%s", beforeRefs, got)
			}
			if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
				t.Fatalf("command advanced HEAD from %s to %s", beforeHead, got)
			}
		})
	}
}

func TestExactMergeAncestryIgnoresShallowMarkerWhenParentsArePresent(t *testing.T) {
	root, clone, base, head := depthOneRepository(t)
	withTestDirectory(t, clone)
	copyCommitAndTreesWithoutBlobs(t, filepath.Join(root, "seed"), base)
	proposal := &StoredEvent{ID: "sha256:" + strings.Repeat("1", 64), Event: Event{Base: base}}
	if err := guardMergeAncestry("proposal merge", proposal, head); err != nil {
		t.Fatalf("complete exact parent graph was blocked by shallow marker: %v", err)
	}
}

func TestExactMergeAncestryNamesFirstMissingIntermediateParent(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", seed)
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "base")
	base := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "hidden intermediate")
	intermediate := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "visible head")
	head := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	copyCommitAndTreesWithoutBlobs(t, seed, base)
	proposalID := "sha256:" + strings.Repeat("2", 64)
	proposal := &StoredEvent{ID: proposalID, Event: Event{Base: base}}
	err := guardMergeAncestry("proposal merge", proposal, head)
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) || gap.Kind != shallowMergeAncestor || gap.MissingID != intermediate {
		t.Fatalf("ancestry = %v, want first missing parent %s", err, intermediate)
	}
	if gap.OwnerID != proposalID || gap.RequiredRef != proposalRef(proposalID) {
		t.Fatalf("ancestry gap has inaccurate selected supplier: %#v", gap)
	}
}

func TestExactMergeAncestryPreservesMalformedCommitError(t *testing.T) {
	root, clone, base, _ := depthOneRepository(t)
	withTestDirectory(t, clone)
	copyCommitAndTreesWithoutBlobs(t, filepath.Join(root, "seed"), base)
	tree := mustGitText(t, "rev-parse", "HEAD^{tree}")
	malformedPayload := []byte("tree " + tree + "\nparent not-an-object-id\n\nmalformed\n")
	output, err := gitInput(malformedPayload, nil, "hash-object", "--literally", "-w", "-t", "commit", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	malformed := strings.TrimSpace(string(output))
	proposal := &StoredEvent{ID: "sha256:" + strings.Repeat("3", 64), Event: Event{Base: base}}
	err = guardMergeAncestry("proposal merge", proposal, malformed)
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "malformed commit parent") {
		t.Fatalf("malformed ancestry = %v, want ordinary malformed-commit error", err)
	}
}

func TestExactMergeAncestryPreservesWrongTypeParentError(t *testing.T) {
	root, clone, base, _ := depthOneRepository(t)
	withTestDirectory(t, clone)
	copyCommitAndTreesWithoutBlobs(t, filepath.Join(root, "seed"), base)
	parentOutput, err := gitInput([]byte("not a commit"), nil, "hash-object", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	wrongParent := strings.TrimSpace(string(parentOutput))
	tree := mustGitText(t, "rev-parse", "HEAD^{tree}")
	payload := []byte("tree " + tree + "\nparent " + wrongParent + "\n\nwrong parent type\n")
	commitOutput, err := gitInput(payload, nil, "hash-object", "--literally", "-w", "-t", "commit", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	current := strings.TrimSpace(string(commitOutput))
	proposal := &StoredEvent{ID: "sha256:" + strings.Repeat("4", 64), Event: Event{Base: base}}
	err = guardMergeAncestry("proposal merge", proposal, current)
	var gap *ShallowDependencyGap
	if err == nil || errors.As(err, &gap) || !strings.Contains(err.Error(), "type blob, want commit") {
		t.Fatalf("wrong-type ancestry = %v, want ordinary type error", err)
	}
}

func TestRunRequestClassifiesRealMissingPipelineBlobBeforeEventAppend(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", seed)
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	runner := strings.Repeat("d", 64)
	writeTestPolicy(t, seed, PolicyDocument{
		Version: policyVersion, Maintainers: []string{runner},
		Proposals: ProposalPolicy{RequiredAccepts: 1},
		Pipelines: map[string]PipelinePolicy{"check": {RequiredResults: 1, TrustedRunners: []string{runner}}},
	})
	mustGit(t, "-C", seed, "add", ".hn/policy.json")
	mustGit(t, "-C", seed, "commit", "-q", "-m", "visible policy base")
	base := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "-C", seed, "switch", "-q", "-c", "candidate")
	writeTestPipeline(t, seed)
	mustGit(t, "-C", seed, "add", ".hn/pipelines/test.json")
	mustGit(t, "-C", seed, "commit", "-q", "-m", "candidate pipeline")
	head := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	pipelineBlob := strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", head+":.hn/pipelines/test.json")))
	mustGit(t, "-C", seed, "switch", "-q", "main")
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	copyCommitAndTreesWithoutBlobs(t, seed, head)
	actor := testIdentity(t, "Pipeline Gap Author")
	proposalEvent := newEvent(actor, "proposal.open", 1, "")
	proposalEvent.Title, proposalEvent.Base, proposalEvent.Head = "missing pipeline blob", base, head
	proposal, err := appendEvent(proposalEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	err = cmdRunRequest([]string{proposal.ID, "test"})
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) || gap.Kind != shallowPipelineDefinition || gap.MissingID != pipelineBlob {
		t.Fatalf("run request = %v, want pipeline gap %s", err, pipelineBlob)
	}
	if gap.OwnerKind != replicationProposal || gap.OwnerID != proposal.ID || gap.RequiredRef != proposalRef(proposal.ID) {
		t.Fatalf("pipeline gap has inaccurate candidate supplier: %#v", gap)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
		t.Fatalf("pipeline gap appended an event or changed a ref:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
}

func TestShallowActorPredecessorGapUsesSignedFullEventID(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	withTestDirectory(t, publisher)
	actor := testIdentity(t, "Shallow Actor")
	first := newEvent(actor, "issue.open", 1, "")
	first.Title = "first"
	storedFirst, err := appendEvent(first, actor)
	if err != nil {
		t.Fatal(err)
	}
	second, err := nextEvent(actor, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	second.Title = "second"
	storedSecond, err := appendEvent(second, actor)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	mustGit(t, "push", "-q", "origin", actorRef(actor.Actor)+":"+actorRef(actor.Actor))
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)

	withTestDirectory(t, clone)
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(actor.Actor)+":"+acceptedActorRef("origin", actor.Actor))
	err = guardShallowEventClosure("proposal status")
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) {
		t.Fatalf("error = %v, want predecessor gap", err)
	}
	if gap.Kind != shallowActorPredecessor || gap.MissingID != storedFirst.ID || gap.OwnerID != actor.Actor {
		t.Fatalf("gap = %#v, want predecessor %s owned by %s", gap, storedFirst.ID, actor.Actor)
	}
	for _, exact := range []string{storedFirst.ID, actor.Actor, actorRef(actor.Actor), "hn sync origin --recover-shallow"} {
		if !strings.Contains(err.Error(), exact) {
			t.Fatalf("diagnostic omitted %q: %v", exact, err)
		}
	}
	if storedSecond.Event.Previous != storedFirst.ID {
		t.Fatal("fixture did not bind the exact predecessor")
	}
}

func TestSelectedShallowRecoveryReusesSavedSelectionAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	withTestDirectory(t, publisher)
	selected := testIdentity(t, "Selected Actor")
	first := newEvent(selected, "issue.open", 1, "")
	first.Title = "first"
	storedFirst, err := appendEvent(first, selected)
	if err != nil {
		t.Fatal(err)
	}
	second, err := nextEvent(selected, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	second.Title = "second"
	storedSecond, err := appendEvent(second, selected)
	if err != nil {
		t.Fatal(err)
	}
	unselected := testIdentity(t, "Unselected Actor")
	unselectedEvent := newEvent(unselected, "issue.open", 1, "")
	unselectedEvent.Title = "must remain outside accepted projection"
	unselectedStored, err := appendEvent(unselectedEvent, unselected)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	for _, ref := range []string{actorRef(selected.Actor), actorRef(unselected.Actor)} {
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	mustGit(t, "config", "user.name", "Receiver")
	mustGit(t, "config", "user.email", "receiver@hn.invalid")
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(selected.Actor)+":"+acceptedActorRef("origin", selected.Actor))
	selection := ReplicationSelection{
		Version: replicationSelectionVersion,
		Remote:  "origin",
		Actors:  []string{selected.Actor, unselected.Actor},
		Budgets: defaultReplicationBudgets(),
	}
	if err := saveReplicationSelection(selection); err != nil {
		t.Fatal(err)
	}
	selectionPath, err := replicationSelectionPath("origin")
	if err != nil {
		t.Fatal(err)
	}
	selectionBefore, err := os.ReadFile(selectionPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := recoverSelectedShallow("origin"); err != nil {
		t.Fatal(err)
	}
	if err := guardShallowEventClosure("proposal status"); err != nil {
		t.Fatalf("fresh verification after recovery failed: %v", err)
	}
	if shallow, err := repositoryIsShallow(); err != nil || !shallow {
		t.Fatalf("selected actor recovery globally unshallowed the repository: shallow=%t err=%v", shallow, err)
	}
	assertRefValue(t, acceptedActorRef("origin", selected.Actor), storedSecond.Commit)
	assertRefAbsent(t, acceptedActorRef("origin", unselected.Actor))
	if _, err := gitOutput("cat-file", "-e", storedFirst.Commit+"^{commit}"); err != nil {
		t.Fatalf("selected predecessor was not recovered: %v", err)
	}
	if _, err := gitOutput("cat-file", "-e", unselectedStored.Commit+"^{commit}"); err == nil {
		t.Fatal("unrelated already-selected actor object entered the repository")
	}
	selectionAfter, err := os.ReadFile(selectionPath)
	if err != nil || !bytes.Equal(selectionBefore, selectionAfter) {
		t.Fatalf("narrow recovery changed saved multi-selection bytes: err=%v", err)
	}
	before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	if err := recoverSelectedShallow("origin"); err != nil {
		t.Fatalf("idempotent retry failed: %v", err)
	}
	after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	if after != before {
		t.Fatalf("idempotent retry changed accepted refs:\nbefore=%s\nafter=%s", before, after)
	}
}

func TestProductionRecoveryRerunsCompleteProposalStatusScope(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	withTestDirectory(t, publisher)
	author := testIdentity(t, "Scoped Recovery Author")
	writeTestPolicy(t, publisher, PolicyDocument{
		Version: policyVersion, Maintainers: []string{author.Actor},
		Proposals: ProposalPolicy{RequiredAccepts: 1},
		Pipelines: map[string]PipelinePolicy{
			"check": {RequiredResults: 1, TrustedRunners: []string{author.Actor}},
		},
	})
	mustGit(t, "add", ".hn/policy.json")
	mustGit(t, "commit", "-q", "-m", "base policy requires absent pipeline")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "candidate head")
	head := mustGitText(t, "rev-parse", "HEAD")
	proposalEvent := newEvent(author, "proposal.open", 1, "")
	proposalEvent.Title, proposalEvent.Base, proposalEvent.Head = "scoped recovery", base, head
	proposal, err := appendEvent(proposalEvent, author)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	for _, ref := range []string{actorRef(author.Actor), proposalRef(proposal.ID)} {
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	mustGit(t, "fetch", "-q", "origin", actorRef(author.Actor)+":"+acceptedActorRef("origin", author.Actor))
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", proposalRef(proposal.ID)+":"+acceptedProposalRef("origin", proposal.ID))
	if err := saveReplicationSelection(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin",
		Proposals: []string{proposal.ID}, Budgets: defaultReplicationBudgets(),
	}); err != nil {
		t.Fatal(err)
	}
	err = cmdProposalStatus(proposal.ID)
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) || gap.Kind != shallowBaseCommit || gap.MissingID != base {
		t.Fatalf("initial status = %v, want base gap %s", err, base)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	beforeShallow := readShallowBytes(t)
	selectionPath, pathErr := replicationSelectionPath("origin")
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	beforeSelection, readErr := os.ReadFile(selectionPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	replicationAfterCopyHook = func() error { return errors.New("injected proposal-base post-copy failure") }
	interruptionErr := cmdSync([]string{"origin", "--recover-shallow"})
	replicationAfterCopyHook = nil
	t.Cleanup(resetReplicationInterruptionHooks)
	if interruptionErr == nil || !strings.Contains(interruptionErr.Error(), "unreferenced object residue may remain") {
		t.Fatalf("post-copy proposal recovery = %v, want explicit residue contract", interruptionErr)
	}
	if _, objectErr := gitOutput("cat-file", "-e", base+"^{commit}"); objectErr != nil {
		t.Fatalf("post-copy fixture did not leave the unavoidable raw object residue: %v", objectErr)
	}
	visibilityErr := cmdProposalStatus(proposal.ID)
	var visibilityGap *ShallowDependencyGap
	if !errors.As(visibilityErr, &visibilityGap) || visibilityGap.Kind != shallowBaseCommit || visibilityGap.MissingID != base {
		t.Fatalf("uncommitted raw base changed trust visibility: %v", visibilityErr)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes"); got != beforeRefs {
		t.Fatalf("post-copy failure changed accepted refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
	if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
		t.Fatalf("post-copy failure changed shallow bytes:\nbefore=%q\nafter=%q", beforeShallow, got)
	}
	if afterSelection, readErr := os.ReadFile(selectionPath); readErr != nil || !bytes.Equal(afterSelection, beforeSelection) {
		t.Fatalf("post-copy failure changed selection bytes: %v", readErr)
	}
	err = cmdSync([]string{"origin", "--recover-shallow"})
	if err == nil || !strings.Contains(err.Error(), "pipeline \"check\" does not exist") {
		t.Fatalf("production recovery = %v, want fresh full-scope pipeline failure", err)
	}
	gapPath, pathErr := shallowGapPath()
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if _, statErr := os.Stat(gapPath); statErr != nil {
		t.Fatalf("full-scope failure cleared durable retry context: %v", statErr)
	}
	afterFirst := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	if afterFirst != beforeRefs {
		t.Fatalf("same-head base recovery changed accepted ref values:\nbefore=%s\nafter=%s", beforeRefs, afterFirst)
	}
	err = cmdSync([]string{"origin", "--recover-shallow"})
	if err == nil || !strings.Contains(err.Error(), "pipeline \"check\" does not exist") {
		t.Fatalf("idempotent production retry = %v, want same fresh full-scope failure", err)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes"); got != afterFirst {
		t.Fatalf("already-satisfied primitive caused another promotion:\nbefore=%s\nafter=%s", afterFirst, got)
	}
}

func TestShallowGapStopsTrustSensitiveCommandsBeforeAdvancement(t *testing.T) {
	fixture := setupShallowRecoveryFixture(t, defaultReplicationBudgets())
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")
	missing := "sha256:" + strings.Repeat("d", 64)
	commands := []struct {
		name string
		run  func() error
	}{
		{"proposal status", func() error { return cmdProposal([]string{"status", missing}) }},
		{"review", func() error { return cmdReview([]string{missing, "--approve"}) }},
		{"run request", func() error { return cmdRunRequest([]string{missing, "check"}) }},
		{"decision", func() error { return cmdDecide([]string{missing, "--accept"}) }},
		{"merge", func() error { return cmdMerge([]string{missing}) }},
	}
	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			err := command.run()
			var gap *ShallowDependencyGap
			if !errors.As(err, &gap) || gap.Kind != shallowActorPredecessor || gap.MissingID != fixture.first.ID {
				t.Fatalf("error = %v, want exact predecessor gap %s", err, fixture.first.ID)
			}
			if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
				t.Fatalf("command advanced refs:\nbefore=%s\nafter=%s", beforeRefs, got)
			}
			if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
				t.Fatalf("command advanced HEAD from %s to %s", beforeHead, got)
			}
		})
	}
}

func TestShallowCommandsClassifyRealUnfetchedEventFacts(t *testing.T) {
	fixture := setupUnfetchedEventFixture(t)
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")
	tests := []struct {
		name    string
		missing string
		kind    ShallowDependencyKind
		run     func() error
	}{
		{"proposal-show", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdProposalShow(fixture.proposal.ID) }},
		{"proposal-status", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdProposalStatus(fixture.proposal.ID) }},
		{"review", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdReview([]string{fixture.proposal.ID, "--approve"}) }},
		{"run-request", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdRunRequest([]string{fixture.proposal.ID, "check"}) }},
		{"decision", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdDecide([]string{fixture.proposal.ID, "--accept"}) }},
		{"merge", fixture.proposal.ID, shallowCandidateEvent, func() error { return cmdMerge([]string{fixture.proposal.ID}) }},
		{"run-show", fixture.request.ID, shallowRunRequest, func() error { return cmdRunShow(fixture.request.ID) }},
		{"run-execute", fixture.request.ID, shallowRunRequest, func() error {
			return cmdRunExecute([]string{fixture.request.ID, "--backend", "host", "--allow-unsafe-host-execution"})
		}},
		{"run-logs", fixture.result.ID, shallowRunResult, func() error { return cmdRunLogs(fixture.result.ID) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			var gap *ShallowDependencyGap
			if !errors.As(err, &gap) || gap.Kind != test.kind || gap.MissingID != test.missing {
				t.Fatalf("error = %v, want %s gap %s", err, test.kind, test.missing)
			}
			if !strings.Contains(gap.Recovery, "actor history") || strings.Contains(gap.RequiredRef, "refs/hn/proposals") {
				t.Fatalf("event gap guessed a candidate-code supplier: %#v", gap)
			}
			if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
				t.Fatalf("command advanced refs:\nbefore=%s\nafter=%s", beforeRefs, got)
			}
			if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
				t.Fatalf("command advanced HEAD from %s to %s", beforeHead, got)
			}
		})
	}
}

type unfetchedEventFixture struct {
	proposal *StoredEvent
	request  *StoredEvent
	result   *StoredEvent
}

func setupUnfetchedEventFixture(t *testing.T) unfetchedEventFixture {
	t.Helper()
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	base := strings.TrimSpace(string(mustGitOutputAt(t, publisher, "rev-parse", "HEAD")))
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "head")
	head := strings.TrimSpace(string(mustGitOutputAt(t, publisher, "rev-parse", "HEAD")))
	withTestDirectory(t, publisher)
	actor := testIdentity(t, "Remote Event Supplier")
	proposalEvent := newEvent(actor, "proposal.open", 1, "")
	proposalEvent.Title = "unfetched candidate"
	proposalEvent.Base = base
	proposalEvent.Head = head
	proposal, err := appendEvent(proposalEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	requestEvent, err := nextEvent(actor, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	requestEvent.Subject = proposal.ID
	requestEvent.Pipeline = "check"
	requestEvent.Definition = eventID([]byte("remote definition"))
	requestEvent.Commit = head
	request, err := appendEvent(requestEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	log := []byte("remote log")
	resultEvent, err := nextEvent(actor, "run.result")
	if err != nil {
		t.Fatal(err)
	}
	resultEvent.Subject = request.ID
	resultEvent.Pipeline = request.Event.Pipeline
	resultEvent.Definition = request.Event.Definition
	resultEvent.Commit = head
	resultEvent.Outcome = "passed"
	resultEvent.Log = eventID(log)
	resultEvent.Backend = "sandbox"
	resultEvent.Platform = "test/test"
	resultEvent.Runner = "hn/test"
	result, err := appendEventWithAttachments(resultEvent, actor, map[string][]byte{"log.txt": log})
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	mustGit(t, "push", "-q", "origin", actorRef(actor.Actor)+":"+actorRef(actor.Actor))
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	return unfetchedEventFixture{proposal: proposal, request: request, result: result}
}

func TestMergeClassifiesRealMissingDecisionEvidenceBeforeAdvancement(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	base := strings.TrimSpace(string(mustGitOutputAt(t, publisher, "rev-parse", "HEAD")))
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "candidate")
	head := strings.TrimSpace(string(mustGitOutputAt(t, publisher, "rev-parse", "HEAD")))
	withTestDirectory(t, publisher)
	proposalActor := testIdentity(t, "Decision Proposal Actor")
	proposalEvent := newEvent(proposalActor, "proposal.open", 1, "")
	proposalEvent.Title, proposalEvent.Base, proposalEvent.Head = "missing decision evidence", base, head
	proposal, err := appendEvent(proposalEvent, proposalActor)
	if err != nil {
		t.Fatal(err)
	}
	decisionActor := testIdentity(t, "Unfetched Decision Actor")
	decisionEvent := newEvent(decisionActor, "proposal.decision", 1, "")
	decisionEvent.Subject, decisionEvent.Verdict = proposal.ID, "accept"
	decisionEvent.Policy = eventID([]byte("decision fixture policy"))
	decision, err := appendEvent(decisionEvent, decisionActor)
	if err != nil {
		t.Fatal(err)
	}
	mergeActor := testIdentity(t, "Merge Evidence Actor")
	mergedEvent := newEvent(mergeActor, "proposal.merged", 1, "")
	mergedEvent.Subject = proposal.ID
	mergedEvent.Policy = decisionEvent.Policy
	mergedEvent.Head, mergedEvent.Commit = head, head
	mergedEvent.Evidence = []string{decision.ID}
	if _, err := appendEvent(mergedEvent, mergeActor); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	for _, identity := range []*Identity{proposalActor, decisionActor, mergeActor} {
		ref := actorRef(identity.Actor)
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	for _, identity := range []*Identity{proposalActor, mergeActor} {
		mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(identity.Actor)+":"+acceptedActorRef("origin", identity.Actor))
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")
	err = cmdMerge([]string{proposal.ID})
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) || gap.Kind != shallowDecision || gap.MissingID != decision.ID {
		t.Fatalf("merge = %v, want decision gap %s", err, decision.ID)
	}
	if !strings.Contains(gap.Recovery, "actor history") || gap.OwnerID != "" || gap.RequiredRef != "" {
		t.Fatalf("unknown decision supplier guidance guessed a ref: %#v", gap)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); got != beforeRefs {
		t.Fatalf("missing decision changed refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
	if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
		t.Fatalf("missing decision moved HEAD from %s to %s", beforeHead, got)
	}
}

func TestUnselectedShallowSupplierRequiresExactSelectionAction(t *testing.T) {
	fixture := setupShallowRecoveryFixture(t, defaultReplicationBudgets())
	if err := recoverSelectedShallow("origin"); err != nil {
		t.Fatalf("recover selected fixture actor: %v", err)
	}
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(fixture.unselected.Actor)+":"+acceptedActorRef("origin", fixture.unselected.Actor))
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	err := guardShallowEventClosure("proposal status")
	var gap *ShallowDependencyGap
	if !errors.As(err, &gap) || gap.MissingID != fixture.unselectedFirst.ID || gap.OwnerID != fixture.unselected.Actor {
		t.Fatalf("error = %v, want real unselected predecessor %s", err, fixture.unselectedFirst.ID)
	}
	for _, exact := range []string{"hn replication select origin --actor " + fixture.unselected.Actor, "preserve existing selectors and budgets", actorRef(fixture.unselected.Actor)} {
		if !strings.Contains(err.Error(), exact) {
			t.Fatalf("unselected diagnostic omitted %q: %v", exact, err)
		}
	}
	if strings.Contains(err.Error(), "--actor "+fixture.actor.Actor) {
		t.Fatalf("diagnostic substituted selected actor: %v", err)
	}
	if err := recoverSelectedShallow("origin"); err == nil || !strings.Contains(err.Error(), "is not in the saved exact selection") {
		t.Fatalf("unselected recovery = %v, want exact-selection block", err)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes"); got != beforeRefs {
		t.Fatalf("unselected recovery changed refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
}

func TestShallowRecoveryHonorsApplicableBudgetsBeforeAcceptance(t *testing.T) {
	tests := []struct {
		name  string
		limit func(*ReplicationBudgets)
	}{
		{"events", func(b *ReplicationBudgets) { b.MaxEvents = 1 }},
		{"objects", func(b *ReplicationBudgets) { b.MaxObjects = 1 }},
		{"object-bytes", func(b *ReplicationBudgets) { b.MaxObjectBytes = 1 }},
		{"total-bytes", func(b *ReplicationBudgets) { b.MaxTotalBytes = 1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			budgets := defaultReplicationBudgets()
			test.limit(&budgets)
			fixture := setupShallowRecoveryFixture(t, budgets)
			before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
			beforeShallow := readShallowBytes(t)
			err := recoverSelectedShallow("origin")
			if err == nil || !strings.Contains(err.Error(), "failed for one or more exact selections") {
				t.Fatalf("over-budget recovery returned %v", err)
			}
			after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
			if after != before {
				t.Fatalf("over-budget recovery advanced accepted refs:\nbefore=%s\nafter=%s", before, after)
			}
			if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
				t.Fatalf("over-budget recovery changed shallow bytes:\nbefore=%q\nafter=%q", beforeShallow, got)
			}
			if _, err := gitOutput("cat-file", "-e", fixture.first.Commit+"^{commit}"); err == nil {
				t.Fatal("over-budget predecessor entered the main object database")
			}
		})
	}
}

func TestShallowRecoveryHonorsAttachmentBudgetBeforeAcceptance(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	withTestDirectory(t, publisher)
	actor := testIdentity(t, "Attachment Budget Actor")
	writeTestPolicy(t, publisher, PolicyDocument{
		Version: policyVersion, Maintainers: []string{actor.Actor},
		Proposals: ProposalPolicy{RequiredAccepts: 1},
		Pipelines: map[string]PipelinePolicy{
			"test": {RequiredResults: 1, TrustedRunners: []string{actor.Actor}},
		},
	})
	writeTestPipeline(t, publisher)
	mustGit(t, "add", ".hn")
	mustGit(t, "commit", "-q", "-m", "pipeline base")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "pipeline code")
	code := mustGitText(t, "rev-parse", "HEAD")
	_, _, definition, err := loadPipeline(code, "test")
	if err != nil {
		t.Fatal(err)
	}
	proposalEvent := newEvent(actor, "proposal.open", 1, "")
	proposalEvent.Title, proposalEvent.Base, proposalEvent.Head = "attachment budget", base, code
	proposal, err := appendEvent(proposalEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	requestEvent, err := nextEvent(actor, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	requestEvent.Subject, requestEvent.Pipeline = proposal.ID, "test"
	requestEvent.Commit, requestEvent.Definition = code, definition
	request, err := appendEvent(requestEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	log := []byte("attachment larger than one byte\n")
	resultEvent, err := nextEvent(actor, "run.result")
	if err != nil {
		t.Fatal(err)
	}
	resultEvent.Subject, resultEvent.Pipeline = request.ID, "test"
	resultEvent.Commit, resultEvent.Definition = code, definition
	resultEvent.Outcome, resultEvent.Log = "passed", eventID(log)
	resultEvent.Backend, resultEvent.Platform, resultEvent.Runner = "host", "test/test", "hn/test"
	result, err := appendEventWithAttachments(resultEvent, actor, map[string][]byte{"log.txt": log})
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, code); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	for _, ref := range []string{actorRef(actor.Actor), proposalRef(proposal.ID)} {
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	mustGit(t, "fetch", "-q", "origin", proposalRef(proposal.ID)+":"+acceptedProposalRef("origin", proposal.ID))
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(actor.Actor)+":"+acceptedActorRef("origin", actor.Actor))
	budgets := defaultReplicationBudgets()
	budgets.MaxAttachmentBytes = 1
	if err := saveReplicationSelection(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Actors: []string{actor.Actor}, Budgets: budgets,
	}); err != nil {
		t.Fatal(err)
	}
	gapErr := guardShallowEventClosure("run logs")
	var gap *ShallowDependencyGap
	if !errors.As(gapErr, &gap) || gap.MissingID != request.ID {
		t.Fatalf("attachment fixture gap = %v, want request %s", gapErr, request.ID)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	beforeShallow := readShallowBytes(t)
	err = recoverSelectedShallow("origin")
	if err == nil || !strings.Contains(err.Error(), "failed for one or more exact selections") {
		t.Fatalf("attachment-over-budget recovery = %v", err)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes"); got != beforeRefs {
		t.Fatalf("attachment budget advanced accepted refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
	if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
		t.Fatalf("attachment budget changed shallow bytes:\nbefore=%q\nafter=%q", beforeShallow, got)
	}
	if _, err := gitOutput("cat-file", "-e", request.Commit+"^{commit}"); err == nil {
		t.Fatal("attachment-over-budget request object entered the main object database")
	}
	assertRefValue(t, acceptedActorRef("origin", actor.Actor), result.Commit)
}

func TestShallowRecoveryInterruptionsPreserveAcceptedRefs(t *testing.T) {
	phases := []struct {
		name          string
		objectResidue bool
		set           func(func() error)
	}{
		{"before-fetch-completion", false, func(hook func() error) { replicationBeforeFetchHook = hook }},
		{"after-fetch", false, func(hook func() error) { replicationAfterFetchHook = hook }},
		{"after-measurement", false, func(hook func() error) { replicationAfterMeasureHook = hook }},
		{"after-object-copy", true, func(hook func() error) { replicationAfterCopyHook = hook }},
		{"before-ref-transaction", true, func(hook func() error) { replicationBeforePromoteHook = hook }},
	}
	for _, phase := range phases {
		t.Run(phase.name, func(t *testing.T) {
			fixture := setupShallowRecoveryFixture(t, defaultReplicationBudgets())
			before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
			beforeHead := mustGitText(t, "rev-parse", "HEAD")
			beforeShallow := readShallowBytes(t)
			selectionPath, err := replicationSelectionPath("origin")
			if err != nil {
				t.Fatal(err)
			}
			beforeSelection, err := os.ReadFile(selectionPath)
			if err != nil {
				t.Fatal(err)
			}
			beforeProjectionErr := guardShallowEventClosure("interruption baseline")
			var beforeGap *ShallowDependencyGap
			if !errors.As(beforeProjectionErr, &beforeGap) {
				t.Fatalf("baseline projection error = %v, want gap", beforeProjectionErr)
			}
			phase.set(func() error { return errors.New("injected " + phase.name) })
			t.Cleanup(resetReplicationInterruptionHooks)
			recoveryErr := recoverSelectedShallow("origin")
			if recoveryErr == nil {
				t.Fatal("interrupted recovery succeeded")
			}
			after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
			if after != before {
				t.Fatalf("interruption advanced accepted refs:\nbefore=%s\nafter=%s", before, after)
			}
			if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
				t.Fatalf("interruption advanced HEAD from %s to %s", beforeHead, got)
			}
			if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
				t.Fatalf("interruption changed shallow boundaries:\nbefore=%q\nafter=%q", beforeShallow, got)
			}
			afterSelection, err := os.ReadFile(selectionPath)
			if err != nil || !bytes.Equal(afterSelection, beforeSelection) {
				t.Fatalf("interruption changed saved selection bytes: err=%v", err)
			}
			afterProjectionErr := guardShallowEventClosure("interruption retry")
			var afterGap *ShallowDependencyGap
			if !errors.As(afterProjectionErr, &afterGap) || afterGap.MissingID != beforeGap.MissingID {
				t.Fatalf("interruption changed projection gap from %v to %v", beforeProjectionErr, afterProjectionErr)
			}
			_, objectErr := gitOutput("cat-file", "-e", fixture.first.Commit+"^{commit}")
			if phase.objectResidue {
				// WP04 must copy verified objects before its ref transaction. A crash in
				// that interval may leave unreachable ODB residue; deleting it could remove
				// objects shared with another transaction. The trust contract therefore
				// preserves accepted roots/projection and denies the residue until retry.
				if objectErr != nil {
					t.Fatalf("%s did not exercise the real post-copy interval: %v", phase.name, objectErr)
				}
				if !strings.Contains(recoveryErr.Error(), "unreferenced object residue may remain") {
					t.Fatalf("post-copy error falsely implied physical rollback: %v", recoveryErr)
				}
				if visible := mustGitText(t, "for-each-ref", "--format=%(refname)", "--contains", fixture.first.Commit, "refs/hn/remotes"); visible != "" {
					t.Fatalf("uncommitted object residue became reachable from accepted roots: %s", visible)
				}
			} else if objectErr == nil {
				t.Fatalf("%s imported an object before the copy phase", phase.name)
			}
			assertRefValue(t, acceptedActorRef("origin", fixture.actor.Actor), fixture.second.Commit)
		})
		resetReplicationInterruptionHooks()
	}
}

func TestShallowRecoveryBoundaryReleaseFailureIsTruthfulAndRetryable(t *testing.T) {
	fixture := setupShallowRecoveryFixture(t, defaultReplicationBudgets())
	baselineErr := guardShallowEventClosure("proposal status")
	var baselineGap *ShallowDependencyGap
	if !errors.As(baselineErr, &baselineGap) {
		t.Fatalf("baseline = %v, want gap", baselineErr)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes")
	beforeShallow := readShallowBytes(t)
	selectionPath, err := replicationSelectionPath("origin")
	if err != nil {
		t.Fatal(err)
	}
	beforeSelection, err := os.ReadFile(selectionPath)
	if err != nil {
		t.Fatal(err)
	}
	previousRelease := replicationReleaseShallow
	replicationReleaseShallow = func(string, []replicationPromotion) error {
		return errors.New("injected shallow marker write failure")
	}
	err = recoverSelectedShallow("origin")
	replicationReleaseShallow = previousRelease
	t.Cleanup(func() { replicationReleaseShallow = releaseRecoveredShallowBoundaries })
	for _, exact := range []string{
		"ref transaction committed", "0 ref value change(s)", acceptedActorRef("origin", fixture.actor.Actor) + "=same-head",
		"required objects were imported", "shallow boundary release failed", "injected shallow marker write failure",
		"accepted fact projection remains fail-closed", "retry hn sync origin --recover-shallow",
	} {
		if err == nil || !strings.Contains(err.Error(), exact) {
			t.Fatalf("post-promotion release error omitted %q: %v", exact, err)
		}
	}
	if strings.Contains(err.Error(), "accepted refs advanced") {
		t.Fatalf("post-promotion release error was not truthful: %v", err)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/remotes"); got != beforeRefs {
		t.Fatalf("same-head recovery unexpectedly changed refs:\nbefore=%s\nafter=%s", beforeRefs, got)
	}
	if got := readShallowBytes(t); !bytes.Equal(got, beforeShallow) {
		t.Fatalf("failed release changed shallow bytes:\nbefore=%q\nafter=%q", beforeShallow, got)
	}
	afterSelection, err := os.ReadFile(selectionPath)
	if err != nil || !bytes.Equal(beforeSelection, afterSelection) {
		t.Fatalf("failed release changed selection bytes: err=%v", err)
	}
	retryGapErr := guardShallowEventClosure("proposal status retry")
	var retryGap *ShallowDependencyGap
	if !errors.As(retryGapErr, &retryGap) || retryGap.MissingID != baselineGap.MissingID {
		t.Fatalf("failed release did not remain fail-closed: %v", retryGapErr)
	}
	if err := recoverSelectedShallow("origin"); err != nil {
		t.Fatalf("retry after marker failure failed: %v", err)
	}
	if err := guardShallowEventClosure("proposal status fresh retry"); err != nil {
		t.Fatalf("fresh projection after retry failed: %v", err)
	}
	assertRefValue(t, acceptedActorRef("origin", fixture.actor.Actor), fixture.second.Commit)
}

func TestShallowRecoveryBoundaryReleaseFailureReportsAdvancedRef(t *testing.T) {
	fixture := setupShallowRecoveryFixture(t, defaultReplicationBudgets())
	if err := os.Chdir(fixture.publisher); err != nil {
		t.Fatal(err)
	}
	thirdEvent, err := nextEvent(fixture.actor, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	thirdEvent.Title = "remote advanced head"
	third, err := appendEvent(thirdEvent, fixture.actor)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "push", "-q", "origin", actorRef(fixture.actor.Actor)+":"+actorRef(fixture.actor.Actor))
	if err := os.Chdir(fixture.clone); err != nil {
		t.Fatal(err)
	}
	previousRelease := replicationReleaseShallow
	replicationReleaseShallow = func(string, []replicationPromotion) error {
		return errors.New("injected advanced marker failure")
	}
	err = recoverSelectedShallow("origin")
	replicationReleaseShallow = previousRelease
	t.Cleanup(func() { replicationReleaseShallow = releaseRecoveredShallowBoundaries })
	for _, exact := range []string{
		"1 ref value change(s)", acceptedActorRef("origin", fixture.actor.Actor) + "=advanced",
		"injected advanced marker failure", "accepted fact projection remains fail-closed",
	} {
		if err == nil || !strings.Contains(err.Error(), exact) {
			t.Fatalf("advanced partial-success error omitted %q: %v", exact, err)
		}
	}
	assertRefValue(t, acceptedActorRef("origin", fixture.actor.Actor), third.Commit)
	projectionErr := guardShallowEventClosure("advanced marker failure")
	var gap *ShallowDependencyGap
	if !errors.As(projectionErr, &gap) || gap.MissingID != fixture.first.ID {
		t.Fatalf("advanced ref with unreleased marker did not stay fail-closed: %v", projectionErr)
	}
	if err := recoverSelectedShallow("origin"); err != nil {
		t.Fatalf("retry after advanced marker failure: %v", err)
	}
}

type shallowRecoveryFixture struct {
	actor           *Identity
	first           *StoredEvent
	second          *StoredEvent
	unselected      *Identity
	unselectedFirst *StoredEvent
	publisher       string
	clone           string
}

func setupShallowRecoveryFixture(t *testing.T, budgets ReplicationBudgets) shallowRecoveryFixture {
	t.Helper()
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "remote.git")
	clone := filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	withTestDirectory(t, publisher)
	actor := testIdentity(t, "Selected Recovery Actor")
	firstEvent := newEvent(actor, "issue.open", 1, "")
	firstEvent.Title = "first"
	first, err := appendEvent(firstEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	secondEvent, err := nextEvent(actor, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	secondEvent.Title = "second"
	second, err := appendEvent(secondEvent, actor)
	if err != nil {
		t.Fatal(err)
	}
	unselected := testIdentity(t, "Unselected Recovery Actor")
	unselectedFirstEvent := newEvent(unselected, "issue.open", 1, "")
	unselectedFirstEvent.Title = "unselected first"
	unselectedFirst, err := appendEvent(unselectedFirstEvent, unselected)
	if err != nil {
		t.Fatal(err)
	}
	unselectedSecondEvent, err := nextEvent(unselected, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	unselectedSecondEvent.Title = "unselected second"
	if _, err := appendEvent(unselectedSecondEvent, unselected); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	for _, identity := range []*Identity{actor, unselected} {
		ref := actorRef(identity.Actor)
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	withTestDirectory(t, clone)
	mustGit(t, "fetch", "-q", "--depth", "1", "origin", actorRef(actor.Actor)+":"+acceptedActorRef("origin", actor.Actor))
	if err := saveReplicationSelection(ReplicationSelection{
		Version: replicationSelectionVersion,
		Remote:  "origin",
		Actors:  []string{actor.Actor},
		Budgets: budgets,
	}); err != nil {
		t.Fatal(err)
	}
	return shallowRecoveryFixture{
		actor: actor, first: first, second: second,
		unselected: unselected, unselectedFirst: unselectedFirst,
		publisher: publisher, clone: clone,
	}
}

func resetReplicationInterruptionHooks() {
	replicationBeforeFetchHook = nil
	replicationAfterFetchHook = nil
	replicationAfterMeasureHook = nil
	replicationAfterCopyHook = nil
	replicationBeforePromoteHook = nil
}

func readShallowBytes(t *testing.T) []byte {
	t.Helper()
	path := mustGitText(t, "rev-parse", "--git-path", "shallow")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func depthOneRepository(t *testing.T) (root, clone, base, head string) {
	t.Helper()
	root = t.TempDir()
	seed := filepath.Join(root, "seed")
	remote := filepath.Join(root, "remote.git")
	clone = filepath.Join(root, "clone")
	mustGit(t, "init", "-q", "-b", "main", seed)
	mustGit(t, "-C", seed, "config", "user.name", "Seed")
	mustGit(t, "-C", seed, "config", "user.email", "seed@hn.invalid")
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "base")
	base = strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "-C", seed, "commit", "--allow-empty", "-q", "-m", "head")
	head = strings.TrimSpace(string(mustGitOutputAt(t, seed, "rev-parse", "HEAD")))
	mustGit(t, "clone", "-q", "--bare", seed, remote)
	mustGit(t, "clone", "-q", "--depth", "1", "file://"+remote, clone)
	return root, clone, base, head
}

func mustGitOutputAt(t *testing.T, directory string, args ...string) []byte {
	t.Helper()
	all := append([]string{"-C", directory}, args...)
	output, err := gitOutput(all...)
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func copyCommitAndTreesWithoutBlobs(t *testing.T, source, commit string) {
	t.Helper()
	copyExactGitObject(t, source, commit, "commit")
	rootTree := strings.TrimSpace(string(mustGitOutputAt(t, source, "rev-parse", commit+"^{tree}")))
	copyExactGitObject(t, source, rootTree, "tree")
	entries := strings.Split(strings.TrimSpace(string(mustGitOutputAt(t, source, "ls-tree", "-r", "-t", commit))), "\n")
	for _, entry := range entries {
		fields := strings.Fields(entry)
		if len(fields) >= 3 && fields[1] == "tree" {
			copyExactGitObject(t, source, fields[2], "tree")
		}
	}
}

func copyExactGitObject(t *testing.T, source, objectID, objectType string) {
	t.Helper()
	contents := mustGitOutputAt(t, source, "cat-file", objectType, objectID)
	output, err := gitInput(contents, nil, "hash-object", "-w", "-t", objectType, "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(output)); got != objectID {
		t.Fatalf("copied %s object = %s, want %s", objectType, got, objectID)
	}
}

func withTestDirectory(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
