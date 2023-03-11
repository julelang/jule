package sema

import (
	"strings"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/build"
)

// Semantic analysis information.
type SemaInfo struct {
	Errors []build.Log
}

// Analyze AST.
// Returns nil if ast is nil.
// Returns nil if pwd is empty.
// Returns nil if pstd is empty.
// Accepts current working directory is pwd.
//
// Parameters:
//   pwd:  working directory path
//   pstd: standard library directory path
//   ast:  abstract syntax tree
func Analyze(pwd string, pstd string, ast *ast.Ast) *SemaInfo {
	if ast == nil {
		return nil
	}

	pwd = strings.TrimSpace(pwd)
	if pwd == "" {
		return nil
	}

	pstd = strings.TrimSpace(pstd)
	if pstd == "" {
		return nil
	}

	sema := &_Sema{
		ast:  ast,
		pwd:  pwd,
		pstd: pstd,
	}
	sema.analyze()

	sinf := &SemaInfo{
		Errors: sema.errors,
	}
	return sinf
}
