package preprocessor

import (
	"github.com/jule-lang/jule/ast/models"
)

// TrimEnofi trims tree by enofi pragma directive.
func TrimEnofi(tree *Tree) {
	for i, obj := range *tree {
		switch t := obj.Data.(type) {
		case models.Preprocessor:
			switch t := t.Command.(type) {
			case models.Directive:
				switch t.Command.(type) {
				case models.DirectiveEnofi:
					*tree = (*tree)[:i]
					return
				}
			}
		}
	}
}
