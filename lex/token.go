package lex

import "github.com/jule-lang/jule/pkg/juleio"

// Token identities.
const ID_NA          = 0
const ID_DT          = 1
const ID_IDENT       = 2
const ID_BRACE       = 3
const ID_RET         = 4
const ID_SEMICOLON   = 5
const ID_LITERAL     = 6
const ID_OP          = 7
const ID_COMMA       = 8
const ID_CONST       = 9
const ID_TYPE        = 10
const ID_COLON       = 11
const ID_ITER        = 12
const ID_BREAK       = 13
const ID_CONTINUE    = 14
const ID_IN          = 15
const ID_IF          = 16
const ID_ELSE        = 17
const ID_COMMENT     = 18
const ID_USE         = 19
const ID_DOT         = 20
const ID_PUB         = 21
const ID_GOTO        = 22
const ID_DBLCOLON    = 23
const ID_ENUM        = 24
const ID_STRUCT      = 25
const ID_CO          = 26
const ID_MATCH       = 27
const ID_CASE        = 28
const ID_DEFAULT     = 29
const ID_SELF        = 30
const ID_TRAIT       = 31
const ID_IMPL        = 32
const ID_CPP         = 33
const ID_FALLTHROUGH = 34
const ID_FN          = 35
const ID_LET         = 36
const ID_UNSAFE      = 37
const ID_MUT         = 38

// Token kinds.
const KND_DBLCOLON     = "::"
const KND_COLON        = ":"
const KND_SEMICOLON    = ";"
const KND_COMMA        = ","
const KND_TRIPLE_DOT   = "..."
const KND_DOT          = "."
const KND_PLUS_EQ      = "+="
const KND_MINUS_EQ     = "-="
const KND_STAR_EQ      = "*="
const KND_SLASH_EQ     = "/="
const KND_PERCENT_EQ   = "%="
const KND_LSHIFT_EQ    = "<<="
const KND_RSHIFT_EQ    = ">>="
const KND_CARET_EQ     = "^="
const KND_AMPER_EQ     = "&="
const KND_VLINE_EQ     = "|="
const KND_EQS          = "=="
const KND_NOT_EQ       = "!="
const KND_GREAT_EQ     = ">="
const KND_LESS_EQ      = "<="
const KND_DBL_AMPER    = "&&"
const KND_DBL_VLINE    = "||"
const KND_LSHIFT       = "<<"
const KND_RSHIFT       = ">>"
const KND_DBL_PLUS     = "++"
const KND_DBL_MINUS    = "--"
const KND_PLUS         = "+"
const KND_MINUS        = "-"
const KND_STAR         = "*"
const KND_SOLIDUS      = "/"
const KND_PERCENT      = "%"
const KND_AMPER        = "&"
const KND_VLINE        = "|"
const KND_CARET        = "^"
const KND_EXCL         = "!"
const KND_LT           = "<"
const KND_GT           = ">"
const KND_EQ           = "="
const KND_LN_COMMENT   = "//"
const KND_RNG_LCOMMENT = "/*"
const KND_RNG_RCOMMENT = "*/"
const KND_LPAREN      = "("
const KND_RPARENT     = ")"
const KND_LBRACKET    = "["
const KND_RBRACKET    = "]"
const KND_LBRACE      = "{"
const KND_RBRACE      = "}"
const KND_I8          = "i8"
const KND_I16         = "i16"
const KND_I32         = "i32"
const KND_I64         = "i64"
const KND_U8          = "u8"
const KND_U16         = "u16"
const KND_U32         = "u32"
const KND_U64         = "u64"
const KND_F32         = "f32"
const KND_F64         = "f64"
const KND_UINT        = "uint"
const KND_INT         = "int"
const KND_UINTPTR     = "uintptr"
const KND_BOOL        = "bool"
const KND_STR         = "str"
const KND_ANY         = "any"
const KND_TRUE        = "true"
const KND_FALSE       = "false"
const KND_NIL         = "nil"
const KND_CONST       = "const"
const KND_RET         = "ret"
const KND_TYPE        = "type"
const KND_ITER        = "for"
const KND_BREAK       = "break"
const KND_CONTINUE    = "continue"
const KND_IN          = "in"
const KND_IF          = "if"
const KND_ELSE        = "else"
const KND_USE         = "use"
const KND_PUB         = "pub"
const KND_GOTO        = "goto"
const KND_ENUM        = "enum"
const KND_STRUCT      = "struct"
const KND_CO          = "co"
const KND_MATCH       = "match"
const KND_CASE        = "case"
const KND_DEFAULT     = "default"
const KND_SELF        = "self"
const KND_TRAIT       = "trait"
const KND_IMPL        = "impl"
const KND_CPP         = "cpp"
const KND_FALLTHROUGH = "fallthrough"
const KND_FN          = "fn"
const KND_LET         = "let"
const KND_UNSAFE      = "unsafe"
const KND_MUT         = "mut"

// Token is lexer token.
type Token struct {
	File   *juleio.File
	Row    int
	Column int
	Kind   string
	Id     uint8
}
