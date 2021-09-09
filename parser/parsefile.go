package parser

import (
	"fmt"
	"sync"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/io"
)

type ParseFileInfo struct {
	X_CXX    string
	Errors   []string
	File     *io.FILE
	Routines *sync.WaitGroup
}

// ParseFile parses file content.
func ParseFile(info *ParseFileInfo) {
	defer info.Routines.Done()
	info.X_CXX = ""
	lexer := lex.New(info.File)
	tokens := lexer.Tokenize()
	if lexer.Errors != nil {
		info.Errors = lexer.Errors
		return
	}
	for _, token := range tokens {
		fmt.Print("'"+token.Value+"'", " ")
	}
}
