package preprocessor

import "github.com/the-xlang/xxc/ast"

// TrimEnofi trims tree by enofi pragma directive.
func TrimEnofi(tree *Tree) {
	for i, obj := range *tree {
		switch t := obj.Value.(type) {
		case ast.Preprocessor:
			switch t := t.Command.(type) {
			case ast.Directive:
				switch t.Command.(type) {
				case ast.EnofiDirective:
					*tree = (*tree)[:i]
					return
				}
			}
		}
	}
}
