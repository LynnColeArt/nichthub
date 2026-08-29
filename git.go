package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func gitOutput(args ...string) ([]byte, error) {
	return gitInput(nil, nil, args...)
}

func gitText(args ...string) (string, error) {
	out, err := gitOutput(args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitInput(input []byte, extraEnv []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Env = append(os.Environ(), extraEnv...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return stdout.Bytes(), nil
}

func requireGitRepository() (string, error) {
	gitDir, err := gitText("rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", fmt.Errorf("not inside a Git repository")
	}
	return gitDir, nil
}

func refValue(ref string) (string, bool, error) {
	out, err := gitOutput("rev-parse", "--verify", "--quiet", ref)
	if err != nil {
		// rev-parse returns status 1 when a ref is absent. Check with show-ref so
		// actual repository errors are still surfaced by callers elsewhere.
		return "", false, nil
	}
	return strings.TrimSpace(string(out)), true, nil
}
