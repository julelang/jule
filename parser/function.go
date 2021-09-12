package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

const entryPointStandard = `
#pragma region X_ENTRY_POINT_STANDARD_CODES
  setlocale(0x0, "");
#pragma endregion

`

type function struct {
	Token      lex.Token
	Name       string
	ReturnType uint8
	Params     []ast.ParameterAST
	Tags       []ast.TagAST
	Block      ast.BlockAST
}

func (f function) String() string {
	var cxx string
	cxx += tagsToString(f.Tags)
	cxx += x.CxxTypeNameFromType(f.ReturnType)
	cxx += " "
	cxx += f.Name
	cxx += "("
	cxx += paramsToCxx(f.Params)
	cxx += ") {"
	cxx += getFunctionStandardCode(f.Name)
	cxx += blockToCxx(f.Block)
	cxx += "\n}"
	return cxx
}

func tagsToString(tags []ast.TagAST) string {
	var cxx strings.Builder
	for _, tag := range tags {
		cxx.WriteString(tag.String())
		cxx.WriteByte(' ')
	}
	return cxx.String()
}

func paramsToCxx(params []ast.ParameterAST) string {
	if len(params) == 0 {
		return ""
	}
	var cxx string
	any := false
	for _, p := range params {
		cxx += p.String()
		cxx += ","
		if !any {
			any = p.Type.Type == x.Any
		}
	}
	cxx = cxx[:len(cxx)-1]
	if any {
		cxx = "template <typename any>\n" + cxx
	}
	return cxx
}

func blockToCxx(block ast.BlockAST) string {
	var cxx strings.Builder
	for _, s := range block.Content {
		cxx.WriteByte('\n')
		cxx.WriteString("  ")
		cxx.WriteString(fmt.Sprint(s.Value))
		cxx.WriteByte(';')
	}
	return cxx.String()
}

func getFunctionStandardCode(name string) string {
	switch name {
	case x.EntryPoint:
		return entryPointStandard
	}
	return ""
}
