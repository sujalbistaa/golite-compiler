package tests

import (
	"testing"

	"golite.dev/mvp/internal/ast"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/parser"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestConstantFolding(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"let x = 5 + 5;",
			"let x = 10;",
		},
		{
			"let x = 10 * 2 - 5;",
			"let x = 15;",
		},
		{
			"let x = (2 + 3) * 4;",
			"let x = 20;",
		},
		{
			"let x = 10 / 0;", // Should not fold division by zero
			"let x = (10 / 0);",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		config := optimizer.Config{EnabledPasses: optimizer.ConstantFolding}
		optimizer.Optimize(program, config)

		if program.String() != tt.expected {
			t.Errorf("constant folding failed for input '%s'.\nexpected=%q\ngot=%q",
				tt.input, tt.expected, program.String())
		}
	}
}

func TestDeadCodeElimination(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"if (true) { print 1; }",
			"iftrue { print 1; }", // The if expression remains, alternative is implicitly nil
		},
		{
			"if (true) { print 1; } else { print 2; }",
			"iftrue { print 1; }else {  }", // Alternative block is emptied
		},
		{
			"if (false) { print 1; }",
			"iffalse {  }", // Consequence block is emptied
		},
		{
			"if (false) { print 1; } else { print 2; }",
			"iffalse {  }else { print 2; }", // Consequence block is emptied
		},
		{
			"let x = 10; if (x > 5) { print x; }", // Condition not constant, no change
			"let x = 10;if(x > 5) { print x; }",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		config := optimizer.Config{EnabledPasses: optimizer.DeadCodeElimination}
		optimizer.Optimize(program, config)
		actual := program.String()

		// Compare strings after removing whitespace to avoid formatting-related failures.
		if removeWhitespace(actual) != removeWhitespace(tt.expected) {
			t.Errorf("DCE failed for input '%s'.\nexpected=%q\ngot=%q",
				tt.input, tt.expected, actual)
		}
	}
}

// removeWhitespace is a test helper to compare AST strings without worrying about spacing.
func removeWhitespace(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			b = append(b, c)
		}
	}
	return string(b)
}

func TestCombinedPasses(t *testing.T) {
	input := "let x = 2 * 5; if (x == 10) { print 1 + 2; }"
	// The optimizer is not smart enough yet to propagate constants.
	// So `x == 10` is not folded, but `1 + 2` is.
	expected := "let x = 10;if((x == 10)) { print 3; }"

	program := parse(input)
	config := optimizer.Config{EnabledPasses: optimizer.AllPasses}
	optimizer.Optimize(program, config)
	actual := program.String()

	if removeWhitespace(actual) != removeWhitespace(expected) {
		t.Errorf("combined passes failed.\nexpected=%q\ngot=%q", expected, actual)
	}
}
