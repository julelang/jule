package parser

import (
	"fmt"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

const entryPointStandard = `
  // Entry point standard codes.
  setlocale(0x0, "");

`

// Function is function define representation.
type Function struct {
	Token      lex.Token
	Name       string
	ReturnType uint8
	Params     []ast.ParameterAST
	Block      ast.BlockAST
}

func (f Function) String() string {
	code := ""
	code += x.CxxTypeNameFromType(f.ReturnType)
	code += " "
	code += f.Name
	code += "("
	if len(f.Params) > 0 {
		for _, p := range f.Params {
			code += p.String()
			code += ","
		}
		code = code[:len(code)-1]
	}
	code += ") {"
	code += getFunctionStandardCode(f.Name)
	for _, s := range f.Block.Content {
		code += "\n"
		code += "  " + fmt.Sprint(s.Value)
		code += ";"
	}
	code += "\n}"
	return code
}

func getFunctionStandardCode(name string) string {
	switch name {
	case x.EntryPoint:
		return entryPointStandard
	}
	return ""
}
