package parser

import (
	"math"
	"strconv"

	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xtype"
)

var i8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.I8, Kind: tokens.I8},
			ExprTag: int64(math.MaxInt8),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I8) + "{" + strconv.FormatInt(math.MaxInt8, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.I8, Kind: tokens.I8},
			ExprTag: int64(math.MinInt8),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I8) + "{" + strconv.FormatInt(math.MinInt8, 10) + "}"},
			},
		},
	},
}

var i16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.I16, Kind: tokens.I16},
			ExprTag: int64(math.MaxInt16),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I16) + "{" + strconv.FormatInt(math.MaxInt16, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.I16, Kind: tokens.I16},
			ExprTag: int64(math.MinInt16),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I16) + "{" + strconv.FormatInt(math.MinInt16, 10) + "}"},
			},
		},
	},
}

var i32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.I32, Kind: tokens.I32},
			ExprTag: int64(math.MaxInt32),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I32) + "{" + strconv.FormatInt(math.MaxInt32, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.I32, Kind: tokens.I32},
			ExprTag: int64(math.MinInt32),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I32) + "{" + strconv.FormatInt(math.MinInt32, 10) + "}"},
			},
		},
	},
}

var i64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.I64, Kind: tokens.I64},
			ExprTag: int64(math.MaxInt64),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I64) + "{" + strconv.FormatInt(math.MaxInt64, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.I64, Kind: tokens.I64},
			ExprTag: int64(math.MinInt64),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.I64) + "{" + strconv.FormatInt(math.MinInt64, 10) + "}"},
			},
		},
	},
}

var u8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.U8, Kind: tokens.U8},
			ExprTag: uint64(math.MaxUint8),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.U8) + "{" + strconv.FormatUint(math.MaxUint8, 10) + "}"},
			},
		},
	},
}

var u16statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.U16, Kind: tokens.U16},
			ExprTag: uint64(math.MaxUint16),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.U16) + "{" + strconv.FormatUint(math.MaxUint16, 10) + "}"},
			},
		},
	},
}

var u32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.U32, Kind: tokens.U32},
			ExprTag: uint64(math.MaxUint32),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.U32) + "{" + strconv.FormatUint(math.MaxUint32, 10) + "}"},
			},
		},
	},
}

var u64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.U64, Kind: tokens.U64},
			ExprTag: uint64(math.MaxUint64),
			Expr: models.Expr{
				Model: exprNode{xtype.CxxId(xtype.U64) + "{" + strconv.FormatUint(math.MaxUint64, 10) + "}"},
			},
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

const f32min = float64(1.17549435082228750796873653722224568e-38)

var f32min_model = exprNode{xtype.CxxId(xtype.F32) + "{1.17549435082228750796873653722224568e-38F}"}

var f32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.F32, Kind: tokens.F32},
			ExprTag: float64(math.MaxFloat32),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(math.MaxFloat32, 'e', -1, 32) + "F"}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.F32, Kind: tokens.F32},
			ExprTag: f32min,
			Expr:    models.Expr{Model: f32min_model},
		},
	},
}

const f64min = float64(2.22507385850720138309023271733240406e-308)

var f64min_model = exprNode{xtype.CxxId(xtype.F64) + "{2.22507385850720138309023271733240406e-308}"}

var f64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: xtype.F64, Kind: tokens.F64},
			ExprTag: float64(math.MaxFloat64),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(math.MaxFloat64, 'e', -1, 64)}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: xtype.F64, Kind: tokens.F64},
			ExprTag: f64min,
			Expr:    models.Expr{Model: f64min_model},
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

var errorTrait = &trait{
	Ast: &models.Trait{
		Id: "Error",
	},
	Defs: &Defmap{
		Funcs: []*function{
			{Ast: &models.Func{
				Pub:     true,
				Id:      "error",
				RetType: models.RetType{Type: DataType{Id: xtype.Str, Kind: tokens.STR}},
			}},
		},
	},
}

var errorType = DataType{
	Id:              xtype.Trait,
	Kind:            errorTrait.Ast.Id,
	Tag:             errorTrait,
	DontUseOriginal: true,
}

var panicFunc = &function{
	Ast: &models.Func{
		Pub: true,
		Id:  "panic",
		Params: []models.Param{
			{
				Id:   "error",
				Type: errorType,
			},
		},
	},
}

