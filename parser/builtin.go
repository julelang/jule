package parser

import (
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

var i8statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.I8, Val: tokens.I8},
			Tag:   "INT8_MAX",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.I8, Val: tokens.I8},
			Tag:   "INT8_MIN",
		},
	},
}

var i16statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.I16, Val: tokens.I16},
			Tag:   "INT16_MAX",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.I16, Val: tokens.I16},
			Tag:   "INT16_MIN",
		},
	},
}

var i32statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.I32, Val: tokens.I32},
			Tag:   "INT32_MAX",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.I32, Val: tokens.I32},
			Tag:   "INT32_MIN",
		},
	},
}

var i64statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.I64, Val: tokens.I64},
			Tag:   "INT64_MAX",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.I64, Val: tokens.I64},
			Tag:   "INT64_MIN",
		},
	},
}

var u8statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.U8, Val: tokens.U8},
			Tag:   "UINT8_MAX",
		},
	},
}

var u16statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.U16, Val: tokens.U16},
			Tag:   "UINT16_MAX",
		},
	},
}

var u32statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.U32, Val: tokens.U32},
			Tag:   "UINT32_MAX",
		},
	},
}

var u64statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.U64, Val: tokens.U64},
			Tag:   "UINT64_MAX",
		},
	},
}

var uintStatics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.UInt, Val: tokens.UINT},
			Tag:   "SIZE_MAX",
		},
	},
}

var intStatics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.Int, Val: tokens.INT},
			Tag:   "",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.Int, Val: tokens.INT},
			Tag:   "",
		},
	},
}

var f32statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.F32, Val: tokens.F32},
			Tag:   "__FLT_MAX__",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.F32, Val: tokens.F32},
			Tag:   "__FLT_MIN__",
		},
	},
}

var f64statics = &Defmap{
	Globals: []*Var{
		{
			Id:    "max",
			Const: true,
			Type:  DataType{Id: xtype.F64, Val: tokens.F64},
			Tag:   "__DBL_MAX__",
		},
		{
			Id:    "min",
			Const: true,
			Type:  DataType{Id: xtype.F64, Val: tokens.F64},
			Tag:   "__DBL_MIN__",
		},
	},
}

var strStatics = &Defmap{
	Globals: []*Var{
		{
			Id:    "npos",
			Const: true,
			Type:  DataType{Id: xtype.UInt, Val: tokens.UINT},
			Tag:   "std::string::npos",
		},
	},
}

var strDefaultFunc = Func{
	Id:      "str",
	Params:  []Param{{Id: "obj", Type: DataType{Id: xtype.Any, Val: "any"}}},
	RetType: DataType{Id: xtype.Str, Val: tokens.STR},
}

var errorStruct = &xstruct{
	Ast: Struct{
		Id: "error",
	},
	Defs: &Defmap{
		Globals: []*Var{
			{
				Id:   "message",
				Type: DataType{Id: xtype.Str, Val: tokens.STR},
			},
		},
	},
	constructor: &Func{
		Params: []Param{
			{
				Id:      "message",
				Type:    DataType{Id: xtype.Str, Val: tokens.STR},
				Default: Expr{Model: exprNode{xapi.ToStr(`"error: undefined error"`)}},
			},
		},
	},
}

var errorType = DataType{Id: xtype.Struct, Val: "error", Tag: errorStruct}

