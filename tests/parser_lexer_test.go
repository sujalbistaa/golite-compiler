package tests

import (
	"testing"

	"golite.dev/mvp/internal/ast"
	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/parser"
)

func TestLexer(t *testing.T) {
	input := `let five = 5;
let ten = 10;
if (5 < 10) {
    print true;
} else {
    print false;
}
10 == 10;
10 != 9;
`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.LET, "let"},
		{lexer.IDENT, "five"},
		{lexer.ASSIGN, "="},
		{lexer.INT, "5"},
		{lexer.SEMICOLON, ";"},
		{lexer.LET, "let"},
		{lexer.IDENT, "ten"},
		{lexer.ASSIGN, "="},
		{lexer.INT, "10"},
		{lexer.SEMICOLON, ";"},
		{lexer.IF, "if"},
		{lexer.LPAREN, "("},
		{lexer.INT, "5"},
		{lexer.LT, "<"},
		{lexer.INT, "10"},
		{lexer.RPAREN, ")"},
		{lexer.LBRACE, "{"},
		{lexer.PRINT, "print"},
		{lexer.TRUE, "true"},
		{lexer.SEMICOLON, ";"},
		{lexer.RBRACE, "}"},
		{lexer.ELSE, "else"},
		{lexer.LBRACE, "{"},
		{lexer.PRINT, "print"},
		{lexer.FALSE, "false"},
		{lexer.SEMICOLON, ";"},
		{lexer.RBRACE, "}"},
		{lexer.INT, "10"},
		{lexer.EQ, "=="},
		{lexer.INT, "10"},
		{lexer.SEMICOLON, ";"},
		{lexer.INT, "10"},
		{lexer.NOT_EQ, "!="},
		{lexer.INT, "9"},
		{lexer.SEMICOLON, ";"},
		{lexer.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestParsingLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = true;
let foobar = y;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}

	return true
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T",
			stmt.Expression)
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", exp.Alternative)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)(-5 * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

// LINES: 236
