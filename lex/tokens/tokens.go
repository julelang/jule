package tokens

// Token kinds.
const (
	SHARP               = "#"
	DOUBLE_COLON        = "::"
	COLON               = ":"
	SEMICOLON           = ";"
	COMMA               = ","
	AT                  = "@"
	TRIPLE_DOT          = "..."
	DOT                 = "."
	PLUS_EQUAL          = "+="
	MINUS_EQUAL         = "-="
	STAR_EQUAL          = "*="
	SLASH_EQUAL         = "/="
	PERCENT_EQUAL       = "%="
	LSHIFT_EQUAL        = "<<="
	RSHIFT_EQUAL        = ">>="
	CARET_EQUAL         = "^="
	AMPER_EQUAL         = "&="
	VLINE_EQUAL         = "|="
	EQUALS              = "=="
	NOT_EQUALS          = "!="
	GREAT_EQUAL         = ">="
	LESS_EQUAL          = "<="
	AND                 = "&&"
	OR                  = "||"
	LSHIFT              = "<<"
	RSHIFT              = ">>"
	PLUS                = "+"
	MINUS               = "-"
	STAR                = "*"
	SLASH               = "/"
	PERCENT             = "%"
	TILDE               = "~"
	AMPER               = "&"
	VLINE               = "|"
	CARET               = "^"
	EXCLAMATION         = "!"
	LESS                = "<"
	GREAT               = ">"
	EQUAL               = "="
	LINE_COMMENT        = "//"
	RANGE_COMMENT_OPEN  = "/*"
	RANGE_COMMENT_CLOSE = "*/"
	LPARENTHESES        = "("
	RPARENTHESES        = ")"
	LBRACKET            = "["
	RBRACKET            = "]"
	LBRACE              = "{"
	RBRACE              = "}"
	I8                  = "i8"
	I16                 = "i16"
	I32                 = "i32"
	I64                 = "i64"
	U8                  = "u8"
	U16                 = "u16"
	U32                 = "u32"
	U64                 = "u64"
	F32                 = "f32"
	F64                 = "f64"
	UINT                = "uint"
	INT                 = "int"
	UINTPTR             = "uintptr"
	INTPTR              = "intptr"
	BOOL                = "bool"
	CHAR                = "char"
	STR                 = "str"
	VOIDPTR             = "voidptr"
	ANY                 = "any"
	TRUE                = "true"
	FALSE               = "false"
	NIL                 = "nil"
	CONST               = "const"
	RET                 = "ret"
	TYPE                = "type"
	ITER                = "iter"
	BREAK               = "break"
	CONTINUE            = "continue"
	IN                  = "in"
	IF                  = "if"
	ELSE                = "else"
	VOLATILE            = "volatile"
	USE                 = "use"
	PUB                 = "pub"
	DEFER               = "defer"
	GOTO                = "goto"
	ENUM                = "enum"
	STRUCT              = "struct"
	CO                  = "co"
	TRY                 = "try"
	CATCH               = "catch"
)

// Token types.
const (
	NA           uint8 = 0
	DataType     uint8 = 1
	Id           uint8 = 2
	Brace        uint8 = 3
	Ret          uint8 = 4
	SemiColon    uint8 = 5
	Value        uint8 = 6
	Operator     uint8 = 7
	Comma        uint8 = 8
	Const        uint8 = 9
	Type         uint8 = 10
	Colon        uint8 = 11
	At           uint8 = 12
	Iter         uint8 = 13
	Break        uint8 = 14
	Continue     uint8 = 15
	In           uint8 = 16
	If           uint8 = 17
	Else         uint8 = 18
	Volatile     uint8 = 19
	Comment      uint8 = 20
	Use          uint8 = 21
	Dot          uint8 = 22
	Pub          uint8 = 23
	Preprocessor uint8 = 24
	Defer        uint8 = 25
	Goto         uint8 = 26
	DoubleColon  uint8 = 27
	Enum         uint8 = 28
	Struct       uint8 = 29
	Co           uint8 = 30
	Try          uint8 = 31
	Catch        uint8 = 32
)
