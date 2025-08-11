package optimizer

import (
	"fmt"

	"golite.dev/mvp/internal/ast"
	"golite.dev/mvp/internal/lexer"
)

// visitorFunc is a function that implements the ast.Visitor interface.
type visitorFunc func(node ast.Node) ast.Node

func (f visitorFunc) Visit(node ast.Node) ast.Node {
	return f(node)
}

// Optimize applies a series of AST-to-AST transformations based on the provided configuration.
func Optimize(program *ast.Program, config Config) *ast.Program {
	// We chain visitors. The output of one pass becomes the input to the next.
	// The Modify function handles the recursive traversal.
	if config.IsEnabled(ConstantFolding) {
		ast.Modify(program, visitorFunc(constantFolding))
	}
	if config.IsEnabled(DeadCodeElimination) {
		ast.Modify(program, visitorFunc(deadCodeElimination))
	}
	return program
}

// constantFolding is a visitor that finds and evaluates constant expressions.
func constantFolding(node ast.Node) ast.Node {
	inf, ok := node.(*ast.InfixExpression)
	if !ok {
		return node
	}

	left, leftOk := inf.Left.(*ast.IntegerLiteral)
	right, rightOk := inf.Right.(*ast.IntegerLiteral)

	if !leftOk || !rightOk {
		return node // Not an infix expression on two integers
	}

	leftVal := left.Value
	rightVal := right.Value

	var newValue int64
	switch inf.Operator {
	case "+":
		newValue = leftVal + rightVal
	case "-":
		newValue = leftVal - rightVal
	case "*":
		newValue = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			// Cannot fold division by zero, leave it for runtime error.
			return node
		}
		newValue = leftVal / rightVal
	default:
		// Not a foldable operator
		return node
	}

	// Create a new IntegerLiteral with BOTH the value and the text literal.
	return &ast.IntegerLiteral{
		Value: newValue,
		Token: lexer.Token{
			Type:    lexer.INT,
			Literal: fmt.Sprintf("%d", newValue), // This was the missing piece
		},
	}
}

// deadCodeElimination is a visitor that removes unreachable code by emptying dead branches.
func deadCodeElimination(node ast.Node) ast.Node {
	ifExp, ok := node.(*ast.IfExpression)
	if !ok {
		return node
	}

	cond, ok := ifExp.Condition.(*ast.Boolean)
	if !ok {
		return node // Condition is not a constant boolean.
	}

	if cond.Value {
		// Condition is `true`, so the alternative branch is dead.
		if ifExp.Alternative != nil {
			// Empty the statements in the else block.
			ifExp.Alternative.Statements = []ast.Statement{}
		}
	} else {
		// Condition is `false`, so the consequence branch is dead.
		// Empty the statements in the if block.
		ifExp.Consequence.Statements = []ast.Statement{}
	}

	// Return the modified IfExpression node. It is not replaced.
	return ifExp
}
