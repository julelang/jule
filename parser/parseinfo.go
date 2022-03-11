package parser

import (
	"sync"

	"github.com/the-xlang/x/pkg/xio"
	"github.com/the-xlang/x/pkg/xlog"
)

type ParseInfo struct {
	Parser   *Parser
	Logs     []xlog.CompilerLog
	File     *xio.File
	Routines *sync.WaitGroup
}

// ParseFileAsync parses file content.
func (info *ParseInfo) ParseAsync(main, justDefs bool) {
	if info.Routines != nil {
		defer info.Routines.Done()
	}
	info.Parser = NewParser(info)
	info.Parser.Parse(main, justDefs)
}
