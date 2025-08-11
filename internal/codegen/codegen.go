package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"golite.dev/mvp/internal/ast"
)

// CGen is the C code generator.
type CGen struct {
	builder strings.Builder
}

// New creates a new C code generator.
func New() *CGen {
	return &CGen{}
}

// Generate takes an AST program and returns a string of equivalent C code.
func (c *CGen) Generate(program *ast.Program) string {
	c.builder.WriteString("#include <stdio.h>\n")
	c.builder.WriteString("#include <stdint.h>\n")
	c.builder.WriteString("#include <stdbool.h>\n\n")
	c.builder.WriteString("int main() {\n")

	for _, stmt := range program.Statements {
		c.genStatement(stmt, 1)
	}

	c.builder.WriteString("    return 0;\n")
	c.builder.WriteString("}\n")
	return c.builder.String()
}

func (c *CGen) writeIndent(level int) {
	c.builder.WriteString(strings.Repeat("    ", level))
}

func (c *CGen) genStatement(stmt ast.Statement, level int) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		c.writeIndent(level)
		// Assuming all variables are 64-bit integers for now.
		c.builder.WriteString("int64_t ")
		c.builder.WriteString(s.Name.Value)
		c.builder.WriteString(" = ")
		c.genExpression(s.Value)
		c.builder.WriteString(";\n")
	case *ast.PrintStatement:
		c.writeIndent(level)
		c.builder.WriteString("printf(\"%lld\\n\", (long long)(")
		c.genExpression(s.Expression)
		c.builder.WriteString("));\n")
	case *ast.ExpressionStatement:
		c.writeIndent(level)
		c.genExpression(s.Expression)
		c.builder.WriteString(";\n")
	}
}

func (c *CGen) genExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		c.builder.WriteString(strconv.FormatInt(e.Value, 10))
	case *ast.Boolean:
		c.builder.WriteString(strconv.FormatBool(e.Value))
	case *ast.Identifier:
		c.builder.WriteString(e.Value)
	case *ast.InfixExpression:
		c.builder.WriteString("(")
		c.genExpression(e.Left)
		c.builder.WriteString(" ")
		// Note: C uses `&&` and `||` for logical ops, but we haven't added those yet.
		// GoLite `==` maps to C `==`, etc.
		c.builder.WriteString(e.Operator)
		c.builder.WriteString(" ")
		c.genExpression(e.Right)
		c.builder.WriteString(")")
	case *ast.IfExpression:
		c.builder.WriteString("if (")
		c.genExpression(e.Condition)
		c.builder.WriteString(") {\n")
		c.genBlock(e.Consequence, 1)
		c.writeIndent(1)
		c.builder.WriteString("}")
		if e.Alternative != nil {
			c.builder.WriteString(" else {\n")
			c.genBlock(e.Alternative, 1)
			c.writeIndent(1)
			c.builder.WriteString("}")
		}
	default:
		panic(fmt.Sprintf("unhandled expression type in C codegen: %T", e))
	}
}

func (c *CGen) genBlock(block *ast.BlockStatement, level int) {
	for _, stmt := range block.Statements {
		c.genStatement(stmt, level+1)
	}
}
