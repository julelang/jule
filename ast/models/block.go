package models

import (
	"strings"
	"sync/atomic"

	"github.com/the-xlang/xxc/pkg/x"
)

// Block is code block.
type Block struct {
	Parent   *Block
	SubIndex int // Anonymous block sub count
	Tree     []Statement
	Gotos    *Gotos
	Labels   *Labels
	Func     *Func
}

func (b Block) String() string {
	AddIndent()
	defer func() { DoneIndent() }()
	return ParseBlock(b)
}

// ParseBlock to cxx.
func ParseBlock(b Block) string {
	// Space count per indent.
	var cxx strings.Builder
	cxx.WriteByte('{')
	for _, s := range b.Tree {
		if s.Val == nil {
			continue
		}
		cxx.WriteByte('\n')
		cxx.WriteString(IndentString())
		cxx.WriteString(s.String())
	}
	cxx.WriteByte('\n')
	cxx.WriteString(strings.Repeat(x.Set.Indent, int(Indent-1)*x.Set.IndentCount))
	cxx.WriteByte('}')
	return cxx.String()
}

// Indent is indention count.
// This should be manuplate atomic.
var Indent uint32 = 0

// IndentString returns indent space of current block.
func IndentString() string {
	return strings.Repeat(x.Set.Indent, int(Indent)*x.Set.IndentCount)
}

// AddIndent adds new indent to IndentString.
func AddIndent() { atomic.AddUint32(&Indent, 1) }

// DoneIndent removes last indent from IndentString.
func DoneIndent() { atomic.SwapUint32(&Indent, Indent-1) }
