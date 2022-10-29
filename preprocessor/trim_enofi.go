package preprocessor

import (
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/pkg/jule"
)

// TrimEnofi trims tree by enofi pragma directive.
func TrimEnofi(tree *Tree) {
	for i, obj := range *tree {
		switch t := obj.Data.(type) {
		case models.Comment:
			if !IsPreprocessorPragma(t.Content) {
				continue
			}
			directive := getDirective(t.Content)
			switch directive {
			case jule.PREPROCESSOR_DIRECTIVE_ENOFI:
				*tree = (*tree)[:i]
				return
			}
		}
	}
}
