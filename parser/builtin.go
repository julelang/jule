package parser

import (
	"strconv"

	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juletype"
)

const maxI8 = 127
const minI8 = -128
const maxI16 = 32767
const minI16 = -32768
const maxI32 = 2147483647
const minI32 = -2147483648
const maxI64 = 9223372036854775807
const minI64 = -9223372036854775808
const maxU8 = 255
const maxU16 = 65535
const maxU32 = 4294967295
const maxU64 = 18446744073709551615

var i8statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: juletype.I8, Kind: tokens.I8},
			ExprTag: int64(maxI8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I8) + "{" + strconv.FormatInt(maxI8, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.I8, Kind: tokens.I8},
			ExprTag: int64(minI8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I8) + "{" + strconv.FormatInt(minI8, 10) + "}"},
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
			Type:    DataType{Id: juletype.I16, Kind: tokens.I16},
			ExprTag: int64(maxI16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I16) + "{" + strconv.FormatInt(maxI16, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.I16, Kind: tokens.I16},
			ExprTag: int64(minI16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I16) + "{" + strconv.FormatInt(minI16, 10) + "}"},
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
			Type:    DataType{Id: juletype.I32, Kind: tokens.I32},
			ExprTag: int64(maxI32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I32) + "{" + strconv.FormatInt(maxI32, 10) + "}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.I32, Kind: tokens.I32},
			ExprTag: int64(minI32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I32) + "{" + strconv.FormatInt(minI32, 10) + "}"},
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
			Type:    DataType{Id: juletype.I64, Kind: tokens.I64},
			ExprTag: int64(maxI64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I64) + "{" + strconv.FormatInt(maxI64, 10) + "LL}"},
			},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.I64, Kind: tokens.I64},
			ExprTag: int64(minI64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.I64) + "{" + strconv.FormatInt(minI64, 10) + "LL}"},
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
			Type:    DataType{Id: juletype.U8, Kind: tokens.U8},
			ExprTag: uint64(maxU8),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U8) + "{" + strconv.FormatUint(maxU8, 10) + "}"},
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
			Type:    DataType{Id: juletype.U16, Kind: tokens.U16},
			ExprTag: uint64(maxU16),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U16) + "{" + strconv.FormatUint(maxU16, 10) + "}"},
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
			Type:    DataType{Id: juletype.U32, Kind: tokens.U32},
			ExprTag: uint64(maxU32),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U32) + "{" + strconv.FormatUint(maxU32, 10) + "}"},
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
			Type:    DataType{Id: juletype.U64, Kind: tokens.U64},
			ExprTag: uint64(maxU64),
			Expr: models.Expr{
				Model: exprNode{juletype.CppId(juletype.U64) + "{" + strconv.FormatUint(maxU64, 10) + "ULL}"},
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
			Type:  DataType{Id: juletype.UInt, Kind: tokens.UINT},
		},
	},
}

var intStatics = &Defmap{
	Globals: []*Var{
		{
			Const: true,
			Id:    "max",
			Type:  DataType{Id: juletype.Int, Kind: tokens.INT},
		},
		{
			Const: true,
			Id:    "min",
			Type:  DataType{Id: juletype.Int, Kind: tokens.INT},
		},
	},
}

const maxF32 = 0x1p127 * (1 + (1 - 0x1p-23))
const minF32 = 1.17549435082228750796873653722224568e-38

var min_modelF32 = exprNode{juletype.CppId(juletype.F32) + "{1.17549435082228750796873653722224568e-38F}"}

var f32statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: juletype.F32, Kind: tokens.F32},
			ExprTag: float64(maxF32),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(maxF32, 'e', -1, 32) + "F"}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.F32, Kind: tokens.F32},
			ExprTag: float64(minF32),
			Expr:    models.Expr{Model: min_modelF32},
		},
	},
}

const maxF64 = 0x1p1023 * (1 + (1 - 0x1p-52))
const minF64 = 2.22507385850720138309023271733240406e-308

var min_modelF64 = exprNode{juletype.CppId(juletype.F64) + "{2.22507385850720138309023271733240406e-308}"}

var f64statics = &Defmap{
	Globals: []*Var{
		{
			Pub:     true,
			Const:   true,
			Id:      "max",
			Type:    DataType{Id: juletype.F64, Kind: tokens.F64},
			ExprTag: float64(maxF64),
			Expr:    models.Expr{Model: exprNode{strconv.FormatFloat(maxF64, 'e', -1, 64)}},
		},
		{
			Pub:     true,
			Const:   true,
			Id:      "min",
			Type:    DataType{Id: juletype.F64, Kind: tokens.F64},
			ExprTag: float64(minF64),
			Expr:    models.Expr{Model: min_modelF64},
		},
	},
}

