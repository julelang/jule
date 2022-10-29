package models

import (
	"strings"

	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/juletype"
)

// Size is the represents data type of sizes (array or etc)
type Size = int

// TypeSize is the represents data type sizes with expression
type TypeSize struct {
	N         Size
	Expr      Expr
	AutoSized bool
}

// Type is data type identifier.
type Type struct {
	// Token used for usually *File comparisons.
	// For this reason, you don't use token as value, identifier or etc.
	Token         lex.Token
	Id            uint8
	Original      any
	Kind          string
	MultiTyped    bool
	ComponentType *Type
	Size          TypeSize
	Tag           any
	Pure          bool
	Generic       bool
	CppLinked     bool
}

// Copy returns deep copy of data type.
func (dt *Type) Copy() Type {
	copy := *dt
	if dt.ComponentType != nil {
		copy.ComponentType = new(Type)
		*copy.ComponentType = dt.ComponentType.Copy()
	}
	return copy
}

// KindWithOriginalId returns dt.Kind with OriginalId.
func (dt *Type) KindWithOriginalId() string {
	if dt.Original == nil {
		return dt.Kind
	}
	_, prefix := dt.KindId()
	original := dt.Original.(Type)
	id, _ := original.KindId()
	return prefix + id
}

// OriginalKindId returns dt.Kind's identifier of official.
//
// Special case is:
//
//	OriginalKindId() -> "" if DataType has not original
func (dt *Type) OriginalKindId() string {
	if dt.Original == nil {
		return ""
	}
	t := dt.Original.(Type)
	id, _ := t.KindId()
	return id
}

// KindId returns dt.Kind's identifier.
func (dt *Type) KindId() (id, prefix string) {
	if dt.Id == juletype.MAP || dt.Id == juletype.FN {
		return dt.Kind, ""
	}
	id = dt.Kind
	runes := []rune(dt.Kind)
	for i, r := range dt.Kind {
		if r == '_' || lex.IsLetter(r) {
			id = string(runes[i:])
			prefix = string(runes[:i])
			break
		}
	}
	for _, dt := range juletype.TYPE_MAP {
		if dt == id {
			return
		}
	}
	runes = []rune(id)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ':' && i+1 < len(runes) && runes[i+1] == ':' { // Namespace?
			i++
			continue
		}
		if r != '_' && !lex.IsLetter(r) && !lex.IsDecimal(byte(r)) {
			id = string(runes[:i])
			break
		}
	}
	return
}

func is_necessary_type(id uint8) bool {
	return id == juletype.TRAIT
}

func (dt *Type) set_to_original_cpp_linked() {
	if dt.Original == nil {
		return
	}
	if dt.Id == juletype.STRUCT {
		id := dt.Id
		tag := dt.Tag
		*dt = dt.Original.(Type)
		dt.Id = id
		dt.Tag = tag
		return
	}
	*dt = dt.Original.(Type)
}

func (dt *Type) SetToOriginal() {
	if dt.CppLinked {
		dt.set_to_original_cpp_linked()
		return
	} else if dt.Pure || dt.Original == nil {
		return
	}
	kind := dt.KindWithOriginalId()
	id := dt.Id
	tok := dt.Token
	generic := dt.Generic
	*dt = dt.Original.(Type)
	dt.Kind = kind
	// Keep original file, generic and necessary type code state
	dt.Token = tok
	dt.Generic = generic
	if is_necessary_type(id) {
		dt.Id = id
	}
	tag := dt.Tag
	switch tag.(type) {
	case Genericable:
		dt.Tag = tag
	}
}

// Modifiers returns pointer and reference marks of data type.
func (dt *Type) Modifiers() string {
	for i, r := range dt.Kind {
		if r != '*' && r != '&' {
			return dt.Kind[:i]
		}
	}
	return ""
}

// Modifiers returns pointer marks of data type.
func (dt *Type) Pointers() string {
	for i, r := range dt.Kind {
		if r != '*' {
			return dt.Kind[:i]
		}
	}
	return ""
}

// Modifiers returns reference marks of data type.
func (dt *Type) References() string {
	for i, r := range dt.Kind {
		if r != '&' {
			return dt.Kind[:i]
		}
	}
	return ""
}

