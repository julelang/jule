package parser

import "github.com/the-xlang/xxc/ast/models"

type value struct {
	ast      models.Value
	constant bool
	volatile bool
	lvalue   bool
	variadic bool
	isType   bool
}
