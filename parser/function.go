package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

const entryPointStandard = `
  // Entry point standard codes.
#if WIN32
  _setmode(0x1, 0x40000);
#else
  setmode(0x1, 0x40000);
#endif

`

// Function is function define representation.
type Function struct {
	Token      lex.Token
	Name       string
	ReturnType uint8
	Block      ast.BlockAST
}

func (f Function) String() string {
	var sb strings.Builder
	sb.WriteString(cxxTypeNameFromType(f.ReturnType))
	sb.WriteByte(' ')
	sb.WriteString(f.Name)
	sb.WriteString("()")
	sb.WriteString(" {")
	sb.WriteString(getFunctionStandardCode(f.Name))
	for _, s := range f.Block.Content {
		sb.WriteByte('\n')
		sb.WriteString("  " + fmt.Sprint(s.Value))
		sb.WriteByte(';')
	}
	sb.WriteString("\n}")
	return sb.String()
}

func getFunctionStandardCode(name string) string {
	switch name {
	case x.EntryPoint:
		return entryPointStandard
	}
	return ""
}