func (dt Type) String() (s string) {
	dt.SetToOriginal()
	if dt.MultiTyped {
		return dt.MultiTypeString()
	}
	// Remove namespace
	i := strings.LastIndex(dt.Kind, lex.KND_DBLCOLON)
	if i != -1 {
		dt.Kind = dt.Kind[i+len(lex.KND_DBLCOLON):]
	}
	modifiers := dt.Modifiers()
	// Apply modifiers.
	defer func() {
		var cpp strings.Builder
		for _, r := range modifiers {
			if r == '&' {
				cpp.WriteString("jule_ref<")
			}
		}
		cpp.WriteString(s)
		for _, r := range modifiers {
			if r == '&' {
				cpp.WriteByte('>')
			}
		}
		for _, r := range modifiers {
			if r == '*' {
				cpp.WriteByte('*')
			}
		}
		s = cpp.String()
	}()
	dt.Kind = dt.Kind[len(modifiers):]
	switch dt.Id {
	case juletype.SLICE:
		return dt.SliceString()
	case juletype.ARRAY:
		return dt.ArrayString()
	case juletype.MAP:
		return dt.MapString()
	}
	switch dt.Tag.(type) {
	case CompiledStruct:
		return dt.StructString()
	}
	switch dt.Id {
	case juletype.ID:
		if dt.CppLinked {
			return dt.Kind
		}
		if dt.Generic {
			return juleapi.AsId(dt.Kind)
		}
		return juleapi.OutId(dt.Kind, dt.Token.File)
	case juletype.ENUM:
		e := dt.Tag.(*Enum)
		return e.Type.String()
	case juletype.TRAIT:
		return dt.TraitString()
	case juletype.STRUCT:
		return dt.StructString()
	case juletype.FN:
		return dt.FnString()
	default:
		return juletype.CppId(dt.Id)
	}
}

// SliceString returns cpp value of slice data type.
func (dt *Type) SliceString() string {
	var cpp strings.Builder
	cpp.WriteString("slice<")
	dt.ComponentType.Pure = dt.Pure
	cpp.WriteString(dt.ComponentType.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// ArrayString returns cpp value of map data type.
func (dt *Type) ArrayString() string {
	var cpp strings.Builder
	cpp.WriteString("array<")
	dt.ComponentType.Pure = dt.Pure
	cpp.WriteString(dt.ComponentType.String())
	cpp.WriteByte(',')
	cpp.WriteString(dt.Size.Expr.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// MapString returns cpp value of map data type.
func (dt *Type) MapString() string {
	var cpp strings.Builder
	types := dt.Tag.([]Type)
	cpp.WriteString("map<")
	key := types[0]
	key.Pure = dt.Pure
	cpp.WriteString(key.String())
	cpp.WriteByte(',')
	value := types[1]
	value.Pure = dt.Pure
	cpp.WriteString(value.String())
	cpp.WriteByte('>')
	return cpp.String()
}

// TraitString returns cpp value of trait data type.
func (dt *Type) TraitString() string {
	var cpp strings.Builder
	id, _ := dt.KindId()
	cpp.WriteString("trait<")
	cpp.WriteString(juleapi.OutId(id, dt.Token.File))
	cpp.WriteByte('>')
	return cpp.String()
}

// StructString returns cpp value of struct data type.
func (dt *Type) StructString() string {
	var cpp strings.Builder
	s := dt.Tag.(CompiledStruct)
	if s.CppLinked() && !Has_attribute(jule.ATTR_TYPEDEF, s.Get_ast().Attributes) {
		cpp.WriteString("struct ")
	}
	cpp.WriteString(s.OutId())
	types := s.Generics()
	if len(types) == 0 {
		return cpp.String()
	}
	cpp.WriteByte('<')
	for _, t := range types {
		t.Pure = dt.Pure
		cpp.WriteString(t.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

// FnString returns cpp value of function DataType.
func (dt *Type) FnString() string {
	var cpp strings.Builder
	cpp.WriteString("fn<std::function<")
	f := dt.Tag.(*Fn)
	f.RetType.Type.Pure = dt.Pure
	cpp.WriteString(f.RetType.String())
	cpp.WriteByte('(')
	if len(f.Params) > 0 {
		for _, param := range f.Params {
			param.Type.Pure = dt.Pure
			cpp.WriteString(param.Prototype())
			cpp.WriteByte(',')
		}
		cppStr := cpp.String()[:cpp.Len()-1]
		cpp.Reset()
		cpp.WriteString(cppStr)
	} else {
		cpp.WriteString("void")
	}
	cpp.WriteString(")>>")
	return cpp.String()
}

// MultiTypeString returns cpp value of muli-typed DataType.
func (dt *Type) MultiTypeString() string {
	var cpp strings.Builder
	cpp.WriteString("std::tuple<")
	types := dt.Tag.([]Type)
	for _, t := range types {
		if !t.Pure {
			t.Pure = dt.Pure
		}
		cpp.WriteString(t.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">" + dt.Modifiers()
}

// MapKind returns data type kind string of map data type.
func (dt *Type) MapKind() string {
	types := dt.Tag.([]Type)
	var kind strings.Builder
	kind.WriteByte('[')
	kind.WriteString(types[0].Kind)
	kind.WriteByte(':')
	kind.WriteString(types[1].Kind)
	kind.WriteByte(']')
	return kind.String()
}
