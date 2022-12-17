package parser

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/julelang/jule/ast"
	"github.com/julelang/jule/ast/models"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/types"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/julelog"
	"github.com/julelang/jule/pkg/juletype"
)

type File = lex.File
type TypeAlias = models.TypeAlias
type Var = models.Var
type Fn = models.Fn
type Arg = models.Arg
type Param = models.Param
type Type = models.Type
type Expr = models.Expr
type Enum = models.Enum
type Struct = models.Struct
type GenericType = models.GenericType
type RetType = models.RetType

// Parser is parser of Jule code.
type Parser struct {
	attributes       []models.Attribute
	docText          strings.Builder
	currentIter      *models.Iter
	currentCase      *models.Case
	wg               sync.WaitGroup
	rootBlock        *models.Block
	nodeBlock        *models.Block
	generics         []*GenericType
	blockTypes       []*TypeAlias
	blockVars        []*Var
	waitingImpls     []*models.Impl
	eval             *eval
	linked_aliases   []*models.TypeAlias
	linked_functions []*models.Fn
	linked_variables []*models.Var
	linked_structs   []*models.Struct
	allowBuiltin     bool
	package_files    *[]*Parser

	NoLocalPkg  bool
	JustDefines bool
	NoCheck     bool
	IsMain      bool
	Used        []*use // All used packages, deep detection
	Uses        []*use // File uses these packages
	Defines     *models.Defmap
	Errors      []julelog.CompilerLog
	Warnings    []julelog.CompilerLog
	File        *File
}

// New returns new instance of Parser.
func New(path string) *Parser {
	p := new(Parser)
	p.File = lex.NewFile(path)
	p.allowBuiltin = true
	p.Defines = new(models.Defmap)
	p.eval = new(eval)
	p.eval.p = p
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok lex.Token, key string, args ...any) {
	p.pusherrmsgtok(tok, jule.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok lex.Token, msg string) {
	p.Errors = append(p.Errors, julelog.CompilerLog{
		Type:    julelog.ERR,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...julelog.CompilerLog) {
	p.Errors = append(p.Errors, errs...)
}

// PushErr appends new error.
func (p *Parser) PushErr(key string, args ...any) {
	p.pusherrmsg(jule.GetError(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errors = append(p.Errors, julelog.CompilerLog{
		Type:    julelog.FLAT_ERR,
		Message: msg,
	})
}

func getTree(toks []lex.Token) ([]models.Object, []julelog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errors
}

func (p *Parser) checkCppUsePath(use *models.UseDecl) bool {
	if build.IsStdHeaderPath(use.Path) {
		return true
	}
	ext := filepath.Ext(use.Path)
	if !juleapi.IsValidHeader(ext) {
		p.pusherrtok(use.Token, "invalid_header_ext", ext)
		return false
	}
	err := os.Chdir(use.Token.File.Dir())
	if err != nil {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || info.IsDir() {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	// Set to absolute path for correct include path
	use.Path, _ = filepath.Abs(use.Path)
	_ = os.Chdir(jule.WORKING_PATH)
	return true
}

func (p *Parser) checkPureUsePath(use *models.UseDecl) bool {
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Token, "use_not_found", use.Path)
		return false
	}
	return true
}

func (p *Parser) checkUsePath(use *models.UseDecl) bool {
	if use.Cpp {
		if !p.checkCppUsePath(use) {
			return false
		}
	} else {
		if !p.checkPureUsePath(use) {
			return false
		}
	}
	return true
}

func (p *Parser) pushSelects(u *use, selectors []lex.Token) (addNs bool) {
	if len(selectors) > 0 && p.Defines.Side == nil {
		p.Defines.Side = new(models.Defmap)
	}
	for i, id := range selectors {
		for j, jid := range selectors {
			if j >= i {
				break
			} else if jid.Kind == id.Kind {
				p.pusherrtok(id, "exist_id", id.Kind)
				i = -1
				break
			}
		}
		if i == -1 {
			break
		}
		if id.Id == lex.ID_SELF {
			addNs = true
			continue
		}
		i, m, def_t := u.Defines.FindById(id.Kind, p.File)
		if i == -1 {
			p.pusherrtok(id, "id_not_exist", id.Kind)
			continue
		}
		switch def_t {
		case 'i':
			p.Defines.Side.Traits = append(p.Defines.Side.Traits, m.Traits[i])
		case 'f':
			p.Defines.Side.Funcs = append(p.Defines.Side.Funcs, m.Funcs[i])
		case 'e':
			p.Defines.Side.Enums = append(p.Defines.Side.Enums, m.Enums[i])
		case 'g':
			p.Defines.Side.Globals = append(p.Defines.Side.Globals, m.Globals[i])
		case 't':
			p.Defines.Side.Types = append(p.Defines.Side.Types, m.Types[i])
		case 's':
			p.Defines.Side.Structs = append(p.Defines.Side.Structs, m.Structs[i])
		}
	}
	return
}

func (p *Parser) pushUse(use *use, selectors []lex.Token) {
	dm, ok := std_builtin_defines[use.LinkString]
	if ok {
		dm.PushDefines(use.Defines)
	}
	if use.FullUse {
		if p.Defines.Side == nil {
			p.Defines.Side = new(models.Defmap)
		}
		use.Defines.PushDefines(p.Defines.Side)
	} else if len(selectors) > 0 {
		if !p.pushSelects(use, selectors) {
			return
		}
	} else if selectors != nil {
		return
	}
	identifiers := strings.SplitN(use.LinkString, lex.KND_DBLCOLON, -1)
	src := p.pushNs(identifiers)
	src.Defines = use.Defines
}

func (p *Parser) compileCppLinkUse(ast *models.UseDecl) (*use, bool) {
	u := new(use)
	u.CppLink = true
	u.Path = ast.Path
	u.Token = ast.Token
	return u, false
}

func make_use_from_ast(ast *models.UseDecl) *use {
	u := new(use)
	u.Defines = new(models.Defmap)
	u.Token = ast.Token
	u.Path = ast.Path
	u.LinkString = ast.LinkString
	u.FullUse = ast.FullUse
	u.Selectors = ast.Selectors
	return u
}

func (p *Parser) wrap_package() {
	for _, fp := range *p.package_files {
		if p == fp {
			continue
		}
		fp.Defines.PushDefines(p.Defines)
	}
}

func (p *Parser) compilePureUse(ast *models.UseDecl) (_ *use, hassErr bool) {
	infos, err := os.ReadDir(ast.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil, true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SRC_EXT) ||
			!build.IsPassFileAnnotation(name) {
			continue
		}
		path := filepath.Join(ast.Path, name)
		psub := New(path)
		psub.SetupPackage()
		psub.Parsef(false, false)
		psub.wrap_package()
		u := make_use_from_ast(ast)
		psub.Defines.PushDefines(u.Defines)
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		p.pushUse(u, ast.Selectors)
		if psub.Errors != nil {
			p.pusherrtok(ast.Token, "use_has_errors")
			return u, true
		}
		return u, false
	}
	return nil, false
}

func (p *Parser) compileUse(useAST *models.UseDecl) (*use, bool) {
	if useAST.Cpp {
		return p.compileCppLinkUse(useAST)
	}
	return p.compilePureUse(useAST)
}

func (p *Parser) use(ast *models.UseDecl, err *bool) {
	if !p.checkUsePath(ast) {
		*err = true
		return
	}
	// Already parsed?
	for _, u := range p.Used {
		if ast.Path == u.Path {
			old := u.FullUse
			u.FullUse = ast.FullUse
			p.pushUse(u, ast.Selectors)
			p.Uses = append(p.Uses, u)
			u.FullUse = old
			return
		}
	}
	var u *use
	u, *err = p.compileUse(ast)
	if u == nil {
		return
	}
	// Already uses?
	for _, pu := range p.Uses {
		if u.Path == pu.Path {
			p.pusherrtok(ast.Token, "already_uses")
			return
		}
	}
	p.Used = append(p.Used, u)
	p.Uses = append(p.Uses, u)
}

func (p *Parser) parseUses(tree *[]models.Object) bool {
	err := false
	for i := range *tree {
		obj := &(*tree)[i]
		switch obj_t := obj.Data.(type) {
		case models.UseDecl:
			if !err {
				p.use(&obj_t, &err)
			}
			obj.Data = nil
		case models.Comment:
			// Ignore beginning comments.
		default:
			goto end
		}
	}
	*tree = nil
end:
	return err
}

func objectIsIgnored(obj *models.Object) bool {
	return obj.Data == nil
}

func (p *Parser) parseSrcTreeObj(obj models.Object) {
	if objectIsIgnored(&obj) {
		return
	}
	switch obj_t := obj.Data.(type) {
	case models.Statement:
		p.St(obj_t)
	case TypeAlias:
		p.Type(obj_t)
	case []GenericType:
		p.Generics(obj_t)
	case Enum:
		p.Enum(obj_t)
	case Struct:
		p.Struct(obj_t)
	case models.Trait:
		p.Trait(obj_t)
	case models.Impl:
		i := new(models.Impl)
		*i = obj_t
		p.waitingImpls = append(p.waitingImpls, i)
	case models.CppLinkFn:
		p.LinkFn(obj_t)
	case models.CppLinkVar:
		p.LinkVar(obj_t)
	case models.CppLinkStruct:
		p.Link_struct(obj_t)
	case models.CppLinkAlias:
		p.Link_alias(obj_t)
	case models.Comment:
		p.Comment(obj_t)
	case models.UseDecl:
		p.pusherrtok(obj.Token, "use_at_content")
	default:
		p.pusherrtok(obj.Token, "invalid_syntax")
	}
}

func (p *Parser) parseSrcTree(tree []models.Object) {
	for _, obj := range tree {
		p.parseSrcTreeObj(obj)
		p.checkDoc(obj)
		p.checkAttribute(obj)
		p.checkGenerics(obj)
	}
}

func (p *Parser) parseTree(tree []models.Object) (ok bool) {
	if p.parseUses(&tree) {
		return false
	}
	p.parseSrcTree(tree)
	return true
}

func (p *Parser) checkParse() {
	if !p.NoCheck {
		p.check_package()
	}
}

// Special case is;
//
//	p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPackage(tree *[]models.Object) (hasErr bool) {
	if p.File == nil {
		return
	}
	dir := p.File.Dir()
	infos, err := os.ReadDir(dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SRC_EXT) ||
			!build.IsPassFileAnnotation(name) ||
			name == p.File.Name() {
			continue
		}
		fp := New(filepath.Join(dir, name))
		fp.package_files = p.package_files
		*p.package_files = append(*p.package_files, fp)
		fp.NoLocalPkg = true
		fp.NoCheck = true

		fp.Parsef(false, true)
		fp.wg.Wait()
		if len(fp.Errors) > 0 {
			p.pusherrs(fp.Errors...)
			return true
		}
	}
	return
}

func (p *Parser) SetupPackage() {
	p.package_files = new([]*Parser)
	*p.package_files = append(*p.package_files, p)
}

// Parses Jule code from object tree.
func (p *Parser) Parset(tree []models.Object, main, justDefines bool) {
	p.IsMain = main
	p.JustDefines = justDefines
	if !p.parseTree(tree) {
		return
	}
	if !p.NoLocalPkg {
		if p.useLocalPackage(&tree) {
			return
		}
	}
	p.checkParse()
	p.wg.Wait()
}

// Parses Jule code from tokens.
func (p *Parser) Parse(toks []lex.Token, main, justDefines bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		p.pusherrs(errors...)
		return
	}
	p.Parset(tree, main, justDefines)
}

// Parses Jule code from file.
func (p *Parser) Parsef(main, justDefines bool) {
	lexer := lex.New(p.File)
	toks := lexer.Lex()
	if lexer.Logs != nil {
		p.pusherrs(lexer.Logs...)
		return
	}
	p.Parse(toks, main, justDefines)
}

func (p *Parser) checkDoc(obj models.Object) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Data.(type) {
	case models.Comment, models.Attribute, []GenericType:
		return
	}
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj models.Object) {
	if p.attributes == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	p.attributes = nil
}

func (p *Parser) checkGenerics(obj models.Object) {
	if p.generics == nil {
		return
	}
	switch obj.Data.(type) {
	case models.Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Token, "generics_not_supports")
	p.generics = nil
}

// Generics parses generics.
func (p *Parser) Generics(generics []GenericType) {
	for i, generic := range generics {
		if juleapi.IsIgnoreId(generic.Id) {
			p.pusherrtok(generic.Token, "ignore_id")
			continue
		}
		for j, cgeneric := range generics {
			if j >= i {
				break
			} else if generic.Id == cgeneric.Id {
				p.pusherrtok(generic.Token, "exist_id", generic.Id)
				break
			}
		}
		g := new(GenericType)
		*g = generic
		p.generics = append(p.generics, g)
	}
}

