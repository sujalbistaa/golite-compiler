package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golite.dev/mvp/internal/codegen"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/parser"
)

func handleBuildCommand() {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	outputFile := buildCmd.String("o", "", "Output file name for the C code.")

	// Correctly parse flags from the arguments that follow the "build" command.
	buildCmd.Parse(os.Args[2:])

	// After parsing, check for the required positional argument (the source file).
	if buildCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: golite build [flags] <file>")
		os.Exit(1)
	}
	filePath := buildCmd.Arg(0) // This is the first non-flag argument.

	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", filePath, err)
		os.Exit(1)
	}

	// Set default output file name if not provided.
	if *outputFile == "" {
		baseName := filepath.Base(filePath)
		ext := filepath.Ext(baseName)
		*outputFile = strings.TrimSuffix(baseName, ext) + ".c"
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printErrors(os.Stderr, "parser errors", p.Errors())
		os.Exit(1)
	}

	generator := codegen.New()
	cCode := generator.Generate(program)

	err = os.WriteFile(*outputFile, []byte(cCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing C code to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled '%s' to '%s'.\n", filePath, *outputFile)
}
