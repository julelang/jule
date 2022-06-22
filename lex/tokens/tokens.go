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
	NEW                 = "new"
	FREE                = "free"
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
	SIZEOF              = "sizeof"
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
	New          uint8 = 13
	Free         uint8 = 14
	Iter         uint8 = 15
	Break        uint8 = 16
	Continue     uint8 = 17
	In           uint8 = 18
	If           uint8 = 19
	Else         uint8 = 20
	Volatile     uint8 = 21
	Comment      uint8 = 22
	Use          uint8 = 23
	Dot          uint8 = 24
	Pub          uint8 = 25
	Preprocessor uint8 = 26
	Defer        uint8 = 27
	Goto         uint8 = 28
	DoubleColon  uint8 = 29
	Enum         uint8 = 30
	Struct       uint8 = 31
	Co           uint8 = 32
	Try          uint8 = 33
	Catch        uint8 = 34
	Sizeof       uint8 = 35
)
