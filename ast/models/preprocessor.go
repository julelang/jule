package models

import "fmt"

// Preprocessor is the AST model of preprocessor directives.
type Preprocessor struct {
	Tok     Tok
	Command any
}

func (pp Preprocessor) String() string {
	return fmt.Sprint(pp.Command)
}

// Directive is the AST model of directives.
type Directive struct {
	Command any
}

func (d Directive) String() string {
	return fmt.Sprint(d.Command)
}

// DirectiveEnofi is the AST model of enofi directive.
type DirectiveEnofi struct{}
