package parser

import (
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xtype"
)

var i8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.I8, Kind: tokens.I8},
			Tag:   "INT8_MAX",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.I8, Kind: tokens.I8},
			Tag:   "INT8_MIN",
		},
	},
}

var i16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.I16, Kind: tokens.I16},
			Tag:   "INT16_MAX",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.I16, Kind: tokens.I16},
			Tag:   "INT16_MIN",
		},
	},
}

var i32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.I32, Kind: tokens.I32},
			Tag:   "INT32_MAX",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.I32, Kind: tokens.I32},
			Tag:   "INT32_MIN",
		},
	},
}

var i64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.I64, Kind: tokens.I64},
			Tag:   "INT64_MAX",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.I64, Kind: tokens.I64},
			Tag:   "INT64_MIN",
		},
	},
}

var u8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.U8, Kind: tokens.U8},
			Tag:   "UINT8_MAX",
		},
	},
}

var u16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.U16, Kind: tokens.U16},
			Tag:   "UINT16_MAX",
		},
	},
}

var u32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.U32, Kind: tokens.U32},
			Tag:   "UINT32_MAX",
		},
	},
}

var u64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.U64, Kind: tokens.U64},
			Tag:   "UINT64_MAX",
		},
	},
}

var uintStatics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "SIZE_MAX",
		},
	},
}

var intStatics = &Defmap{
	Globals: []*Var{
		{
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.Int, Kind: tokens.INT},
		},
		{
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.Int, Kind: tokens.INT},
		},
	},
}

var f32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.F32, Kind: tokens.F32},
			Tag:   "__FLT_MAX__",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.F32, Kind: tokens.F32},
			Tag:   "__FLT_MIN__",
		},
	},
}

var f64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "max",
			Type:  DataType{Id: xtype.F64, Kind: tokens.F64},
			Tag:   "__DBL_MAX__",
		},
		{
			Pub:   true,
			Const: true,
			Id:    "min",
			Type:  DataType{Id: xtype.F64, Kind: tokens.F64},
			Tag:   "__DBL_MIN__",
		},
	},
}

var strStatics = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "npos",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "std::string::npos",
		},
	},
}

var strDefaultFunc = Func{
	Pub:     true,
	Id:      "str",
	Params:  []Param{{Id: "obj", Type: DataType{Id: xtype.Any, Kind: tokens.ANY}}},
	RetType: RetType{Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
}

var errorStruct = &xstruct{
	Ast: Struct{
		Id: "Error",
	},
	Defs: &Defmap{
		Globals: []*Var{
			{
				Pub:  true,
				Id:   "message",
				Type: DataType{Id: xtype.Str, Kind: tokens.STR},
			},
		},
	},
	constructor: &Func{
		Pub: true,
		Params: []Param{
			{
				Id:      "message",
				Type:    DataType{Id: xtype.Str, Kind: tokens.STR},
				Default: Expr{Model: exprNode{}},
			},
		},
	},
}

var errorType = DataType{
	Id:   xtype.Struct,
	Kind: errorStruct.Ast.Id,
	Tag:  errorStruct,
}

var panicFunc = &function{
	Ast: &models.Func{
		Pub: true,
		Id:  "panic",
		Params: []models.Param{
			{
				Const: true,
				Id:    "error",
				Type:  errorType,
			},
		},
	},
}

var errorHandlerFunc = &models.Func{
	Id: "handler",
	Params: []models.Param{
		{
			Const: true,
			Id:    "error",
			Type:  errorType,
		},
	},
	RetType: models.RetType{
		Type: models.DataType{
			Id:   xtype.Void,
			Kind: xtype.VoidTypeStr,
		},
	},
}

var recoverFunc = &function{
	Ast: &models.Func{
		Pub: true,
		Id:  "recover",
		Params: []models.Param{
			{
				Id: "handler",
				Type: models.DataType{
					Id:   xtype.Func,
					Kind: errorHandlerFunc.DataTypeString(),
					Tag:  errorHandlerFunc,
				},
			},
		},
	},
}

// Builtin definitions.
var Builtin = &Defmap{
	Funcs: []*function{
		panicFunc,
		recoverFunc,
		{
			Ast: &Func{
				Pub:     true,
				Id:      "out",
				RetType: RetType{Type: DataType{Id: xtype.Void, Kind: xtype.VoidTypeStr}},
				Params: []Param{{
					Const:   true,
					Id:      "v",
					Type:    DataType{Id: xtype.Any, Kind: tokens.ANY},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
		{
			Ast: &Func{
				Pub:     true,
				Id:      "outln",
				RetType: RetType{Type: DataType{Id: xtype.Void, Kind: xtype.VoidTypeStr}},
				Params: []Param{{
					Const:   true,
					Id:      "v",
					Type:    DataType{Id: xtype.Any, Kind: tokens.ANY},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
	},
	Structs: []*xstruct{
		errorStruct,
	},
}

var strDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "len",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "len()",
		},
	},
	Funcs: []*function{
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has_prefix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has_suffix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "find",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.UInt, Kind: tokens.UINT}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "rfind",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.UInt, Kind: tokens.UINT}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "trim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "rtrim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: xtype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "split",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
				{
					Id:      "n",
					Type:    DataType{Id: xtype.I64, Kind: tokens.I64},
					Default: Expr{Model: exprNode{"-1"}},
				},
			},
			RetType: RetType{Type: DataType{Id: xtype.Str, Kind: x.Prefix_Slice + tokens.STR}},
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "replace",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
				{Id: "new", Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
				{
					Id:      "n",
					Type:    DataType{Id: xtype.I64, Kind: tokens.I64},
					Default: Expr{Model: exprNode{"-1"}},
				},
			},
			RetType: RetType{Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
		}},
	},
}

var sliceDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "len",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "len()",
		},
	},
	Funcs: []*function{
		{Ast: &Func{
			Pub: true,
			Id:  "clear",
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "find",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "rfind",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "erase",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "erase_all",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "append",
			Params: []Param{{Id: "values", Variadic: true}},
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "insert",
			Params: []Param{
				{Id: "start", Type: DataType{Id: xtype.UInt, Kind: tokens.UINT}},
				{Id: "values", Variadic: true},
			},
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
	},
}

var arrayDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "len",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "len()",
		},
	},
	Funcs: []*function{
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
	},
}

