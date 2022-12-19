package models

import (
	"strings"

	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
)

// GenericType is the AST model of generic data-type.
type GenericType struct {
	Token lex.Token
	Id    string
}

func (gt GenericType) String() string {
	var cpp strings.Builder
	cpp.WriteString("typename ")
	cpp.WriteString(build.AsId(gt.Id))
	return cpp.String()
}

func GenericsToCpp(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("template<")
	for _, g := range generics {
		cpp.WriteString(g.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}
