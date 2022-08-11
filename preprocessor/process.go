package preprocessor

import "github.com/jule-lang/jule/ast/models"

// Tree is the AST tree.
type Tree = []models.Object

// Process all preprocessor directives and commands.
func Process(tree *Tree, includeEnofi bool) {
	if includeEnofi {
		TrimEnofi(tree)
	}
}
