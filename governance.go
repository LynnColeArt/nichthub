package main

import (
	"fmt"
	"strings"
)

func cmdDecide(args []string) error {
	if len(args) < 1 {
		return usageError("usage: nh decide PROPOSAL <--accept|--reject> [--body TEXT]")
	}
	query := args[0]
	flags := quietFlags("decide")
	accept := flags.Bool("accept", false, "accept the proposal")
	reject := flags.Bool("reject", false, "reject the proposal")
	body := flags.String("body", "", "decision explanation")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || *accept == *reject {
		return usageError("choose exactly one of --accept or --reject")
	}
	if *reject && strings.TrimSpace(*body) == "" {
		return usageError("a rejection requires --body TEXT")
	}
	if err := prepareShallowVerification(shallowVerificationScope{Operation: "proposal decision", Subject: query}); err != nil {
		return err
	}
	if err := guardShallowEventClosure("proposal decision"); err != nil {
		return err
	}
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEventDependency("proposal decision", query, shallowCandidateEvent, events)
	if err != nil {
		return err
	}
	if err := guardProposalEvaluationDependencies("proposal decision", proposal, events); err != nil {
		return err
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		return err
	}
	if *accept {
		if err := requireLineageTerminalEligibility("accept", proposal.ID, evaluation.Lineage); err != nil {
			return err
		}
	}
	if evaluation.Merged {
		return fmt.Errorf("proposal %s is already merged", proposal.ID)
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	if !actorListed(identity.Actor, evaluation.Policy.Maintainers) {
		return fmt.Errorf("identity %s is not a maintainer under policy %s", shortID(identity.Actor), shortID(evaluation.PolicyDigest))
	}
	if *accept && !evaluation.Ready {
		return fmt.Errorf("proposal is not ready; run 'nh proposal status %s' for missing evidence", shortID(proposal.ID))
	}
	if *accept {
		currentEvents, err := collectEvents()
		if err != nil {
			return err
		}
		currentProposal, err := resolveEvent(currentEvents, proposal.ID)
		if err != nil {
			return err
		}
		if err := guardProposalEvaluationDependencies("proposal decision", currentProposal, currentEvents); err != nil {
			return err
		}
		currentEvaluation, err := evaluateProposal(currentProposal, currentEvents)
		if err != nil {
			return err
		}
		if err := requireLineageTerminalEligibility("accept", proposal.ID, currentEvaluation.Lineage); err != nil {
			return err
		}
		if !currentEvaluation.Ready {
			return fmt.Errorf("proposal is not ready; run 'nh proposal status %s' for missing evidence", shortID(proposal.ID))
		}
		evaluation = currentEvaluation
	}
	event, err := nextEvent(identity, "proposal.decision")
	if err != nil {
		return err
	}
	event.Subject = proposal.ID
	event.Policy = evaluation.PolicyDigest
	event.Body = *body
	if *accept {
		event.Verdict = "accept"
		event.Evidence = append([]string(nil), evaluation.Evidence...)
	} else {
		event.Verdict = "reject"
	}
	stored, err := appendEvent(event, identity)
	if err != nil {
		return err
	}
	fmt.Printf("Recorded %s decision %s on proposal %s under policy %s\n", event.Verdict, shortID(stored.ID), shortID(proposal.ID), shortID(event.Policy))
	if len(event.Evidence) > 0 {
		fmt.Printf("Evidence: %d signed event(s)\n", len(event.Evidence))
	}
	return nil
}

func cmdMerge(args []string) error {
	if len(args) != 1 {
		return usageError("usage: nh merge PROPOSAL")
	}
	if err := prepareShallowVerification(shallowVerificationScope{Operation: "proposal merge", Subject: args[0]}); err != nil {
		return err
	}
	if err := guardShallowEventClosure("proposal merge"); err != nil {
		return err
	}
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEventDependency("proposal merge", args[0], shallowCandidateEvent, events)
	if err != nil {
		return err
	}
	if err := guardProposalEvaluationDependencies("proposal merge", proposal, events); err != nil {
		return err
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		return err
	}
	if err := requireLineageTerminalEligibility("merge", proposal.ID, evaluation.Lineage); err != nil {
		return err
	}
	if evaluation.Merged {
		return fmt.Errorf("proposal %s is already recorded as merged", proposal.ID)
	}
	if evaluation.Rejected {
		return fmt.Errorf("proposal has a current maintainer rejection")
	}
	if !evaluation.Accepted {
		return fmt.Errorf("proposal lacks %d required acceptance decision(s)", evaluation.Policy.Proposals.RequiredAccepts)
	}
	if err := requireProposalCode(proposal); err != nil {
		return err
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	if !actorListed(identity.Actor, evaluation.Policy.Maintainers) {
		return fmt.Errorf("identity %s is not a maintainer under proposal policy", shortID(identity.Actor))
	}
	branch, err := gitText("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("merge requires a checked-out target branch")
	}
	worktree, err := gitText("status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return err
	}
	if worktree != "" {
		return fmt.Errorf("worktree must be clean before merging")
	}
	current, err := resolveCommit("HEAD")
	if err != nil {
		return err
	}
	if err := guardMergeAncestry("proposal merge", proposal, current); err != nil {
		return err
	}
	contained, missing, err := exactCommitAncestorUntil(proposal.Event.Head, current, proposal.Event.Base)
	if err != nil {
		return err
	}
	if missing != "" {
		return classifyShallowDependency(exactDependency{
			Operation: "proposal merge", Kind: shallowMergeAncestor, MissingID: missing,
			ObjectType: "commit", OwnerKind: replicationProposal, OwnerID: proposal.ID,
		}, fmt.Errorf("exact proposal containment requires unavailable parent %s", missing))
	}
	if contained {
		return fmt.Errorf("proposal head is already contained in branch %s", branch)
	}
	message := fmt.Sprintf("Merge NH proposal %s: %s", shortID(proposal.ID), oneLine(evaluation.DisplayTitle))
	if _, err := gitOutput("merge", "--no-ff", "-m", message, proposal.Event.Head); err != nil {
		if _, abortErr := gitOutput("merge", "--abort"); abortErr != nil {
			return fmt.Errorf("merge failed and automatic abort also failed: %v; abort error: %v", err, abortErr)
		}
		return fmt.Errorf("merge failed and was aborted for proposal %s; recover with 'nh proposal revise %s --base REV --head REV': %w", proposal.ID, proposal.ID, err)
	}
	mergeCommit, err := resolveCommit("HEAD")
	if err != nil {
		return fmt.Errorf("code merged but resulting commit could not be resolved: %w", err)
	}
	event, err := nextEvent(identity, "proposal.merged")
	if err != nil {
		return fmt.Errorf("code merged at %s but merge event creation failed: %w", shortOID(mergeCommit), err)
	}
	event.Subject = proposal.ID
	event.Policy = evaluation.PolicyDigest
	event.Head = proposal.Event.Head
	event.Commit = mergeCommit
	event.Evidence = sortedDecisionIDs(evaluation.AcceptDecisions)
	stored, err := appendEvent(event, identity)
	if err != nil {
		return fmt.Errorf("code merged at %s but recording failed: %w", shortOID(mergeCommit), err)
	}
	fmt.Printf("Merged proposal %s into %s at %s\n", shortID(proposal.ID), branch, shortOID(mergeCommit))
	fmt.Printf("Recorded merge event %s with %d acceptance decision(s)\n", shortID(stored.ID), len(event.Evidence))
	return nil
}

func requireLineageTerminalEligibility(operation, proposalID string, state proposalLineageState) error {
	statusCommand := fmt.Sprintf("nh proposal status %s", proposalID)
	if state.MergeConflict {
		return fmt.Errorf("cannot %s proposal %s: lineage has competing merges at %s; run '%s'", operation, proposalID, strings.Join(state.MergedCandidateIDs, ", "), statusCommand)
	}
	if state.Superseded {
		return fmt.Errorf("cannot %s proposal %s: superseded by %s; run '%s'", operation, proposalID, strings.Join(state.SuccessorIDs, ", "), statusCommand)
	}
	if state.LineageClosed {
		return fmt.Errorf("cannot %s proposal %s: lineage is closed by merged proposal %s; run '%s'", operation, proposalID, strings.Join(state.MergedCandidateIDs, ", "), statusCommand)
	}
	return nil
}
