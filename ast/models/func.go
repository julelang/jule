package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

// Func is function declaration AST model.
type Func struct {
	Pub        bool
	Tok        Tok
	Id         string
	Generics   []*GenericType
	Combines   *[][]DataType
	Attributes []Attribute
	Params     []Param
	RetType    RetType
	Block      *Block
	Receiver   *DataType
}

// FindAttribute returns attribute if exist, nil if not.
func (f *Func) FindAttribute(kind string) *Attribute {
	for i := range f.Attributes {
		attribute := &f.Attributes[i]
		if attribute.Tag.Kind == kind {
			return attribute
		}
	}
	return nil
}

// DataTypeString returns data type string of function.
func (f *Func) DataTypeString() string {
	var cxx strings.Builder
	cxx.WriteByte('(')
	if len(f.Params) > 0 {
		for _, p := range f.Params {
			if p.Variadic {
				cxx.WriteString("...")
			}
			cxx.WriteString(p.Type.Kind)
			cxx.WriteString(", ")
		}
		cxxStr := cxx.String()[:cxx.Len()-2]
		cxx.Reset()
		cxx.WriteString(cxxStr)
	}
	cxx.WriteByte(')')
	if f.RetType.Type.MultiTyped {
		cxx.WriteByte('[')
		for _, t := range f.RetType.Type.Tag.([]DataType) {
			cxx.WriteString(t.Kind)
			cxx.WriteByte(',')
		}
		return cxx.String()[:cxx.Len()-1] + "]"
	} else if f.RetType.Type.Id != xtype.Void {
		cxx.WriteString(f.RetType.Type.Kind)
	}
	return cxx.String()
}

// OutId returns xapi.OutId result of function.
func (f *Func) OutId() string {
	if f.Receiver != nil {
		return f.Id
	}
	return xapi.OutId(f.Id, f.Tok.File)
}

// DefString returns define string of function.
func (f *Func) DefString() string {
	return f.Id + f.DataTypeString()
}
