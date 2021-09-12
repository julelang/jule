package ast

// AST Object types.
const (
	NA         uint8 = 0
	Identifier uint8 = 1
	Statement  uint8 = 2
	Range      uint8 = 3
	Block      uint8 = 4
	Type       uint8 = 5
	Tag        uint8 = 6
)

// AST Identifier types.
const (
	IdentifierName uint8 = 1
)

// AST Statement types.
const (
	StatementFunction     uint8 = 1
	StatementReturn       uint8 = 2
	StatementFunctionCall uint8 = 3
)

// AST Range types.
const (
	RangeBrace       uint8 = 1
	RangeParentheses uint8 = 2
)

// AST Expression node types.
const (
	ExpressionNodeValue    uint8 = 1
	ExpressionNodeOperator uint8 = 2
	ExpressionNodeBrace    uint8 = 3
)

// AST Value types.
const (
	ValueNumeric uint8 = 1
	ValueName    uint8 = 2
)
