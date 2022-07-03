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
func (gt genericableTypes) Generics() []DataType { return gt.types }

// SetGenerics sets generics.
func (gt genericableTypes) SetGenerics([]DataType) {}

// DataType is data type identifier.
type DataType struct {
	// Tok used for usually *File comparisons.
	// For this reason, you don't use token as val, identifier or etc.
	Tok        Tok
	Id         uint8
	Original   any
	Val        string
	MultiTyped bool
	Tag        any
}

// ValWithOriginalId returns dt.Val with OriginalId.
func (dt *DataType) ValWithOriginalId() string {
	if dt.Original == nil {
		return dt.Val
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
	id = dt.Val
	runes := []rune(dt.Val)
	for i, r := range dt.Val {
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

func (dt DataType) String() string {
	var cxx strings.Builder
	if dt.Original != nil {
		val := dt.ValWithOriginalId()
		tok := dt.Tok
		dt = dt.Original.(DataType)
		dt.Val = val
		dt.Tok = tok
	}
	for i, run := range dt.Val {
		if run == '*' {
			cxx.WriteRune(run)
			continue
		}
		dt.Val = dt.Val[i:]
		break
	}
	if dt.MultiTyped {
		return dt.MultiTypeString() + cxx.String()
	}
	if dt.Val != "" {
		switch {
		case strings.HasPrefix(dt.Val, "[]"):
			pointers := cxx.String()
			cxx.Reset()
			cxx.WriteString("array<")
			dt.Val = dt.Val[2:]
			cxx.WriteString(dt.String())
			cxx.WriteByte('>')
			cxx.WriteString(pointers)
			return cxx.String()
		case dt.Id == xtype.Map && dt.Val[0] == '[':
			pointers := cxx.String()
			types := dt.Tag.([]DataType)
			cxx.Reset()
			cxx.WriteString("map<")
			cxx.WriteString(types[0].String())
			cxx.WriteByte(',')
			cxx.WriteString(types[1].String())
			cxx.WriteByte('>')
			cxx.WriteString(pointers)
			return cxx.String()
		}
	}
	if dt.Tag != nil {
		switch t := dt.Tag.(type) {
		case Genericable:
			return dt.StructString() + cxx.String()
		case []DataType:
			dt.Tag = genericableTypes{t}
			return dt.StructString() + cxx.String()
		}
	}
	switch dt.Id {
	case xtype.Id, xtype.Enum:
		return xapi.OutId(dt.Val, dt.Tok.File) + cxx.String()
	case xtype.Struct:
		return dt.StructString() + cxx.String()
	case xtype.Func:
		return dt.FuncString() + cxx.String()
	default:
		return xtype.CxxTypeIdFromType(dt.Id) + cxx.String()
	}
}

// StructString returns cxx value of struct DataType.
func (dt *DataType) StructString() string {
	var cxx strings.Builder
	cxx.WriteString(xapi.OutId(dt.Val, dt.Tok.File))
	s := dt.Tag.(Genericable)
	types := s.Generics()
	if len(types) == 0 {
		return cxx.String()
	}
	cxx.WriteByte('<')
	for _, t := range types {
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// FuncString returns cxx value of function DataType.
func (dt *DataType) FuncString() string {
	var cxx strings.Builder
	cxx.WriteString("func<")
	fun := dt.Tag.(*Func)
	cxx.WriteString(fun.RetType.String())
	cxx.WriteByte('(')
	if len(fun.Params) > 0 {
		for _, param := range fun.Params {
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
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}