func readySliceDefs(t DataType) {
	elemType := typeOfSliceComponents(t)

	findFunc, _, _ := sliceDefs.funcById("find", nil)
	findFunc.Ast.Params[0].Type = elemType
	findFunc.Ast.RetType.Type = elemType
	findFunc.Ast.RetType.Type.Kind = tokens.STAR + findFunc.Ast.RetType.Type.Kind

	rfindFunc, _, _ := sliceDefs.funcById("rfind", nil)
	rfindFunc.Ast.Params[0].Type = elemType
	rfindFunc.Ast.RetType.Type = elemType
	rfindFunc.Ast.RetType.Type.Kind = tokens.STAR + rfindFunc.Ast.RetType.Type.Kind

	eraseFunc, _, _ := sliceDefs.funcById("erase", nil)
	eraseFunc.Ast.Params[0].Type = elemType

	eraseAllFunc, _, _ := sliceDefs.funcById("erase_all", nil)
	eraseAllFunc.Ast.Params[0].Type = elemType

	appendFunc, _, _ := sliceDefs.funcById("append", nil)
	appendFunc.Ast.Params[0].Type = elemType

	insertFunc, _, _ := sliceDefs.funcById("insert", nil)
	insertFunc.Ast.Params[1].Type = elemType
}

var mapDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:   true,
			Const: true,
			Id:    "len",
			Type:  DataType{Id: xtype.UInt, Kind: tokens.UINT},
			Tag:   "size()",
		},
	},
	Funcs: []*function{
		{Ast: &Func{
			Pub: true,
			Id:  "clear",
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "keys",
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "values",
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has",
			Params:  []Param{{Id: "key", Const: true}},
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "del",
			Params: []Param{{Id: "key", Const: true}},
		}},
	},
}

// Use this at before use mapDefs if necessary.
// Because some definitions is responsive for map data-types.
func readyMapDefs(mapt DataType) {
	types := mapt.Tag.([]DataType)
	keyt := types[0]
	valt := types[1]

	keysFunc, _, _ := mapDefs.funcById("keys", nil)
	keysFunc.Ast.RetType.Type = keyt
	keysFunc.Ast.RetType.Type.Kind = x.Prefix_Slice + keysFunc.Ast.RetType.Type.Kind

	valuesFunc, _, _ := mapDefs.funcById("values", nil)
	valuesFunc.Ast.RetType.Type = valt
	valuesFunc.Ast.RetType.Type.Kind = x.Prefix_Slice + valuesFunc.Ast.RetType.Type.Kind

	hasFunc, _, _ := mapDefs.funcById("has", nil)
	hasFunc.Ast.Params[0].Type = keyt

	delFunc, _, _ := mapDefs.funcById("del", nil)
	delFunc.Ast.Params[0].Type = keyt
}

func init() {
	intMax := intStatics.Globals[0]
	intMin := intStatics.Globals[1]
	switch xtype.BitSize {
	case 8:
		intMax.Tag = i8statics.Globals[0].Tag
		intMin.Tag = i8statics.Globals[1].Tag
	case 16:
		intMax.Tag = i16statics.Globals[0].Tag
		intMin.Tag = i16statics.Globals[1].Tag
	case 32:
		intMax.Tag = i32statics.Globals[0].Tag
		intMin.Tag = i32statics.Globals[1].Tag
	case 64:
		intMax.Tag = i64statics.Globals[0].Tag
		intMin.Tag = i64statics.Globals[1].Tag
	}

	errorStruct.constructor.Id = errorStruct.Ast.Id
	errorStruct.constructor.RetType.Type = errorType
}
