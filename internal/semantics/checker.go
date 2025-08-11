package semantics

import (
	"fmt"

	"golite.dev/mvp/internal/ast"
	"golite.dev/mvp/internal/object"
)

type Checker struct {
	errors []string
	table  *SymbolTable
}

func New() *Checker {
	return &Checker{
		errors: []string{},
		table:  NewSymbolTable(),
	}
}

func (c *Checker) Errors() []string {
	return c.errors
}

func (c *Checker) Check(node ast.Node) object.ObjectType {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return c.checkProgram(node)
	case *ast.LetStatement:
		return c.checkLetStatement(node)
	case *ast.ExpressionStatement:
		return c.Check(node.Expression)
	case *ast.BlockStatement:
		return c.checkBlockStatement(node)
	case *ast.PrintStatement:
		c.Check(node.Expression)
		return object.NULL_OBJ // Statements don't have a type

	// Expressions
	case *ast.Identifier:
		return c.checkIdentifier(node)
	case *ast.IntegerLiteral:
		return object.INTEGER_OBJ
	case *ast.Boolean:
		return object.BOOLEAN_OBJ
	case *ast.InfixExpression:
		return c.checkInfixExpression(node)
	case *ast.PrefixExpression:
		return c.checkPrefixExpression(node)
	case *ast.IfExpression:
		return c.checkIfExpression(node)
	case *ast.FunctionLiteral:
		return c.checkFunctionLiteral(node)
	case *ast.CallExpression:
		return c.checkCallExpression(node)
	}
	return object.NULL_OBJ
}

func (c *Checker) addError(format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}

func (c *Checker) checkProgram(program *ast.Program) object.ObjectType {
	for _, stmt := range program.Statements {
		c.Check(stmt)
	}
	return object.NULL_OBJ
}

func (c *Checker) checkBlockStatement(block *ast.BlockStatement) object.ObjectType {
	// Scoping for block statements can be added here if needed
	for _, stmt := range block.Statements {
		c.Check(stmt)
	}
	return object.NULL_OBJ
}

func (c *Checker) checkLetStatement(stmt *ast.LetStatement) object.ObjectType {
	valType := c.Check(stmt.Value)
	if valType == object.ERROR_OBJ {
		return object.ERROR_OBJ
	}
	c.table.Define(stmt.Name.Value, valType)
	return object.NULL_OBJ
}

func (c *Checker) checkIdentifier(ident *ast.Identifier) object.ObjectType {
	symbol, ok := c.table.Resolve(ident.Value)
	if !ok {
		c.addError("identifier not found: %s", ident.Value)
		return object.ERROR_OBJ
	}
	return symbol.Type
}

func (c *Checker) checkPrefixExpression(node *ast.PrefixExpression) object.ObjectType {
	rightType := c.Check(node.Right)
	if rightType == object.ERROR_OBJ {
		return object.ERROR_OBJ
	}

	switch node.Operator {
	case "!":
		if rightType != object.BOOLEAN_OBJ {
			c.addError("unknown operator: %s%s", node.Operator, rightType)
			return object.ERROR_OBJ
		}
		return object.BOOLEAN_OBJ
	case "-":
		if rightType != object.INTEGER_OBJ {
			c.addError("unknown operator: %s%s", node.Operator, rightType)
			return object.ERROR_OBJ
		}
		return object.INTEGER_OBJ
	default:
		c.addError("unknown operator: %s%s", node.Operator, rightType)
		return object.ERROR_OBJ
	}
}

func (c *Checker) checkInfixExpression(node *ast.InfixExpression) object.ObjectType {
	leftType := c.Check(node.Left)
	rightType := c.Check(node.Right)

	if leftType == object.ERROR_OBJ || rightType == object.ERROR_OBJ {
		return object.ERROR_OBJ
	}

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		switch node.Operator {
		case "+", "-", "*", "/":
			return object.INTEGER_OBJ
		case "<", ">", "==", "!=":
			return object.BOOLEAN_OBJ
		default:
			c.addError("unknown operator: %s %s %s", leftType, node.Operator, rightType)
			return object.ERROR_OBJ
		}
	case leftType == object.BOOLEAN_OBJ && rightType == object.BOOLEAN_OBJ:
		switch node.Operator {
		case "==", "!=":
			return object.BOOLEAN_OBJ
		default:
			c.addError("unknown operator: %s %s %s", leftType, node.Operator, rightType)
			return object.ERROR_OBJ
		}
	case leftType != rightType:
		c.addError("type mismatch: %s %s %s", leftType, node.Operator, rightType)
		return object.ERROR_OBJ
	default:
		c.addError("unknown operator: %s %s %s", leftType, node.Operator, rightType)
		return object.ERROR_OBJ
	}
}

func (c *Checker) checkIfExpression(ie *ast.IfExpression) object.ObjectType {
	condType := c.Check(ie.Condition)
	if condType != object.BOOLEAN_OBJ {
		c.addError("if condition must be a boolean, got %s", condType)
	}

	c.Check(ie.Consequence)
	if ie.Alternative != nil {
		c.Check(ie.Alternative)
	}
	// In a real language, we'd check if consequence/alternative have compatible return types.
	// For now, if-expressions don't return values.
	return object.NULL_OBJ
}

func (c *Checker) checkFunctionLiteral(fl *ast.FunctionLiteral) object.ObjectType {
	// Create a new scope for the function body
	enclosedTable := NewEnclosedSymbolTable(c.table)
	originalTable := c.table
	c.table = enclosedTable
	defer func() { c.table = originalTable }() // Restore the original table after checking

	for _, p := range fl.Parameters {
		// In a typed language, parameters would have types. Here we assume they can be anything
		// until they are used. A more robust checker would handle this differently.
		c.table.Define(p.Value, "ANY") // A placeholder type
	}

	c.Check(fl.Body)
	return object.FUNCTION_OBJ
}

func (c *Checker) checkCallExpression(ce *ast.CallExpression) object.ObjectType {
	// Check that the function being called is actually a function
	fnType := c.Check(ce.Function)
	if fnType != object.FUNCTION_OBJ {
		c.addError("not a function: %s", ce.Function.String())
		return object.ERROR_OBJ
	}
	// This is a simplified check. A full check would verify argument types.
	fn, ok := ce.Function.(*ast.Identifier)
	if ok {
		// A more complex check would look up the function literal itself.
		// For now, we'll just check based on the identifier.
		sym, ok := c.table.Resolve(fn.Value)
		if !ok {
			// This should have been caught by checkIdentifier already.
			return object.ERROR_OBJ
		}

		// A more advanced checker would resolve the identifier to its FunctionLiteral
		// and check the arity there. This is a simplification for now.
		if sym.Type != object.FUNCTION_OBJ {
			c.addError("not a function: %s", fn.Value)
			return object.ERROR_OBJ
		}
	}

	for _, arg := range ce.Arguments {
		c.Check(arg)
	}
	// A full implementation would return the function's return type.
	return object.NULL_OBJ
}
