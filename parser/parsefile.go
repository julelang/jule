package parser

import (
	"sync"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/io"
	"github.com/the-xlang/x/pkg/xlog"
)

type ParseFileInfo struct {
	Parser   *Parser
	Errors   []xlog.CompilerLog
	File     *io.File
	Routines *sync.WaitGroup
}

// ParseFileAsync parses file content.
func (info *ParseFileInfo) ParseAsync(justDefs bool) {
	defer info.Routines.Done()
	lexer := lex.NewLex(info.File)
	tokens := lexer.Tokenize()
	if lexer.Errors != nil {
		info.Errors = lexer.Errors
		return
	}
	info.Parser = NewParser(tokens, info)
	info.Parser.Parse(justDefs)
}
