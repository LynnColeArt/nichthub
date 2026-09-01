package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

func quietFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	return flags
}

func cmdInit(args []string) error {
	flags := quietFlags("init")
	name := flags.String("name", "", "display name")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return usageError("usage: hn init [--name NAME]")
	}
	identity, path, err := createIdentity(*name)
	if err != nil {
		return err
	}
	fmt.Printf("Created identity %s (%s)\n", oneLine(identity.Name), shortID(identity.Actor))
	fmt.Printf("Private identity stored in %s\n", path)
	return nil
}

func cmdIdentity(args []string) error {
	if len(args) == 0 {
		return usageError("usage: hn identity <show|list|public|authorize|accept|rotate>")
	}
	switch args[0] {
	case "show":
		return cmdIdentityShow(args[1:])
	case "list":
		return cmdIdentityList(args[1:])
	case "public":
		return cmdIdentityPublic(args[1:])
	case "authorize":
		return cmdIdentityAuthorize(args[1:])
	case "accept":
		return cmdIdentityAccept(args[1:])
	case "rotate":
		return cmdIdentityRotate(args[1:])
	default:
		return fmt.Errorf("unknown identity command %q", args[0])
	}
}

func cmdIssue(args []string) error {
	if len(args) == 0 {
		return usageError("usage: hn issue <open|comment|list|show>")
	}
	switch args[0] {
	case "open":
		return cmdIssueOpen(args[1:])
	case "comment":
		return cmdIssueComment(args[1:])
	case "list":
		if len(args) != 1 {
			return usageError("usage: hn issue list")
		}
		return cmdIssueList()
	case "show":
		if len(args) != 2 {
			return usageError("usage: hn issue show ISSUE")
		}
		return cmdIssueShow(args[1])
	default:
		return fmt.Errorf("unknown issue command %q", args[0])
	}
}

