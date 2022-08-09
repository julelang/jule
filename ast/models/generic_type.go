package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// GenericType is the AST model of generic data-type.
type GenericType struct {
	Tok Tok
	Id  string
}

func (gt GenericType) String() string {
	var cpp strings.Builder
	cpp.WriteString("typename ")
	cpp.WriteString(xapi.AsId(gt.Id))
	return cpp.String()
}
