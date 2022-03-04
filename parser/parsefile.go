package parser

import (
	"sync"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/io"
)

type ParseFileInfo struct {
	X_CXX    string
	Errors   []string
	File     *io.File
	Routines *sync.WaitGroup
}

// ParseFileAsync parses file content.
func ParseFileAsync(info *ParseFileInfo) {
	defer info.Routines.Done()
	info.X_CXX = ""
	lexer := lex.NewLex(info.File)
	tokens := lexer.Tokenize()
	if lexer.Errors != nil {
		info.Errors = lexer.Errors
		return
	}
	parser := NewParser(tokens, info)
	parser.Parse()
	info.X_CXX += parser.Cxx()
}
