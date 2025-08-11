package tests

import (
	"strings"
	"testing"

	"golite.dev/mvp/internal/codegen"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/parser"
)

func TestCCodeGen(t *testing.T) {
	input := `
	let x = 10;
	let y = 20;
	let z = x + y;
	print z;

	if (z > 25) {
		print 1;
	} else {
		print 0;
	}
	`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	generator := codegen.New()
	cCode := generator.Generate(program)

	expectedSnippets := []string{
		"#include <stdio.h>",
		"#include <stdint.h>",
		"int main() {",
		"int64_t x = 10;",
		"int64_t y = 20;",
		"int64_t z = (x + y);",
		"printf(\"%lld\\n\", (long long)(z));",
		"if ((z > 25)) {",
		"printf(\"%lld\\n\", (long long)(1));",
		"} else {",
		"printf(\"%lld\\n\", (long long)(0));",
		"}",
		"return 0;",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(cCode, snippet) {
			t.Errorf("Generated C code did not contain expected snippet: %q", snippet)
			t.Logf("Full generated code:\n%s", cCode)
		}
	}
}
