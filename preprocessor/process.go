package preprocessor

import (
	"github.com/the-xlang/xxc/ast"
)

// Tree is the AST tree.
type Tree = []ast.Obj

// Process all preprocessor directives and commands.
func Process(tree *Tree) {
	TrimEnofi(tree)
}
