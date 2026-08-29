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
		return usageError("usage: nh init [--name NAME]")
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
	if len(args) != 1 || args[0] != "show" {
		return usageError("usage: nh identity show")
	}
	identity, err := loadIdentity()
	if err != nil {
		return err
	}
	fmt.Printf("Name:   %s\n", oneLine(identity.Name))
	fmt.Printf("Actor:  %s\n", identity.Actor)
	fmt.Printf("Ref:    %s\n", actorRef(identity.Actor))
	return nil
}

func cmdIssue(args []string) error {
	if len(args) == 0 {
		return usageError("usage: nh issue <open|comment|list|show>")
	}
	switch args[0] {
	case "open":
		return cmdIssueOpen(args[1:])
	case "comment":
		return cmdIssueComment(args[1:])
	case "list":
		if len(args) != 1 {
			return usageError("usage: nh issue list")
		}
		return cmdIssueList()
	case "show":
		if len(args) != 2 {
			return usageError("usage: nh issue show ISSUE")
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
		return usageError("usage: nh issue open [--body TEXT] TITLE")
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
		return usageError("usage: nh issue comment ISSUE [--body TEXT] [TEXT]")
	}
	subjectQuery := args[0]
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
	if len(args) > 1 {
		return usageError("usage: nh sync [REMOTE]")
	}
	remote := "origin"
	if len(args) == 1 {
		remote = args[0]
	}
	if _, err := gitOutput("remote", "get-url", remote); err != nil {
		return fmt.Errorf("remote %q does not exist", remote)
	}

	actorRefspec := "+refs/nh/actors/*:refs/nh/remotes/" + remote + "/actors/*"
	proposalRefspec := "+refs/nh/proposals/*:refs/nh/remotes/" + remote + "/proposals/*"
	if _, err := gitOutput("fetch", "--no-tags", remote, actorRefspec, proposalRefspec); err != nil {
		return fmt.Errorf("fetch Nichthub refs: %w", err)
	}

	identity, identityErr := loadIdentity()
	if identityErr == nil {
		ref := actorRef(identity.Actor)
		if _, exists, err := refValue(ref); err != nil {
			return err
		} else if exists {
			pushSpec := ref + ":" + ref
			if _, err := gitOutput("push", remote, pushSpec); err != nil {
				return fmt.Errorf("publish actor events: %w", err)
			}
		}
	}
	proposalRefs, err := gitText("for-each-ref", "--format=%(refname)", "refs/nh/proposals")
	if err != nil {
		return err
	}
	for _, ref := range strings.Fields(proposalRefs) {
		if _, err := gitOutput("push", remote, ref+":"+ref); err != nil {
			return fmt.Errorf("publish proposal code: %w", err)
		}
	}
	if _, err := gitOutput("fetch", "--no-tags", remote, actorRefspec, proposalRefspec); err != nil {
		return fmt.Errorf("refresh Nichthub refs: %w", err)
	}

	events, err := collectEvents()
	if err != nil {
		return err
	}
	fmt.Printf("Synchronized %d verified events with %s\n", len(events), remote)
	return nil
}

func cmdLog(args []string) error {
	if len(args) != 0 {
		return usageError("usage: nh log")
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
