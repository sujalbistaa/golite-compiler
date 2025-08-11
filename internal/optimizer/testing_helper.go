package optimizer

import "golite.dev/mvp/internal/ast"

// DeadCodeEliminationForTest exposes the internal deadCodeElimination function for testing.
// In a production compiler, you might structure the packages differently to avoid this.
func DeadCodeEliminationForTest(node ast.Node) ast.Node {
	return deadCodeElimination(node)
}

// LINES: 8
