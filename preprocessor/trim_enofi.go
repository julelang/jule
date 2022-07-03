package preprocessor

import (
	"github.com/the-xlang/xxc/ast/models"
)

// TrimEnofi trims tree by enofi pragma directive.
func TrimEnofi(tree *Tree) {
	for i, obj := range *tree {
		switch t := obj.Value.(type) {
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
