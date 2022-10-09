package preprocessor

import (
	"strings"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/pkg/jule"
)

// Tree is the AST tree.
type Tree = []models.Object

// IsPreprocessorPragma reports pragma is preprocessor pragma or not.
func IsPreprocessorPragma(s string) bool {
	if !strings.HasPrefix(s, jule.PRAGMA_COMMENT_PREFIX) {
		return false
	}
	switch getDirective(s) {
	case jule.PREPROCESSOR_DIRECTIVE_ENOFI:
		return true
	default:
		return false
	}
}

func getDirective(s string) string {
	return s[len(jule.PRAGMA_COMMENT_PREFIX):]
}

// Process all preprocessor directives and commands.
func Process(tree *Tree, includeEnofi bool) {
	if includeEnofi {
		TrimEnofi(tree)
	}
}