// Builtin definitions.
var Builtin = &Defmap{
	Funcs: []*function{
		{
			Ast: Func{
				Id:      "out",
				RetType: DataType{Id: xtype.Void, Val: xtype.VoidTypeStr},
				Params: []Param{{
					Id:      "v",
					Const:   true,
					Type:    DataType{Id: xtype.Any, Val: "any"},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
		{
			Ast: Func{
				Id:      "outln",
				RetType: DataType{Id: xtype.Void, Val: xtype.VoidTypeStr},
				Params: []Param{{
					Id:      "v",
					Const:   true,
					Type:    DataType{Id: xtype.Any, Val: "any"},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
	},
}

var strDefs = &Defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: xtype.UInt, Val: tokens.UINT},
			Tag:   "len()",
		},
	},
	Funcs: []*function{
		{Ast: Func{
			Id:      "empty",
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
			Id:      "has_prefix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
			Id:      "has_suffix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
			Id:      "find",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.UInt, Val: tokens.UINT},
		}},
		{Ast: Func{
			Id:      "rfind",
			Params:  []Param{{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.UInt, Val: tokens.UINT},
		}},
		{Ast: Func{
			Id:      "trim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.Str, Val: tokens.STR},
		}},
		{Ast: Func{
			Id:      "rtrim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: xtype.Str, Val: tokens.STR}}},
			RetType: DataType{Id: xtype.Str, Val: tokens.STR},
		}},
		{Ast: Func{
			Id: "split",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}},
				{
					Id:      "n",
					Type:    DataType{Id: xtype.I64, Val: tokens.I64},
					Default: Expr{Model: exprNode{"-1"}},
				},
			},
			RetType: DataType{Id: xtype.Str, Val: "[]" + tokens.STR},
		}},
		{Ast: Func{
			Id: "replace",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: xtype.Str, Val: tokens.STR}},
				{Id: "new", Type: DataType{Id: xtype.Str, Val: tokens.STR}},
				{
					Id:      "n",
					Type:    DataType{Id: xtype.I64, Val: tokens.I64},
					Default: Expr{Model: exprNode{"-1"}},
				},
			},
			RetType: DataType{Id: xtype.Str, Val: tokens.STR},
		}},
	},
}

var arrDefs = &Defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: xtype.UInt, Val: tokens.UINT},
			Tag:   "len()",
		},
	},
	Funcs: []*function{
		{Ast: Func{Id: "clear"}},
		{Ast: Func{
			Id:      "empty",
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
			Id:     "find",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "rfind",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "erase",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "erase_all",
			Params: []Param{{Id: "value"}},
		}},
		{Ast: Func{
			Id:     "append",
			Params: []Param{{Id: "values", Variadic: true}},
		}},
		{Ast: Func{
			Id: "insert",
			Params: []Param{
				{Id: "start", Type: DataType{Id: xtype.UInt, Val: tokens.UINT}},
				{Id: "values", Variadic: true},
			},
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
	},
}

func readyArrDefs(arrt DataType) {
	elemType := typeOfArrayItems(arrt)

	findFunc, _, _ := arrDefs.funcById("find", nil)
	findFunc.Ast.Params[0].Type = elemType
	findFunc.Ast.RetType = elemType
	findFunc.Ast.RetType.Val = tokens.STAR + findFunc.Ast.RetType.Val

	rfindFunc, _, _ := arrDefs.funcById("rfind", nil)
	rfindFunc.Ast.Params[0].Type = elemType
	rfindFunc.Ast.RetType = elemType
	rfindFunc.Ast.RetType.Val = tokens.STAR + rfindFunc.Ast.RetType.Val

	eraseFunc, _, _ := arrDefs.funcById("erase", nil)
	eraseFunc.Ast.Params[0].Type = elemType

	eraseAllFunc, _, _ := arrDefs.funcById("erase_all", nil)
	eraseAllFunc.Ast.Params[0].Type = elemType

	appendFunc, _, _ := arrDefs.funcById("append", nil)
	appendFunc.Ast.Params[0].Type = elemType

	insertFunc, _, _ := arrDefs.funcById("insert", nil)
	insertFunc.Ast.Params[1].Type = elemType
}

var mapDefs = &Defmap{
	Globals: []*Var{
		{
			Id:    "len",
			Const: true,
			Type:  DataType{Id: xtype.UInt, Val: tokens.UINT},
			Tag:   "size()",
		},
	},
	Funcs: []*function{
		{Ast: Func{Id: "clear"}},
		{Ast: Func{Id: "keys"}},
		{Ast: Func{Id: "values"}},
		{Ast: Func{
			Id:      "empty",
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
			Id:      "has",
			Params:  []Param{{Id: "key", Const: true}},
			RetType: DataType{Id: xtype.Bool, Val: tokens.BOOL},
		}},
		{Ast: Func{
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
	keysFunc.Ast.RetType = keyt
	keysFunc.Ast.RetType.Val = "[]" + keysFunc.Ast.RetType.Val

	valuesFunc, _, _ := mapDefs.funcById("values", nil)
	valuesFunc.Ast.RetType = valt
	valuesFunc.Ast.RetType.Val = "[]" + valuesFunc.Ast.RetType.Val

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
}
