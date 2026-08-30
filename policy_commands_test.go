package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	policyActorA = "1111111111111111111111111111111111111111111111111111111111111111"
	policyActorB = "2222222222222222222222222222222222222222222222222222222222222222"
	policyActorC = "3333333333333333333333333333333333333333333333333333333333333333"
	policyActorD = "4444444444444444444444444444444444444444444444444444444444444444"
)

func TestPolicyShowAndCheckContract(t *testing.T) {
	// Boundary: the policy command handler plus the live main.go route in a real
	// Git repository. WP06 owns that route and formally handed off this wiring.
	root := enterPolicyTestRepository(t)
	baseBytes := []byte(`{
  "version": "nh.policy/0",
  "maintainers": ["` + policyActorB + `", "` + policyActorA + `"],
  "proposals": {
    "requiredApprovals": 1,
    "requiredAccepts": 1,
    "trustedReviewers": ["` + policyActorC + `", "` + policyActorA + `"],
    "allowAuthorApproval": false
  },
  "pipelines": {
    "zeta": {"requiredResults": 1, "trustedRunners": ["` + policyActorD + `"]},
    "alpha": {"requiredResults": 1, "trustedRunners": ["` + policyActorC + `", "` + policyActorB + `"]}
  }
}
`)
	base := commitPolicyBytes(t, root, baseBytes, "base policy")

	show, err := captureTestOutput(t, func() error { return run([]string{"policy", "show"}) })
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, show,
		"Policy commit: "+base,
		"Policy digest: "+eventID(baseBytes),
		"Maintainers (required accepts 1):",
		policyActorA,
		policyActorB,
		"Trusted reviewers (required approvals 1; author approval false):",
		"alpha (required results 1):",
		"zeta (required results 1):",
	)
	assertBefore(t, show, policyActorA, policyActorB)
	assertBefore(t, show, "alpha (required results", "zeta (required results")

	explicit, err := captureTestOutput(t, func() error { return cmdPolicy([]string{"show", base}) })
	if err != nil {
		t.Fatal(err)
	}
	if explicit != show {
		t.Fatalf("explicit show differs from HEAD show\nHEAD:\n%s\nexplicit:\n%s", show, explicit)
	}

	headBytes := []byte(`{
  "pipelines": {
    "beta": {"trustedRunners": ["` + policyActorD + `"], "requiredResults": 1},
    "alpha": {"trustedRunners": ["` + policyActorA + `", "` + policyActorC + `"], "requiredResults": 2}
  },
  "proposals": {
    "allowAuthorApproval": true,
    "trustedReviewers": ["` + policyActorD + `", "` + policyActorA + `"],
    "requiredAccepts": 2,
    "requiredApprovals": 1
  },
  "maintainers": ["` + policyActorC + `", "` + policyActorA + `"],
  "version": "nh.policy/0"
}
`)
	head := commitPolicyBytes(t, root, headBytes, "proposed policy")
	check, err := captureTestOutput(t, func() error {
		return cmdPolicy([]string{"check", "--base", base, "--head", head})
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, check,
		"Base commit: "+base,
		"Base policy digest: "+eventID(baseBytes),
		"Proposed commit: "+head,
		"Proposed policy digest: "+eventID(headBytes),
		"The base policy governs this amendment candidate.",
		"exact policy bytes: changed",
		"maintainers added: "+policyActorC,
		"maintainers removed: "+policyActorB,
		"required accepts: changed (1 -> 2)",
		"trusted reviewers added: "+policyActorD,
		"trusted reviewers removed: "+policyActorC,
		"required approvals: unchanged (1)",
		"author approval: changed (false -> true)",
		"pipelines added: beta",
		"pipelines removed: zeta",
		"pipeline alpha required results: changed (1 -> 2)",
		"pipeline alpha trusted runners added: "+policyActorA,
		"pipeline alpha trusted runners removed: "+policyActorB,
	)
	assertBefore(t, check, "pipeline alpha required results", "pipeline beta required results")
	assertBefore(t, check, "pipeline beta required results", "pipeline zeta required results")

	draftBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	draftPath := filepath.Join(root, "draft-policy.json")
	if err := os.WriteFile(draftPath, draftBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	fileCheck, err := captureTestOutput(t, func() error {
		return cmdPolicy([]string{"check", "--base", base, "--file", draftPath})
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, fileCheck,
		"Proposed file: "+draftPath,
		"Proposed policy digest: "+eventID(draftBytes),
		"The base policy governs this amendment candidate.",
	)
}

func TestPolicyCheckSemanticallyEqualExactBytesAndUsage(t *testing.T) {
	root := enterPolicyTestRepository(t)
	baseBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	base := commitPolicyBytes(t, root, baseBytes, "base")
	headBytes := []byte("{\n  \"version\": \"nh.policy/0\",\n  \"maintainers\": [\"" + policyActorA + "\"],\n  \"proposals\": {\"requiredApprovals\": 0, \"requiredAccepts\": 1, \"trustedReviewers\": [], \"allowAuthorApproval\": false},\n  \"pipelines\": {}\n}\n")
	head := commitPolicyBytes(t, root, headBytes, "same semantics")

	output, err := captureTestOutput(t, func() error {
		return cmdPolicy([]string{"check", "--base", base, "--head", head})
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, output,
		"Base policy digest: "+eventID(baseBytes),
		"Proposed policy digest: "+eventID(headBytes),
		"exact policy bytes: changed",
		"maintainers added: (none)",
		"maintainers removed: (none)",
		"required accepts: unchanged (1)",
		"required approvals: unchanged (0)",
		"author approval: unchanged (false)",
		"pipelines added: (none)",
		"pipelines removed: (none)",
	)

	usageCases := []struct {
		args []string
		want string
	}{
		{[]string{"show", "HEAD", "extra"}, "usage: nh policy show [REV]"},
		{[]string{"check", "--base", base}, "exactly one of --head and --file is required"},
		{[]string{"check", "--base", base, "--head", head, "--file", "draft"}, "exactly one of --head and --file is required"},
		{[]string{"check", "--head", head}, "usage: nh policy check --base REV <--head REV|--file PATH>"},
		{[]string{"check", "--base", base, "--head", head, "extra"}, "usage: nh policy check --base REV <--head REV|--file PATH>"},
		{[]string{"unknown"}, "unknown policy command \"unknown\""},
	}
	for _, test := range usageCases {
		if err := cmdPolicy(test.args); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("cmdPolicy(%q) error = %v, want %q", test.args, err, test.want)
		}
	}
}

func TestPolicyCheckRejectsInvalidSidesWithoutMutation(t *testing.T) {
	// Boundary: the policy command handler plus real Git state and files; no
	// policy parser or validator implementation detail is invoked directly.
	root := enterPolicyTestRepository(t)
	validBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	base := commitPolicyBytes(t, root, validBytes, "valid")

	tests := []struct {
		name       string
		policy     string
		wantReason string
	}{
		{"empty maintainers", `{"version":"nh.policy/0","maintainers":[],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "requires at least one maintainer"},
		{"malformed actor", `{"version":"nh.policy/0","maintainers":["bad"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "contains invalid actor \"bad\""},
		{"duplicate actor", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `","` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "contains duplicate actor " + policyActorA},
		{"accept lockout", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":2,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "requiredAccepts must be between 1 and the number of maintainers"},
		{"approval lockout", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":1,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "requiredApprovals exceeds the number of trusted reviewers"},
		{"invalid pipeline", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{"bad/name":{"requiredResults":1,"trustedRunners":["` + policyActorB + `"]}}}`, "invalid pipeline name \"bad/name\""},
		{"result lockout", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{"test":{"requiredResults":2,"trustedRunners":["` + policyActorB + `"]}}}`, "requiredResults exceeds its trusted runner count"},
		{"unsupported version", `{"version":"nh.policy/9","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}`, "unsupported policy version \"nh.policy/9\""},
		{"unknown field", `{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{},"admin":true}`, "unknown field \"admin\""},
		{"malformed JSON", `{"version":"nh.policy/0"`, "unexpected EOF"},
		{"trailing JSON", string(validBytes) + `{}`, "contains more than one JSON value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			draft := filepath.Join(root, "invalid-policy.json")
			if err := os.WriteFile(draft, []byte(test.policy), 0o644); err != nil {
				t.Fatal(err)
			}
			beforeHead := mustGitText(t, "rev-parse", "HEAD")
			beforeStatus := mustGitText(t, "status", "--porcelain", "--untracked-files=no")
			beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh")
			beforeDraft, err := os.ReadFile(draft)
			if err != nil {
				t.Fatal(err)
			}

			err = cmdPolicy([]string{"check", "--base", base, "--file", draft})
			if err == nil || !strings.Contains(err.Error(), "proposed policy") || !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("error = %v, want proposed policy and %q", err, test.wantReason)
			}
			if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
				t.Fatalf("HEAD changed from %s to %s", beforeHead, got)
			}
			if got := mustGitText(t, "status", "--porcelain", "--untracked-files=no"); got != beforeStatus {
				t.Fatalf("tracked working tree changed: before %q, after %q", beforeStatus, got)
			}
			if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh"); got != beforeRefs {
				t.Fatalf("Nichthub refs changed: before %q, after %q", beforeRefs, got)
			}
			afterDraft, err := os.ReadFile(draft)
			if err != nil {
				t.Fatal(err)
			}
			if string(afterDraft) != string(beforeDraft) {
				t.Fatal("draft policy was modified")
			}

			invalidCommit := commitPolicyBytes(t, root, []byte(test.policy), "invalid "+test.name)
			err = cmdPolicy([]string{"check", "--base", invalidCommit, "--head", base})
			if err == nil || !strings.Contains(err.Error(), "base policy") || !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("invalid base error = %v, want base policy and %q", err, test.wantReason)
			}
			err = cmdPolicy([]string{"show", invalidCommit})
			if err == nil || !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("malformed show error = %v, want %q", err, test.wantReason)
			}
		})
	}
	oversized := filepath.Join(root, "oversized-policy.json")
	if err := os.WriteFile(oversized, make([]byte, maxPolicySize+1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmdPolicy([]string{"check", "--base", base, "--file", oversized}); err == nil || !strings.Contains(err.Error(), "proposed policy") || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized draft error = %v", err)
	}

	missingCommit := commitWithoutPolicy(t, root)
	if err := cmdPolicy([]string{"show", missingCommit}); err == nil || !strings.Contains(err.Error(), "policy") || !strings.Contains(err.Error(), missingCommit) || !strings.Contains(err.Error(), ".nh/policy.json") {
		t.Fatalf("missing policy error = %v", err)
	}
	if err := cmdPolicy([]string{"show", "HEAD:.nh/policy.json"}); err == nil || !strings.Contains(err.Error(), "not a commit") {
		t.Fatalf("non-commit revision error = %v", err)
	}
	if err := cmdPolicy([]string{"check", "--base", missingCommit, "--head", base}); err == nil || !strings.Contains(err.Error(), "base policy") {
		t.Fatalf("invalid base error = %v", err)
	}
}

