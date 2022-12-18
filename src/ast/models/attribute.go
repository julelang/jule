package models

import "github.com/julelang/jule/lex"

// Attribute is attribtue AST model.
type Attribute struct {
	Token lex.Token
	Tag   string
}

// Has_attribute returns true attribute if exist, false if not.
func Has_attribute(kind string, attributes []Attribute) bool {
	for i := range attributes {
		attribute := attributes[i]
		if attribute.Tag == kind {
			return true
		}
	}
	return false
}
