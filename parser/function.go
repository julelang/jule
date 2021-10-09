package parser

import (
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

const entryPointStandard = `
#pragma region X_ENTRY_POINT_STANDARD_CODES
  setlocale(0x0, "");
#pragma endregion X_ENTRY_POINT_STANDARD_CODES

`

const entryPointStandardEnd = `


#pragma region X_ENTRY_POINT_END_STANDARD_CODES
  return EXIT_SUCCESS;
#pragma endregion X_ENTRY_POINT_END_STANDARD_CODES`

type function struct {
	Token      lex.Token
	Name       string
	ReturnType ast.TypeAST
	Params     []ast.ParameterAST
	Attributes []ast.AttributeAST
	Block      ast.BlockAST
}

func (f function) String() string {
	f.readyCxx()
	var cxx strings.Builder
	cxx.WriteString("template <typename any>\n")
	cxx.WriteString(attributesToString(f.Attributes))
	cxx.WriteString(f.typeString())
	cxx.WriteByte(' ')
	cxx.WriteString(f.nameString())
	cxx.WriteByte('(')
	cxx.WriteString(paramsToCxx(f.Params))
	cxx.WriteString(") {")
	cxx.WriteString(getFunctionStandardCode(f.Name))
	cxx.WriteString(f.Block.String())
	cxx.WriteString(getFunctionStandardEndCode(f.Name))
	cxx.WriteString("\n}")
	return cxx.String()
}

func (f *function) typeString() string {
	if f.Name == "_"+x.EntryPoint {
		return "int"
	}
	return f.ReturnType.String()
}

func (f *function) nameString() string {
	if f.Name == "_"+x.EntryPoint {
		return x.EntryPoint
	}
	return f.Name
}

func (f *function) readyCxx() {
	switch f.Name {
	case x.EntryPoint:
		f.ReturnType.Code = x.Int32
	}
}

func attributesToString(attributes []ast.AttributeAST) string {
	var cxx strings.Builder
	for _, attribute := range attributes {
		cxx.WriteString(attribute.String())
		cxx.WriteByte(' ')
	}
	return cxx.String()
}

func paramsToCxx(params []ast.ParameterAST) string {
	if len(params) == 0 {
		return ""
	}
	var cxx strings.Builder
	for _, p := range params {
		cxx.WriteString(p.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1]
}

func getFunctionStandardCode(name string) string {
	switch name {
	case "_" + x.EntryPoint:
		return entryPointStandard
	}
	return ""
}

func getFunctionStandardEndCode(name string) string {
	switch name {
	case "_" + x.EntryPoint:
		return entryPointStandardEnd
	}
	return ""
}
