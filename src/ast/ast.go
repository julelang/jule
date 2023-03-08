package ast

import "github.com/julelang/jule/lex"

// Type of AST Node's data.
type NodeData = any

// AST Node.
type Node struct {
	Token lex.Token
	Data  any
}

// Group for AST model of comments.
type CommentGroup struct {
	Comments []*Comment
}

// AST model of just comment lines.
type Comment struct {
	Token   lex.Token
	Text string
}

// Directive AST.
type Directive struct {
	Token lex.Token
	Tag   string
}
