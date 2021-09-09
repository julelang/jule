package parser

import (
	"fmt"
	"strings"

	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
)

// CxxParser.
type CxxParser struct {
	Functions []*Function

	Position int
	Tokens   []lex.Token
	PFI      *ParseFileInfo
}

// NewParser returns new instance of CxxParser.
func NewParser(tokens []lex.Token, PFI *ParseFileInfo) *CxxParser {
	parser := new(CxxParser)
	parser.Tokens = tokens
	parser.PFI = PFI
	parser.Position = 0
	return parser
}

// PushError is appends new error by parser fields.
func (cp *CxxParser) PushError(err string) {
	message := x.Errors[err]
	cp.PFI.Errors = append(cp.PFI.Errors, fmt.Sprintf(
		"%s %s: %d", cp.PFI.File.Path, message, cp.Tokens[cp.Position].Line))
}

// String is return full C++ code of parsed objects.
func (cp CxxParser) String() string {
	var sb strings.Builder
	for _, function := range cp.Functions {
		sb.WriteString(function.String())
	}
	return sb.String()
}

// Ended reports position is at end of tokens or not.
func (cp *CxxParser) Ended() bool {
	return cp.Position >= len(cp.Tokens)
}

// Parse is parse X code to C++ code.
//
//! This function is main point of parsing.
func (cp *CxxParser) Parse() {
	for cp.Position != -1 && !cp.Ended() {
		firstToken := cp.Tokens[cp.Position]
		switch firstToken.Type {
		case lex.Name:
			cp.processName()
		default:
			cp.PushError("invalid_syntax")
		}
	}
}

// ParseFunction parse X function to C++ code.
func (cp *CxxParser) ParseFunction() {
	defineToken := cp.Tokens[cp.Position]
	function := new(Function)
	function.Name = defineToken.Value
	function.Line = defineToken.Line
	function.FILE = defineToken.File
	function.ReturnType = x.Void
	// Skip function parentheses.
	//! Fix here at after.
	cp.Position += 3
	if cp.Ended() {
		cp.Position--
		cp.PushError("function_body")
		cp.Position = -1 // Stop parsing.
		return
	}
	token := cp.Tokens[cp.Position]
	if token.Type == lex.Type {
		function.ReturnType = typeFromName(token.Value)
		cp.Position++
		if cp.Ended() {
			cp.Position--
			cp.PushError("function_body")
			cp.Position = -1 // Stop parsing.
			return
		}
		token = cp.Tokens[cp.Position]
	}
	switch token.Type {
	case lex.Brace:
		if token.Value != "{" {
			cp.PushError("invalid_syntax")
			cp.Position = -1 // Stop parsing.
			return
		}
		// Skip function braces.
		//! Fix here at after.
		cp.Position += 3
	default:
		cp.PushError("invalid_syntax")
		cp.Position = -1 // Stop parsing.
		return
	}
	cp.Functions = append(cp.Functions, function)
}

func (cp *CxxParser) processName() {
	cp.Position++
	if cp.Ended() {
		cp.Position--
		cp.PushError("invalid_syntax")
		return
	}
	cp.Position--
	secondToken := cp.Tokens[cp.Position+1]
	switch secondToken.Type {
	case lex.Brace:
		switch secondToken.Value {
		case "(": // Function.
			cp.ParseFunction()
		default:
			cp.PushError("invalid_syntax")
		}
	}
}
