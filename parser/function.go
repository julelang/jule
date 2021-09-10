package parser

import (
	"strings"

	"github.com/the-xlang/x/pkg/io"
)

// Function is function define representation.
type Function struct {
	FILE       *io.FILE
	Line       int
	Name       string
	ReturnType uint8
}

func (f Function) String() string {
	var sb strings.Builder
	sb.WriteString(cxxTypeNameFromType(f.ReturnType))
	sb.WriteByte(' ')
	sb.WriteString(f.Name)
	sb.WriteString("()")
	sb.WriteString(" {")
	sb.WriteByte('}')
	return sb.String()
}