func cmdIssueOpen(args []string) error {
	flags := quietFlags("issue open")
	body := flags.String("body", "", "issue description")
	if err := flags.Parse(args); err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if title == "" {
		return usageError("usage: hn issue open [--body TEXT] TITLE")
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	event, err := nextEvent(identity, "issue.open")
	if err != nil {
		return err
	}
	event.Title = title
	event.Body = *body
	stored, err := appendEvent(event, identity)
	if err != nil {
		return err
	}
	fmt.Printf("Opened issue %s: %s\n", shortID(stored.ID), oneLine(title))
	return nil
}

func cmdIssueComment(args []string) error {
	if len(args) < 1 {
		return usageError("usage: hn issue comment ISSUE [--body TEXT] [TEXT]")
	}
	subjectQuery := args[0]
	if err := requireFullEventID(subjectQuery); err != nil {
		return err
	}
	flags := quietFlags("issue comment")
	bodyFlag := flags.String("body", "", "comment body")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	events, err := collectEvents()
	if err != nil {
		return err
	}
	subject, err := resolveEvent(events, subjectQuery)
	if err != nil {
		return err
	}
	if subject.Event.Kind != "issue.open" {
		return fmt.Errorf("%s is not an issue", shortID(subject.ID))
	}
	body := *bodyFlag
	if body == "" && flags.NArg() > 0 {
		body = strings.Join(flags.Args(), " ")
	}
	if strings.TrimSpace(body) == "" {
		return usageError("comment body cannot be empty")
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	event, err := nextEvent(identity, "issue.comment")
	if err != nil {
		return err
	}
	event.Subject = subject.ID
	event.Body = body
	stored, err := appendEvent(event, identity)
	if err != nil {
		return err
	}
	fmt.Printf("Commented on %s with event %s\n", shortID(subject.ID), shortID(stored.ID))
	return nil
}

func cmdIssueList() error {
	events, err := collectEvents()
	if err != nil {
		return err
	}
	count := 0
	for _, stored := range events {
		if stored.Event.Kind == "issue.open" {
			fmt.Printf("%s  %s  (%s)\n", shortID(stored.ID), oneLine(stored.Event.Title), oneLine(stored.Event.ActorName))
			count++
		}
	}
	if count == 0 {
		fmt.Println("No issues.")
	}
	return nil
}

func cmdIssueShow(query string) error {
	events, err := collectEvents()
	if err != nil {
		return err
	}
	issue, err := resolveEvent(events, query)
	if err != nil {
		return err
	}
	if issue.Event.Kind != "issue.open" {
		return fmt.Errorf("%s is not an issue", shortID(issue.ID))
	}
	fmt.Printf("%s  %s\n", shortID(issue.ID), oneLine(issue.Event.Title))
	fmt.Printf("Opened by %s at %s\n", oneLine(issue.Event.ActorName), issue.Event.Timestamp)
	if issue.Event.Body != "" {
		fmt.Printf("\n%s\n", safeText(issue.Event.Body))
	}
	comments := make([]StoredEvent, 0)
	for _, stored := range events {
		if stored.Event.Kind == "issue.comment" && stored.Event.Subject == issue.ID {
			comments = append(comments, stored)
		}
	}
	if len(comments) > 0 {
		fmt.Println("\nComments:")
		for _, comment := range comments {
			fmt.Printf("\n%s (%s):\n%s\n", oneLine(comment.Event.ActorName), shortID(comment.ID), safeText(comment.Event.Body))
		}
	}
	return nil
}

func cmdSync(args []string) error {
	remote := "origin"
	remoteSet := false
	recoverShallow := false
	for _, argument := range args {
		switch {
		case argument == "--recover-shallow":
			if recoverShallow {
				return usageError("usage: hn sync [REMOTE] [--recover-shallow]")
			}
			recoverShallow = true
		case strings.HasPrefix(argument, "-"):
			return usageError("usage: hn sync [REMOTE] [--recover-shallow]")
		case !remoteSet:
			remote = argument
			remoteSet = true
		default:
			return usageError("usage: hn sync [REMOTE] [--recover-shallow]")
		}
	}
	if !validReplicationRemote(remote) {
		return fmt.Errorf("invalid remote name %q", remote)
	}
	if recoverShallow {
		return recoverSelectedShallow(remote)
	}
	if _, err := gitOutput("remote", "get-url", remote); err != nil {
		return fmt.Errorf("remote %q does not exist", remote)
	}

	selection, _, err := loadReplicationSelection(remote)
	if err != nil {
		return err
	}
	result, importErr := runReplicationTransaction(selection)
	for _, outcome := range result.Outcomes {
		if outcome.Status == replicationPromoted {
			fmt.Printf("Replication %s %s: promoted\n", outcome.Kind, outcome.ID)
		} else {
			fmt.Printf("Replication %s %s: failed (%s): %s\n", outcome.Kind, outcome.ID, outcome.Status, oneLine(outcome.Diagnostic))
		}
	}

	publicationErr := publishLocalFacts(remote)
	events, collectErr := collectEvents()
	if collectErr == nil {
		fmt.Printf("Synchronized %d verified events with %s; promoted %d selections\n", len(events), remote, result.Promoted)
	}
	if importErr != nil {
		if publicationErr != nil {
			return fmt.Errorf("import phase failed: %v; publication phase failed: %w", importErr, publicationErr)
		}
		return fmt.Errorf("import phase failed (publication phase completed): %w", importErr)
	}
	if publicationErr != nil {
		return fmt.Errorf("publication phase failed after replication: %w", publicationErr)
	}
	if collectErr != nil {
		return collectErr
	}
	if result.hasFailures() {
		return fmt.Errorf("replication import phase failed for one or more exact selections; independently valid selections were promoted")
	}
	return nil
}

func publishLocalFacts(remote string) error {
	actorRefs, err := gitText("for-each-ref", "--format=%(refname)", "refs/hn/actors")
	if err != nil {
		return replicationPhaseError(remote, "actor publication inspection")
	}
	for _, ref := range strings.Fields(actorRefs) {
		if _, err := gitOutput("push", remote, ref+":"+ref); err != nil {
			return replicationPhaseError(remote, "actor publication")
		}
	}
	proposalRefs, err := gitText("for-each-ref", "--format=%(refname)", "refs/hn/proposals")
	if err != nil {
		return replicationPhaseError(remote, "proposal publication inspection")
	}
	for _, ref := range strings.Fields(proposalRefs) {
		if _, err := gitOutput("push", remote, ref+":"+ref); err != nil {
			return replicationPhaseError(remote, "proposal publication")
		}
	}
	memoryRefs, err := gitText("for-each-ref", "--format=%(refname)", memoryRefPrefix)
	if err != nil {
		return replicationPhaseError(remote, "memory publication inspection")
	}
	for _, ref := range strings.Fields(memoryRefs) {
		if _, _, ok := parseMemoryRef(ref); !ok {
			return replicationPhaseError(remote, "memory publication inspection")
		}
		if _, err := gitOutput("push", remote, ref+":"+ref); err != nil {
			return replicationPhaseError(remote, "memory publication")
		}
	}
	return nil
}

func cmdLog(args []string) error {
	if len(args) != 0 {
		return usageError("usage: hn log")
	}
	if err := prepareShallowVerification(shallowVerificationScope{Operation: "log"}); err != nil {
		return err
	}
	if err := guardShallowEventClosure("log"); err != nil {
		return err
	}
	events, err := collectEvents()
	if err != nil {
		return err
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Event.Timestamp != events[j].Event.Timestamp {
			return events[i].Event.Timestamp < events[j].Event.Timestamp
		}
		return events[i].ID < events[j].ID
	})
	for _, stored := range events {
		detail := stored.Event.Title
		if detail == "" {
			detail = stored.Event.Body
		}
		fmt.Printf("%s  %-13s  %-12s  %s\n", shortID(stored.ID), stored.Event.Kind, oneLine(stored.Event.ActorName), oneLine(detail))
	}
	return nil
}

func oneLine(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}

func safeText(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return '�'
		}
		return r
	}, value)
}
