package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: golite <command> [arguments]")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "parse":
		handleParseCommand()
	case "run":
		handleRunCommand()
	case "optimize":
		handleOptimizeCommand()
	case "build":
		handleBuildCommand()
	case "profile":
		handleProfileCommand()
	case "evolve":
		handleEvolveCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