func (p *Parser) make_type_alias(alias models.TypeAlias) *models.TypeAlias {
	a := new(models.TypeAlias)
	*a = alias
	alias.Desc = p.docText.String()
	p.docText.Reset()
	return a
}

// Type parses Jule type define statement.
func (p *Parser) Type(alias TypeAlias) {
	if juleapi.IsIgnoreId(alias.Id) {
		p.pusherrtok(alias.Token, "ignore_id")
		return
	}
	_, tok, canshadow := p.defined_by_id(alias.Id)
	if tok.Id != lex.ID_NA && !canshadow {
		p.pusherrtok(alias.Token, "exist_id", alias.Id)
		return
	}
	p.Defines.Types = append(p.Defines.Types, p.make_type_alias(alias))
}

func (p *Parser) parse_enum_items_str(e *Enum) {
	for _, item := range e.Items {
		if juleapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := p.evalExpr(item.Expr, nil)
			if !val.constExpr && !p.eval.has_error {
				p.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				p:         p,
				expr_t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Token,
			}.check()
		} else {
			expr := value{constExpr: true, expr: item.Id}
			item.ExprTag = expr.expr
			item.Expr.Model = strModel(expr)
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.ExprTag = item.ExprTag
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		itemVar.Token = e.Token
		p.Defines.Globals = append(p.Defines.Globals, itemVar)
	}
}

func (p *Parser) parse_enum_items_integer(e *Enum) {
	max := juletype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			p.pusherrtok(item.Token, "overflow_limits")
		} else {
			max--
		}
		if juleapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Token, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Token, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Tokens != nil {
			val, model := p.evalExpr(item.Expr, nil)
			if !val.constExpr && !p.eval.has_error {
				p.pusherrtok(item.Expr.Tokens[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assign_checker{
				p:         p,
				expr_t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Token,
			}.check()
		} else {
			expr := max - (max - uint64(i))
			item.ExprTag = uint64(expr)
			item.Expr.Model = exprNode{strconv.FormatUint(expr, 16)}
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.ExprTag = item.ExprTag
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		itemVar.Token = e.Token
		p.Defines.Globals = append(p.Defines.Globals, itemVar)
	}
}

// Enum parses Jule enumerator statement.
func (p *Parser) Enum(e Enum) {
	if juleapi.IsIgnoreId(e.Id) {
		p.pusherrtok(e.Token, "ignore_id")
		return
	}
	_, tok, _ := p.defined_by_id(e.Id)
	if tok.Id != lex.ID_NA {
		p.pusherrtok(e.Token, "exist_id", e.Id)
		return
	}
	e.Desc = p.docText.String()
	p.docText.Reset()
	e.Type, _ = p.realType(e.Type, true)
	if !types.IsPure(e.Type) {
		p.pusherrtok(e.Token, "invalid_type_source")
		return
	}
	pdefs := p.Defines
	puses := p.Uses
	p.Defines = new(models.Defmap)
	defer func() {
		p.Defines = pdefs
		p.Uses = puses
		p.Defines.Enums = append(p.Defines.Enums, &e)
	}()
	switch {
	case e.Type.Id == juletype.STR:
		p.parse_enum_items_str(&e)
	case juletype.IsInteger(e.Type.Id):
		p.parse_enum_items_integer(&e)
	default:
		p.pusherrtok(e.Token, "invalid_type_source")
	}
}

func (p *Parser) pushField(s *Struct, f **Var, i int) {
	for _, cf := range s.Fields {
		if *f == cf {
			break
		}
		if (*f).Id == cf.Id {
			p.pusherrtok((*f).Token, "exist_id", (*f).Id)
			break
		}
	}
	if len(s.Generics) == 0 {
		p.parseField(s, f, i)
	} else {
		p.parseNonGenericType(s.Generics, &(*f).Type)
		param := models.Param{Id: (*f).Id, Type: (*f).Type}
		param.Default.Model = exprNode{juleapi.DEFAULT_EXPR}
		s.Constructor.Params[i] = param
	}
}

func (p *Parser) parseFields(s *Struct) {
	s.Defines.Globals = make([]*models.Var, len(s.Fields))
	for i, f := range s.Fields {
		p.pushField(s, &f, i)
		s.Defines.Globals[i] = f
	}
}

func make_constructor(s *Struct) *models.Fn {
	constructor := new(models.Fn)
	constructor.Id = s.Id
	constructor.Token = s.Token
	constructor.Params = make([]models.Param, len(s.Fields))
	constructor.RetType.Type = Type{
		Id:    juletype.STRUCT,
		Kind:  s.Id,
		Token: s.Token,
		Tag:   s,
	}
	if len(s.Generics) > 0 {
		constructor.Generics = make([]*models.GenericType, len(s.Generics))
		copy(constructor.Generics, s.Generics)
		constructor.Combines = new([][]models.Type)
	}
	return constructor
}

func (p *Parser) make_struct(model models.Struct) *Struct {
	s := new(Struct)
	*s = model
	s.Desc = p.docText.String()
	p.docText.Reset()
	//s.Traits = new([]*trait)
	//s.depends = new([]*structure)
	s.Owner = p
	s.Generics = p.generics
	p.generics = nil
	s.Attributes = p.attributes
	p.attributes = nil
	s.Defines = new(models.Defmap)
	s.Constructor = make_constructor(s)
	s.Origin = s
	return s
}

// Struct parses Jule structure.
func (p *Parser) Struct(model Struct) {
	if juleapi.IsIgnoreId(model.Id) {
		p.pusherrtok(model.Token, "ignore_id")
		return
	} else if def, _, _ := p.defined_by_id(model.Id); def != nil {
		p.pusherrtok(model.Token, "exist_id", model.Id)
		return
	}
	s := p.make_struct(model)
	p.Defines.Structs = append(p.Defines.Structs, s)
}

// LinkFn parses cpp link function.
func (p *Parser) LinkFn(link models.CppLinkFn) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := p.linkById(link.Link.Id)
	if def_t != ' ' {
		p.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	linkf := link.Link
	linkf.Owner = p
	setGenerics(linkf, p.generics)
	p.generics = nil
	linkf.Attributes = p.attributes
	p.attributes = nil
	p.linked_functions = append(p.linked_functions, linkf)
}

// Link_alias parses cpp link structure.
func (p *Parser) Link_alias(link models.CppLinkAlias) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := p.linkById(link.Link.Id)
	if def_t != ' ' {
		p.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	ta := p.make_type_alias(link.Link)
	p.linked_aliases = append(p.linked_aliases, ta)
}

// Link_struct parses cpp link structure.
func (p *Parser) Link_struct(link models.CppLinkStruct) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := p.linkById(link.Link.Id)
	if def_t != ' ' {
		p.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	s := p.make_struct(link.Link)
	s.CppLinked = true
	p.linked_structs = append(p.linked_structs, s)
}

// LinkVar parses cpp link function.
func (p *Parser) LinkVar(link models.CppLinkVar) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Token, "ignore_id")
		return
	}
	_, def_t := p.linkById(link.Link.Id)
	if def_t != ' ' {
		p.pusherrtok(link.Token, "exist_id", link.Link.Id)
		return
	}
	p.linked_variables = append(p.linked_variables, link.Link)
}

// Trait parses Jule trait.
func (p *Parser) Trait(model models.Trait) {
	if juleapi.IsIgnoreId(model.Id) {
		p.pusherrtok(model.Token, "ignore_id")
		return
	} else if def, _, _ := p.defined_by_id(model.Id); def != nil {
		p.pusherrtok(model.Token, "exist_id", model.Id)
		return
	}
	trait := new(models.Trait)
	*trait = model
	trait.Desc = p.docText.String()
	p.docText.Reset()
	trait.Defines = new(models.Defmap)
	trait.Defines.Funcs = make([]*Fn, len(model.Funcs))
	for i, f := range trait.Funcs {
		if juleapi.IsIgnoreId(f.Id) {
			p.pusherrtok(f.Token, "ignore_id")
		}
		for j, jf := range trait.Funcs {
			if j >= i {
				break
			} else if f.Id == jf.Id {
				p.pusherrtok(f.Token, "exist_id", f.Id)
			}
		}
		_ = p.check_param_dup(f.Params)
		p.parseTypesNonGenerics(f)
		trait.Defines.Funcs[i] = f
	}
	p.Defines.Traits = append(p.Defines.Traits, trait)
}

func (p *Parser) implTrait(model *models.Impl) {
	trait_def, _, _ := p.trait_by_id(model.Base.Kind)
	if trait_def == nil {
		p.pusherrtok(model.Base, "id_not_exist", model.Base.Kind)
		return
	}
	trait_def.Used = true
	sid, _ := model.Target.KindId()
	side := p.Defines.Side
	p.Defines.Side = nil
	s, _, _ := p.struct_by_id(model.Target.Kind)
	p.Defines.Side = side
	if s == nil {
		p.pusherrtok(model.Target.Token, "id_not_exist", sid)
		return
	}
	model.Target.Tag = s
	s.Origin.Traits = append(s.Origin.Traits, trait_def)
	for _, obj := range model.Tree {
		switch obj_t := obj.Data.(type) {
		case models.Comment:
			p.Comment(obj_t)
		case *Fn:
			if trait_def.FindFunc(obj_t.Id) == nil {
				p.pusherrtok(model.Target.Token, "trait_hasnt_id", trait_def.Id, obj_t.Id)
				break
			}
			i, _, _ := s.Defines.FindById(obj_t.Id, nil)
			if i != -1 {
				p.pusherrtok(obj_t.Token, "exist_id", obj_t.Id)
				continue
			}
			obj_t.Receiver.Token = s.Token
			obj_t.Receiver.Tag = s
			obj_t.Attributes = p.attributes
			obj_t.Owner = p
			p.attributes = nil
			obj_t.Desc = p.docText.String()
			p.docText.Reset()
			_ = p.check_param_dup(obj_t.Params)
			p.check_ret_variables(obj_t)
			obj_t.Used = true
			if len(s.Generics) == 0 {
				p.parseTypesNonGenerics(obj_t)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, obj_t)
		}
	}
	for _, tf := range trait_def.Defines.Funcs {
		ok := false
		ds := tf.DefString()
		sf, _, _ := s.Defines.FnById(tf.Id, nil)
		if sf != nil {
			ok = tf.Pub == sf.Pub && ds == sf.DefString()
		}
		if !ok {
			p.pusherrtok(model.Target.Token, "not_impl_trait_def", trait_def.Id, ds)
		}
	}
}

func (p *Parser) implStruct(model *models.Impl) {
	side := p.Defines.Side
	p.Defines.Side = nil
	s, _, _ := p.struct_by_id(model.Base.Kind)
	p.Defines.Side = side
	if s == nil {
		p.pusherrtok(model.Base, "id_not_exist", model.Base.Kind)
		return
	}
	for _, obj := range model.Tree {
		switch obj_t := obj.Data.(type) {
		case []GenericType:
			p.Generics(obj_t)
		case models.Comment:
			p.Comment(obj_t)
		case *Fn:
			i, _, _ := s.Defines.FindById(obj_t.Id, nil)
			if i != -1 {
				p.pusherrtok(obj_t.Token, "exist_id", obj_t.Id)
				continue
			}
			sf := new(Fn)
			*sf = *obj_t
			sf.Receiver.Token = s.Token
			sf.Receiver.Tag = s
			sf.Attributes = p.attributes
			sf.Desc = p.docText.String()
			sf.Owner = p
			p.docText.Reset()
			p.attributes = nil
			setGenerics(sf, p.generics)
			p.generics = nil
			_ = p.check_param_dup(sf.Params)
			p.check_ret_variables(sf)
			for _, generic := range obj_t.Generics {
				if types.FindGeneric(generic.Id, s.Generics) != nil {
					p.pusherrtok(generic.Token, "exist_id", generic.Id)
				}
			}
			if len(s.Generics) == 0 {
				p.parseTypesNonGenerics(sf)
			}
			s.Defines.Funcs = append(s.Defines.Funcs, sf)
		}
	}
}

