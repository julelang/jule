package sema

import (
	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
)

// Semantic analysis information.
type SemaInfo struct {
	Errors []build.Log
}

// Analyze AST.
func Analysis(ast *ast.Ast) *SemaInfo {
	return nil
}
