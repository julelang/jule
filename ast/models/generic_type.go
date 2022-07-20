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
	var cxx strings.Builder
	cxx.WriteString("typename ")
	cxx.WriteString(xapi.OutId(gt.Id, gt.Tok.File))
	return cxx.String()
}
