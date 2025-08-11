package ast

// Visitor defines the interface for a visitor that can modify the AST.
// The Visit method is called for each node in the AST. If the visitor
// returns a non-nil node, the original node is replaced with the new one.
// If the visitor returns nil, the node may be removed from its parent list.
type Visitor interface {
	Visit(node Node) Node
}

// Modify traverses an AST node and applies a visitor to each node in the tree.
// It allows for the modification of the AST in place.
func Modify(node Node, visitor Visitor) Node {
	if node == nil {
		return nil
	}

	// Apply the visitor to the current node first (pre-order traversal)
	node = visitor.Visit(node)

	switch n := node.(type) {
	case *Program:
		for i, stmt := range n.Statements {
			n.Statements[i], _ = Modify(stmt, visitor).(Statement)
		}
		// Filter out nil statements from the program
		n.Statements = filterNilStatements(n.Statements)
	case *BlockStatement:
		for i, stmt := range n.Statements {
			n.Statements[i], _ = Modify(stmt, visitor).(Statement)
		}
		// Filter out nil statements from the block
		n.Statements = filterNilStatements(n.Statements)
	case *LetStatement:
		n.Value = Modify(n.Value, visitor).(Expression)
	case *PrintStatement:
		n.Expression = Modify(n.Expression, visitor).(Expression)
	case *ExpressionStatement:
		n.Expression = Modify(n.Expression, visitor).(Expression)
	case *PrefixExpression:
		n.Right = Modify(n.Right, visitor).(Expression)
	case *InfixExpression:
		n.Left = Modify(n.Left, visitor).(Expression)
		n.Right = Modify(n.Right, visitor).(Expression)
	case *IfExpression:
		n.Condition = Modify(n.Condition, visitor).(Expression)
		n.Consequence = Modify(n.Consequence, visitor).(*BlockStatement)
		if n.Alternative != nil {
			n.Alternative = Modify(n.Alternative, visitor).(*BlockStatement)
		}
	case *FunctionLiteral:
		for i, param := range n.Parameters {
			n.Parameters[i] = Modify(param, visitor).(*Identifier)
		}
		n.Body = Modify(n.Body, visitor).(*BlockStatement)
	case *CallExpression:
		n.Function = Modify(n.Function, visitor).(Expression)
		for i, arg := range n.Arguments {
			n.Arguments[i] = Modify(arg, visitor).(Expression)
		}
	// Literals and identifiers have no children to modify, so we do nothing.
	case *Identifier, *IntegerLiteral, *Boolean:
		// No children to traverse
	}

	// Apply the visitor again (post-order traversal) in case the children's
	// modification enables a modification on the parent.
	return visitor.Visit(node)
}

// filterNilStatements creates a new slice of statements without any nil values.
func filterNilStatements(stmts []Statement) []Statement {
	newStmts := make([]Statement, 0, len(stmts))
	for _, stmt := range stmts {
		if stmt != nil {
			newStmts = append(newStmts, stmt)
		}
	}
	return newStmts
}

// LINES: 86
