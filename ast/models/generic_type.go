package models

import (
	"strings"

	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/pkg/juleapi"
)

// GenericType is the AST model of generic data-type.
type GenericType struct {
	Token lex.Token
	Id    string
}

func (gt GenericType) String() string {
	var cpp strings.Builder
	cpp.WriteString("typename ")
	cpp.WriteString(juleapi.AsId(gt.Id))
	return cpp.String()
}
