package main

import (
	"fmt"
	"io"
	"os"

	"golite.dev/mvp/internal/evaluator"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/object"
	"golite.dev/mvp/internal/parser"
	"golite.dev/mvp/internal/semantics"
)

func handleRunCommand() {
	var input []byte
	var err error

	if len(os.Args) > 2 {
		filePath := os.Args[2]
		input, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", filePath, err)
			os.Exit(1)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %s\n", err)
			os.Exit(1)
		}
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printErrors(os.Stderr, "parser errors", p.Errors())
		os.Exit(1)
	}

	checker := semantics.New()
	checker.Check(program)
	if len(checker.Errors()) != 0 {
		printErrors(os.Stderr, "semantic errors", checker.Errors())
		os.Exit(1)
	}

	env := object.NewEnvironment()
	evaluated := evaluator.Eval(program, env)
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		errors := []string{evaluated.Inspect()}
		printErrors(os.Stderr, "runtime error", errors)
		os.Exit(1)
	}
}

func printErrors(out io.Writer, errorType string, errors []string) {
	fmt.Fprintf(out, "Encountered %s:\n", errorType)
	for _, msg := range errors {
		fmt.Fprintln(out, "\t"+msg)
	}
}

// LINES: 66
