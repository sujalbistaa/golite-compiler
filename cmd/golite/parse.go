package main

import (
	"fmt"
	"io"
	"os"

	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/parser"
)

func handleParseCommand() {
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
		printParserErrors(os.Stderr, p.Errors())
		os.Exit(1)
	}

	fmt.Println(program.String())
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, "parser errors:")
	for _, msg := range errors {
		fmt.Fprintln(out, "\t"+msg)
	}
}

// LINES: 50
