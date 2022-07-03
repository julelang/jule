package preprocessor

import "github.com/the-xlang/xxc/ast/models"

// Tree is the AST tree.
type Tree = []models.Object

// Process all preprocessor directives and commands.
func Process(tree *Tree) {
	TrimEnofi(tree)
}
