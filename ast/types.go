package ast

// AST Object types.
const (
	NA         uint8 = 0
	Identifier uint8 = 1
	Statement  uint8 = 2
	Range      uint8 = 3
	Block      uint8 = 4
	Type       uint8 = 5
)

// AST Identifier types.
const (
	IdentifierNA   uint8 = 0
	IdentifierName uint8 = 1
)

// AST Statement types.
const (
	StatementNA       uint8 = 0
	StatementFunction uint8 = 1
	StatementReturn   uint8 = 2
)

// AST Range types.
const (
	RangeNA          uint8 = 0
	RangeBrace       uint8 = 1
	RangeParentheses uint8 = 2
)

// AST Expression types.
const (
	ExpressionNA      uint8 = 0
	ExpressionNumeric uint8 = 1
)

// AST Expression node types.
const (
	ExpressionNodeNA       uint8 = 0
	ExpressionNodeValue    uint8 = 1
	ExpressionNodeOperator uint8 = 2
)

// AST Value types.
const (
	ValueNA      uint8 = 0
	ValueNumeric uint8 = 1
)
