package models

import (
	"strings"
	"unicode"

	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

type genericableTypes struct {
	types []DataType
}

// Generics returns generic types.
func (gt genericableTypes) Generics() []DataType {
	return gt.types
}

// SetGenerics sets generics.
func (gt genericableTypes) SetGenerics([]DataType) {}

// DataType is data type identifier.
type DataType struct {
	// Tok used for usually *File comparisons.
	// For this reason, you don't use token as value, identifier or etc.
	Tok             Tok
	Id              uint8
	Original        any
	Kind            string
	MultiTyped      bool
	Tag             any
	DontUseOriginal bool
}

// ValWithOriginalId returns dt.Val with OriginalId.
func (dt *DataType) ValWithOriginalId() string {
	if dt.Original == nil {
		return dt.Kind
	}
	_, prefix := dt.GetValId()
	original := dt.Original.(DataType)
	return prefix + original.Tok.Kind
}

// OriginalValId returns dt.Val's identifier of official.
//
// Special case is:
//   OriginalValId() -> "" if DataType has not original
func (dt *DataType) OriginalValId() string {
	if dt.Original == nil {
		return ""
	}
	t := dt.Original.(DataType)
	id, _ := t.GetValId()
	return id
}

// GetValId returns dt.Val's identifier.
func (dt *DataType) GetValId() (id, prefix string) {
	id = dt.Kind
	runes := []rune(dt.Kind)
	for i, r := range dt.Kind {
		if r == '_' || unicode.IsLetter(r) {
			id = string(runes[i:])
			prefix = string(runes[:i])
			break
		}
	}
	runes = []rune(id)
	for i, r := range runes {
		if r != '_' && !unicode.IsLetter(r) {
			id = string(runes[:i])
			break
		}
	}
	return
}

func (dt *DataType) setToOriginal() {
	if dt.DontUseOriginal || dt.Original == nil {
		return
	}
	val := dt.ValWithOriginalId()
	tok := dt.Tok
	*dt = dt.Original.(DataType)
	dt.Kind = val
	dt.Tok = tok
}

// Pointers returns pointer marks of data type.
func (dt *DataType) Pointers() string {
	for i, run := range dt.Kind {
		if run != '*' {
			return dt.Kind[:i]
		}
	}
	return ""
}

func (dt DataType) String() string {
	dt.setToOriginal()
	if dt.MultiTyped {
		return dt.MultiTypeString()
	}
	pointers := dt.Pointers()
	dt.Kind = dt.Kind[len(pointers):]
	if dt.Kind != "" {
		switch {
		case strings.HasPrefix(dt.Kind, "[]"):
			return dt.ArrayString() + pointers
		case dt.Id == xtype.Map && dt.Kind[0] == '[':
			return dt.MapString() + pointers
		}
	}
	if dt.Tag != nil {
		switch t := dt.Tag.(type) {
		case []DataType:
			dt.Tag = genericableTypes{t}
			return dt.StructString()
		case Genericable:
			return dt.StructString()
		}
	}
	switch dt.Id {
	case xtype.Id, xtype.Enum:
		return xapi.OutId(dt.Kind, dt.Tok.File) + pointers
	case xtype.Struct:
		return dt.StructString() + pointers
	case xtype.Func:
		return dt.FuncString() + pointers
	default:
		return xtype.CxxTypeIdFromType(dt.Id) + pointers
	}
}

// ArrayString returns cxx value of array data type.
func (dt DataType) ArrayString() string {
	var cxx strings.Builder
	cxx.WriteString("array<")
	dt.Kind = dt.Kind[2:]
	cxx.WriteString(dt.String())
	cxx.WriteByte('>')
	return cxx.String()
}

// MapString returns cxx value of map data type.
func (dt *DataType) MapString() string {
	var cxx strings.Builder
	types := dt.Tag.([]DataType)
	cxx.WriteString("map<")
	key := types[0]
	key.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(key.String())
	cxx.WriteByte(',')
	value := types[1]
	value.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(value.String())
	cxx.WriteByte('>')
	return cxx.String()
}

// StructString returns cxx value of struct data type.
func (dt *DataType) StructString() string {
	var cxx strings.Builder
	cxx.WriteString(xapi.OutId(dt.Kind, dt.Tok.File))
	s := dt.Tag.(Genericable)
	types := s.Generics()
	if len(types) == 0 {
		return cxx.String()
	}
	cxx.WriteByte('<')
	for _, t := range types {
		t.DontUseOriginal = dt.DontUseOriginal
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// FuncString returns cxx value of function DataType.
func (dt *DataType) FuncString() string {
	var cxx strings.Builder
	cxx.WriteString("func<")
	f := dt.Tag.(*Func)
	f.RetType.Type.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(f.RetType.String())
	cxx.WriteByte('(')
	if len(f.Params) > 0 {
		for _, param := range f.Params {
			param.Type.DontUseOriginal = dt.DontUseOriginal
			cxx.WriteString(param.Prototype())
			cxx.WriteByte(',')
		}
		cxxStr := cxx.String()[:cxx.Len()-1]
		cxx.Reset()
		cxx.WriteString(cxxStr)
	} else {
		cxx.WriteString("void")
	}
	cxx.WriteString(")>")
	return cxx.String()
}

// MultiTypeString returns cxx value of muli-typed DataType.
func (dt *DataType) MultiTypeString() string {
	types := dt.Tag.([]DataType)
	var cxx strings.Builder
	cxx.WriteString("std::tuple<")
	for _, t := range types {
		t.DontUseOriginal = dt.DontUseOriginal
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">" + dt.Pointers()
}
