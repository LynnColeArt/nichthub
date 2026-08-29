package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type runnerOptions struct {
	Remote          string
	Pipeline        string
	AcceptedActor   string
	Backend         executionBackend
	Interval        time.Duration
	ReportNoPending bool
}

func cmdRunner(args []string) error {
	if len(args) == 0 {
		return usageError("usage: nh runner <once|watch> --accept-pipeline NAME --accept-actor ACTOR")
	}
	switch args[0] {
	case "once":
		options, err := parseRunnerOptions("runner once", args[1:], false)
		if err != nil {
			return err
		}
		options.ReportNoPending = true
		_, err = runnerOnce(options)
		return err
	case "watch":
		options, err := parseRunnerOptions("runner watch", args[1:], true)
		if err != nil {
			return err
		}
		return runnerWatch(options)
	default:
		return fmt.Errorf("unknown runner command %q", args[0])
	}
}

func parseRunnerOptions(name string, args []string, watching bool) (runnerOptions, error) {
	flags := quietFlags(name)
	remote := flags.String("remote", "origin", "Git remote to synchronize")
	pipeline := flags.String("accept-pipeline", "", "only execute this pipeline name")
	actor := flags.String("accept-actor", "", "only execute requests signed by this actor")
	backendName := flags.String("backend", "sandbox", "execution backend: sandbox or host")
	allowHost := flags.Bool("allow-unsafe-host-execution", false, "permit an unsandboxed host backend")
	interval := flags.Duration("interval", 30*time.Second, "watch polling interval")
	if err := flags.Parse(args); err != nil {
		return runnerOptions{}, err
	}
	if flags.NArg() != 0 || !validPipelineName(*pipeline) || !validActorID(*actor) {
		return runnerOptions{}, usageError("runner requires --accept-pipeline NAME and a full --accept-actor ACTOR fingerprint")
	}
	if watching && *interval < time.Second {
		return runnerOptions{}, fmt.Errorf("runner watch interval must be at least 1s")
	}
	backend, err := selectBackend(*backendName, *allowHost)
	if err != nil {
		return runnerOptions{}, err
	}
	if _, err := loadIdentity(); err != nil {
		return runnerOptions{}, err
	}
	return runnerOptions{
		Remote:        *remote,
		Pipeline:      *pipeline,
		AcceptedActor: *actor,
		Backend:       backend,
		Interval:      *interval,
	}, nil
}

func validActorID(actor string) bool {
	decoded, err := hex.DecodeString(actor)
	return err == nil && len(decoded) == 32
}

func runnerOnce(options runnerOptions) (bool, error) {
	return runnerOnceContext(context.Background(), options)
}

func runnerOnceContext(ctx context.Context, options runnerOptions) (bool, error) {
	if err := cmdSync([]string{options.Remote}); err != nil {
		return false, err
	}
	events, err := collectEvents()
	if err != nil {
		return false, err
	}
	identity, err := loadIdentity()
	if err != nil {
		return false, err
	}
	var firstError error
	for index := range events {
		request := &events[index]
		if request.Event.Kind != "run.request" || request.Event.Pipeline != options.Pipeline || request.Event.Actor != options.AcceptedActor {
			continue
		}
		alreadyRan := false
		for _, result := range currentRunResults(events, request.ID) {
			if result.Event.Actor == identity.Actor {
				alreadyRan = true
				break
			}
		}
		if alreadyRan {
			continue
		}
		fmt.Printf("Runner accepted request %s from %s for pipeline %s\n", shortID(request.ID), shortID(request.Event.Actor), oneLine(request.Event.Pipeline))
		stored, err := executeRunRequest(ctx, events, request, identity, options.Backend, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nh runner: skipped %s: %v\n", shortID(request.ID), err)
			if firstError == nil {
				firstError = err
			}
			continue
		}
		fmt.Printf("Runner recorded %s result %s with %s backend\n", stored.Event.Outcome, shortID(stored.ID), options.Backend.Name())
		if err := cmdSync([]string{options.Remote}); err != nil {
			return true, err
		}
		return true, nil
	}
	if firstError != nil {
		return false, firstError
	}
	if options.ReportNoPending {
		fmt.Printf("No pending %s requests from actor %s.\n", oneLine(options.Pipeline), shortID(options.AcceptedActor))
	}
	return false, nil
}

func runnerWatch(options runnerOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fmt.Printf("Watching %s for pipeline %s requests from actor %s with %s backend\n", options.Remote, oneLine(options.Pipeline), shortID(options.AcceptedActor), options.Backend.Name())
	for {
		ran, err := runnerOnceContext(ctx, options)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Println("Runner stopped.")
				return nil
			}
			fmt.Fprintf(os.Stderr, "nh runner: %v\n", err)
			ran = false
		}
		if ran {
			continue
		}
		timer := time.NewTimer(options.Interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			fmt.Println("Runner stopped.")
			return nil
		case <-timer.C:
		}
	}
}
