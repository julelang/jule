package models

import (
	"strings"
	"sync/atomic"
)

// Block is code block.
type Block struct {
	IsUnsafe bool
	Deferred bool
	Parent   *Block
	SubIndex int // Index of statement in parent block
	Tree     []Statement
	Func     *Fn

	// If block is the root block, has all labels and gotos of all sub blocks.
	Gotos  *Gotos
	Labels *Labels
}

func (b Block) String() string {
	AddIndent()
	s := ""
	if b.Deferred {
		s = "__JULEC_DEFER("
	}
	s += ParseBlock(b)
	DoneIndent()
	if b.Deferred {
		s += ");"
	}
	return s
}

// ParseBlock to cpp.
func ParseBlock(b Block) string {
	// Space count per indent.
	var cpp strings.Builder
	cpp.WriteByte('{')
	for _, s := range b.Tree {
		if s.Data == nil {
			continue
		}
		cpp.WriteByte('\n')
		cpp.WriteString(IndentString())
		cpp.WriteString(s.String())
	}
	cpp.WriteByte('\n')
	indent := strings.Repeat(Indentation, int(Indent-1))
	cpp.WriteString(indent)
	cpp.WriteByte('}')
	return cpp.String()
}

// Indent is indention count.
// This should be manuplate atomic.
var Indent uint32 = 0

// Indentation.
var Indentation = "\t"

// IndentString returns indent space of current block.
func IndentString() string {
	return strings.Repeat(Indentation, int(Indent))
}

// AddIndent adds new indent to IndentString.
func AddIndent() { atomic.AddUint32(&Indent, 1) }

// DoneIndent removes last indent from IndentString.
func DoneIndent() {
	atomic.SwapUint32(&Indent, atomic.LoadUint32(&Indent)-1)
}
