package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/profiler"
)

// RealExecutor is an implementation of the Executor interface that runs real commands.
type RealExecutor struct{}

func (r *RealExecutor) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func handleProfileCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: golite profile <file>")
		os.Exit(1)
	}
	sourceFile := os.Args[2]

	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: source file does not exist: %s\n", sourceFile)
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "golite-profile-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// By default, the profile command runs with all optimizations enabled.
	optConfig := optimizer.Config{EnabledPasses: optimizer.AllPasses}

	prof := profiler.New(&RealExecutor{}, tempDir)
	metrics, err := prof.Run(sourceFile, optConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during profiling: %v\n", err)
		os.Exit(1)
	}

	jsonOutput, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting metrics to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}

// LINES: 59