// Impl parses Jule impl.
func (p *Parser) Impl(impl *models.Impl) {
	if !types.IsVoid(impl.Target) {
		p.implTrait(impl)
		return
	}
	p.implStruct(impl)
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (p *Parser) pushNs(identifiers []string) *models.Namespace {
	var src *models.Namespace
	prev := p.Defines
	for _, id := range identifiers {
		src = prev.NsById(id)
		if src == nil {
			src = new(models.Namespace)
			src.Id = id
			src.Defines = new(models.Defmap)
			prev.Namespaces = append(prev.Namespaces, src)
		}
		prev = src.Defines
	}
	return src
}

// Comment parses Jule documentation comments line.
func (p *Parser) Comment(c models.Comment) {
	if strings.HasPrefix(c.Content, jule.PRAGMA_COMMENT_PREFIX) {
		p.PushAttribute(c)
		return
	}
	p.docText.WriteString(c.Content)
	p.docText.WriteByte('\n')
}

// PushAttribute process and appends to attribute list.
func (p *Parser) PushAttribute(c models.Comment) {
	var attr models.Attribute
	// Skip attribute prefix
	attr.Tag = c.Content[len(jule.PRAGMA_COMMENT_PREFIX):]
	attr.Token = c.Token
	ok := false
	for _, kind := range jule.ATTRS {
		if attr.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		return
	}
	for _, attr2 := range p.attributes {
		if attr.Tag == attr2.Tag {
			return
		}
	}
	p.attributes = append(p.attributes, attr)
}

// St parse Jule statement.
func (p *Parser) St(s models.Statement) {
	switch data_t := s.Data.(type) {
	case Fn:
		p.Func(data_t)
	case Var:
		p.Global(data_t)
	default:
		p.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (p *Parser) parseFuncNonGenericType(generics []*GenericType, dt *Type) {
	f := dt.Tag.(*Fn)
	for i := range f.Params {
		p.parseNonGenericType(generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(generics, &f.RetType.Type)
}

func (p *Parser) parseMultiNonGenericType(generics []*GenericType, dt *Type) {
	types := dt.Tag.([]Type)
	for i := range types {
		mt := &types[i]
		p.parseNonGenericType(generics, mt)
	}
}

func (p *Parser) parseMapNonGenericType(generics []*GenericType, dt *Type) {
	p.parseMultiNonGenericType(generics, dt)
}

func (p *Parser) parseCommonNonGenericType(generics []*GenericType, dt *Type) {
	if dt.Id == juletype.ID {
		id, prefix := dt.KindId()
		def, _, _ := p.defined_by_id(id)
		switch deft := def.(type) {
		case *Struct:
			deft = p.structConstructorInstance(deft)
			if dt.Tag != nil {
				deft.SetGenerics(dt.Tag.([]Type))
			}
			dt.Kind = prefix + deft.AsTypeKind()
			dt.Id = juletype.STRUCT
			dt.Tag = deft
			dt.Pure = true
			dt.Original = nil
			goto tagcheck
		}
	}
	if types.IsGeneric(generics, *dt) {
		return
	}
tagcheck:
	if dt.Tag != nil {
		switch t := dt.Tag.(type) {
		case *Struct:
			for _, ct := range t.GetGenerics() {
				if types.IsGeneric(generics, ct) {
					return
				}
			}
		case []Type:
			for _, ct := range t {
				if types.IsGeneric(generics, ct) {
					return
				}
			}
		}
	}
	p.fn_parse_type(dt)
}

func (p *Parser) parseNonGenericType(generics []*GenericType, dt *Type) {
	switch {
	case dt.MultiTyped:
		p.parseMultiNonGenericType(generics, dt)
	case types.IsFn(*dt):
		p.parseFuncNonGenericType(generics, dt)
	case types.IsMap(*dt):
		p.parseMapNonGenericType(generics, dt)
	case types.IsArray(*dt):
		p.parseNonGenericType(generics, dt.ComponentType)
		dt.Kind = jule.PREFIX_ARRAY + dt.ComponentType.Kind
	case types.IsSlice(*dt):
		p.parseNonGenericType(generics, dt.ComponentType)
		dt.Kind = jule.PREFIX_SLICE + dt.ComponentType.Kind
	default:
		p.parseCommonNonGenericType(generics, dt)
	}
}

func (p *Parser) parseTypesNonGenerics(f *Fn) {
	for i := range f.Params {
		p.parseNonGenericType(f.Generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(f.Generics, &f.RetType.Type)
}

func (p *Parser) check_ret_variables(f *Fn) {
	for i, v := range f.RetType.Identifiers {
		if juleapi.IsIgnoreId(v.Kind) {
			continue
		}
		for _, generic := range f.Generics {
			if v.Kind == generic.Id {
				goto exist
			}
		}
		for _, param := range f.Params {
			if v.Kind == param.Id {
				goto exist
			}
		}
		for j, jv := range f.RetType.Identifiers {
			if j >= i {
				break
			}
			if jv.Kind == v.Kind {
				goto exist
			}
		}
		continue
	exist:
		p.pusherrtok(v, "exist_id", v.Kind)

	}
}

func setGenerics(f *Fn, generics []*models.GenericType) {
	f.Generics = generics
	if len(f.Generics) > 0 {
		f.Combines = new([][]models.Type)
	}
}

// Func parse Jule function.
func (p *Parser) Func(ast Fn) {
	_, tok, canshadow := p.defined_by_id(ast.Id)
	if tok.Id != lex.ID_NA && !canshadow {
		p.pusherrtok(ast.Token, "exist_id", ast.Id)
	} else if juleapi.IsIgnoreId(ast.Id) {
		p.pusherrtok(ast.Token, "ignore_id")
	}
	f := new(Fn)
	*f = ast
	f.Attributes = p.attributes
	p.attributes = nil
	f.Owner = p
	f.Desc = p.docText.String()
	p.docText.Reset()
	setGenerics(f, p.generics)
	p.generics = nil
	p.check_ret_variables(f)
	_ = p.check_param_dup(f.Params)
	f.Used = f.Id == jule.INIT_FN
	p.Defines.Funcs = append(p.Defines.Funcs, f)
}

// ParseVariable parse Jule global variable.
func (p *Parser) Global(vast Var) {
	def, _, _ := p.defined_by_id(vast.Id)
	if def != nil {
		p.pusherrtok(vast.Token, "exist_id", vast.Id)
		return
	} else {
		for _, g := range p.Defines.Globals {
			if vast.Id == g.Id {
				p.pusherrtok(vast.Token, "exist_id", vast.Id)
				return
			}
		}
	}
	vast.Desc = p.docText.String()
	p.docText.Reset()
	v := new(Var)
	*v = vast
	p.Defines.Globals = append(p.Defines.Globals, v)
}

// Var parse Jule variable.
func (p *Parser) Var(model Var) *Var {
	if juleapi.IsIgnoreId(model.Id) {
		p.pusherrtok(model.Token, "ignore_id")
	}
	v := new(Var)
	*v = model
	if v.Type.Id != juletype.VOID {
		vt, ok := p.realType(v.Type, true)
		if ok {
			v.Type = vt
		} else {
			v.Type = models.Type{}
		}
	}
	var val value
	switch tag_t := v.Tag.(type) {
	case value:
		val = tag_t
	default:
		if v.SetterTok.Id != lex.ID_NA {
			val, v.Expr.Model = p.evalExpr(v.Expr, &v.Type)
		}
	}
	if val.data.Type.MultiTyped {
		p.pusherrtok(model.Token, "missing_multi_assign_identifiers")
		return v
	}
	if v.Type.Id != juletype.VOID {
		if v.SetterTok.Id != lex.ID_NA {
			if v.Type.Size.AutoSized && v.Type.Id == juletype.ARRAY {
				v.Type.Size = val.data.Type.Size
			}
			assign_checker{
				p:                p,
				expr_t:           v.Type,
				v:                val,
				errtok:           v.Token,
				not_allow_assign: types.IsRef(v.Type),
			}.check()
		}
	} else {
		if v.SetterTok.Id == lex.ID_NA {
			p.pusherrtok(v.Token, "missing_autotype_value")
		} else {
			p.eval.has_error = p.eval.has_error || val.data.Value == ""
			v.Type = val.data.Type
			p.check_valid_init_expr(v.Mutable, val, v.SetterTok)
			p.checkValidityForAutoType(v.Type, v.SetterTok)
		}
	}
	if !v.IsField && types.IsRef(v.Type) && v.SetterTok.Id == lex.ID_NA {
		p.pusherrtok(v.Token, "reference_not_initialized")
	}
	if !v.IsField && v.SetterTok.Id == lex.ID_NA {
		p.pusherrtok(v.Token, "variable_not_initialized")
	}
	if v.Const {
		v.ExprTag = val.expr
		if !types.IsAllowForConst(v.Type) {
			p.pusherrtok(v.Token, "invalid_type_for_const", v.Type.Kind)
		} else if v.SetterTok.Id != lex.ID_NA && !validExprForConst(val) {
			p.eval.pusherrtok(v.Token, "expr_not_const")
		}
	}
	return v
}

func (p *Parser) varsFromParams(f *Fn) []*Var {
	length := len(f.Params)
	vars := make([]*Var, length)
	for i, param := range f.Params {
		v := new(models.Var)
		v.Owner = f.Block
		v.Mutable = param.Mutable
		v.Id = param.Id
		v.Token = param.Token
		v.Type = param.Type
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Token, "variadic_parameter_not_last")
			}
			v.Type.Original = nil
			v.Type.ComponentType = new(models.Type)
			*v.Type.ComponentType = param.Type
			v.Type.Id = juletype.SLICE
			v.Type.Kind = jule.PREFIX_SLICE + v.Type.Kind
		}
		vars[i] = v
	}
	return vars
}

func (p *Parser) linked_alias_by_id(id string) *models.TypeAlias {
	for _, fp := range *p.package_files {
		for _, link := range fp.linked_aliases {
			if link.Id == id {
				return link
			}
		}
	}
	return nil
}

func (p *Parser) linked_struct_by_id(id string) *Struct {
	for _, fp := range *p.package_files {
		for _, link := range fp.linked_structs {
			if link.Id == id {
				return link
			}
		}
	}
	return nil
}

func (p *Parser) linkedVarById(id string) *Var {
	for _, fp := range *p.package_files {
		for _, link := range fp.linked_variables {
			if link.Id == id {
				return link
			}
		}
	}
	return nil
}

func (p *Parser) linkedFnById(id string) *models.Fn {
	for _, fp := range *p.package_files {
		for _, link := range fp.linked_functions {
			if link.Id == id {
				return link
			}
		}
	}
	return nil
}

// Returns link by identifier.
//
// Types:
//  ' ' -> not found
//  'f' -> function
//  'v' -> variable
//  's' -> struct
//  't' -> type alias
func (p *Parser) linkById(id string) (any, byte) {
	f := p.linkedFnById(id)
	if f != nil {
		return f, 'f'
	}
	v := p.linkedVarById(id)
	if v != nil {
		return v, 'v'
	}
	s := p.linked_struct_by_id(id)
	if s != nil {
		return s, 's'
	}
	ta := p.linked_alias_by_id(id)
	if ta != nil {
		return ta, 't'
	}
	return nil, ' '
}

// fn_by_id returns function by specified id.
//
// Special case:
//
//	fn_by_id(id) -> nil: if function is not exist.
func (p *Parser) fn_by_id(id string) (*Fn, *models.Defmap, bool) {
	if p.allowBuiltin {
		f, _, _ := Builtin.FnById(id, nil)
		if f != nil {
			return f, nil, false
		}
	}
	for _, fp := range *p.package_files {
		f, dm, can_shadow := fp.Defines.FnById(id, fp.File)
		if f != nil && p.is_accessible_define(fp, dm) {
			return f, dm, can_shadow
		}
	}
	return nil, nil, false
}

func (p *Parser) global_by_id(id string) (*Var, *models.Defmap, bool) {
	for _, fp := range *p.package_files {
		g, dm, _ := fp.Defines.GlobalById(id, fp.File)
		if g != nil && p.is_accessible_define(fp, dm) {
			return g, dm, true
		}
	}
	return nil, nil, false
}

func (p *Parser) NsById(id string) *models.Namespace { return p.Defines.NsById(id) }

// Reports identifier is shadowed or not.
func (p *Parser) is_shadowed(id string) bool {
	def, _, _ := p.block_define_by_id(id)
	return def != nil
}

func (p *Parser) is_accessible_define(fp *Parser, dm *models.Defmap) bool {
	// Description of this condition
	// Parameters:
	//   fp: package representer, which package is provides this define
	//   dm: *Defmap of define
	//
	// For example:
	//  Our package has a two file.
	//  These two file uses an another package named as X.
	//  First package file uses like: X::{A, B, C}
	//  Second package file uses like: X::*
	//
	//  In this case, first package file can access all defines of X package.
	//  For this reason, uses the condition below for ignore this case.
	//
	//  The Logic
	//  The current package (p) is not equals to fp, this indicates that
	//  p found the definition from another package file. For this reason
	//  dm should be equals to fp.Defines because otherwise
	//  it means a definition from outside the package. By this logic,
	//  definitions of other package files that they have outside
	//  of the package are ignored.
	//
	// If current package is equals to fp, no problem.
	// Definition founds in p.
	return p == fp || dm == fp.Defines
}

func (p *Parser) type_by_id(id string) (*TypeAlias, *models.Defmap, bool) {
	alias, canshadow := p.block_type_by_id(id)
	if alias != nil {
		return alias, nil, canshadow
	}
	if p.allowBuiltin {
		alias, _, _ = Builtin.TypeById(id, nil)
		if alias != nil {
			return alias, nil, false
		}
	}
	for _, fp := range *p.package_files {
		a, dm, can_shadow := fp.Defines.TypeById(id, fp.File)
		if a != nil && p.is_accessible_define(fp, dm) {
			return a, dm, can_shadow
		}
	}
	return nil, nil, false
}

func (p *Parser) enum_by_id(id string) (*Enum, *models.Defmap, bool) {
	if p.allowBuiltin {
		e, _, _ := Builtin.EnumById(id, nil)
		if e != nil {
			return e, nil, false
		}
	}
	for _, fp := range *p.package_files {
		e, dm, can_shadow := fp.Defines.EnumById(id, fp.File)
		if e != nil && p.is_accessible_define(fp, dm) {
			return e, dm, can_shadow
		}
	}
	return nil, nil, false
}

func (p *Parser) struct_by_id(id string) (*Struct, *models.Defmap, bool) {
	if p.allowBuiltin {
		s, _, _ := Builtin.StructById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	for _, fp := range *p.package_files {
		s, dm, can_shadow := fp.Defines.StructById(id, fp.File)
		if s != nil && p.is_accessible_define(fp, dm) {
			return s, dm, can_shadow
		}
	}
	return nil, nil, false
}

func (p *Parser) trait_by_id(id string) (*models.Trait, *models.Defmap, bool) {
	if p.allowBuiltin {
		trait_def, _, _ := Builtin.TraitById(id, nil)
		if trait_def != nil {
			return trait_def, nil, false
		}
	}
	for _, fp := range *p.package_files {
		t, dm, can_shadow := fp.Defines.TraitById(id, fp.File)
		if t != nil && p.is_accessible_define(fp, dm) {
			return t, dm, can_shadow
		}
	}
	return nil, nil, false
}

func (p *Parser) block_type_by_id(id string) (_ *TypeAlias, can_shadow bool) {
	for i := len(p.blockTypes) - 1; i >= 0; i-- {
		alias := p.blockTypes[i]
		if alias != nil && alias.Id == id {
			return alias, !alias.Generic && alias.Owner != p.nodeBlock
		}
	}
	return nil, false

}

func (p *Parser) block_var_by_id(id string) (_ *Var, can_shadow bool) {
	for i := len(p.blockVars) - 1; i >= 0; i-- {
		v := p.blockVars[i]
		if v != nil && v.Id == id {
			return v, v.Owner != p.nodeBlock
		}
	}
	return nil, false
}

func (p *Parser) defined_by_id(id string) (def any, tok lex.Token, canshadow bool) {
	var a *TypeAlias
	a, _, canshadow = p.type_by_id(id)
	if a != nil {
		return a, a.Token, canshadow
	}
	var e *Enum
	e, _, canshadow = p.enum_by_id(id)
	if e != nil {
		return e, e.Token, canshadow
	}
	var s *Struct
	s, _, canshadow = p.struct_by_id(id)
	if s != nil {
		return s, s.Token, canshadow
	}
	var trait *models.Trait
	trait, _, canshadow = p.trait_by_id(id)
	if trait != nil {
		return trait, trait.Token, canshadow
	}
	var f *Fn
	f, _, canshadow = p.fn_by_id(id)
	if f != nil {
		return f, f.Token, canshadow
	}
	bv, canshadow := p.block_var_by_id(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	g, _, _ := p.global_by_id(id)
	if g != nil {
		return g, g.Token, true
	}
	return
}

func (p *Parser) block_define_by_id(id string) (def any, tok lex.Token, canshadow bool) {
	bv, canshadow := p.block_var_by_id(id)
	if bv != nil {
		return bv, bv.Token, canshadow
	}
	alias, canshadow := p.block_type_by_id(id)
	if alias != nil {
		return alias, alias.Token, canshadow
	}
	return
}

func (p *Parser) precheck_package() {
	p.check_aliases()
	p.parse_package_structs()
	p.parse_package_waiting_fns()
	p.parse_package_waiting_impls()
	p.parse_package_waiting_globals()
	p.check_package_cpp_links()
}

func (p *Parser) parse_package_defines() {
	for _, pf := range *p.package_files {
		pf.parse_defines()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) parse_defines() {
	p.check_structs()
	p.check_fns()
}

func (p *Parser) check_package() {
	if p.IsMain && !p.JustDefines {
		f, _, _ := p.Defines.FnById(jule.ENTRY_POINT, nil)
		if f == nil {
			p.PushErr("no_entry_point")
		} else {
			f.IsEntryPoint = true
			f.Used = true
		}
	}
	p.precheck_package()
	if !p.JustDefines {
		p.parse_package_defines()
	}
}

func (p *Parser) parse_struct(s *Struct) {
	p.parseFields(s)
}

func (p *Parser) parse_package_structs() {
	for _, pf := range *p.package_files {
		pf.parse_structs()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) parse_structs() {
	for _, s := range p.Defines.Structs {
		p.parse_struct(s)
	}
}

func (p *Parser) parse_package_linked_structs() {
	for _, pf := range *p.package_files {
		pf.parse_linked_structs()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) check_package_linked_aliases() {
	for _, pf := range *p.package_files {
		pf.check_linked_aliases()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) check_package_linked_vars() {
	for _, pf := range *p.package_files {
		pf.check_linked_vars()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) check_package_linked_fns() {
	for _, pf := range *p.package_files {
		pf.check_linked_fns()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

func (p *Parser) parse_linked_structs() {
	for _, link := range p.linked_structs {
		p.parse_struct(link)
	}
}

func (p *Parser) check_linked_aliases() {
	for _, link := range p.linked_aliases {
		link.Type, _ = p.realType(link.Type, true)
	}
}

func (p *Parser) check_linked_vars() {
	for _, link := range p.linked_variables {
		vt, ok := p.realType(link.Type, true)
		if ok {
			link.Type = vt
		}
	}
}

func (p *Parser) check_linked_fns() {
	for _, link := range p.linked_functions {
		if len(link.Generics) == 0 {
			p.reload_fn_types(link)
		}
	}
}

func (p *Parser) check_package_cpp_links() {
	p.check_package_linked_aliases()
	p.parse_package_linked_structs()
	p.check_package_linked_vars()
	p.check_package_linked_fns()
}

func (p *Parser) parse_package_waiting_fns() {
	for _, pf := range *p.package_files {
		pf.ParseWaitingFns()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

// ParseWaitingFns parses Jule global functions for waiting to parsing.
func (p *Parser) ParseWaitingFns() {
	for _, f := range p.Defines.Funcs {
		owner := p // f.Ast.Owner.(*Parser) == p
		if len(f.Generics) > 0 {
			owner.parseTypesNonGenerics(f)
		} else {
			owner.reload_fn_types(f)
		}
	}
}

func (p *Parser) check_aliases() {
	for i, alias := range p.Defines.Types {
		p.Defines.Types[i].Type, _ = p.realType(alias.Type, true)
	}
}

func (p *Parser) parse_package_waiting_globals() {
	for _, pf := range *p.package_files {
		pf.ParseWaitingGlobals()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

// ParseWaitingGlobals parses Jule global variables for waiting to parsing.
func (p *Parser) ParseWaitingGlobals() {
	for _, g := range p.Defines.Globals {
		*g = *p.Var(*g)
	}
}

func (p *Parser) parse_package_waiting_impls() {
	for _, pf := range *p.package_files {
		pf.ParseWaitingImpls()
		if p != pf {
			pf.wg.Wait()
			p.pusherrs(pf.Errors...)
		}
	}
}

// ParseWaitingImpls parses Jule impls for waiting to parsing.
func (p *Parser) ParseWaitingImpls() {
	for _, i := range p.waitingImpls {
		p.Impl(i)
	}
	p.waitingImpls = nil
}

func (p *Parser) checkParamDefaultExprWithDefault(param *Param) {
	if types.IsFn(param.Type) {
		p.pusherrtok(param.Token, "invalid_type_for_default_arg", param.Type.Kind)
	}
}

func (p *Parser) checkParamDefaultExpr(f *Fn, param *Param) {
	if !paramHasDefaultArg(param) || param.Token.Id == lex.ID_NA {
		return
	}
	// Skip default argument with default value
	if param.Default.Model != nil {
		if param.Default.Model.String() == juleapi.DEFAULT_EXPR {
			p.checkParamDefaultExprWithDefault(param)
			return
		}
	}
	dt := param.Type
	if param.Variadic {
		dt.Id = juletype.SLICE
		dt.Kind = jule.PREFIX_SLICE + dt.Kind
		dt.ComponentType = new(models.Type)
		*dt.ComponentType = param.Type
		dt.Original = nil
		dt.Pure = true
	}
	v, model := p.evalExpr(param.Default, nil)
	param.Default.Model = model
	p.checkArgType(param, v, param.Token)
}

func (p *Parser) param(f *Fn, param *Param) (err bool) {
	p.checkParamDefaultExpr(f, param)
	return
}

func (p *Parser) check_param_dup(params []models.Param) (err bool) {
	for i, param := range params {
		for j, jparam := range params {
			if j >= i {
				break
			} else if param.Id == jparam.Id {
				err = true
				p.pusherrtok(param.Token, "exist_id", param.Id)
			}
		}
	}
	return
}

func (p *Parser) params(f *Fn) (err bool) {
	hasDefaultArg := false
	for i := range f.Params {
		param := &f.Params[i]
		err = err || p.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Token, "param_must_have_default_arg", param.Id)
			err = true
		}
	}
	return
}

func (p *Parser) block_variables_of_fn(f *Fn) []*Var {
	vars := p.varsFromParams(f)
	vars = append(vars, f.RetType.Vars(f.Block)...)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*Struct)
		vars = append(vars, s.SelfVar(f.Receiver))
	}
	return vars
}

func (p *Parser) parse_pure_fn(f *Fn) (err bool) {
	hasError := p.eval.has_error
	owner := f.Owner.(*Parser)
	err = owner.params(f)
	if err {
		return
	}
	owner.blockVars = owner.block_variables_of_fn(f)
	owner.check_fn(f)
	if owner != p {
		owner.wg.Wait()
		p.pusherrs(owner.Errors...)
		owner.Errors = nil
	}
	owner.blockTypes = nil
	owner.blockVars = nil
	p.eval.has_error = hasError
	return
}

func (p *Parser) parse_fn(f *Fn) (err bool) {
	if len(f.Generics) > 0 {
		return false
	}
	return p.parse_pure_fn(f)
}

func (p *Parser) check_fns() {
	err := false
	check := func(f *Fn) {
		if len(f.Generics) > 0 {
			return
		}
		p.check_fn_special_cases(f)
		if err {
			return
		}
		p.blockTypes = nil
		err = p.parse_fn(f)
	}
	for _, f := range p.Defines.Funcs {
		check(f)
	}
}

func (p *Parser) parseStructFunc(s *Struct, f *Fn) (err bool) {
	if len(f.Generics) > 0 {
		return
	}
	if len(s.Generics) == 0 {
		p.parseTypesNonGenerics(f)
		return p.parse_fn(f)
	}
	return
}

func (p *Parser) checkStruct(xs *Struct) (err bool) {
	for _, f := range xs.Defines.Funcs {
		p.blockTypes = nil
		err = p.parseStructFunc(xs, f)
		if err {
			break
		}
	}
	return
}

func (p *Parser) check_structs() {
	err := false
	check := func(xs *Struct) {
		if err {
			return
		}
		p.checkStruct(xs)
	}
	for _, s := range p.Defines.Structs {
		check(s)
	}
}

func (p *Parser) check_fn_special_cases(f *Fn) {
	switch f.Id {
	case jule.ENTRY_POINT, jule.INIT_FN:
		p.checkSolidFuncSpecialCases(f)
	}
}

func (p *Parser) call_fn(f *Fn, data callData, m *exprModel) value {
	v := p.parse_fn_call_toks(f, data.generics, data.args, m)
	v.lvalue = types.IsLvalue(v.data.Type)
	return v
}

func (p *Parser) callStructConstructor(s *Struct, argsToks []lex.Token, m *exprModel) (v value) {
	f := s.Constructor
	s = f.RetType.Type.Tag.(*Struct)
	v.data.Type = f.RetType.Type.Copy()
	v.data.Type.Kind = s.AsTypeKind()
	v.is_type = false
	v.lvalue = false
	v.constExpr = false
	v.data.Value = s.Id

	// Set braces to parentheses
	argsToks[0].Kind = lex.KND_LPAREN
	argsToks[len(argsToks)-1].Kind = lex.KND_RPARENT

	args := p.get_args(argsToks, true)
	if s.CppLinked {
		m.append_sub(exprNode{lex.KND_LPAREN})
		m.append_sub(exprNode{f.RetType.String()})
		m.append_sub(exprNode{lex.KND_RPARENT})
	} else {
		m.append_sub(exprNode{f.RetType.String()})
	}
	if s.CppLinked {
		m.append_sub(exprNode{lex.KND_LBRACE})
	} else {
		m.append_sub(exprNode{lex.KND_LPAREN})
	}
	p.parseArgs(f, args, m, f.Token)
	if m != nil {
		m.append_sub(argsExpr{args.Src})
	}
	if s.CppLinked {
		m.append_sub(exprNode{lex.KND_RBRACE})
	} else {
		m.append_sub(exprNode{lex.KND_RPARENT})
	}
	return v
}

func (p *Parser) parseField(s *Struct, f **Var, i int) {
	*f = p.Var(**f)
	v := *f
	param := models.Param{Id: v.Id, Type: v.Type}
	if !types.IsPtr(v.Type) && types.IsStruct(v.Type) {
		ts := v.Type.Tag.(*Struct)
		if s.IsSameBase(ts) || ts.IsDependedTo(s) {
			p.pusherrtok(v.Type.Token, "illegal_cycle_in_declaration", s.Id)
		} else {
			s.Origin.Depends = append(s.Origin.Depends, ts)
		}
	}
	if has_expr(v.Expr) {
		param.Default = v.Expr
	} else {
		param.Default.Model = exprNode{juleapi.DEFAULT_EXPR}
	}
	s.Constructor.Params[i] = param
}

func (p *Parser) structConstructorInstance(as *Struct) *Struct {
	s := new(Struct)
	*s = *as
	s.Origin = as
	s.Constructor = new(Fn)
	*s.Constructor = *as.Constructor
	s.Constructor.RetType.Type.Tag = s
	s.Defines = as.Defines
	for i := range s.Defines.Funcs {
		f := &s.Defines.Funcs[i]
		nf := new(Fn)
		*nf = **f
		nf.Receiver.Tag = s
		*f = nf
	}
	return s
}

func (p *Parser) check_anon_fn(f *Fn) {
	_ = p.check_param_dup(f.Params)
	p.check_ret_variables(f)
	p.reload_fn_types(f)
	globals := p.Defines.Globals
	blockVariables := p.blockVars
	p.Defines.Globals = append(blockVariables, p.Defines.Globals...)
	p.blockVars = p.block_variables_of_fn(f)
	rootBlock := p.rootBlock
	nodeBlock := p.nodeBlock
	p.check_fn(f)
	p.rootBlock = rootBlock
	p.nodeBlock = nodeBlock
	p.Defines.Globals = globals
	p.blockVars = blockVariables
}

// Returns nil if has error.
func (p *Parser) get_args(toks []lex.Token, targeting bool) *models.Args {
	toks, _ = p.get_range(lex.KND_LPAREN, lex.KND_RPARENT, toks)
	if toks == nil {
		toks = make([]lex.Token, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks, targeting)
	if len(b.Errors) > 0 {
		p.pusherrs(b.Errors...)
		args = nil
	}
	return args
}

// Should toks include brackets.
func (p *Parser) get_generics(toks []lex.Token) (_ []Type, err bool) {
	if len(toks) == 0 {
		return nil, false
	}
	// Remove braces
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, lex.ID_COMMA, true)
	generics := make([]Type, len(parts))
	p.pusherrs(errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		b := ast.NewBuilder(nil)
		j := 0
		generic, _ := b.DataType(part, &j, true)
		b.Wait()
		if j+1 < len(part) {
			p.pusherrtok(part[j+1], "invalid_syntax")
		}
		p.pusherrs(b.Errors...)
		generics[i] = generic
		ok := p.fn_parse_type(&generics[i])
		if !ok {
			err = true
		}
	}
	return generics, err
}

func (p *Parser) checkGenericsQuantity(required, given int, errTok lex.Token) bool {
	// n = length of required generic type source
	switch {
	case required == 0 && given > 0:
		p.pusherrtok(errTok, "not_has_generics")
		return false
	case required > 0 && given == 0:
		p.pusherrtok(errTok, "has_generics")
		return false
	case required < given:
		p.pusherrtok(errTok, "generics_overflow")
		return false
	case required > given:
		p.pusherrtok(errTok, "missing_generics")
		return false
	default:
		return true
	}
}

func (p *Parser) pushGeneric(generic *GenericType, source Type) {
	alias := &TypeAlias{
		Id:      generic.Id,
		Token:   generic.Token,
		Type:    source,
		Used:    true,
		Generic: true,
	}
	p.blockTypes = append(p.blockTypes, alias)
}

func (p *Parser) pushGenerics(generics []*GenericType, sources []Type) {
	for i, generic := range generics {
		p.pushGeneric(generic, sources[i])
	}
}

func (p *Parser) fn_parse_type(t *Type) bool {
	pt, ok := p.realType(*t, true)
	if ok && types.IsArray(pt) && pt.Size.AutoSized {
		p.pusherrtok(pt.Token, "invalid_type")
		ok = false
	}
	*t = pt
	return ok
}

func (p *Parser) reload_fn_types(f *Fn) {
	for i := range f.Params {
		_ = p.fn_parse_type(&f.Params[i].Type)
	}
	ok := p.fn_parse_type(&f.RetType.Type)
	if ok && types.IsArray(f.RetType.Type) {
		p.pusherrtok(f.RetType.Type.Token, "invalid_type")
	}
}

func itsCombined(f *Fn, generics []Type) bool {
	if f.Combines == nil { // Built-in
		return true
	}
	for _, combine := range *f.Combines {
		for i, gt := range generics {
			ct := combine[i]
			if types.Equals(gt, ct) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) parseGenericFunc(f *Fn, generics []Type, errtok lex.Token) {
	owner := f.Owner.(*Parser)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*Struct)
		owner.pushGenerics(s.Generics, s.GetGenerics())
	}
	owner.reload_fn_types(f)
	if f.Block == nil {
		return
	} else if itsCombined(f, generics) {
		return
	}
	*f.Combines = append(*f.Combines, generics)
	p.parse_pure_fn(f)
}

func (p *Parser) parseGenerics(f *Fn, args *models.Args, errTok lex.Token) bool {
	if len(f.Generics) > 0 && len(args.Generics) == 0 {
		for _, generic := range f.Generics {
			ok := false
			for _, param := range f.Params {
				if types.HasThisGeneric(generic, param.Type) {
					ok = true
					break
				}
			}
			if !ok {
				goto check
			}
		}
		args.DynamicGenericAnnotation = true
		goto ok
	}
check:
	if !p.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok) {
		return false
	} else {
		owner := f.Owner.(*Parser)
		owner.pushGenerics(f.Generics, args.Generics)
		owner.reload_fn_types(f)
	}
ok:
	return true
}

func (p *Parser) parse_fn_call(f *Fn, args *models.Args, m *exprModel, errTok lex.Token) (v value) {
	args.NeedsPureType = p.rootBlock == nil || len(p.rootBlock.Func.Generics) == 0
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		for i := range params {
			param := &params[i]
			fparam := &f.Params[i]
			*param = *fparam
			param.Type = fparam.Type.Copy()
		}
		retType := f.RetType.Type.Copy()
		owner := f.Owner.(*Parser)
		rootBlock := owner.rootBlock
		nodeBlock := owner.nodeBlock
		blockVars := owner.blockVars
		blockTypes := owner.blockTypes
		defer func() {
			owner.rootBlock = rootBlock
			owner.nodeBlock = nodeBlock
			owner.blockVars = blockVars
			owner.blockTypes = blockTypes

			// Remember generics
			for i := range params {
				params[i].Type.Generic = f.Params[i].Type.Generic
			}
			retType.Generic = f.RetType.Type.Generic

			f.Params = params
			f.RetType.Type = retType
		}()
		if !p.parseGenerics(f, args, errTok) {
			return
		}
	} else {
		_ = p.checkGenericsQuantity(len(f.Generics), len(args.Generics), errTok)
		if f.Receiver != nil {
			switch f.Receiver.Tag.(type) {
			case *Struct:
				owner := f.Owner.(*Parser)
				s := f.Receiver.Tag.(*Struct)
				generics := s.GetGenerics()
				if len(generics) > 0 {
					owner.pushGenerics(s.Generics, generics)
					owner.reload_fn_types(f)
				}
			}
		}
	}
	if args == nil {
		goto end
	}
	p.parseArgs(f, args, m, errTok)
	if len(args.Generics) > 0 {
		p.parseGenericFunc(f, args.Generics, errTok)
	}
	if m != nil {
		m.append_sub(callExpr{
			generics: genericsExpr{args.Generics},
			args:     argsExpr{args.Src},
			f:        f,
		})
	}
end:
	v.data.Value = " "
	v.data.Type = f.RetType.Type.Copy()
	if args.NeedsPureType {
		v.data.Type.Pure = true
		v.data.Type.Original = nil
	}
	return
}

func (p *Parser) parse_fn_call_toks(f *Fn, genericsToks, argsToks []lex.Token, m *exprModel) (v value) {
	var generics []Type
	var args *models.Args
	var err bool
	generics, err = p.get_generics(genericsToks)
	if err {
		p.eval.has_error = true
		return
	}
	args = p.get_args(argsToks, false)
	args.Generics = generics
	return p.parse_fn_call(f, args, m, argsToks[0])
}

func (p *Parser) parseStructArgs(f *Fn, args *models.Args, errTok lex.Token) {
	sap := structArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	sap.parse()
}

func (p *Parser) parsePureArgs(f *Fn, args *models.Args, m *exprModel, errTok lex.Token) {
	pap := pureArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func (p *Parser) parseArgs(f *Fn, args *models.Args, m *exprModel, errTok lex.Token) {
	if args.Targeted {
		p.parseStructArgs(f, args, errTok)
		return
	}
	p.parsePureArgs(f, args, m, errTok)
}

func has_expr(expr Expr) bool {
	return len(expr.Tokens) > 0 || expr.Model != nil
}

func paramHasDefaultArg(param *Param) bool {
	return has_expr(param.Default)
}

// [identifier]
type paramMap map[string]*paramMapPair
type paramMapPair struct {
	param *Param
	arg   *Arg
}

func (p *Parser) pushGenericByFunc(f *Fn, pair *paramMapPair, args *models.Args, gt Type) bool {
	tf := gt.Tag.(*Fn)
	cf := pair.param.Type.Tag.(*Fn)
	if len(tf.Params) != len(cf.Params) {
		return false
	}
	for i, param := range tf.Params {
		pair := *pair
		pair.param = &cf.Params[i]
		ok := p.pushGenericByArg(f, &pair, args, param.Type)
		if !ok {
			return ok
		}
	}
	{
		pair := *pair
		pair.param = &models.Param{
			Type: cf.RetType.Type,
		}
		return p.pushGenericByArg(f, &pair, args, tf.RetType.Type)
	}
}

func (p *Parser) pushGenericByMultiTyped(f *Fn, pair *paramMapPair, args *models.Args, gt Type) bool {
	_types := gt.Tag.([]Type)
	for _, mt := range _types {
		for _, generic := range f.Generics {
			if types.HasThisGeneric(generic, pair.param.Type) {
				p.pushGenericByType(f, generic, args, mt)
				break
			}
		}
	}
	return true
}

func (p *Parser) pushGenericByCommonArg(f *Fn, pair *paramMapPair, args *models.Args, t Type) bool {
	for _, generic := range f.Generics {
		if types.IsThisGeneric(generic, pair.param.Type) {
			p.pushGenericByType(f, generic, args, t)
			return true
		}
	}
	return false
}

func (p *Parser) pushGenericByType(f *Fn, generic *GenericType, args *models.Args, gt Type) {
	owner := f.Owner.(*Parser)
	// Already added
	alias, _ := owner.block_type_by_id(generic.Id)
	if alias != nil {
		return
	}
	id, _ := gt.KindId()
	gt.Kind = id
	f.Owner.(*Parser).pushGeneric(generic, gt)
	args.Generics = append(args.Generics, gt)
}

func (p *Parser) pushGenericByComponent(f *Fn, pair *paramMapPair, args *models.Args, argType Type) bool {
	for argType.ComponentType != nil {
		argType = *argType.ComponentType
	}
	return p.pushGenericByCommonArg(f, pair, args, argType)
}

func (p *Parser) pushGenericByArg(f *Fn, pair *paramMapPair, args *models.Args, argType Type) bool {
	_, prefix := pair.param.Type.KindId()
	_, tprefix := argType.KindId()
	if prefix != tprefix {
		return false
	}
	switch {
	case types.IsFn(argType):
		return p.pushGenericByFunc(f, pair, args, argType)
	case argType.MultiTyped, types.IsMap(argType):
		return p.pushGenericByMultiTyped(f, pair, args, argType)
	case types.IsArray(argType), types.IsSlice(argType):
		return p.pushGenericByComponent(f, pair, args, argType)
	default:
		return p.pushGenericByCommonArg(f, pair, args, argType)
	}
}

func (p *Parser) parseArg(f *Fn, pair *paramMapPair, args *models.Args, variadiced *bool) {
	var value value
	var model iExpr
	if pair.param.Variadic {
		t := types.VariadicToSlice(pair.param.Type)
		value, model = p.evalExpr(pair.arg.Expr, &t)
	} else {
		value, model = p.evalExpr(pair.arg.Expr, &pair.param.Type)
	}
	pair.arg.Expr.Model = model
	if !value.variadic && !pair.param.Variadic &&
		!models.Has_attribute(jule.ATTR_CDEF, f.Attributes) &&
		types.IsPure(pair.param.Type) && juletype.IsNumeric(pair.param.Type.Id) {
		pair.arg.CastType = new(Type)
		*pair.arg.CastType = pair.param.Type.Copy()
		pair.arg.CastType.Original = nil
		pair.arg.CastType.Pure = true
	}
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	if args.DynamicGenericAnnotation &&
		types.HasGenerics(f.Generics, pair.param.Type) {
		ok := p.pushGenericByArg(f, pair, args, value.data.Type)
		if !ok {
			p.pusherrtok(pair.arg.Token, "dynamic_type_annotation_failed")
		}
		return
	}
	p.checkArgType(pair.param, value, pair.arg.Token)
}

func (p *Parser) checkArgType(param *Param, val value, errTok lex.Token) {
	p.check_valid_init_expr(param.Mutable, val, errTok)
	assign_checker{
		p:      p,
		expr_t: param.Type,
		v:      val,
		errtok: errTok,
	}.check()
}

// get_range returns between of brackets.
//
// Special case is:
//
//	get_range(open, close, tokens) = nil, false if fail
func (p *Parser) get_range(open, close string, toks []lex.Token) (_ []lex.Token, ok bool) {
	i := 0
	toks = ast.Range(&i, open, close, toks)
	return toks, toks != nil
}

func (p *Parser) checkSolidFuncSpecialCases(f *Fn) {
	if len(f.Params) > 0 {
		p.pusherrtok(f.Token, "fn_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != juletype.VOID {
		p.pusherrtok(f.RetType.Type.Token, "fn_have_ret", f.Id)
	}
	f.Attributes = nil
	if f.IsUnsafe {
		p.pusherrtok(f.Token, "fn_is_unsafe", f.Id)
	}
}

func (p *Parser) checkNewBlockCustom(b *models.Block, oldBlockVars []*Var) {
	b.Gotos = new(models.Gotos)
	b.Labels = new(models.Labels)
	if p.rootBlock == nil {
		p.rootBlock = b
		p.nodeBlock = b
		defer func() {
			p.checkLabelNGoto()
			p.rootBlock = nil
			p.nodeBlock = nil
		}()
	} else {
		b.Parent = p.nodeBlock
		b.SubIndex = p.nodeBlock.SubIndex + 1
		b.Func = p.nodeBlock.Func
		oldNode := p.nodeBlock
		old_unsafe := b.IsUnsafe
		b.IsUnsafe = b.IsUnsafe || oldNode.IsUnsafe
		p.nodeBlock = b
		defer func() {
			p.nodeBlock = oldNode
			b.IsUnsafe = old_unsafe
			*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, *b.Gotos...)
			*p.rootBlock.Labels = append(*p.rootBlock.Labels, *b.Labels...)
		}()
	}
	blockTypes := p.blockTypes
	p.checkBlock(b)

	vars := p.blockVars[len(oldBlockVars):]
	aliases := p.blockTypes[len(blockTypes):]
	for _, v := range vars {
		if !v.Used {
			p.pusherrtok(v.Token, "declared_but_not_used", v.Id)
		}
	}
	for _, a := range aliases {
		if !a.Used {
			p.pusherrtok(a.Token, "declared_but_not_used", a.Id)
		}
	}
	p.blockVars = oldBlockVars
	p.blockTypes = blockTypes
}

func (p *Parser) checkNewBlock(b *models.Block) {
	p.checkNewBlockCustom(b, p.blockVars)
}

func (p *Parser) st(s *models.Statement, recover bool) bool {
	switch data := s.Data.(type) {
	case models.ExprStatement:
		p.expr_st(&data, recover)
		s.Data = data
	case Var:
		p.var_st(&data, false)
		s.Data = data
	case models.Assign:
		p.assign(&data)
		s.Data = data
	case models.Break:
		p.break_st(&data)
		s.Data = data
	case models.Continue:
		p.continue_st(&data)
		s.Data = data
	case *models.Match:
		p.matchcase(data)
	case TypeAlias:
		def, _, canshadow := p.block_define_by_id(data.Id)
		if def != nil && !canshadow {
			p.pusherrtok(data.Token, "exist_id", data.Id)
			break
		} else if juleapi.IsIgnoreId(data.Id) {
			p.pusherrtok(data.Token, "ignore_id")
			break
		}
		data.Type, _ = p.realType(data.Type, true)
		p.blockTypes = append(p.blockTypes, &data)
	case *models.Block:
		p.checkNewBlock(data)
		s.Data = data
	case models.ConcurrentCall:
		p.concurrentCall(&data)
		s.Data = data
	case models.Comment:
	default:
		return false
	}
	return true
}

func (p *Parser) fallthrough_st(f *models.Fallthrough, b *models.Block, i *int) {
	switch {
	case p.currentCase == nil || *i+1 < len(b.Tree):
		p.pusherrtok(f.Token, "fallthrough_wrong_use")
		return
	case p.currentCase.Next == nil:
		p.pusherrtok(f.Token, "fallthrough_into_final_case")
		return
	}
	f.Case = p.currentCase
}

func (p *Parser) check_st(b *models.Block, i *int) {
	s := &b.Tree[*i]
	if p.st(s, true) {
		return
	}
	switch data := s.Data.(type) {
	case models.Iter:
		data.Parent = b
		s.Data = data
		p.iter(&data)
		s.Data = data
	case models.Fallthrough:
		p.fallthrough_st(&data, b, i)
		s.Data = data
	case models.Conditional:
		p.conditional(&data)
		s.Data = data
	case models.Ret:
		rc := retChecker{t: p, ret_ast: &data, f: b.Func}
		rc.check()
		s.Data = data
	case models.Goto:
		obj := new(models.Goto)
		*obj = data
		obj.Index = *i
		obj.Block = b
		*b.Gotos = append(*b.Gotos, obj)
	case models.Label:
		if find_label_parent(data.Label, b) != nil {
			p.pusherrtok(data.Token, "label_exist", data.Label)
			break
		}
		obj := new(models.Label)
		*obj = data
		obj.Index = *i
		obj.Block = b
		*b.Labels = append(*b.Labels, obj)
	default:
		p.pusherrtok(s.Token, "invalid_syntax")
	}
}

func (p *Parser) checkBlock(b *models.Block) {
	for i := 0; i < len(b.Tree); i++ {
		p.check_st(b, &i)
	}
}

func (p *Parser) recoverFuncExprSt(s *models.ExprStatement) {
	errtok := s.Expr.Tokens[0]
	callToks := s.Expr.Tokens[1:]
	args := p.get_args(callToks, false)
	handleParam := recoverFunc.Params[0]
	if len(args.Src) == 0 {
		p.pusherrtok(errtok, "missing_expr_for", handleParam.Id)
		return
	} else if len(args.Src) > 1 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	v, _ := p.evalExpr(args.Src[0].Expr, nil)
	if v.data.Type.Kind != handleParam.Type.Kind {
		p.eval.pusherrtok(errtok, "incompatible_types",
			handleParam.Type.Kind, v.data.Type.Kind)
		return
	}
	handler := v.data.Type.Tag.(*Fn)
	s.Expr.Model = exprNode{"try{\n"}
	var catcher serieExpr
	catcher.exprs = append(catcher.exprs, "} catch(trait<JULEC_ID(Error)> ")
	catcher.exprs = append(catcher.exprs, handler.Params[0].OutId())
	catcher.exprs = append(catcher.exprs, ") ")
	r, _ := utf8.DecodeRuneInString(v.data.Value)
	if r == '_' || lex.IsLetter(r) { // Function source
		catcher.exprs = append(catcher.exprs, "{")
		catcher.exprs = append(catcher.exprs, handler.OutId())
		catcher.exprs = append(catcher.exprs, "(")
		catcher.exprs = append(catcher.exprs, handler.Params[0].OutId())
		catcher.exprs = append(catcher.exprs, "); }")
	} else {
		catcher.exprs = append(catcher.exprs, handler.Block)
	}
	catchExpr := models.Statement{
		Data: models.ExprStatement{
			Expr: models.Expr{Model: catcher},
		},
	}
	p.nodeBlock.Tree = append(p.nodeBlock.Tree, catchExpr)
}

func (p *Parser) expr_st(s *models.ExprStatement, recover bool) {
	if s.Expr.IsNotBinop() {
		expr := s.Expr.Op.(models.BinopExpr)
		tok := expr.Tokens[0]
		if tok.Id == lex.ID_IDENT && tok.Kind == recoverFunc.Id {
			if ast.IsFnCall(s.Expr.Tokens) != nil {
				if !recover {
					p.pusherrtok(tok, "invalid_syntax")
				}
				def, _, _ := p.defined_by_id(tok.Kind)
				if def == recoverFunc {
					p.recoverFuncExprSt(s)
					return
				}
			}
		}
	}
	if s.Expr.Model == nil {
		_, s.Expr.Model = p.evalExpr(s.Expr, nil)
	}
}

func (p *Parser) parseCase(c *models.Case, expr_t Type) {
	for i := range c.Exprs {
		expr := &c.Exprs[i]
		value, model := p.evalExpr(*expr, nil)
		expr.Model = model
		assign_checker{
			p:      p,
			expr_t: expr_t,
			v:      value,
			errtok: expr.Tokens[0],
		}.check()
	}
	oldCase := p.currentCase
	p.currentCase = c
	p.checkNewBlock(c.Block)
	p.currentCase = oldCase
}

func (p *Parser) cases(m *models.Match, expr_t Type) {
	for i := range m.Cases {
		p.parseCase(&m.Cases[i], expr_t)
	}
}

func (p *Parser) matchcase(m *models.Match) {
	if !m.Expr.IsEmpty() {
		value, expr_model := p.evalExpr(m.Expr, nil)
		m.Expr.Model = expr_model
		m.ExprType = value.data.Type
	} else {
		m.ExprType.Id = juletype.BOOL
		m.ExprType.Kind = juletype.TYPE_MAP[m.ExprType.Id]
	}
	p.cases(m, m.ExprType)
	if m.Default != nil {
		p.parseCase(m.Default, m.ExprType)
	}
}

func find_label(id string, b *models.Block) *models.Label {
	for _, label := range *b.Labels {
		if label.Label == id {
			return label
		}
	}
	return nil
}

func (p *Parser) checkLabels() {
	labels := p.rootBlock.Labels
	for _, label := range *labels {
		if !label.Used {
			p.pusherrtok(label.Token, "declared_but_not_used", label.Label+":")
		}
	}
}

func stIsDef(s *models.Statement) bool {
	switch t := s.Data.(type) {
	case Var:
		return true
	case models.Assign:
		for _, selector := range t.Left {
			if selector.Var.New {
				return true
			}
		}
	}
	return false
}

func (p *Parser) checkSameScopeGoto(gt *models.Goto, label *models.Label) {
	if label.Index < gt.Index { // Label at above.
		return
	}
	for i := label.Index; i > gt.Index; i-- {
		s := &label.Block.Tree[i]
		if stIsDef(s) {
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
}

func (p *Parser) checkLabelParents(gt *models.Goto, label *models.Label) bool {
	block := label.Block
parent_scopes:
	if block.Parent != nil && block.Parent != gt.Block {
		block = block.Parent
		for i := 0; i < len(block.Tree); i++ {
			s := &block.Tree[i]
			switch {
			case s.Token.Row >= label.Token.Row:
				return true
			case stIsDef(s):
				p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
				return false
			}
		}
		goto parent_scopes
	}
	return true
}

func (p *Parser) checkGotoScope(gt *models.Goto, label *models.Label) {
	for i := gt.Index; i < len(gt.Block.Tree); i++ {
		s := &gt.Block.Tree[i]
		switch {
		case s.Token.Row >= label.Token.Row:
			return
		case stIsDef(s):
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			return
		}
	}
}

func (p *Parser) checkDiffScopeGoto(gt *models.Goto, label *models.Label) {
	switch {
	case label.Block.SubIndex > 0 && gt.Block.SubIndex == 0:
		if !p.checkLabelParents(gt, label) {
			return
		}
	case label.Block.SubIndex < gt.Block.SubIndex: // Label at parent blocks.
		return
	}
	block := label.Block
	for i := label.Index - 1; i >= 0; i-- {
		s := &block.Tree[i]
		switch s.Data.(type) {
		case models.Block:
			if s.Token.Row <= gt.Token.Row {
				return
			}
		}
		if stIsDef(s) {
			p.pusherrtok(gt.Token, "goto_jumps_declarations", gt.Label)
			break
		}
	}
	// Parent Scopes
	if block.Parent != nil && block.Parent != gt.Block {
		_ = p.checkLabelParents(gt, label)
	} else { // goto Scope
		p.checkGotoScope(gt, label)
	}
}

func (p *Parser) checkGoto(gt *models.Goto, label *models.Label) {
	switch {
	case gt.Block == label.Block:
		p.checkSameScopeGoto(gt, label)
	case label.Block.SubIndex > 0:
		p.checkDiffScopeGoto(gt, label)
	}
}

func (p *Parser) checkGotos() {
	for _, gt := range *p.rootBlock.Gotos {
		label := find_label(gt.Label, p.rootBlock)
		if label == nil {
			p.pusherrtok(gt.Token, "label_not_exist", gt.Label)
			continue
		}
		label.Used = true
		p.checkGoto(gt, label)
	}
}

func (p *Parser) checkLabelNGoto() {
	p.checkGotos()
	p.checkLabels()
}

func matchHasRet(m *models.Match) (ok bool) {
	if m.Default == nil {
		return
	}
	ok = true
	fall := false
	for _, c := range m.Cases {
		falled := fall
		ok, fall = hasRet(c.Block)
		if falled && !ok && !fall {
			return false
		}
		switch {
		case !ok:
			if !fall {
				return false
			}
			fallthrough
		case fall:
			if c.Next == nil {
				return false
			}
			continue
		}
		fall = false
	}
	ok, _ = hasRet(m.Default.Block)
	return ok
}

func hasRet(b *models.Block) (ok bool, fall bool) {
	if b == nil {
		return false, false
	}
	for _, s := range b.Tree {
		switch t := s.Data.(type) {
		case *models.Block:
			ok, fall = hasRet(t)
			if ok {
				return true, fall
			}
		case models.Fallthrough:
			fall = true
		case models.Ret:
			return true, fall
		case *models.Match:
			if matchHasRet(t) {
				return true, false
			}
		}
	}
	return false, fall
}

func (p *Parser) checkRets(f *Fn) {
	ok, _ := hasRet(f.Block)
	if ok {
		return
	}
	if !types.IsVoid(f.RetType.Type) {
		p.pusherrtok(f.Token, "missing_ret")
	}
}

func (p *Parser) check_fn(f *Fn) {
	if f.Block == nil || f.Block.Tree == nil {
		goto always
	} else {
		rootBlock := p.rootBlock
		nodeBlock := p.nodeBlock
		p.rootBlock = nil
		p.nodeBlock = nil
		f.Block.Func = f
		p.checkNewBlock(f.Block)
		p.rootBlock = rootBlock
		p.nodeBlock = nodeBlock
	}
always:
	p.checkRets(f)
}

func (p *Parser) var_st(v *Var, noParse bool) {
	def, _, canshadow := p.block_define_by_id(v.Id)
	if !canshadow && def != nil {
		p.pusherrtok(v.Token, "exist_id", v.Id)
		return
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	p.blockVars = append(p.blockVars, v)
}

func (p *Parser) concurrentCall(cc *models.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, cc.Expr.Model = p.evalExpr(cc.Expr, nil)
}

func (p *Parser) assignment(left value, errtok lex.Token) bool {
	state := true
	if !left.lvalue {
		p.eval.pusherrtok(errtok, "assign_require_lvalue")
		state = false
	}
	if left.constExpr {
		p.pusherrtok(errtok, "assign_const")
		state = false
	} else if !left.mutable {
		p.pusherrtok(errtok, "assignment_to_non_mut")
	}
	switch left.data.Type.Tag.(type) {
	case Fn:
		f, _, _ := p.fn_by_id(left.data.Token.Kind)
		if f != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) singleAssign(assign *models.Assign, l, r []value) {
	left := l[0]
	switch {
	case juleapi.IsIgnoreId(left.data.Value):
		return
	case !p.assignment(left, assign.Setter):
		return
	}
	right := r[0]
	if assign.Setter.Kind != lex.KND_EQ && !lex.IsLiteral(right.data.Value) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:         p,
			l:  left,
			r: right,
			op:  assign.Setter,
		}
		right = solver.solve()
		assign.Setter.Kind += lex.KND_EQ
	}
	assign_checker{
		p:      p,
		expr_t:      left.data.Type,
		v:      right,
		errtok: assign.Setter,
	}.check()
}

func (p *Parser) assignExprs(vsAST *models.Assign) (l []value, r []value) {
	l = make([]value, len(vsAST.Left))
	r = make([]value, len(vsAST.Right))
	n := len(l)
	if n < len(r) {
		n = len(r)
	}
	for i := 0; i < n; i++ {
		var r_type *Type = nil
		if i < len(l) {
			left := &vsAST.Left[i]
			if !left.Var.New && !(len(left.Expr.Tokens) == 1 &&
				juleapi.IsIgnoreId(left.Expr.Tokens[0].Kind)) {
				v, model := p.evalExpr(left.Expr, nil)
				left.Expr.Model = model
				l[i] = v
				r_type = &v.data.Type
			} else {
				l[i].data.Value = juleapi.IGNORE
			}
		}
		if i < len(r) {
			left := &vsAST.Right[i]
			v, model := p.evalExpr(*left, r_type)
			left.Model = model
			r[i] = v
		}
	}
	return
}

func (p *Parser) funcMultiAssign(vsAST *models.Assign, l, r []value) {
	types := r[0].data.Type.Tag.([]Type)
	if len(types) > len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "missing_multi_assign_identifiers")
		return
	} else if len(types) < len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "overflow_multi_assign_identifiers")
		return
	}
	rights := make([]value, len(types))
	for i, t := range types {
		rights[i] = value{data: models.Data{Token: t.Token, Type: t}}
	}
	p.multiAssign(vsAST, l, rights)
}

func (p *Parser) check_valid_init_expr(left_mutable bool, right value, errtok lex.Token) {
	if p.unsafe_allowed() || !lex.IsIdentifierRune(right.data.Value) {
		return
	}
	if left_mutable && !right.mutable && types.IsMut(right.data.Type) {
		p.pusherrtok(errtok, "assignment_non_mut_to_mut")
		return
	}
	checker := assign_checker{
		p:      p,
		v:      right,
		errtok: errtok,
	}
	_ = checker.check_validity()
}

func (p *Parser) multiAssign(assign *models.Assign, l, r []value) {
	for i := range assign.Left {
		left := &assign.Left[i]
		left.Ignore = juleapi.IsIgnoreId(left.Var.Id)
		right := r[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr := l[i]
			if !p.assignment(leftExpr, assign.Setter) {
				return
			}
			p.check_valid_init_expr(leftExpr.mutable, right, assign.Setter)
			assign_checker{
				p:      p,
				expr_t:      leftExpr.data.Type,
				v:      right,
				errtok: assign.Setter,
			}.check()
			continue
		}
		left.Var.Tag = right
		p.var_st(&left.Var, false)
	}
}

func (p *Parser) unsafe_allowed() bool {
	return (p.rootBlock != nil && p.rootBlock.IsUnsafe) ||
		(p.nodeBlock != nil && p.nodeBlock.IsUnsafe)
}

func (p *Parser) postfix(assign *models.Assign, l, r []value) {
	if len(r) > 0 {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	left := l[0]
	_ = p.assignment(left, assign.Setter)
	if types.IsExplicitPtr(left.data.Type) {
		if !p.unsafe_allowed() {
			p.pusherrtok(assign.Left[0].Expr.Tokens[0], "unsafe_behavior_at_out_of_unsafe_scope")
		}
		return
	}
	checkType := left.data.Type
	if types.IsRef(checkType) {
		checkType = types.DerefPtrOrRef(checkType)
	}
	if types.IsPure(checkType) && juletype.IsNumeric(checkType.Id) {
		return
	}
	p.pusherrtok(assign.Setter, "operator_not_for_juletype", assign.Setter.Kind, left.data.Type.Kind)
}

func (p *Parser) assign(assign *models.Assign) {
	ln := len(assign.Left)
	rn := len(assign.Right)
	l, r := p.assignExprs(assign)
	switch {
	case rn == 0 && ast.IsPostfixOp(assign.Setter.Kind):
		p.postfix(assign, l, r)
		return
	case ln == 1 && !assign.Left[0].Var.New:
		p.singleAssign(assign, l, r)
		return
	case assign.Setter.Kind != lex.KND_EQ:
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	case rn == 1:
		right := r[0]
		if right.data.Type.MultiTyped {
			assign.MultipleRet = true
			p.funcMultiAssign(assign, l, r)
			return
		}
	}
	switch {
	case ln > rn:
		p.pusherrtok(assign.Setter, "overflow_multi_assign_identifiers")
		return
	case ln < rn:
		p.pusherrtok(assign.Setter, "missing_multi_assign_identifiers")
		return
	}
	p.multiAssign(assign, l, r)
}

func (p *Parser) whileProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterWhile)
	val, model := p.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	iter.Profile = profile
	if !p.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		p.pusherrtok(iter.Token, "iter_while_require_bool_expr")
	}
	if profile.Next.Data != nil {
		_ = p.st(&profile.Next, false)
	}
	p.checkNewBlock(iter.Block)
}

func (p *Parser) foreachProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterForeach)
	val, model := p.evalExpr(profile.Expr, nil)
	profile.Expr.Model = model
	profile.ExprType = val.data.Type
	if !p.eval.has_error && val.data.Value != "" && !isForeachIterExpr(val) {
		p.pusherrtok(iter.Token, "iter_foreach_require_enumerable_expr")
	} else {
		fc := foreachChecker{p, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := p.blockVars
	if !juleapi.IsIgnoreId(profile.KeyA.Id) {
		p.blockVars = append(p.blockVars, &profile.KeyA)
	}
	if !juleapi.IsIgnoreId(profile.KeyB.Id) {
		p.blockVars = append(p.blockVars, &profile.KeyB)
	}
	p.checkNewBlockCustom(iter.Block, blockVars)
}

func (p *Parser) iter(iter *models.Iter) {
	oldCase := p.currentCase
	oldIter := p.currentIter
	p.currentCase = nil
	p.currentIter = iter
	switch iter.Profile.(type) {
	case models.IterWhile:
		p.whileProfile(iter)
	case models.IterForeach:
		p.foreachProfile(iter)
	default:
		p.checkNewBlock(iter.Block)
	}
	p.currentCase = oldCase
	p.currentIter = oldIter
}

func (p *Parser) conditional_node(node *models.If) {
	val, model := p.evalExpr(node.Expr, nil)
	node.Expr.Model = model
	if !p.eval.has_error && val.data.Value != "" && !isBoolExpr(val) {
		p.pusherrtok(node.Token, "if_require_bool_expr")
	}
	p.checkNewBlock(node.Block)
}

func (p *Parser) conditional(model *models.Conditional) {
	p.conditional_node(model.If)
	for _, elif := range model.Elifs {
		p.conditional_node(elif)
	}
	if model.Default != nil {
		p.checkNewBlock(model.Default.Block)
	}
}

func find_label_parent(id string, b *models.Block) *models.Label {
	label := find_label(id, b)
	for label == nil {
		if b.Parent == nil {
			return nil
		}
		b = b.Parent
		label = find_label(id, b)
	}
	return label
}

func (p *Parser) breakWithLabel(ast *models.Break) {
	if p.currentIter == nil && p.currentCase == nil {
		p.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
		return
	}
	var label *models.Label
	switch {
	case p.currentCase != nil && p.currentIter != nil:
		if p.currentCase.Block.Parent.SubIndex < p.currentIter.Parent.SubIndex {
			label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
			}
		} else {
			label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
			if label == nil {
				label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
			}
		}
	case p.currentCase != nil:
		label = find_label_parent(ast.LabelToken.Kind, p.currentCase.Block.Parent)
	case p.currentIter != nil:
		label = find_label_parent(ast.LabelToken.Kind, p.currentIter.Parent)
	}
	if label == nil {
		p.pusherrtok(ast.LabelToken, "label_not_exist", ast.LabelToken.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		p.pusherrtok(ast.LabelToken, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch data := obj.Data.(type) {
		case models.Comment:
			continue
		case *models.Match:
			label.Used = true
			ast.Label = data.EndLabel()
		case models.Iter:
			label.Used = true
			ast.Label = data.EndLabel()
		default:
			p.pusherrtok(ast.LabelToken, "invalid_label")
		}
		break
	}
}

func (p *Parser) continueWithLabel(ast *models.Continue) {
	if p.currentIter == nil {
		p.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
		return
	}
	label := find_label_parent(ast.LoopLabel.Kind, p.currentIter.Parent)
	if label == nil {
		p.pusherrtok(ast.LoopLabel, "label_not_exist", ast.LoopLabel.Kind)
		return
	} else if label.Index+1 >= len(label.Block.Tree) {
		p.pusherrtok(ast.LoopLabel, "invalid_label")
		return
	}
	label.Used = true
	for i := label.Index + 1; i < len(label.Block.Tree); i++ {
		obj := &label.Block.Tree[i]
		if obj.Data == nil {
			continue
		}
		switch data := obj.Data.(type) {
		case models.Comment:
			continue
		case models.Iter:
			label.Used = true
			ast.Label = data.NextLabel()
		default:
			p.pusherrtok(ast.LoopLabel, "invalid_label")
		}
		break
	}
}

func (p *Parser) break_st(ast *models.Break) {
	switch {
	case ast.LabelToken.Id != lex.ID_NA:
		p.breakWithLabel(ast)
	case p.currentCase != nil:
		ast.Label = p.currentCase.Match.EndLabel()
	case p.currentIter != nil:
		ast.Label = p.currentIter.EndLabel()
	default:
		p.pusherrtok(ast.Token, "break_at_out_of_valid_scope")
	}
}

func (p *Parser) continue_st(ast *models.Continue) {
	switch {
	case p.currentIter == nil:
		p.pusherrtok(ast.Token, "continue_at_out_of_valid_scope")
	case ast.LoopLabel.Id != lex.ID_NA:
		p.continueWithLabel(ast)
	default:
		ast.Label = p.currentIter.NextLabel()
	}
}

func (p *Parser) checkValidityForAutoType(expr_t Type, errtok lex.Token) {
	if p.eval.has_error {
		return
	}
	switch expr_t.Id {
	case juletype.NIL:
		p.pusherrtok(errtok, "nil_for_autotype")
	case juletype.VOID:
		p.pusherrtok(errtok, "void_for_autotype")
	}
}

func (p *Parser) typeSourceOfMultiTyped(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	ok := false
	for i, mt := range types {
		mt, ok = p.typeSource(mt, err)
		types[i] = mt
	}
	dt.Tag = types
	return dt, ok
}

func (p *Parser) typeSourceIsAlias(dt Type, alias *TypeAlias, err bool) (Type, bool) {
	original := dt.Original
	old := dt
	dt = alias.Type
	dt.Token = alias.Token
	dt.Generic = alias.Generic
	dt.Original = original
	dt, ok := p.typeSource(dt, err)
	dt.Pure = false
	if ok && old.Tag != nil && !types.IsStruct(alias.Type) { // Has generics
		p.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, ok
}

func (p *Parser) typeSourceIsEnum(e *Enum, tag any) (dt Type, _ bool) {
	dt.Id = juletype.ENUM
	dt.Kind = e.Id
	dt.Tag = e
	dt.Token = e.Token
	dt.Pure = true
	if tag != nil {
		p.pusherrtok(dt.Token, "invalid_type_source")
	}
	return dt, true
}

func (p *Parser) typeSourceIsFn(dt Type, err bool) (Type, bool) {
	f := dt.Tag.(*Fn)
	p.reload_fn_types(f)
	dt.Kind = f.TypeKind()
	return dt, true
}

func (p *Parser) typeSourceIsMap(dt Type, err bool) (Type, bool) {
	types := dt.Tag.([]Type)
	key := &types[0]
	*key, _ = p.realType(*key, err)
	value := &types[1]
	*value, _ = p.realType(*value, err)
	dt.Kind = dt.MapKind()
	return dt, true
}

func (p *Parser) typeSourceIsStruct(s *Struct, st Type) (dt Type, _ bool) {
	generics := s.GetGenerics()
	if len(generics) > 0 {
		if !p.checkGenericsQuantity(len(s.Generics), len(generics), st.Token) {
			goto end
		}
		for i, g := range generics {
			var ok bool
			g, ok = p.realType(g, true)
			generics[i] = g
			if !ok {
				goto end
			}
		}
		*s.Constructor.Combines = append(*s.Constructor.Combines, generics)
		owner := s.Owner.(*Parser)
		blockTypes := owner.blockTypes
		owner.blockTypes = nil
		owner.pushGenerics(s.Generics, generics)
		for i, f := range s.Fields {
			owner.parseField(s, &f, i)
		}
		if len(s.Defines.Funcs) > 0 {
			for _, f := range s.Defines.Funcs {
				if len(f.Generics) == 0 {
					blockVars := owner.blockVars
					blockTypes := owner.blockTypes
					owner.reload_fn_types(f)
					_ = p.parse_pure_fn(f)
					owner.blockVars = blockVars
					owner.blockTypes = blockTypes
				}
			}
		}
		if owner != p {
			owner.wg.Wait()
			p.pusherrs(owner.Errors...)
			owner.Errors = nil
		}
		owner.blockTypes = blockTypes
	} else if len(s.Generics) > 0 {
		p.pusherrtok(st.Token, "has_generics")
	}
end:
	dt.Id = juletype.STRUCT
	dt.Kind = s.AsTypeKind()
	dt.Tag = s
	dt.Token = s.Token
	return dt, true
}

func (p *Parser) typeSourceIsTrait(trait_def *models.Trait, tag any, errTok lex.Token) (dt Type, _ bool) {
	if tag != nil {
		p.pusherrtok(errTok, "invalid_type_source")
	}
	trait_def.Used = true
	dt.Id = juletype.TRAIT
	dt.Kind = trait_def.Id
	dt.Tag = trait_def
	dt.Token = trait_def.Token
	dt.Pure = true
	return dt, true
}

func (p *Parser) tokenizeDataType(id string) []lex.Token {
	parts := strings.SplitN(id, lex.KND_DBLCOLON, -1)
	var toks []lex.Token
	for i, part := range parts {
		toks = append(toks, lex.Token{
			Id:   lex.ID_IDENT,
			Kind: part,
			File: p.File,
		})
		if i < len(parts)-1 {
			toks = append(toks, lex.Token{
				Id:   lex.ID_DBLCOLON,
				Kind: lex.KND_DBLCOLON,
				File: p.File,
			})
		}
	}
	return toks
}

func (p *Parser) typeSourceIsArrayType(arr_t *Type) (ok bool) {
	ok = true
	arr_t.Original = nil
	arr_t.Pure = true
	*arr_t.ComponentType, ok = p.realType(*arr_t.ComponentType, true)
	if !ok {
		return
	} else if types.IsArray(*arr_t.ComponentType) && arr_t.ComponentType.Size.AutoSized {
		p.pusherrtok(arr_t.Token, "invalid_type")
	}
	modifiers := arr_t.Modifiers()
	arr_t.Kind = modifiers + jule.PREFIX_ARRAY + arr_t.ComponentType.Kind
	if arr_t.Size.AutoSized || arr_t.Size.Expr.Model != nil {
		return
	}
	val, model := p.evalExpr(arr_t.Size.Expr, nil)
	arr_t.Size.Expr.Model = model
	if val.constExpr {
		arr_t.Size.N = models.Size(tonumu(val.expr))
	} else {
		p.eval.pusherrtok(arr_t.Token, "expr_not_const")
	}
	assign_checker{
		p:      p,
		expr_t:      Type{Id: juletype.UINT, Kind: juletype.TYPE_MAP[juletype.UINT]},
		v:      val,
		errtok: arr_t.Size.Expr.Tokens[0],
	}.check()
	return
}

func (p *Parser) typeSourceIsSliceType(slc_t *Type) (ok bool) {
	*slc_t.ComponentType, ok = p.realType(*slc_t.ComponentType, true)
	if ok && types.IsArray(*slc_t.ComponentType) && slc_t.ComponentType.Size.AutoSized {
		p.pusherrtok(slc_t.Token, "invalid_type")
	}
	modifiers := slc_t.Modifiers()
	slc_t.Kind = modifiers + jule.PREFIX_SLICE + slc_t.ComponentType.Kind
	return
}

func (p *Parser) check_type_validity(expr_t Type, errtok lex.Token) {
	modifiers := expr_t.Modifiers()
	if strings.Contains(modifiers, "&&") ||
		(strings.Contains(modifiers, "*") && strings.Contains(modifiers, "&")) {
		p.pusherrtok(expr_t.Token, "invalid_type")
		return
	}
	if types.IsRef(expr_t) && !types.ValidForRef(types.DerefPtrOrRef(expr_t)) {
		p.pusherrtok(errtok, "invalid_type")
		return
	}
	if expr_t.Id == juletype.UNSAFE {
		n := len(expr_t.Kind) - len(lex.KND_UNSAFE) - 1
		if n < 0 || expr_t.Kind[n] != '*' {
			p.pusherrtok(errtok, "invalid_type")
		}
	}
}

func (p *Parser) get_define(id string, cpp_linked bool) any {
	var def any = nil
	if cpp_linked {
		def, _ = p.linkById(id)
	} else if strings.Contains(id, lex.KND_DBLCOLON) { // Has namespace?
		toks := p.tokenizeDataType(id)
		defs := p.eval.getNs(&toks)
		if defs == nil {
			return nil
		}
		i, m, def_t := defs.FindById(toks[0].Kind, p.File)
		switch def_t {
		case 't':
			def = m.Types[i]
		case 's':
			def = m.Structs[i]
		case 'e':
			def = m.Enums[i]
		case 'i':
			def = m.Traits[i]
		}
	} else {
		def, _, _ = p.defined_by_id(id)
	}
	return def
}

func (p *Parser) typeSource(dt Type, err bool) (ret Type, ok bool) {
	if dt.Kind == "" {
		return dt, true
	}
	original := dt.Original
	defer func() {
		ret.CppLinked = (original != nil && original.(Type).CppLinked) || dt.CppLinked
		ret.Original = original
		p.check_type_validity(ret, dt.Token)
	}()
	dt.SetToOriginal()
	switch {
	case dt.MultiTyped:
		return p.typeSourceOfMultiTyped(dt, err)
	case dt.Id == juletype.MAP:
		return p.typeSourceIsMap(dt, err)
	case dt.Id == juletype.ARRAY:
		ok = p.typeSourceIsArrayType(&dt)
		return dt, ok
	case dt.Id == juletype.SLICE:
		ok = p.typeSourceIsSliceType(&dt)
		return dt, ok
	}
	switch dt.Id {
	case juletype.STRUCT:
		_, prefix := dt.KindId()
		ret, ok = p.typeSourceIsStruct(dt.Tag.(*Struct), dt)
		ret.Kind = prefix + ret.Kind
		return
	case juletype.ID:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		def := p.get_define(id, dt.CppLinked)
		switch def := def.(type) {
		case *TypeAlias:
			def.Used = true
			return p.typeSourceIsAlias(dt, def, err)
		case *Enum:
			def.Used = true
			return p.typeSourceIsEnum(def, dt.Tag)
		case *Struct:
			def.Used = true
			def = p.structConstructorInstance(def)
			switch tagt := dt.Tag.(type) {
			case []models.Type:
				def.SetGenerics(tagt)
			}
			return p.typeSourceIsStruct(def, dt)
		case *models.Trait:
			def.Used = true
			return p.typeSourceIsTrait(def, dt.Tag, dt.Token)
		default:
			if err {
				p.pusherrtok(dt.Token, "invalid_type_source")
			}
			return dt, false
		}
	case juletype.FN:
		return p.typeSourceIsFn(dt, err)
	}
	return dt, true
}

func (p *Parser) realType(dt Type, err bool) (ret Type, _ bool) {
	original := dt.Original
	defer func() {
		ret.CppLinked = (original != nil && original.(Type).CppLinked) || dt.CppLinked
		ret.Original = original
	}()
	dt.SetToOriginal()
	return p.typeSource(dt, err)
}

func (p *Parser) checkMultiType(real, check Type, ignoreAny bool, errTok lex.Token) {
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	realTypes := real.Tag.([]Type)
	checkTypes := real.Tag.([]Type)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		p.check_type(realType, checkType, ignoreAny, true, errTok)
	}
}

func (p *Parser) check_type(real, check Type, ignoreAny, allow_assign bool, errTok lex.Token) {
	if types.IsVoid(check) {
		p.eval.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
		return
	}
	if !ignoreAny && real.Id == juletype.ANY {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.checkMultiType(real, check, ignoreAny, errTok)
		return
	}
	checker := types.Checker{
		ErrTok:      errTok,
		L:           real,
		R:           check,
		IgnoreAny:   ignoreAny,
		AllowAssign: allow_assign,
	}
	ok := checker.Check()
	if ok || checker.ErrorLogged {
		p.pusherrs(checker.Errors...)
		return
	}
	if real.Kind != check.Kind {
		p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
	} else if types.IsArray(real) || types.IsArray(check) {
		if types.IsArray(real) != types.IsArray(check) {
			p.pusherrtok(errTok, "incompatible_types", real.Kind, check.Kind)
			return
		}
		realKind := strings.Replace(real.Kind, jule.MARK_ARRAY, strconv.Itoa(real.Size.N), 1)
		checkKind := strings.Replace(check.Kind, jule.MARK_ARRAY, strconv.Itoa(check.Size.N), 1)
		p.pusherrtok(errTok, "incompatible_types", realKind, checkKind)
	}
}

func (p *Parser) evalExpr(expr Expr, prefix *models.Type) (value, iExpr) {
	p.eval.has_error = false
	p.eval.type_prefix = prefix
	return p.eval.eval_expr(expr)
}

func (p *Parser) evalToks(toks []lex.Token, prefix *models.Type) (value, iExpr) {
	p.eval.has_error = false
	p.eval.type_prefix = prefix
	return p.eval.eval_toks(toks)
}