var errorHandlerFunc = &models.Func{
	Id: "handler",
	Params: []models.Param{
		{
			Id:   "error",
			Type: errorType,
		},
	},
	RetType: models.RetType{
		Type: models.DataType{
			Id:   xtype.Void,
			Kind: xtype.TypeMap[xtype.Void],
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
	Types: []*models.Type{
		{
			Pub:  true,
			Id:   "byte",
			Type: DataType{Id: xtype.U8, Kind: xtype.TypeMap[xtype.U8]},
		},
		{
			Pub:  true,
			Id:   "rune",
			Type: DataType{Id: xtype.I32, Kind: xtype.TypeMap[xtype.I32]},
		},
	},
	Funcs: []*function{
		panicFunc,
		recoverFunc,
		{
			Ast: &Func{
				Pub: true,
				Id:  "out",
				RetType: RetType{
					Type: DataType{Id: xtype.Void, Kind: xtype.TypeMap[xtype.Void]},
				},
				Params: []Param{{
					Id:      "v",
					Type:    DataType{Id: xtype.Any, Kind: tokens.ANY},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
		{
			Ast: &Func{
				Pub: true,
				Id:  "outln",
				RetType: RetType{
					Type: DataType{Id: xtype.Void, Kind: xtype.TypeMap[xtype.Void]},
				},
				Params: []Param{{
					Id:      "v",
					Type:    DataType{Id: xtype.Any, Kind: tokens.ANY},
					Default: Expr{Model: exprNode{`""`}},
				}},
			},
		},
		{
			Ast: &Func{
				Pub: true,
				Id:  "new",
				Attributes: []models.Attribute{
					{Tag: lex.Tok{Kind: x.Attribute_TypeArg}},
				},
				Generics: []*models.GenericType{
					{Id: "T"},
				},
				RetType: RetType{
					Type: DataType{Id: xtype.Id, Kind: "*T"},
				},
			},
		},
	},
	Traits: []*trait{
		errorTrait,
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

	findFunc, _ := sliceDefs.funcById("find", nil)
	findFunc.Ast.Params[0].Type = elemType
	findFunc.Ast.RetType.Type = elemType
	findFunc.Ast.RetType.Type.Kind = tokens.STAR + findFunc.Ast.RetType.Type.Kind

	rfindFunc, _ := sliceDefs.funcById("rfind", nil)
	rfindFunc.Ast.Params[0].Type = elemType
	rfindFunc.Ast.RetType.Type = elemType
	rfindFunc.Ast.RetType.Type.Kind = tokens.STAR + rfindFunc.Ast.RetType.Type.Kind

	eraseFunc, _ := sliceDefs.funcById("erase", nil)
	eraseFunc.Ast.Params[0].Type = elemType

	eraseAllFunc, _ := sliceDefs.funcById("erase_all", nil)
	eraseAllFunc.Ast.Params[0].Type = elemType

	appendFunc, _ := sliceDefs.funcById("append", nil)
	appendFunc.Ast.Params[0].Type = elemType

	insertFunc, _ := sliceDefs.funcById("insert", nil)
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
			Params:  []Param{{Id: "key"}},
			RetType: RetType{Type: DataType{Id: xtype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:    true,
			Id:     "del",
			Params: []Param{{Id: "key"}},
		}},
	},
}

// Use this at before use mapDefs if necessary.
// Because some definitions is responsive for map data-types.
func readyMapDefs(mapt DataType) {
	types := mapt.Tag.([]DataType)
	keyt := types[0]
	valt := types[1]

	keysFunc, _ := mapDefs.funcById("keys", nil)
	keysFunc.Ast.RetType.Type = keyt
	keysFunc.Ast.RetType.Type.Kind = x.Prefix_Slice + keysFunc.Ast.RetType.Type.Kind

	valuesFunc, _ := mapDefs.funcById("values", nil)
	valuesFunc.Ast.RetType.Type = valt
	valuesFunc.Ast.RetType.Type.Kind = x.Prefix_Slice + valuesFunc.Ast.RetType.Type.Kind

	hasFunc, _ := mapDefs.funcById("has", nil)
	hasFunc.Ast.Params[0].Type = keyt

	delFunc, _ := mapDefs.funcById("del", nil)
	delFunc.Ast.Params[0].Type = keyt
}

func init() {
	intMax := intStatics.Globals[0]
	intMin := intStatics.Globals[1]
	uintMax := uintStatics.Globals[0]
	switch xtype.BitSize {
	case 8:
		intMax.Expr = i8statics.Globals[0].Expr
		intMax.ExprTag = i8statics.Globals[0].ExprTag
		intMin.Expr = i8statics.Globals[1].Expr
		intMin.ExprTag = i8statics.Globals[1].ExprTag

		uintMax.Expr = u8statics.Globals[0].Expr
		uintMax.ExprTag = u8statics.Globals[0].ExprTag
	case 16:
		intMax.Expr = i16statics.Globals[0].Expr
		intMax.ExprTag = i16statics.Globals[0].ExprTag
		intMin.Expr = i16statics.Globals[1].Expr
		intMin.ExprTag = i16statics.Globals[1].ExprTag

		uintMax.Expr = u16statics.Globals[0].Expr
		uintMax.ExprTag = u16statics.Globals[0].ExprTag
	case 32:
		intMax.Expr = i32statics.Globals[0].Expr
		intMax.ExprTag = i32statics.Globals[0].ExprTag
		intMin.Expr = i32statics.Globals[1].Expr
		intMin.ExprTag = i32statics.Globals[1].ExprTag

		uintMax.Expr = u32statics.Globals[0].Expr
		uintMax.ExprTag = u32statics.Globals[0].ExprTag
	case 64:
		intMax.Expr = i64statics.Globals[0].Expr
		intMax.ExprTag = i64statics.Globals[0].ExprTag
		intMin.Expr = i64statics.Globals[1].Expr
		intMin.ExprTag = i64statics.Globals[1].ExprTag

		uintMax.Expr = u64statics.Globals[0].Expr
		uintMax.ExprTag = u64statics.Globals[0].ExprTag
	}
}
