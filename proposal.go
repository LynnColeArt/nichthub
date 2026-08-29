package main

import (
	"fmt"
	"sort"
	"strings"
)

func cmdProposal(args []string) error {
	if len(args) == 0 {
		return usageError("usage: nh proposal <open|list|show|status>")
	}
	switch args[0] {
	case "open":
		return cmdProposalOpen(args[1:])
	case "list":
		if len(args) != 1 {
			return usageError("usage: nh proposal list")
		}
		return cmdProposalList()
	case "show":
		if len(args) != 2 {
			return usageError("usage: nh proposal show PROPOSAL")
		}
		return cmdProposalShow(args[1])
	case "status":
		if len(args) != 2 {
			return usageError("usage: nh proposal status PROPOSAL")
		}
		return cmdProposalStatus(args[1])
	default:
		return fmt.Errorf("unknown proposal command %q", args[0])
	}
}

func cmdProposalOpen(args []string) error {
	flags := quietFlags("proposal open")
	baseRevision := flags.String("base", "", "base Git revision")
	headRevision := flags.String("head", "", "proposed Git revision")
	body := flags.String("body", "", "proposal description")
	if err := flags.Parse(args); err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if *baseRevision == "" || *headRevision == "" || title == "" {
		return usageError("usage: nh proposal open --base REV --head REV [--body TEXT] TITLE")
	}
	base, err := resolveCommit(*baseRevision)
	if err != nil {
		return fmt.Errorf("resolve base: %w", err)
	}
	head, err := resolveCommit(*headRevision)
	if err != nil {
		return fmt.Errorf("resolve head: %w", err)
	}
	if base == head {
		return fmt.Errorf("proposal base and head resolve to the same commit")
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	event, err := nextEvent(identity, "proposal.open")
	if err != nil {
		return err
	}
	event.Title = title
	event.Body = *body
	event.Base = base
	event.Head = head
	stored, err := appendEvent(event, identity)
	if err != nil {
		return err
	}
	if err := createProposalRef(stored.ID, head); err != nil {
		return fmt.Errorf("proposal event %s was created but its code ref failed: %w", shortID(stored.ID), err)
	}
	fmt.Printf("Opened proposal %s: %s\n", shortID(stored.ID), oneLine(title))
	fmt.Printf("Code: %s..%s\n", shortOID(base), shortOID(head))
	return nil
}

func resolveCommit(revision string) (string, error) {
	if strings.HasPrefix(revision, "-") {
		return "", fmt.Errorf("invalid revision %q", revision)
	}
	commit, err := gitText("rev-parse", "--verify", "--end-of-options", revision+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("Git revision %q is not a commit", revision)
	}
	return commit, nil
}

func shortOID(oid string) string {
	if len(oid) > 12 {
		return oid[:12]
	}
	return oid
}

func proposalEvents(events []StoredEvent) []StoredEvent {
	proposals := make([]StoredEvent, 0)
	for _, stored := range events {
		if stored.Event.Kind == "proposal.open" {
			proposals = append(proposals, stored)
		}
	}
	return proposals
}

func currentReviews(events []StoredEvent, proposalID string) []StoredEvent {
	byActor := make(map[string]StoredEvent)
	for _, stored := range events {
		if stored.Event.Kind != "review.submit" || stored.Event.Subject != proposalID {
			continue
		}
		previous, exists := byActor[stored.Event.Actor]
		if !exists || stored.Event.Sequence > previous.Event.Sequence {
			byActor[stored.Event.Actor] = stored
		}
	}
	reviews := make([]StoredEvent, 0, len(byActor))
	for _, stored := range byActor {
		reviews = append(reviews, stored)
	}
	sortStoredEvents(reviews)
	return reviews
}

func sortStoredEvents(events []StoredEvent) {
	// Sequence is actor-local, so timestamp is only a presentation order.
	// Event identity and validity never depend on this ordering.
	sort.Slice(events, func(i, j int) bool {
		if events[i].Event.Timestamp != events[j].Event.Timestamp {
			return events[i].Event.Timestamp < events[j].Event.Timestamp
		}
		return events[i].ID < events[j].ID
	})
}

func cmdProposalList() error {
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposals := proposalEvents(events)
	if len(proposals) == 0 {
		fmt.Println("No proposals.")
		return nil
	}
	for _, proposal := range proposals {
		reviews := currentReviews(events, proposal.ID)
		approvals, changes := reviewCounts(reviews)
		availability := "code-missing"
		if head, exists, err := proposalHead(proposal.ID); err != nil {
			return err
		} else if exists && head == proposal.Event.Head {
			availability = "code-ready"
		} else if exists {
			availability = "code-mismatch"
		}
		fmt.Printf("%s  %s  +%d/-%d  %s\n", shortID(proposal.ID), oneLine(proposal.Event.Title), approvals, changes, availability)
	}
	return nil
}

func cmdProposalShow(query string) error {
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEvent(events, query)
	if err != nil {
		return err
	}
	if proposal.Event.Kind != "proposal.open" {
		return fmt.Errorf("%s is not a proposal", shortID(proposal.ID))
	}
	fmt.Printf("%s  %s\n", shortID(proposal.ID), oneLine(proposal.Event.Title))
	fmt.Printf("Proposed by %s at %s\n", oneLine(proposal.Event.ActorName), proposal.Event.Timestamp)
	fmt.Printf("Base: %s\nHead: %s\n", proposal.Event.Base, proposal.Event.Head)
	head, exists, err := proposalHead(proposal.ID)
	if err != nil {
		return err
	}
	switch {
	case !exists:
		fmt.Println("Code: unavailable")
	case head != proposal.Event.Head:
		fmt.Printf("Code: REF MISMATCH (%s)\n", head)
	default:
		fmt.Println("Code: available and matched")
	}
	if proposal.Event.Body != "" {
		fmt.Printf("\n%s\n", safeText(proposal.Event.Body))
	}
	reviews := currentReviews(events, proposal.ID)
	if len(reviews) > 0 {
		fmt.Println("\nCurrent reviews:")
		for _, review := range reviews {
			fmt.Printf("\n%s: %s (%s)\n", oneLine(review.Event.ActorName), review.Event.Verdict, shortID(review.ID))
			if review.Event.Body != "" {
				fmt.Printf("%s\n", safeText(review.Event.Body))
			}
		}
	}
	return nil
}

func reviewCounts(reviews []StoredEvent) (approvals, changes int) {
	for _, review := range reviews {
		switch review.Event.Verdict {
		case "approve":
			approvals++
		case "request-changes":
			changes++
		}
	}
	return approvals, changes
}

func cmdReview(args []string) error {
	if len(args) < 1 {
		return usageError("usage: nh review PROPOSAL <--approve|--request-changes> [--body TEXT]")
	}
	proposalQuery := args[0]
	flags := quietFlags("review")
	approve := flags.Bool("approve", false, "approve the proposal")
	requestChanges := flags.Bool("request-changes", false, "request changes")
	body := flags.String("body", "", "review explanation")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || *approve == *requestChanges {
		return usageError("choose exactly one of --approve or --request-changes")
	}
	events, err := collectEvents()
	if err != nil {
		return err
	}
	proposal, err := resolveEvent(events, proposalQuery)
	if err != nil {
		return err
	}
	if proposal.Event.Kind != "proposal.open" {
		return fmt.Errorf("%s is not a proposal", shortID(proposal.ID))
	}
	head, exists, err := proposalHead(proposal.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("proposal code is unavailable; sync before reviewing")
	}
	if head != proposal.Event.Head {
		return fmt.Errorf("proposal code ref does not match the signed proposal")
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	event, err := nextEvent(identity, "review.submit")
	if err != nil {
		return err
	}
	event.Subject = proposal.ID
	event.Body = *body
	if *approve {
		event.Verdict = "approve"
	} else {
		event.Verdict = "request-changes"
	}
	stored, err := appendEvent(event, identity)
	if err != nil {
		return err
	}
	fmt.Printf("Recorded %s review %s on proposal %s\n", event.Verdict, shortID(stored.ID), shortID(proposal.ID))
	return nil
}
