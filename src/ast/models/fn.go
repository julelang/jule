package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/juletype"
)

// Fn is function declaration AST model.
type Fn struct {
	Pub           bool
	IsUnsafe      bool
	IsEntryPoint  bool
	Used          bool
	Token         lex.Token
	Id            string
	Generics      []*GenericType
	Combines      *[][]Type
	Attributes    []Attribute
	Params        []Param
	RetType       RetType
	Block         *Block
	Receiver      *Var
	Owner         any
	BuiltinCaller any
	Desc          string
}

func (f *Fn) plainTypeString() string {
	var s strings.Builder
	s.WriteByte('(')
	n := len(f.Params)
	if f.Receiver != nil {
		s.WriteString(f.Receiver.ReceiverTypeString())
		if n > 0 {
			s.WriteString(", ")
		}
	}
	if n > 0 {
		for _, p := range f.Params {
			if p.Variadic {
				s.WriteString("...")
			}
			s.WriteString(p.TypeString())
			s.WriteString(", ")
		}
		cppStr := s.String()[:s.Len()-2]
		s.Reset()
		s.WriteString(cppStr)
	}
	s.WriteByte(')')
	if f.RetType.Type.MultiTyped {
		s.WriteByte('(')
		for _, t := range f.RetType.Type.Tag.([]Type) {
			s.WriteString(t.Kind)
			s.WriteByte(',')
		}
		return s.String()[:s.Len()-1] + ")"
	} else if f.RetType.Type.Id != juletype.VOID {
		s.WriteString(f.RetType.Type.Kind)
	}
	return s.String()
}

// TypeKind returns data type string of function.
func (f *Fn) TypeKind() string {
	var cpp strings.Builder
	if f.IsUnsafe {
		cpp.WriteString("unsafe ")
	}
	cpp.WriteString("fn")
	cpp.WriteString(f.plainTypeString())
	return cpp.String()
}

// OutId returns juleapi.OutId result of function.
func (f *Fn) OutId() string {
	if f.IsEntryPoint {
		return juleapi.OutId(f.Id, nil)
	}
	if f.Receiver != nil {
		return f.Id
	}
	return juleapi.OutId(f.Id, f.Token.File)
}

// DefString returns define string of function.
func (f *Fn) DefString() string {
	var s strings.Builder
	if f.IsUnsafe {
		s.WriteString("unsafe ")
	}
	s.WriteString("fn ")
	s.WriteString(f.Id)
	s.WriteString(f.plainTypeString())
	return s.String()
}

// PrototypeParams returns prototype cpp code of function parameters.
func (f *Fn) PrototypeParams() string {
	if len(f.Params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range f.Params {
		cpp.WriteString(p.Prototype())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

func (f Fn) String() string {
	return f.StringOwner("")
}

func ParamsToCpp(params []Param) string {
	if len(params) == 0 {
		return "(void)"
	}
	var cpp strings.Builder
	cpp.WriteByte('(')
	for _, p := range params {
		cpp.WriteString(p.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ")"
}

// Head returns declaration head of function.
func (f *Fn) Head(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.DeclHead(owner))
	cpp.WriteString(ParamsToCpp(f.Params))
	return cpp.String()
}

// Prototype returns prototype cpp code of function.
func (f *Fn) Prototype(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.DeclHead(owner))
	cpp.WriteString(f.PrototypeParams())
	cpp.WriteByte(';')
	return cpp.String()
}

func (f *Fn) DeclHead(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(GenericsToCpp(f.Generics))
	if cpp.Len() > 0 {
		cpp.WriteByte('\n')
		cpp.WriteString(IndentString())
	}
	if !f.IsEntryPoint {
		cpp.WriteString("inline ")
	}
	cpp.WriteString(f.RetType.String())
	cpp.WriteByte(' ')
	if owner != "" {
		cpp.WriteString(owner)
		cpp.WriteString(lex.KND_DBLCOLON)
	}
	cpp.WriteString(f.OutId())
	return cpp.String()
}

func (f *Fn) StringOwner(owner string) string {
	var cpp strings.Builder
	cpp.WriteString(f.Head(owner))
	cpp.WriteByte(' ')
	vars := f.RetType.Vars(f.Block)
	cpp.WriteString(FnBlockToString(vars, f.Block))
	return cpp.String()
}

func FnBlockToString(vars []*Var, b *Block) string {
	var cpp strings.Builder
	if vars != nil {
		statements := make([]Statement, len(vars))
		for i, v := range vars {
			statements[i] = Statement{Token: v.Token, Data: *v}
		}
		b.Tree = append(statements, b.Tree...)
	}
	cpp.WriteString(b.String())
	return cpp.String()
}
