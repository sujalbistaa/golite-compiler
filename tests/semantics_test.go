package tests

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"golite.dev/mvp/internal/evaluator"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/object"
	"golite.dev/mvp/internal/parser"
	"golite.dev/mvp/internal/semantics"
)

func TestTypeCheckerErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{
			"let x = 10; let y = true; x + y;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"foobar;",
			"identifier not found: foobar",
		},
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true;",
			"unknown operator: -BOOLEAN",
		},
		{
			"if (10) { 1 }",
			"if condition must be a boolean, got INTEGER",
		},
		{
			"let add = 10; add(1,2);",
			"not a function: add",
		},
		{
			"let x = 1; let x = 2;",
			"", // Shadowing is allowed for now
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("parser error on input '%s': %s", tt.input, p.Errors()[0])
		}

		checker := semantics.New()
		checker.Check(program)
		errors := checker.Errors()

		if len(errors) == 0 {
			if tt.expectedError != "" {
				t.Errorf("expected error but got none for input: %s", tt.input)
			}
			continue
		}

		if !strings.Contains(errors[0], tt.expectedError) {
			t.Errorf("wrong error message for input '%s'.\nexpected: %q\ngot:      %q",
				tt.input, tt.expectedError, errors[0])
		}
	}
}

func TestEvaluator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print 5;", "5\n"},
		{"print true;", "true\n"},
		{"print 10 + 5 * 2;", "20\n"},
		{"print (10 + 5) * 2;", "30\n"},
		{"let x = 10; print x;", "10\n"},
		{"let x = 10; let y = 20; print x + y;", "30\n"},
		{
			`let multiply = func(x, y) { x * y; };
             print multiply(5, 10);`,
			"50\n",
		},
		{
			`if (10 > 1) { print 1; } else { print 0; }`,
			"1\n",
		},
		{
			`if (10 < 1) { print 1; } else { print 0; }`,
			"0\n",
		},
		{
			`if (true) { print 99; }`,
			"99\n",
		},
	}

	for _, tt := range tests {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute test
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		evaluator.Eval(program, env)

		// Restore stdout and read captured output
		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)

		if buf.String() != tt.expected {
			t.Errorf("unexpected output for '%s'. expected=%q, got=%q", tt.input, tt.expected, buf.String())
		}
	}
}

// LINES: 125
