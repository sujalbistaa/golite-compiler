package main

import (
	"flag"
	"fmt"
	"os"

	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/parser"
)

func handleOptimizeCommand() {
	optCmd := flag.NewFlagSet("optimize", flag.ExitOnError)
	dumpAST := optCmd.Bool("dump-ast", false, "Print the optimized Abstract Syntax Tree.")
	constFold := optCmd.Bool("const-fold", false, "Enable constant folding.")
	dce := optCmd.Bool("dce", false, "Enable dead code elimination.")

	// The first arg is the command name, so we parse from the 2nd arg onwards.
	optCmd.Parse(os.Args[2:])

	if optCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: golite optimize [flags] <file>")
		os.Exit(1)
	}

	filePath := optCmd.Arg(0)
	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", filePath, err)
		os.Exit(1)
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printErrors(os.Stderr, "parser errors", p.Errors())
		os.Exit(1)
	}

	var enabledPasses optimizer.Pass
	if *constFold {
		enabledPasses |= optimizer.ConstantFolding
	}
	if *dce {
		enabledPasses |= optimizer.DeadCodeElimination
	}

	// If no flags are specified, enable all passes by default.
	if !*constFold && !*dce {
		enabledPasses = optimizer.AllPasses
	}

	config := optimizer.Config{EnabledPasses: enabledPasses}
	optimizedProgram := optimizer.Optimize(program, config)

	if *dumpAST {
		fmt.Println(optimizedProgram.String())
	} else {
		fmt.Println("Optimization complete. Use --dump-ast to view the result.")
	}
}
