package main

import (
	"flag"
	"fmt"
	"os"

	"golite.dev/mvp/internal/selfevolve"
)

func handleEvolveCommand() {
	evolveCmd := flag.NewFlagSet("evolve", flag.ExitOnError)
	generations := evolveCmd.Int("generations", 10, "Number of generations to run.")

	evolveCmd.Parse(os.Args[2:])

	if evolveCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: golite evolve [flags] <corpus-dir>")
		os.Exit(1)
	}
	corpusDir := evolveCmd.Arg(0)

	// Check if corpus directory exists
	info, err := os.Stat(corpusDir)
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: corpus directory does not exist: %s\n", corpusDir)
		os.Exit(1)
	}

	// The runner needs a temporary directory for the profiler.
	tempDir, err := os.MkdirTemp("", "golite-evolve-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// The `RealExecutor` is defined in profile.go.
	runner := selfevolve.NewRunner(&RealExecutor{}, tempDir)

	bestIndividual, err := runner.Run(corpusDir, *generations)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Evolution process failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nEvolution complete.")
	fmt.Println("Best configuration found:")
	fmt.Printf("  - Passes: %s\n", bestIndividual.PassNames())
	fmt.Printf("  - Fitness Score: %.2f\n", bestIndividual.Fitness)
}

// LINES: 55