var strDefaultFunc = Func{
	Pub:     true,
	Id:      "str",
	Params:  []Param{{Id: "obj", Type: DataType{Id: juletype.Any, Kind: tokens.ANY}}},
	RetType: RetType{Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
}

var errorTrait = &trait{
	Ast: &models.Trait{
		Id: "Error",
	},
	Defs: &Defmap{
		Funcs: []*Fn{
			{Ast: &models.Func{
				Pub:     true,
				Id:      "error",
				RetType: models.RetType{Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
			}},
		},
	},
}

var errorType = DataType{
	Id:   juletype.Trait,
	Kind: errorTrait.Ast.Id,
	Tag:  errorTrait,
	Pure: true,
}

var panicFunc = &Fn{
	Ast: &models.Func{
		Pub: true,
		Id:  "panic",
		Params: []models.Param{
			{
				Id:   "error",
				Type: DataType{Id: juletype.Any, Kind: juletype.TypeMap[juletype.Any]},
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
			Id:   juletype.Void,
			Kind: juletype.TypeMap[juletype.Void],
		},
	},
}

var unsafe_sizeof_fn = &Fn{Ast: &models.Func{Id: "sizeof"}}
var unsafe_alignof_fn = &Fn{Ast: &models.Func{Id: "alignof"}}

var unsafe_funcs = []*Fn{
	unsafe_sizeof_fn,
	unsafe_alignof_fn,
}

var recoverFunc = &Fn{
	Ast: &models.Func{
		Pub: true,
		Id:  "recover",
		Params: []models.Param{
			{
				Id: "handler",
				Type: models.DataType{
					Id:   juletype.Func,
					Kind: errorHandlerFunc.DataTypeString(),
					Tag:  errorHandlerFunc,
				},
			},
		},
	},
}

// Parser instance for built-in generics.
var genericFile = &Parser{}

// Builtin definitions.
var Builtin = &Defmap{
	Types: []*models.Type{
		{
			Pub:  true,
			Id:   "byte",
			Type: DataType{Id: juletype.U8, Kind: juletype.TypeMap[juletype.U8]},
		},
		{
			Pub:  true,
			Id:   "rune",
			Type: DataType{Id: juletype.I32, Kind: juletype.TypeMap[juletype.I32]},
		},
	},
	Funcs: []*Fn{
		panicFunc,
		recoverFunc,
		{Ast: &Func{
			Pub: true,
			Id:  "out",
			RetType: RetType{
				Type: DataType{Id: juletype.Void, Kind: juletype.TypeMap[juletype.Void]},
			},
			Params: []Param{{
				Id:   "expr",
				Type: DataType{Id: juletype.Any, Kind: tokens.ANY},
			}},
		}},
		{Ast: &Func{
			Pub:      true,
			Id:       "new",
			Owner:    genericFile,
			Generics: []*GenericType{{Id: "T"}},
			Attributes: []models.Attribute{
				{Tag: jule.Attribute_TypeArg},
			},
			RetType: RetType{Type: DataType{Id: juletype.Id, Kind: tokens.STAR + "T"}},
		}},
		{Ast: &Func{
			Pub:      true,
			Id:       "make",
			Owner:    genericFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType: models.RetType{
				Type: DataType{
					Id:            juletype.Slice,
					Kind:          jule.Prefix_Slice + "Item",
					ComponentType: &DataType{Id: juletype.Id, Kind: "Item"},
				},
			},
			Params: []models.Param{
				{
					Id:   "n",
					Type: DataType{Id: juletype.Int, Kind: juletype.TypeMap[juletype.Int]},
				},
			},
		}},
		{Ast: &Func{
			Pub:      true,
			Id:       "copy",
			Owner:    genericFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType:  models.RetType{Type: DataType{Id: juletype.Int, Kind: juletype.TypeMap[juletype.Int]}},
			Params: []models.Param{
				{
					Id: "dest",
					Type: DataType{
						Id:            juletype.Slice,
						Kind:          jule.Prefix_Slice + "Item",
						ComponentType: &DataType{Id: juletype.Id, Kind: "Item"},
					},
				},
				{
					Id: "src",
					Type: DataType{
						Id:            juletype.Slice,
						Kind:          jule.Prefix_Slice + "Item",
						ComponentType: &DataType{Id: juletype.Id, Kind: "Item"},
					},
				},
			},
		}},
		{Ast: &Func{
			Pub:      true,
			Id:       "append",
			Owner:    genericFile,
			Generics: []*GenericType{{Id: "Item"}},
			RetType: models.RetType{
				Type: DataType{
					Id:            juletype.Slice,
					Kind:          jule.Prefix_Slice + "Item",
					ComponentType: &DataType{Id: juletype.Id, Kind: "Item"},
				},
			},
			Params: []models.Param{
				{
					Id: "src",
					Type: DataType{
						Id:            juletype.Slice,
						Kind:          jule.Prefix_Slice + "Item",
						ComponentType: &DataType{Id: juletype.Id, Kind: "Item"},
					},
				},
				{
					Id:       "components",
					Type:     DataType{Id: juletype.Id, Kind: "Item"},
					Variadic: true,
				},
			},
		}},
	},
	Traits: []*trait{
		errorTrait,
	},
}

var strDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: DataType{Id: juletype.Int, Kind: tokens.INT},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has_prefix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has_suffix",
			Params:  []Param{{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "find",
			Params:  []Param{{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Int, Kind: tokens.INT}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "rfind",
			Params:  []Param{{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Int, Kind: tokens.INT}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "trim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "rtrim",
			Params:  []Param{{Id: "bytes", Type: DataType{Id: juletype.Str, Kind: tokens.STR}}},
			RetType: RetType{Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "split",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
				{
					Id:   "n",
					Type: DataType{Id: juletype.Int, Kind: tokens.INT},
				},
			},
			RetType: RetType{Type: DataType{Id: juletype.Str, Kind: jule.Prefix_Slice + tokens.STR}},
		}},
		{Ast: &Func{
			Pub: true,
			Id:  "replace",
			Params: []Param{
				{Id: "sub", Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
				{Id: "new", Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
				{
					Id:   "n",
					Type: DataType{Id: juletype.Int, Kind: tokens.INT},
				},
			},
			RetType: RetType{Type: DataType{Id: juletype.Str, Kind: tokens.STR}},
		}},
	},
}

// Use this at before use strDefs if necessary.
// Because some definitions is responsive for str data types.
func readyStrDefs(s value) {
	lenVar := strDefs.Globals[0]
	lenVar.Const = s.constExpr
	if lenVar.Const {
		lenVar.ExprTag = int64(len(s.expr.(string)))
		lenVar.Expr.Model = getModel(value{
			expr: lenVar.ExprTag,
			data: models.Data{Type: lenVar.Type},
		})
	}
}

var sliceDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: DataType{Id: juletype.Int, Kind: tokens.INT},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
	},
}

var arrayDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: DataType{Id: juletype.Int, Kind: tokens.INT},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
		{Ast: &Func{
			Pub:     true,
			Id:      "empty",
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
	},
}

var mapDefs = &Defmap{
	Globals: []*Var{
		{
			Pub:  true,
			Id:   "len",
			Type: DataType{Id: juletype.Int, Kind: tokens.INT},
			Tag:  "len()",
		},
	},
	Funcs: []*Fn{
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
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
		}},
		{Ast: &Func{
			Pub:     true,
			Id:      "has",
			Params:  []Param{{Id: "key"}},
			RetType: RetType{Type: DataType{Id: juletype.Bool, Kind: tokens.BOOL}},
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

	keysFunc, _, _ := mapDefs.funcById("keys", nil)
	keysFunc.Ast.RetType.Type = keyt
	keysFunc.Ast.RetType.Type.Kind = jule.Prefix_Slice + keysFunc.Ast.RetType.Type.Kind

	valuesFunc, _, _ := mapDefs.funcById("values", nil)
	valuesFunc.Ast.RetType.Type = valt
	valuesFunc.Ast.RetType.Type.Kind = jule.Prefix_Slice + valuesFunc.Ast.RetType.Type.Kind

	hasFunc, _, _ := mapDefs.funcById("has", nil)
	hasFunc.Ast.Params[0].Type = keyt

	delFunc, _, _ := mapDefs.funcById("del", nil)
	delFunc.Ast.Params[0].Type = keyt
}

func init() {
	// Copy out function as outln
	outFunc, _, _ := Builtin.funcById("out", nil)
	outlnFunc := new(Fn)
	*outlnFunc = *outFunc
	outlnFunc.Ast = new(models.Func)
	*outlnFunc.Ast = *outFunc.Ast
	outlnFunc.Ast.Id = "outln"
	Builtin.Funcs = append(Builtin.Funcs, outlnFunc)

	// Set bits of platform-dependent types
	intMax := intStatics.Globals[0]
	intMin := intStatics.Globals[1]
	uintMax := uintStatics.Globals[0]
	switch juletype.BitSize {
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
