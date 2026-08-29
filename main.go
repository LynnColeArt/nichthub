package main

import (
	"errors"
	"fmt"
	"os"
)

const version = "0.0.1-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nh: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "--version":
		fmt.Println(version)
		return nil
	case "init":
		return cmdInit(args[1:])
	case "identity":
		return cmdIdentity(args[1:])
	case "issue":
		return cmdIssue(args[1:])
	case "proposal":
		return cmdProposal(args[1:])
	case "review":
		return cmdReview(args[1:])
	case "run":
		return cmdRun(args[1:])
	case "runner":
		return cmdRunner(args[1:])
	case "decide":
		return cmdDecide(args[1:])
	case "merge":
		return cmdMerge(args[1:])
	case "sync":
		return cmdSync(args[1:])
	case "log":
		return cmdLog(args[1:])
	default:
		return fmt.Errorf("unknown command %q; run 'nh help'", args[0])
	}
}

func printUsage() {
	fmt.Print(`Nichthub distributes collaboration with a Git repository.

Usage:
  nh init [--name NAME]
  nh identity show
  nh issue open [--body TEXT] TITLE
  nh issue comment ISSUE [--body TEXT] [TEXT]
  nh issue list
  nh issue show ISSUE
  nh proposal open --base REV --head REV [--body TEXT] TITLE
  nh proposal list
  nh proposal show PROPOSAL
  nh proposal status PROPOSAL
  nh review PROPOSAL <--approve|--request-changes> [--body TEXT]
  nh run request PROPOSAL PIPELINE
  nh run list
  nh run show REQUEST
  nh run execute REQUEST [--backend sandbox|host]
  nh run logs RESULT
  nh runner once --accept-pipeline NAME --accept-actor ACTOR
  nh runner watch --accept-pipeline NAME --accept-actor ACTOR
  nh decide PROPOSAL <--accept|--reject> [--body TEXT]
  nh merge PROPOSAL
  nh sync [REMOTE]
  nh log
  nh version
`)
}

func usageError(message string) error {
	return errors.New(message)
}
