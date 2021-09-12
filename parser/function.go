package parser

import (
	"fmt"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

const entryPointStandard = `
#pragma region X_ENTRY_POINT_STANDARD_CODES
  setlocale(0x0, "");
#pragma endregion

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
		any := false
		for _, p := range f.Params {
			code += p.String()
			code += ","
			if !any {
				any = p.Type.Type == x.Any
			}
		}
		code = code[:len(code)-1]
		if any {
			code = "template <typename any>\n" + code
		}
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
