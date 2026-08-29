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
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEvent(events, query)
	if err != nil {
		return err
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		return err
	}
	if evaluation.Merged {
		return fmt.Errorf("proposal %s is already merged", shortID(proposal.ID))
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
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEvent(events, args[0])
	if err != nil {
		return err
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		return err
	}
	if evaluation.Merged {
		return fmt.Errorf("proposal %s is already recorded as merged", shortID(proposal.ID))
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
	if _, err := gitOutput("merge-base", "--is-ancestor", proposal.Event.Base, current); err != nil {
		return fmt.Errorf("current branch %s does not descend from proposal base %s", branch, shortOID(proposal.Event.Base))
	}
	if _, err := gitOutput("merge-base", "--is-ancestor", proposal.Event.Head, current); err == nil {
		return fmt.Errorf("proposal head is already contained in branch %s", branch)
	}
	message := fmt.Sprintf("Merge NH proposal %s: %s", shortID(proposal.ID), oneLine(proposal.Event.Title))
	if _, err := gitOutput("merge", "--no-ff", "-m", message, proposal.Event.Head); err != nil {
		if _, abortErr := gitOutput("merge", "--abort"); abortErr != nil {
			return fmt.Errorf("merge failed and automatic abort also failed: %v; abort error: %v", err, abortErr)
		}
		return fmt.Errorf("merge failed and was aborted: %w", err)
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