func TestPolicyEvaluationUsesExactBaseAcrossAmendment(t *testing.T) {
	root := enterPolicyTestRepository(t)
	maintainer := testIdentity(t, "Maintainer")
	oldReviewer := testIdentity(t, "Old Reviewer")
	oldRunner := testIdentity(t, "Old Runner")
	newReviewer := testIdentity(t, "New Reviewer")
	newRunner := testIdentity(t, "New Runner")
	basePolicy := PolicyDocument{
		Version: policyVersion, Maintainers: []string{maintainer.Actor},
		Proposals: ProposalPolicy{RequiredApprovals: 1, RequiredAccepts: 1, TrustedReviewers: []string{oldReviewer.Actor}},
		Pipelines: map[string]PipelinePolicy{"test": {RequiredResults: 1, TrustedRunners: []string{oldRunner.Actor}}},
	}
	writeTestPolicy(t, root, basePolicy)
	writeTestPipeline(t, root)
	mustGit(t, "add", ".nh")
	mustGit(t, "commit", "-q", "-m", "base policy")
	base := mustGitText(t, "rev-parse", "HEAD")
	baseBytes, err := gitOutput("show", base+":.nh/policy.json")
	if err != nil {
		t.Fatal(err)
	}
	baseDigest := eventID(baseBytes)

	headPolicy := basePolicy
	headPolicy.Proposals.TrustedReviewers = []string{oldReviewer.Actor, newReviewer.Actor}
	headPolicy.Proposals.AllowAuthorApproval = true
	headPolicy.Pipelines = map[string]PipelinePolicy{"test": {RequiredResults: 1, TrustedRunners: []string{oldRunner.Actor, newRunner.Actor}}}
	writeTestPolicy(t, root, headPolicy)
	mustGit(t, "add", ".nh/policy.json")
	mustGit(t, "commit", "-q", "-m", "add policy actors")
	amended := mustGitText(t, "rev-parse", "HEAD")

	authorizationEvent, err := nextEvent(maintainer, "identity.authorize")
	if err != nil {
		t.Fatal(err)
	}
	authorizationEvent.Relationship = identityRelationshipDevice
	authorizationEvent.TargetActor = newReviewer.Actor
	authorizationEvent.TargetKey = newReviewer.PublicKey
	authorization, err := appendEvent(authorizationEvent, maintainer)
	if err != nil {
		t.Fatal(err)
	}
	verifiedAuthorization, authorizationID, err := verifyEvent(authorization.Payload, authorization.Signature)
	if err != nil {
		t.Fatal(err)
	}
	if authorizationID != authorization.ID || verifiedAuthorization.Actor != maintainer.Actor || verifiedAuthorization.TargetActor != newReviewer.Actor {
		t.Fatal("stored authorization does not preserve its verified actor bindings")
	}
	acceptanceEvent, err := nextEvent(newReviewer, "identity.accept")
	if err != nil {
		t.Fatal(err)
	}
	acceptanceEvent.Subject = authorization.ID
	acceptance, err := appendEvent(acceptanceEvent, newReviewer)
	if err != nil {
		t.Fatal(err)
	}
	continuityEvents, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	continuity, err := ProjectIdentityContinuity(continuityEvents)
	if err != nil {
		t.Fatal(err)
	}
	if len(continuity.Edges) != 1 {
		t.Fatalf("continuity edges = %#v", continuity.Edges)
	}
	edge := continuity.Edges[0]
	if edge.State != IdentityEdgeAccepted || edge.Relationship != identityRelationshipDevice || edge.AuthorizingActor != maintainer.Actor || edge.TargetActor != newReviewer.Actor || len(edge.AcceptanceIDs) != 1 || edge.AcceptanceIDs[0] != acceptance.ID {
		t.Fatalf("accepted continuity edge = %#v", edge)
	}

	first := appendPolicyTestProposal(t, maintainer, base, amended, "Amend policy")
	appendPolicyTestReviewAndResult(t, maintainer, newReviewer, newRunner, first, amended)
	firstEvents, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	firstResolved, err := resolveFullEvent(firstEvents, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	firstEvaluation, err := evaluateProposal(firstResolved, firstEvents)
	if err != nil {
		t.Fatal(err)
	}
	if firstEvaluation.PolicyDigest != baseDigest || firstEvaluation.Ready || len(firstEvaluation.Approvals) != 0 || len(firstEvaluation.Pipelines) != 1 || len(firstEvaluation.Pipelines[0].Passed) != 0 {
		t.Fatalf("continuity target inherited policy authority: %#v", firstEvaluation)
	}
	decisionEvent, err := nextEvent(newReviewer, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	decisionEvent.Subject = first.ID
	decisionEvent.Policy = baseDigest
	decisionEvent.Verdict = "reject"
	decisionEvent.Body = "continuity must not confer maintainer authority"
	decisionPayload, decisionSignature, err := encodeAndSign(decisionEvent, newReviewer)
	if err != nil {
		t.Fatal(err)
	}
	verifiedDecision, decisionID, err := verifyEvent(decisionPayload, decisionSignature)
	if err != nil {
		t.Fatal(err)
	}
	decision := StoredEvent{ID: decisionID, Event: verifiedDecision, Payload: decisionPayload, Signature: decisionSignature}
	byID := make(map[string]StoredEvent, len(firstEvents))
	for _, stored := range firstEvents {
		byID[stored.ID] = stored
	}
	if err := validateDecisionEvent(decision, *firstResolved, byID); err == nil || !strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("accepted device inherited maintainer authority: %v", err)
	}
	firstStatus, err := captureTestOutput(t, func() error { return cmdProposalStatus(first.ID) })
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, firstStatus,
		"Policy:   "+shortID(baseDigest)+" from base "+shortOID(base),
		"Reviews:  0/1 trusted approvals",
		"Pipeline: test 0/1 trusted passes",
		"Status:   blocked",
	)

	mustGit(t, "commit", "--allow-empty", "-q", "-m", "later change")
	laterHead := mustGitText(t, "rev-parse", "HEAD")
	second := appendPolicyTestProposal(t, maintainer, amended, laterHead, "Later candidate")
	appendPolicyTestReviewAndResult(t, maintainer, newReviewer, newRunner, second, laterHead)
	amendedBytes, err := gitOutput("show", amended+":.nh/policy.json")
	if err != nil {
		t.Fatal(err)
	}
	amendedDigest := eventID(amendedBytes)
	secondEvents, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	secondResolved, err := resolveFullEvent(secondEvents, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	secondEvaluation, err := evaluateProposal(secondResolved, secondEvents)
	if err != nil {
		t.Fatal(err)
	}
	if secondEvaluation.PolicyDigest != amendedDigest || !secondEvaluation.Ready || len(secondEvaluation.Approvals) != 1 || len(secondEvaluation.Pipelines) != 1 || len(secondEvaluation.Pipelines[0].Passed) != 1 {
		t.Fatalf("explicit policy actors did not qualify: %#v", secondEvaluation)
	}
	secondStatus, err := captureTestOutput(t, func() error { return cmdProposalStatus(second.ID) })
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, secondStatus,
		"Policy:   "+shortID(amendedDigest)+" from base "+shortOID(amended),
		"Reviews:  1/1 trusted approvals",
		"Pipeline: test 1/1 trusted passes",
		"Status:   ready for decision",
	)
}

