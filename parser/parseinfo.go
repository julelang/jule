package parser

import (
	"sync"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/io"
	"github.com/the-xlang/x/pkg/xlog"
)

type ParseInfo struct {
	Parser   *Parser
	Logs     []xlog.CompilerLog
	File     *io.File
	Routines *sync.WaitGroup
}

// ParseFileAsync parses file content.
func (info *ParseInfo) ParseAsync(justDefs bool) {
	defer info.Routines.Done()
	lexer := lex.NewLex(info.File)
	tokens := lexer.Tokenize()
	if lexer.Logs != nil {
		info.Logs = lexer.Logs
		return
	}
	info.Parser = NewParser(tokens, info)
	info.Parser.Parse(justDefs)
}