func TestProposalOpenUsesPolicyAmendmentDiagnostic(t *testing.T) {
	root := enterPolicyTestRepository(t)
	author, _, err := createIdentity("Policy Author")
	if err != nil {
		t.Fatal(err)
	}
	baseBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + author.Actor + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	base := commitPolicyBytes(t, root, baseBytes, "base")
	headBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + author.Actor + `","` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	head := commitPolicyBytes(t, root, headBytes, "amend policy")

	output, err := captureTestOutput(t, func() error {
		return run([]string{"proposal", "open", "--base", base, "--head", head, "Amend policy"})
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, output,
		"Opened proposal",
		"Policy amendment: base digest "+eventID(baseBytes),
		"governs this candidate; proposed digest "+eventID(headBytes),
		"applies only to later candidates based on commit "+head,
	)

	mustGit(t, "commit", "--allow-empty", "-q", "-m", "ordinary change")
	ordinaryHead := mustGitText(t, "rev-parse", "HEAD")
	ordinaryOutput, err := captureTestOutput(t, func() error {
		return run([]string{"proposal", "open", "--base", head, "--head", ordinaryHead, "Ordinary change"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ordinaryOutput, "Policy amendment:") {
		t.Fatalf("ordinary proposal printed a policy diagnostic:\n%s", ordinaryOutput)
	}
}

func TestProposalOpenRejectsInvalidPolicyBeforeEventsOrRefs(t *testing.T) {
	root := enterPolicyTestRepository(t)
	author, _, err := createIdentity("Policy Author")
	if err != nil {
		t.Fatal(err)
	}
	baseBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + author.Actor + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	base := commitPolicyBytes(t, root, baseBytes, "base")
	invalidHead := commitPolicyBytes(t, root, []byte(`{"version":"nh.policy/0","maintainers":[]}`+"\n"), "invalid amendment")
	identityPaths, err := identityKeyringPaths()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(identityPaths.active); err != nil {
		t.Fatal(err)
	}
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh")

	_, err = captureTestOutput(t, func() error {
		return run([]string{"proposal", "open", "--base", base, "--head", invalidHead, "Invalid policy"})
	})
	if err == nil || !strings.Contains(err.Error(), "proposed policy") || !strings.Contains(err.Error(), "requires at least one maintainer") {
		t.Fatalf("proposal error = %v, want proposed policy lockout rejection", err)
	}
	if afterRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh"); afterRefs != beforeRefs {
		t.Fatalf("rejected proposal changed refs: before %q, after %q", beforeRefs, afterRefs)
	}
}

func TestPolicyAmendmentDiagnosticIsExactAndReadOnly(t *testing.T) {
	root := enterPolicyTestRepository(t)
	baseBytes := []byte(`{"version":"nh.policy/0","maintainers":["` + policyActorA + `"],"proposals":{"requiredApprovals":0,"requiredAccepts":1,"trustedReviewers":[],"allowAuthorApproval":false},"pipelines":{}}` + "\n")
	base := commitPolicyBytes(t, root, baseBytes, "base")
	message, err := policyAmendmentDiagnostic(base, base)
	if err != nil {
		t.Fatal(err)
	}
	if message != "" {
		t.Fatalf("unchanged policy diagnostic = %q, want empty", message)
	}
	headBytes := append([]byte("\n"), baseBytes...)
	head := commitPolicyBytes(t, root, headBytes, "byte-only policy change")
	beforeHead := mustGitText(t, "rev-parse", "HEAD")
	beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh")
	message, err = policyAmendmentDiagnostic(base, head)
	if err != nil {
		t.Fatal(err)
	}
	assertPolicyOutputContains(t, message, eventID(baseBytes), eventID(headBytes), head, "base digest", "governs")
	if got := mustGitText(t, "rev-parse", "HEAD"); got != beforeHead {
		t.Fatalf("diagnostic changed HEAD from %s to %s", beforeHead, got)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh"); got != beforeRefs {
		t.Fatalf("diagnostic changed refs: before %q, after %q", beforeRefs, got)
	}
}

func enterPolicyTestRepository(t *testing.T) string {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Policy Test")
	mustGit(t, "config", "user.email", "policy@nh.invalid")
	return root
}

func commitPolicyBytes(t *testing.T, root string, encoded []byte, message string) string {
	t.Helper()
	path := filepath.Join(root, ".nh", "policy.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".nh/policy.json")
	mustGit(t, "commit", "-q", "-m", message)
	return mustGitText(t, "rev-parse", "HEAD")
}

func commitWithoutPolicy(t *testing.T, root string) string {
	t.Helper()
	if err := os.Remove(filepath.Join(root, ".nh", "policy.json")); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".nh/policy.json")
	mustGit(t, "commit", "-q", "-m", "remove policy")
	return mustGitText(t, "rev-parse", "HEAD")
}

func assertPolicyOutputContains(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(output, value) {
			t.Errorf("output missing %q:\n%s", value, output)
		}
	}
}

func assertBefore(t *testing.T, output, first, second string) {
	t.Helper()
	firstIndex, secondIndex := strings.Index(output, first), strings.Index(output, second)
	if firstIndex < 0 || secondIndex < 0 || firstIndex >= secondIndex {
		t.Errorf("%q does not precede %q in output:\n%s", first, second, output)
	}
}

func writeTestPipeline(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, ".nh", "pipelines", "test.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":"nh.pipeline/0","steps":[{"name":"test","command":"true"}]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendPolicyTestProposal(t *testing.T, author *Identity, base, head, title string) *StoredEvent {
	t.Helper()
	event, err := nextEvent(author, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	event.Title, event.Base, event.Head = title, base, head
	stored, err := appendEvent(event, author)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(stored.ID, head); err != nil {
		t.Fatal(err)
	}
	return stored
}

func appendPolicyTestReviewAndResult(t *testing.T, maintainer, reviewer, runner *Identity, proposal *StoredEvent, head string) {
	t.Helper()
	review, err := nextEvent(reviewer, "review.submit")
	if err != nil {
		t.Fatal(err)
	}
	review.Subject, review.Verdict = proposal.ID, "approve"
	if _, err := appendEvent(review, reviewer); err != nil {
		t.Fatal(err)
	}
	_, _, definition, err := loadPipeline(head, "test")
	if err != nil {
		t.Fatal(err)
	}
	request, err := nextEvent(maintainer, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	request.Subject, request.Pipeline, request.Definition, request.Commit = proposal.ID, "test", definition, head
	storedRequest, err := appendEvent(request, maintainer)
	if err != nil {
		t.Fatal(err)
	}
	result, err := nextEvent(runner, "run.result")
	if err != nil {
		t.Fatal(err)
	}
	log := []byte("passed\n")
	result.Subject, result.Pipeline, result.Definition, result.Commit = storedRequest.ID, "test", definition, head
	result.Outcome, result.Log, result.Backend, result.Platform, result.Runner = "passed", eventID(log), "sandbox", "test/test", "nh/test"
	if _, err := appendEventWithAttachments(result, runner, map[string][]byte{"log.txt": log}); err != nil {
		t.Fatal(err)
	}
}
