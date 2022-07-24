package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/ast/models"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xtype"
	"github.com/the-xlang/xxc/preprocessor"
)

type File = xio.File
type Type = models.Type
type Var = models.Var
type Func = models.Func
type Arg = models.Arg
type Param = models.Param
type DataType = models.DataType
type Expr = models.Expr
type Tok = ast.Tok
type Toks = ast.Toks
type Attribute = models.Attribute
type Enum = models.Enum
type Struct = models.Struct
type GenericType = models.GenericType
type RetType = models.RetType

var used []*use

type waitingGlobal struct {
	Var  *Var
	Defs *Defmap
}

// Parser is parser of X code.
type Parser struct {
	attributes      []Attribute
	docText         strings.Builder
	iterCount       int
	caseCount       int
	wg              sync.WaitGroup
	rootBlock       *models.Block
	nodeBlock       *models.Block
	generics        []*GenericType
	blockTypes      []*Type
	blockVars       []*Var
	embeds          strings.Builder
	waitingGlobals  []waitingGlobal
	eval            *eval
	receiveredFuncs []*function

	NoLocalPkg bool
	JustDefs   bool
	NoCheck    bool
	IsMain     bool
	Uses       []*use
	Defs       *Defmap
	Errors     []xlog.CompilerLog
	Warnings   []xlog.CompilerLog
	File       *File
}

// New returns new instance of Parser.
func New(f *File) *Parser {
	p := new(Parser)
	p.File = f
	p.Defs = new(Defmap)
	p.eval = new(eval)
	p.eval.p = p
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok Tok, key string, args ...any) {
	p.pusherrmsgtok(tok, x.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok Tok, msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok Tok, key string, args ...any) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.Warning,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: x.GetWarning(key, args...),
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...xlog.CompilerLog) {
	p.Errors = append(p.Errors, errs...)
}

// PushErr appends new error.
func (p *Parser) PushErr(key string, args ...any) {
	p.pusherrmsg(x.GetError(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.FlatError,
		Message: msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string, args ...any) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.FlatWarning,
		Message: x.GetWarning(key, args...),
	})
}

// CxxEmbeds returns C++ code of cxx embeds.
func (p *Parser) CxxEmbeds() string {
	var cxx strings.Builder
	cxx.WriteString(p.embeds.String())
	return cxx.String()
}

func cxxTypes(dm *Defmap) string {
	var cxx strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Tok.Id != tokens.NA {
			cxx.WriteString(t.String())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxTypes returns C++ code of types.
func (p *Parser) CxxTypes() string {
	var cxx strings.Builder
	cxx.WriteString(cxxTypes(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxTypes(use.defs))
	}
	cxx.WriteString(cxxTypes(p.Defs))
	return cxx.String()
}

func cxxEnums(dm *Defmap) string {
	var cxx strings.Builder
	for _, e := range dm.Enums {
		if e.Used && e.Tok.Id != tokens.NA {
			cxx.WriteString(e.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxEnums returns C++ code of enums.
func (p *Parser) CxxEnums() string {
	var cxx strings.Builder
	cxx.WriteString(cxxEnums(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxEnums(use.defs))
	}
	cxx.WriteString(cxxEnums(p.Defs))
	return cxx.String()
}

func cxxTraits(dm *Defmap) string {
	var cxx strings.Builder
	for _, t := range dm.Traits {
		if t.Used && t.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(t.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxTraits returns C++ code of traits.
func (p *Parser) CxxTraits() string {
	var cxx strings.Builder
	cxx.WriteString(cxxTraits(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxTraits(use.defs))
	}
	cxx.WriteString(cxxTraits(p.Defs))
	return cxx.String()
}

func cxxStructs(dm *Defmap) string {
	var cxx strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(s.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxTraits returns C++ code of structures.
func (p *Parser) CxxStructs() string {
	var cxx strings.Builder
	cxx.WriteString(cxxStructs(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxStructs(use.defs))
	}
	cxx.WriteString(cxxStructs(p.Defs))
	return cxx.String()
}

func cxxNamespaces(dm *Defmap) string {
	var cxx strings.Builder
	for _, ns := range dm.Namespaces {
		cxx.WriteString(ns.String())
		cxx.WriteString("\n\n")
	}
	return cxx.String()
}

// CxxNamespaces returns C++ code of namespaces.
func (p *Parser) CxxNamespaces() string {
	var cxx strings.Builder
	cxx.WriteString(cxxNamespaces(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxNamespaces(use.defs))
	}
	cxx.WriteString(cxxNamespaces(p.Defs))
	return cxx.String()
}

func cxxPrototypes(dm *Defmap) string {
	var cxx strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(f.Prototype())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxPrototypes returns C++ code of prototypes of C++ code.
func (p *Parser) CxxPrototypes() string {
	var cxx strings.Builder
	cxx.WriteString(cxxPrototypes(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxPrototypes(use.defs))
	}
	cxx.WriteString(cxxPrototypes(p.Defs))
	return cxx.String()
}

func cxxGlobals(dm *Defmap) string {
	var cxx strings.Builder
	for _, g := range dm.Globals {
		if !g.Const && g.Used && g.IdTok.Id != tokens.NA {
			cxx.WriteString(g.String())
			cxx.WriteByte('\n')
		}
	}
	return cxx.String()
}

// CxxGlobals returns C++ code of global variables.
func (p *Parser) CxxGlobals() string {
	var cxx strings.Builder
	cxx.WriteString(cxxGlobals(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxGlobals(use.defs))
	}
	cxx.WriteString(cxxGlobals(p.Defs))
	return cxx.String()
}

func cxxFuncs(dm *Defmap) string {
	var cxx strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
	}
	return cxx.String()
}

// CxxFuncs returns C++ code of functions.
func (p *Parser) CxxFuncs() string {
	var cxx strings.Builder
	cxx.WriteString(cxxFuncs(Builtin))
	for _, use := range used {
		cxx.WriteString(cxxFuncs(use.defs))
	}
	cxx.WriteString(cxxFuncs(p.Defs))
	return cxx.String()
}

// CxxInitializerCaller returns C++ code of initializer caller.
func (p *Parser) CxxInitializerCaller() string {
	var cxx strings.Builder
	cxx.WriteString("void ")
	cxx.WriteString(xapi.InitializerCaller)
	cxx.WriteString("(void) {")
	models.AddIndent()
	indent := models.IndentString()
	models.DoneIndent()
	pushInit := func(defs *Defmap) {
		f, _, _ := defs.funcById(x.InitializerFunction, nil)
		if f == nil {
			return
		}
		cxx.WriteByte('\n')
		cxx.WriteString(indent)
		cxx.WriteString(f.outId())
		cxx.WriteString("();")
	}
	for _, use := range used {
		pushInit(use.defs)
	}
	pushInit(p.Defs)
	cxx.WriteString("\n}")
	return cxx.String()
}

// Cxx returns full C++ code of parsed objects.
func (p *Parser) Cxx() string {
	var cxx strings.Builder
	cxx.WriteString(p.CxxEmbeds())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxTypes())
	cxx.WriteByte('\n')
	cxx.WriteString(p.CxxEnums())
	cxx.WriteString(p.CxxTraits())
	cxx.WriteString(p.CxxStructs())
	cxx.WriteString(p.CxxPrototypes())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxGlobals())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxNamespaces())
	cxx.WriteString(p.CxxFuncs())
	cxx.WriteString(p.CxxInitializerCaller())
	return cxx.String()
}

func getTree(toks Toks) ([]models.Object, []xlog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errors
}

func (p *Parser) checkUsePath(use *models.Use) bool {
	info, err := os.Stat(use.Path)
	// Exists directory?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Tok, "use_not_found", use.Path)
		return false
	}
	// Already uses?
	for _, puse := range p.Uses {
		if use.Path == puse.Path {
			p.pusherrtok(use.Tok, "already_uses")
			return false
		}
	}
	return true
}

func (p *Parser) compileUse(useAST *models.Use) (_ *use, hasErr bool) {
	infos, err := ioutil.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil, true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, x.SrcExt) ||
			!xio.IsUseable(name) {
			continue
		}
		f, err := xio.Openfx(filepath.Join(useAST.Path, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			continue
		}
		psub := New(f)
		psub.Parsef(false, false)
		use := new(use)
		use.defs = new(Defmap)
		use.tok = useAST.Tok
		use.Path = useAST.Path
		use.LinkString = useAST.LinkString
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		p.embeds.WriteString(psub.embeds.String())
		p.pushUseDefs(use, psub.Defs)
		if psub.Errors != nil {
			p.pusherrtok(useAST.Tok, "use_has_errors")
			return use, true
		}
		return use, false
	}
	return nil, false
}

func (p *Parser) pushUseNamespaces(use, dm *Defmap) {
	for _, ns := range dm.Namespaces {
		def := p.nsById(ns.Id, false)
		if def == nil {
			use.Namespaces = append(use.Namespaces, ns)
			continue
		}
	}
}

func (p *Parser) pushUseDefs(use *use, dm *Defmap) {
	p.pushUseNamespaces(use.defs, dm)
	use.defs.Types = append(use.defs.Types, dm.Types...)
	use.defs.Structs = append(use.defs.Structs, dm.Structs...)
	use.defs.Enums = append(use.defs.Enums, dm.Enums...)
	use.defs.Globals = append(use.defs.Globals, dm.Globals...)
	use.defs.Funcs = append(use.defs.Funcs, dm.Funcs...)
}

func (p *Parser) use(useAST *models.Use) (err bool) {
	if !p.checkUsePath(useAST) {
		return true
	}
	// Already parsed?
	for _, use := range used {
		if useAST.Path == use.Path {
			p.Uses = append(p.Uses, use)
			return
		}
	}
	use, err := p.compileUse(useAST)
	if use == nil {
		return err
	}
	used = append(used, use)
	p.Uses = append(p.Uses, use)
	return err
}

func (p *Parser) parseUses(tree *[]models.Object) (err bool) {
	for i, obj := range *tree {
		switch t := obj.Data.(type) {
		case models.Use:
			// || operator used for ignore compiling of other packages
			// if already have errors
			err = err || p.use(&t)
		case models.Comment: // Ignore beginning comments.
		default:
			*tree = (*tree)[i:]
			return
		}
	}
	*tree = nil
	return
}

func (p *Parser) parseSrcTreeObj(obj models.Object) {
	switch t := obj.Data.(type) {
	case Attribute:
		p.PushAttribute(t)
	case models.Statement:
		p.Statement(t)
	case Type:
		p.Type(t)
	case []GenericType:
		p.Generics(t)
	case Enum:
		p.Enum(t)
	case Struct:
		p.Struct(t)
	case models.Trait:
		p.Trait(t)
	case models.Impl:
		// Parse at end
	case models.CxxEmbed:
		p.embeds.WriteString(t.String())
		p.embeds.WriteByte('\n')
	case models.Comment:
		p.Comment(t)
	case models.Namespace:
		p.Namespace(t)
	case models.Use:
		p.pusherrtok(obj.Tok, "use_at_content")
	case models.Preprocessor:
	default:
		p.pusherrtok(obj.Tok, "invalid_syntax")
	}
}

func (p *Parser) parseSrcTreeEndObj(obj models.Object) {
	switch t := obj.Data.(type) {
	case models.Impl:
		p.Impl(t)
	}
}

func (p *Parser) parseSrcTree(tree []models.Object, parser func(models.Object)) {
	for _, obj := range tree {
		parser(obj)
		p.checkDoc(obj)
		p.checkAttribute(obj)
		p.checkGenerics(obj)
	}
}

func (p *Parser) parseTree(tree []models.Object) (ok bool) {
	if p.parseUses(&tree) {
		return false
	}
	p.parseSrcTree(tree, p.parseSrcTreeObj)
	p.parseSrcTree(tree, p.parseSrcTreeEndObj)
	return true
}

func (p *Parser) checkParse() {
	if p.NoCheck {
		return
	}
	if p.docText.Len() > 0 {
		p.pushwarn("exist_undefined_doc")
	}
	p.wg.Add(1)
	go p.check()
}

// Special case is;
//  p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPackage(tree *[]models.Object) (hasErr bool) {
	if p.File == nil {
		return
	}
	infos, err := ioutil.ReadDir(p.File.Dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, x.SrcExt) ||
			!xio.IsUseable(name) ||
			name == p.File.Name {
			continue
		}
		f, err := xio.Openfx(filepath.Join(p.File.Dir, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			return true
		}
		fp := New(f)
		fp.NoLocalPkg = true
		fp.NoCheck = true
		fp.Defs = p.Defs
		fp.Parsef(false, true)
		fp.wg.Wait()
		if len(fp.Errors) > 0 {
			p.pusherrs(fp.Errors...)
			return true
		}
		p.waitingGlobals = append(p.waitingGlobals, fp.waitingGlobals...)
	}
	return
}

// Parses X code from object tree.
func (p *Parser) Parset(tree []models.Object, main, justDefs bool) {
	p.IsMain = main
	p.JustDefs = justDefs
	preprocessor.Process(&tree, !main)
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

// Parses X code from tokens.
func (p *Parser) Parse(toks Toks, main, justDefs bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		p.pusherrs(errors...)
		return
	}
	p.Parset(tree, main, justDefs)
}

// Parses X code from file.
func (p *Parser) Parsef(main, justDefs bool) {
	lexer := lex.NewLex(p.File)
	toks := lexer.Lex()
	if lexer.Logs != nil {
		p.pusherrs(lexer.Logs...)
		return
	}
	p.Parse(toks, main, justDefs)
}

func (p *Parser) checkDoc(obj models.Object) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Data.(type) {
	case models.Comment, Attribute, []GenericType:
		return
	}
	p.pushwarntok(obj.Tok, "doc_ignored")
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj models.Object) {
	if p.attributes == nil {
		return
	}
	switch obj.Data.(type) {
	case Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Tok, "attribute_not_supports")
	p.attributes = nil
}

func (p *Parser) checkGenerics(obj models.Object) {
	if p.generics == nil {
		return
	}
	switch obj.Data.(type) {
	case Attribute, models.Comment, []GenericType:
		return
	}
	p.pusherrtok(obj.Tok, "generics_not_supports")
	p.generics = nil
}

// Generics parses generics.
func (p *Parser) Generics(generics []GenericType) {
	for i, generic := range generics {
		if xapi.IsIgnoreId(generic.Id) {
			p.pusherrtok(generic.Tok, "ignore_id")
			continue
		}
		for j, cgeneric := range generics {
			if j >= i {
				break
			} else if generic.Id == cgeneric.Id {
				p.pusherrtok(generic.Tok, "exist_id", generic.Id)
				break
			}
		}
		g := new(GenericType)
		*g = generic
		p.generics = append(p.generics, g)
	}
}

// Type parses X type define statement.
func (p *Parser) Type(t Type) {
	if _, tok, _, canshadow := p.defById(t.Id); tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return
	} else if xapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return
	}
	t.Desc = p.docText.String()
	p.docText.Reset()
	p.Defs.Types = append(p.Defs.Types, &t)
}

// Enum parses X enumerator statement.
func (p *Parser) Enum(e Enum) {
	if xapi.IsIgnoreId(e.Id) {
		p.pusherrtok(e.Tok, "ignore_id")
		return
	} else if _, tok, _, _ := p.defById(e.Id); tok.Id != tokens.NA {
		p.pusherrtok(e.Tok, "exist_id", e.Id)
		return
	}
	e.Desc = p.docText.String()
	p.docText.Reset()
	e.Type, _ = p.realType(e.Type, true)
	if !typeIsPure(e.Type) || !xtype.IsInteger(e.Type.Id) {
		p.pusherrtok(e.Type.Tok, "invalid_type_source")
		return
	}
	pdefs := p.Defs
	uses := p.Uses
	p.Defs = nil
	p.Uses = nil
	p.Defs = new(Defmap)
	defer func() {
		p.Defs = nil
		p.Defs = pdefs
		p.Uses = uses
		p.Defs.Enums = append(p.Defs.Enums, &e)
	}()
	max := xtype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			p.pusherrtok(item.Tok, "overflow_limits")
		} else {
			max--
		}
		if xapi.IsIgnoreId(item.Id) {
			p.pusherrtok(item.Tok, "ignore_id")
		} else {
			for _, checkItem := range e.Items {
				if item == checkItem {
					break
				}
				if item.Id == checkItem.Id {
					p.pusherrtok(item.Tok, "exist_id", item.Id)
					break
				}
			}
		}
		if item.Expr.Toks != nil {
			val, model := p.evalExpr(item.Expr)
			item.Expr.Model = model
			p.wg.Add(1)
			go assignChecker{
				p:         p,
				t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Tok,
			}.checkAssignType()
		} else {
			item.Expr.Model = exprNode{strconv.Itoa(i)}
		}
		itemVar := new(Var)
		itemVar.Const = true
		itemVar.Id = item.Id
		itemVar.Type = e.Type
		p.Defs.Globals = append(p.Defs.Globals, itemVar)
	}
}

func (p *Parser) pushField(s *xstruct, f *Var, i int) {
	for _, cf := range s.Ast.Fields {
		if f == cf {
			break
		}
		if f.Id == cf.Id {
			p.pusherrtok(f.IdTok, "exist_id", f.Id)
			break
		}
	}
	if len(s.Ast.Generics) == 0 {
		p.parseField(s, &f, i)
	} else {
		p.parseNonGenericType(s.Ast.Generics, &f.Type)
		param := models.Param{Id: f.Id, Type: f.Type}
		param.Default.Model = exprNode{xapi.DefaultExpr}
		s.constructor.Params[i] = param
	}
}

func (p *Parser) parseFields(s *xstruct) {
	s.constructor = new(Func)
	s.constructor.Id = s.Ast.Id
	s.constructor.Tok = s.Ast.Tok
	s.constructor.Params = make([]models.Param, len(s.Ast.Fields))
	s.constructor.RetType.Type = DataType{
		Id:   xtype.Struct,
		Kind: s.Ast.Id,
		Tok:  s.Ast.Tok,
		Tag:  s,
	}
	if len(s.Ast.Generics) > 0 {
		s.constructor.Generics = make([]*models.GenericType, len(s.Ast.Generics))
		copy(s.constructor.Generics, s.Ast.Generics)
		s.constructor.Combines = new([][]models.DataType)
	}
	s.Defs.Globals = make([]*models.Var, len(s.Ast.Fields))
	for i, f := range s.Ast.Fields {
		s.Defs.Globals[i] = f
		p.pushField(s, f, i)
	}
}

// Struct parses X structure.
func (p *Parser) Struct(s Struct) {
	if xapi.IsIgnoreId(s.Id) {
		p.pusherrtok(s.Tok, "ignore_id")
		return
	} else if _, tok, _, _ := p.defById(s.Id); tok.Id != tokens.NA {
		p.pusherrtok(s.Tok, "exist_id", s.Id)
		return
	}
	xs := new(xstruct)
	p.Defs.Structs = append(p.Defs.Structs, xs)
	xs.Desc = p.docText.String()
	p.docText.Reset()
	xs.Ast = s
	xs.Ast.Generics = p.generics
	p.generics = nil
	xs.Defs = new(Defmap)
	p.parseFields(xs)
}

// Trait parses X trait.
func (p *Parser) Trait(t models.Trait) {
	if xapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return
	} else if _, tok, _, _ := p.defById(t.Id); tok.Id != tokens.NA {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return
	}
	trait := new(trait)
	trait.Defs = new(Defmap)
	trait.Defs.Funcs = make([]*function, len(t.Funcs))
	trait.Desc = p.docText.String()
	p.docText.Reset()
	trait.Ast = &t
	p.Defs.Traits = append(p.Defs.Traits, trait)
}

// Impl parses X impl.
func (p *Parser) Impl(impl models.Impl) {
	trait, _, _ := p.traitById(impl.Trait.Kind)
	if trait == nil {
		p.pusherrtok(impl.Trait, "id_noexist", impl.Trait.Kind)
		return
	}
	trait.Used = true
	sid, _ := impl.Target.KindId()
	xs, _, _ := p.structById(sid)
	if xs == nil {
		p.pusherrtok(impl.Target.Tok, "id_noexist", sid)
		return
	}
	impl.Target.Tag = xs
	xs.traits = append(xs.traits, trait)
	for _, tf := range trait.Ast.Funcs {
		ok := false
		ds := tf.DefString()
		for _, f := range impl.Funcs {
			if tf.Pub == f.Pub && ds == f.DefString() {
				ok = true
				break
			}
		}
		if !ok {
			p.pusherrtok(impl.Target.Tok, "notimpl_trait_def", trait.Ast.Id, ds)
		}
	}
	for _, f := range impl.Funcs {
		if trait.FindFunc(f.Id) == nil {
			p.pusherrtok(impl.Target.Tok, "trait_hasnt_id", trait.Ast.Id, f.Id)
			break
		}
		sf := new(function)
		sf.Ast = f
		sf.Ast.Receiver = &impl.Target
		sf.used = true
		xs.Defs.Funcs = append(xs.Defs.Funcs, sf)
	}
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (p *Parser) pushNs(ns *models.Namespace) *namespace {
	var src *namespace
	prev := p.Defs
	for _, id := range ns.Ids {
		src = p.nsById(id, false)
		if src == nil {
			src = new(namespace)
			src.Id = id
			src.Tok = ns.Tok
			src.Defs = new(Defmap)
			src.Defs.parent = prev
			prev.Namespaces = append(prev.Namespaces, src)
		}
		prev = src.Defs
	}
	return src
}

// Namespace parses namespace statement.
func (p *Parser) Namespace(ns models.Namespace) {
	src := p.pushNs(&ns)
	pdefs := p.Defs
	p.Defs = src.Defs
	p.parseTree(ns.Tree)
	p.Defs = pdefs
}

// Comment parses X documentation comments line.
func (p *Parser) Comment(c models.Comment) {
	c.Content = strings.TrimSpace(c.Content)
	if p.docText.Len() == 0 {
		if strings.HasPrefix(c.Content, x.DocPrefix) {
			c.Content = c.Content[4:]
			if c.Content == "" {
				c.Content = " "
			}
			goto write
		}
		return
	}
	p.docText.WriteByte('\n')
write:
	p.docText.WriteString(c.Content)
}

// PushAttribute processes and appends to attribute list.
func (p *Parser) PushAttribute(attribute Attribute) {
	ok := false
	for _, kind := range x.Attributes {
		if attribute.Tag.Kind == kind {
			ok = true
			break
		}
	}
	if !ok {
		p.pusherrtok(attribute.Tag, "undefined_attribute")
	}
	for _, attr := range p.attributes {
		if attr.Tag.Kind == attribute.Tag.Kind {
			p.pusherrtok(attribute.Tag, "attribute_repeat")
			return
		}
	}
	p.attributes = append(p.attributes, attribute)
}

func genericsToCxx(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cxx strings.Builder
	cxx.WriteString("template<")
	for _, generic := range generics {
		cxx.WriteString(generic.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// Statement parse X statement.
func (p *Parser) Statement(s models.Statement) {
	switch t := s.Data.(type) {
	case Func:
		p.Func(t)
	case Var:
		p.Global(t)
	default:
		p.pusherrtok(s.Tok, "invalid_syntax")
	}
}

func (p *Parser) parseFuncNonGenericType(generics []*GenericType, t *DataType) {
	f := t.Tag.(*Func)
	for i := range f.Params {
		p.parseNonGenericType(generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(generics, &f.RetType.Type)
}

func (p *Parser) parseMultiNonGenericType(generics []*GenericType, t *DataType) {
	types := t.Tag.([]DataType)
	for i := range types {
		mt := &types[i]
		p.parseNonGenericType(generics, mt)
	}
}

func (p *Parser) parseMapNonGenericType(generics []*GenericType, t *DataType) {
	p.parseMultiNonGenericType(generics, t)
}

func (p *Parser) parseCommonNonGenericType(generics []*GenericType, t *DataType) {
	if !typeIsGeneric(generics, *t) {
		*t, _ = p.realType(*t, true)
	}
}

func (p *Parser) parseNonGenericType(generics []*GenericType, t *DataType) {
	switch {
	case t.MultiTyped:
		p.parseMultiNonGenericType(generics, t)
	case typeIsFunc(*t):
		p.parseFuncNonGenericType(generics, t)
	case typeIsMap(*t):
		p.parseMapNonGenericType(generics, t)
	default:
		p.parseCommonNonGenericType(generics, t)

	}
}

func (p *Parser) parseTypesNonGenerics(f *Func) {
	for i := range f.Params {
		p.parseNonGenericType(f.Generics, &f.Params[i].Type)
	}
	p.parseNonGenericType(f.Generics, &f.RetType.Type)
}

func (p *Parser) checkRetVars(f *function) {
	for i, v := range f.Ast.RetType.Identifiers {
		if xapi.IsIgnoreId(v.Kind) {
			continue
		}
		for _, generic := range f.Ast.Generics {
			if v.Kind == generic.Id {
				goto exist
			}
		}
		for _, param := range f.Ast.Params {
			if v.Kind == param.Id {
				goto exist
			}
		}
		for j, jv := range f.Ast.RetType.Identifiers {
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

// Func parse X function.
func (p *Parser) Func(fast Func) {
	_, tok, _, canshadow := p.defById(fast.Id)
	if tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(fast.Tok, "exist_id", fast.Id)
	} else if xapi.IsIgnoreId(fast.Id) {
		p.pusherrtok(fast.Tok, "ignore_id")
	}
	f := new(function)
	f.Ast = new(Func)
	*f.Ast = fast
	f.Ast.Attributes = p.attributes
	p.attributes = nil
	f.Desc = p.docText.String()
	p.docText.Reset()
	f.Ast.Generics = p.generics
	if len(f.Ast.Generics) > 0 {
		f.Ast.Combines = new([][]models.DataType)
	}
	p.generics = nil
	p.checkRetVars(f)
	p.checkFuncAttributes(f)
	f.used = f.Ast.Id == x.InitializerFunction
	if f.Ast.Receiver == nil {
		p.Defs.Funcs = append(p.Defs.Funcs, f)
		p.parseTypesNonGenerics(f.Ast)
	} else {
		p.receiveredFuncs = append(p.receiveredFuncs, f)
	}
}

// ParseVariable parse X global variable.
func (p *Parser) Global(vast Var) {
	def, _, _, _ := p.defById(vast.Id)
	if def != nil {
		p.pusherrtok(vast.IdTok, "exist_id", vast.Id)
		return
	} else {
		for _, g := range p.waitingGlobals {
			if vast.Id == g.Var.Id {
				p.pusherrtok(vast.IdTok, "exist_id", vast.Id)
				return
			}
		}
	}
	vast.Desc = p.docText.String()
	p.docText.Reset()
	v := new(Var)
	*v = vast
	wg := waitingGlobal{Var: v, Defs: p.Defs}
	p.waitingGlobals = append(p.waitingGlobals, wg)
	p.Defs.Globals = append(p.Defs.Globals, v)
}

func (p *Parser) checkArrayType(t *DataType) {
	exprs := t.Tag.([][]any)
	for i := range exprs {
		exprSlice := exprs[i]
		expr := exprSlice[1].(models.Expr)
		if expr.Model != nil {
			continue
		}
		if arrayExprIsAutoSized(expr) {
			p.eval.pusherrtok(t.Tok, "invalid_syntax")
			continue
		}
		val, model := p.evalExpr(expr)
		expr.Model = model
		exprSlice[1] = expr
		if val.constExpr {
			exprSlice[0] = tonumu(val.expr)
		} else {
			p.eval.pusherrtok(t.Tok, "expr_not_const")
		}
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      DataType{Id: xtype.UInt, Kind: xtype.TypeMap[xtype.UInt]},
			v:      val,
			errtok: expr.Toks[0],
		}.checkAssignType()
	}
}

// Var parse X variable.
func (p *Parser) Var(v Var) *Var {
	if xapi.IsIgnoreId(v.Id) {
		p.pusherrtok(v.IdTok, "ignore_id")
	}
	var val value
	switch t := v.Tag.(type) {
	case value:
		val = t
	default:
		if v.SetterTok.Id != tokens.NA {
			val, v.Expr.Model = p.evalExpr(v.Expr)
		}
	}
	if v.Type.Id != xtype.Void {
		t, ok := p.realType(v.Type, true)
		if ok {
			v.Type = t
			if v.SetterTok.Id != tokens.NA {
				p.wg.Add(1)
				go assignChecker{
					p:        p,
					constant: v.Const,
					t:        v.Type,
					v:        val,
					errtok:   v.IdTok,
				}.checkAssignType()
			}
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.IdTok, "missing_autotype_value")
		} else {
			p.eval.hasError = p.eval.hasError || val.data.Value == ""
			v.Type = val.data.Type
			if typeIsPure(v.Type) && xtype.IsInteger(v.Type.Id) {
				dt := DataType{
					Id:   xtype.I64,
					Kind: xtype.TypeMap[xtype.I64],
				}
				if integerAssignable(dt, val) {
					v.Type.Id = xtype.Int
					v.Type.Kind = xtype.TypeMap[v.Type.Id]
				}
			}
			p.checkValidityForAutoType(v.Type, v.SetterTok)
			p.checkAssignConst(v.Const, v.Type, val, v.SetterTok)
		}
	}
	if v.Const {
		v.ExprTag = val.expr
		if v.Volatile {
			p.pusherrtok(v.IdTok, "const_volatile")
		}
		if !typeIsAllowForConst(v.Type) {
			p.pusherrtok(v.IdTok, "invalid_type_for_const", v.Type.Kind)
		}
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.IdTok, "missing_const_value")
		} else {
			if !validExprForConst(val) {
				p.eval.pusherrtok(v.IdTok, "expr_not_const")
			}
		}
	}
	return &v
}

func (p *Parser) checkTypeParam(f *function) {
	if len(f.Ast.Generics) == 0 {
		p.pusherrtok(f.Ast.Tok, "func_must_have_generics_if_has_attribute", x.Attribute_TypeArg)
	}
	if len(f.Ast.Params) != 0 {
		p.pusherrtok(f.Ast.Tok, "func_cant_have_params_if_has_attribute", x.Attribute_TypeArg)
	}
}

func (p *Parser) checkFuncAttributes(f *function) {
	for _, attribute := range f.Ast.Attributes {
		switch attribute.Tag.Kind {
		case x.Attribute_Inline:
		case x.Attribute_TypeArg:
			p.checkTypeParam(f)
		default:
			p.pusherrtok(attribute.Tok, "invalid_attribute")
		}
	}
}

func (p *Parser) varsFromParams(params []Param) []*Var {
	length := len(params)
	vars := make([]*Var, length)
	for i, param := range params {
		v := new(models.Var)
		v.Id = param.Id
		v.IdTok = param.Tok
		v.Type = param.Type
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Tok, "variadic_parameter_notlast")
			}
			v.Type.Kind = x.Prefix_Slice + v.Type.Kind
		}
		vars[i] = v
	}
	return vars
}

// FuncById returns function by specified id.
//
// Special case:
//  FuncById(id) -> nil: if function is not exist.
func (p *Parser) FuncById(id string) (*function, *Defmap, bool) {
	f, _, _ := Builtin.funcById(id, nil)
	if f != nil {
		return f, nil, false
	}
	for _, use := range p.Uses {
		f, m, _ := use.defs.funcById(id, p.File)
		if f != nil {
			use.used = true
			return f, m, false
		}
	}
	return p.Defs.funcById(id, p.File)
}

func (p *Parser) varById(id string) (*Var, *Defmap) {
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, p.Defs
	}
	return p.globalById(id)
}

func (p *Parser) globalById(id string) (*Var, *Defmap) {
	for _, use := range p.Uses {
		g, m, _ := use.defs.globalById(id, p.File)
		if g != nil {
			use.used = true
			return g, m
		}
	}
	g, m, _ := p.Defs.globalById(id, p.File)
	return g, m
}

func (p *Parser) nsById(id string, parent bool) *namespace {
	ns := Builtin.nsById(id, parent)
	if ns != nil {
		return ns
	}
	for _, use := range p.Uses {
		ns := use.defs.nsById(id, parent)
		if ns != nil {
			use.used = true
			return ns
		}
	}
	return p.Defs.nsById(id, parent)
}

func (p *Parser) typeById(id string) (*Type, *Defmap, bool) {
	t := p.blockTypesById(id)
	if t != nil {
		return t, p.Defs, false
	}
	t, _, _ = Builtin.typeById(id, nil)
	if t != nil {
		return t, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.typeById(id, p.File)
		if t != nil {
			use.used = true
			return t, m, false
		}
	}
	return p.Defs.typeById(id, p.File)
}

func (p *Parser) enumById(id string) (*Enum, *Defmap, bool) {
	s, _, _ := Builtin.enumById(id, nil)
	if s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.enumById(id, p.File)
		if t != nil {
			use.used = true
			return t, m, false
		}
	}
	return p.Defs.enumById(id, p.File)
}

func (p *Parser) structById(id string) (*xstruct, *Defmap, bool) {
	s, _, _ := Builtin.structById(id, nil)
	if s != nil {
		return s, nil, false
	}
	for _, use := range p.Uses {
		s, m, _ := use.defs.structById(id, p.File)
		if s != nil {
			use.used = true
			return s, m, false
		}
	}
	return p.Defs.structById(id, p.File)
}

func (p *Parser) traitById(id string) (*trait, *Defmap, bool) {
	t, _, _ := Builtin.traitById(id, nil)
	if t != nil {
		return t, nil, false
	}
	for _, use := range p.Uses {
		t, m, _ := use.defs.traitById(id, p.File)
		if t != nil {
			use.used = true
			return t, m, false
		}
	}
	return p.Defs.traitById(id, p.File)
}

func (p *Parser) blockTypesById(id string) *Type {
	for _, t := range p.blockTypes {
		if t != nil && t.Id == id {
			return t
		}
	}
	return nil
}

func (p *Parser) blockVarById(id string) *Var {
	for _, v := range p.blockVars {
		if v != nil && v.Id == id {
			return v
		}
	}
	return nil
}

func (p *Parser) defById(id string) (def any, tok Tok, m *Defmap, canshadow bool) {
	var t *Type
	t, m, canshadow = p.typeById(id)
	if t != nil {
		return t, t.Tok, m, canshadow
	}
	var e *Enum
	e, m, canshadow = p.enumById(id)
	if e != nil {
		return e, e.Tok, m, canshadow
	}
	var s *xstruct
	s, m, canshadow = p.structById(id)
	if s != nil {
		return s, s.Ast.Tok, m, canshadow
	}
	var trait *trait
	trait, m, canshadow = p.traitById(id)
	if trait != nil {
		return trait, trait.Ast.Tok, m, canshadow
	}
	var f *function
	f, m, canshadow = p.FuncById(id)
	if f != nil {
		return f, f.Ast.Tok, m, canshadow
	}
	if bv := p.blockVarById(id); bv != nil {
		return bv, bv.IdTok, p.Defs, false
	}
	g, m := p.globalById(id)
	if g != nil {
		return g, g.IdTok, m, true
	}
	return
}

func (p *Parser) blockDefById(id string) (def any, tok Tok) {
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, bv.IdTok
	}
	t := p.blockTypesById(id)
	if t != nil {
		return t, t.Tok
	}
	return
}

func (p *Parser) checkUses() {
	for _, use := range p.Uses {
		if !use.used {
			p.pusherrtok(use.tok, "declared_but_not_used", use.LinkString)
		}
	}
}

func (p *Parser) check() {
	defer p.wg.Done()
	if p.IsMain && !p.JustDefs {
		f, _, _ := p.Defs.funcById(x.EntryPoint, nil)
		if f == nil {
			p.PushErr("no_entry_point")
		} else {
			f.isEntryPoint = true
			f.used = true
		}
	}
	p.checkTypes()
	p.WaitingGlobals()
	p.waitingGlobals = nil
	p.checkReceivers()
	if !p.JustDefs {
		p.checkTraits()
		p.checkFuncs()
		p.checkStructs()
		p.checkUses()
	}
}

func (p *Parser) checkTrait(t *trait) {
	for i, f := range t.Ast.Funcs {
		if xapi.IsIgnoreId(f.Id) {
			p.pusherrtok(f.Tok, "ignore_id")
		}
		for j, jf := range t.Ast.Funcs {
			if j >= i {
				break
			} else if f.Id == jf.Id {
				p.pusherrtok(f.Tok, "exist_id", f.Id)
			}
		}
		p.parseTypesNonGenerics(f)
		tf := new(function)
		tf.Ast = f
		t.Defs.Funcs[i] = tf
	}
}

func (p *Parser) checkTraits() {
	for _, ns := range p.Defs.Namespaces {
		p.Defs = ns.Defs
		for _, t := range ns.Defs.Traits {
			p.checkTrait(t)
		}
		p.Defs = p.Defs.parent
	}
	for _, t := range p.Defs.Traits {
		p.checkTrait(t)
	}
}

func (p *Parser) checkTypes() {
	for i, t := range p.Defs.Types {
		p.Defs.Types[i].Type, _ = p.realType(t.Type, true)
	}
}

// WaitingGlobals parses X global variables for waiting to parsing.
func (p *Parser) WaitingGlobals() {
	pdefs := p.Defs
	for _, g := range p.waitingGlobals {
		p.Defs = g.Defs // Set defs for namespaces
		*g.Var = *p.Var(*g.Var)
	}
	p.Defs = pdefs
}

func (p *Parser) checkParamDefaultExprWithDefault(param *Param) {
	if typeIsFunc(param.Type) {
		p.pusherrtok(param.Tok, "invalid_type_for_default_arg", param.Type.Kind)
	}
}

func (p *Parser) checkParamDefaultExpr(f *Func, param *Param) {
	if !paramHasDefaultArg(param) || param.Tok.Id == tokens.NA {
		return
	}
	// Skip default argument with default value
	if param.Default.Model != nil {
		if param.Default.Model.String() == xapi.DefaultExpr {
			p.checkParamDefaultExprWithDefault(param)
			return
		}
	}
	dt := param.Type
	if param.Variadic {
		dt.Kind = x.Prefix_Array + dt.Kind // For slice.
	}
	v, model := p.evalExpr(param.Default)
	param.Default.Model = model
	p.wg.Add(1)
	go p.checkArgType(*param, v, param.Tok)
}

func (p *Parser) param(f *Func, param *Param) (err bool) {
	param.Type, err = p.realType(param.Type, true)
	// Assign to !err because p.realType
	// returns true if success, false if not.
	err = !err
	if param.Reference {
		if param.Variadic {
			p.pusherrtok(param.Tok, "variadic_reference_param")
			err = true
		}
		if typeIsPtr(param.Type) {
			p.pusherrtok(param.Tok, "pointer_reference")
			err = true
		}
	}
	p.checkParamDefaultExpr(f, param)
	return
}

func (p *Parser) params(f *Func) (err bool) {
	hasDefaultArg := false
	for i := range f.Params {
		param := &f.Params[i]
		err = err || p.param(f, param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Tok, "param_must_have_default_arg", param.Id)
			err = true
		}
	}
	return
}

func (p *Parser) blockVarsOfFunc(f *Func) []*Var {
	vars := p.varsFromParams(f.Params)
	vars = append(vars, f.RetType.Vars()...)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*xstruct)
		vars = append(vars, s.selfVar(*f.Receiver))
	}
	return vars
}

func (p *Parser) receiver(f *Func) {
	if f.Receiver == nil {
		return
	}
	if f.Receiver.Id != xtype.Id {
		p.pusherrtok(f.Receiver.Tok, "invalid_type_source")
		return
	}
	id, _ := f.Receiver.KindId()
	s, _, _ := p.Defs.structById(id, nil)
	if s == nil {
		p.pusherrtok(f.Receiver.Tok, "invalid_type_source")
		return
	}
	f.Receiver.Tag = s
	d, _, _ := s.Defs.defById(f.Id, nil)
	if d != -1 {
		p.pusherrtok(f.Receiver.Tok, "exist_id", f.Id)
		return
	}
	for _, generic := range f.Generics {
		if findGeneric(generic.Id, s.Ast.Generics) != nil {
			p.pusherrtok(generic.Tok, "exist_id", generic.Id)
		}
	}
	fn := new(function)
	fn.Ast = f
	s.Defs.Funcs = append(s.Defs.Funcs, fn)
}

func (p *Parser) parsePureFunc(f *Func) (err bool) {
	hasError := p.eval.hasError
	defer func() { p.eval.hasError = hasError }()
	err = p.params(f)
	if err {
		return
	}
	p.blockVars = p.blockVarsOfFunc(f)
	p.checkFunc(f)
	p.blockTypes = nil
	p.blockVars = nil
	return
}

func (p *Parser) parseFunc(f *function) (err bool) {
	if f.checked ||
		(len(f.Ast.Generics) > 0 && len(*f.Ast.Combines) == 0) {
		return false
	}
	return p.parsePureFunc(f.Ast)
}

func (p *Parser) checkFuncs() {
	err := false
	check := func(f *function) {
		p.wg.Add(1)
		go p.checkFuncSpecialCases(f.Ast)
		if err {
			return
		}
		p.blockTypes = nil
		err = p.parseFunc(f)
		f.checked = true
	}
	for _, use := range p.Uses {
		for _, ns := range use.defs.Namespaces {
			pdefs := p.Defs
			p.Defs = ns.Defs
			for _, f := range ns.Defs.Funcs {
				check(f)
			}
			p.Defs = pdefs
		}
	}
	for _, ns := range p.Defs.Namespaces {
		p.Defs = ns.Defs
		for _, f := range ns.Defs.Funcs {
			check(f)
		}
		p.Defs = p.Defs.parent
	}
	for _, f := range p.Defs.Funcs {
		check(f)
	}
}

func (p *Parser) parseStructFunc(s *xstruct, f *function) (err bool) {
	if len(f.Ast.Generics) > 0 {
		return
	}
	if len(s.Ast.Generics) == 0 {
		p.parseTypesNonGenerics(f.Ast)
		return p.parseFunc(f)
	}
	return
}

func (p *Parser) checkStruct(xs *xstruct) (err bool) {
	for _, f := range xs.Defs.Funcs {
		if f.checked {
			continue
		}
		p.blockTypes = nil
		err = p.parseStructFunc(xs, f)
		if err {
			break
		}
		f.checked = true
	}
	return
}

func (p *Parser) checkStructs() {
	err := false
	check := func(xs *xstruct) {
		if err {
			return
		}
		p.checkStruct(xs)
	}
	for _, ns := range p.Defs.Namespaces {
		p.Defs = ns.Defs
		for _, s := range ns.Defs.Structs {
			check(s)
		}
		p.Defs = p.Defs.parent
	}
	for _, s := range p.Defs.Structs {
		check(s)
	}
}

func (p *Parser) checkReceivers() {
	for _, f := range p.receiveredFuncs {
		p.receiver(f.Ast)
	}
}

func (p *Parser) checkFuncSpecialCases(f *Func) {
	defer p.wg.Done()
	switch f.Id {
	case x.EntryPoint, x.InitializerFunction:
		p.checkSolidFuncSpecialCases(f)
	}
}

func (p *Parser) callFunc(f *Func, genericsToks, argsToks Toks, m *exprModel) value {
	v := p.parseFuncCallToks(f, genericsToks, argsToks, m)
	v.lvalue = typeIsLvalue(v.data.Type)
	return v
}

func (p *Parser) callStructConstructor(s *xstruct, genericsToks, argsToks Toks, m *exprModel) value {
	v := p.parseFuncCallToks(s.constructor, genericsToks, argsToks, m)
	v.isType = false
	v.lvalue = false
	return v
}

func (p *Parser) parseField(s *xstruct, f **Var, i int) {
	*f = p.Var(**f)
	v := *f
	param := models.Param{Id: v.Id, Type: v.Type}
	if v.Type.Id == xtype.Struct && v.Type.Tag == s && typeIsPure(v.Type) {
		p.pusherrtok(v.Type.Tok, "invalid_type_source")
	}
	if hasExpr(v.Expr) {
		param.Default = v.Expr
	} else {
		param.Default.Model = exprNode{xapi.DefaultExpr}
	}
	s.constructor.Params[i] = param
}

func (p *Parser) structConstructorInstance(as *xstruct) *xstruct {
	if as == errorStruct {
		return as
	}
	s := new(xstruct)
	s.Ast = as.Ast
	s.constructor = new(Func)
	*s.constructor = *as.constructor
	s.constructor.RetType.Type.Tag = s
	s.Defs = new(Defmap)
	*s.Defs = *as.Defs
	if len(as.Ast.Generics) > 0 { // Parse if has generics
		for i := range s.Ast.Fields {
			p.parseField(s, &s.Defs.Globals[i], i)
		}
	}
	for i := range s.Defs.Funcs {
		f := &s.Defs.Funcs[i]
		nf := new(function)
		*nf = **f
		nf.Ast.Receiver.Tag = s
		*f = nf
	}
	return s
}

func (p *Parser) checkAnonFunc(f *Func) {
	p.reloadFuncTypes(f)
	globals := p.Defs.Globals
	blockVariables := p.blockVars
	p.Defs.Globals = append(blockVariables, p.Defs.Globals...)
	p.blockVars = p.varsFromParams(f.Params)
	rootBlock := p.rootBlock
	nodeBlock := p.nodeBlock
	p.rootBlock = nil
	p.nodeBlock = nil
	p.checkFunc(f)
	p.rootBlock = rootBlock
	p.nodeBlock = nodeBlock
	p.Defs.Globals = globals
	p.blockVars = blockVariables
}

func (p *Parser) getArgs(toks Toks) *models.Args {
	toks, _ = p.getRange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make(Toks, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks)
	if len(b.Errors) > 0 {
		p.pusherrs(b.Errors...)
		args = nil
	}
	return args
}

// Should toks include brackets.
func (p *Parser) getGenerics(toks Toks) []DataType {
	if len(toks) == 0 {
		return nil
	}
	// Remove braces
	toks = toks[1 : len(toks)-1]
	parts, errs := ast.Parts(toks, tokens.Comma, true)
	generics := make([]DataType, len(parts))
	p.pusherrs(errs...)
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		b := ast.NewBuilder(nil)
		index := 0
		generic, _ := b.DataType(part, &index, false, true)
		b.Wait()
		if index+1 < len(part) {
			p.pusherrtok(part[index+1], "invalid_syntax")
		}
		p.pusherrs(b.Errors...)
		generics[i], _ = p.realType(generic, true)
	}
	return generics
}

func (p *Parser) checkGenericsQuantity(n int, generics []DataType, errTok Tok) bool {
	// n = length of required generic type source.
	switch {
	case n == 0 && len(generics) > 0:
		p.pusherrtok(errTok, "not_has_generics")
		return false
	case n > 0 && len(generics) == 0:
		p.pusherrtok(errTok, "has_generics")
		return false
	case n < len(generics):
		p.pusherrtok(errTok, "generics_overflow")
		return false
	case n > len(generics):
		p.pusherrtok(errTok, "missing_generics")
		return false
	default:
		return true
	}
}

func (p *Parser) pushGenerics(generics []*GenericType, sources []DataType) {
	for i, generic := range generics {
		p.blockTypes = append(p.blockTypes, &Type{
			Id:   generic.Id,
			Tok:  generic.Tok,
			Type: sources[i],
		})
	}
}

func (p *Parser) reloadFuncTypes(f *Func) {
	for i, param := range f.Params {
		f.Params[i].Type, _ = p.realType(param.Type, true)
	}
	f.RetType.Type, _ = p.realType(f.RetType.Type, true)
}

func itsCombined(f *Func, generics []DataType) bool {
	for _, combine := range *f.Combines {
		for i, gt := range generics {
			ct := combine[i]
			if typesEquals(gt, ct) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) parseGenerics(f *Func, generics []DataType, m *exprModel, errTok Tok) bool {
	if !p.checkGenericsQuantity(len(f.Generics), generics, errTok) {
		return false
	}
	// Add generic types to call expression
	var cxx strings.Builder
	cxx.WriteByte('<')
	for _, generic := range generics {
		cxx.WriteString(generic.String())
		cxx.WriteByte(',')
	}
	m.appendSubNode(exprNode{cxx.String()[:cxx.Len()-1] + ">"})
	// Apply generics
	blockTypes := p.blockTypes
	blockVars := p.blockVars
	p.blockTypes = nil
	defer func() { p.blockTypes, p.blockVars = blockTypes, blockVars }()
	p.pushGenerics(f.Generics, generics)
	if isConstructor(f) {
		p.readyConstructor(&f)
		s := f.RetType.Type.Tag.(*xstruct)
		s.SetGenerics(generics)
	} else if f.Receiver != nil {
		s := f.Receiver.Tag.(*xstruct)
		p.pushGenerics(s.Ast.Generics, s.Generics())
	}
	p.reloadFuncTypes(f)
	if itsCombined(f, generics) {
		return true
	}
	*f.Combines = append(*f.Combines, generics)
	if isConstructor(f) {
		return true
	}
	rootBlock := p.rootBlock
	nodeBlock := p.nodeBlock
	defer func() { p.rootBlock, p.nodeBlock = rootBlock, nodeBlock }()
	p.rootBlock = nil
	p.nodeBlock = nil
	p.parsePureFunc(f)
	return true
}

func isConstructor(f *Func) bool {
	if !typeIsStruct(f.RetType.Type) {
		return false
	}
	s := f.RetType.Type.Tag.(*xstruct)
	return f.Id == s.Ast.Id
}

func (p *Parser) readyConstructor(f **Func) {
	s := (*f).RetType.Type.Tag.(*xstruct)
	s = p.structConstructorInstance(s)
	*f = s.constructor
}

func (p *Parser) parseFuncCall(f *Func, generics []DataType, args *models.Args, m *exprModel, errTok Tok) (v value) {
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		copy(params, f.Params)
		retType := f.RetType
		defer func() { f.Params, f.RetType = params, retType }()
		if !p.parseGenerics(f, generics, m, errTok) {
			return
		}
		f.RetType.Type.DontUseOriginal = true
	} else {
		_ = p.checkGenericsQuantity(len(f.Generics), generics, errTok)
		if f.Receiver != nil {
			s := f.Receiver.Tag.(*xstruct)
			generics := s.Generics()
			if len(generics) > 0 {
				blockTypes := p.blockTypes
				p.blockTypes = nil
				p.pushGenerics(s.Ast.Generics, generics)
				p.reloadFuncTypes(f)
				p.blockTypes = blockTypes
			}
		}
	}
	v.data.Type = f.RetType.Type
	v.data.Value = f.Id
	if isConstructor(f) {
		if len(f.Generics) == 0 {
			p.readyConstructor(&f)
		}
		s := f.RetType.Type.Tag.(*xstruct)
		s.SetGenerics(generics)
		v.data.Type.Kind = s.dataTypeString()
	}
	if m != nil {
		m.appendSubNode(exprNode{tokens.LPARENTHESES})
		defer m.appendSubNode(exprNode{tokens.RPARENTHESES})
	}
	if args == nil {
		return
	}
	p.parseArgs(f, args, m, errTok)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
	return
}

func (p *Parser) parseFuncCallToks(f *Func, genericsToks, argsToks Toks, m *exprModel) (v value) {
	var generics []DataType
	var args *models.Args
	if f.FindAttribute(x.Attribute_TypeArg) != nil {
		if len(genericsToks) > 0 {
			p.pusherrtok(genericsToks[0], "invalid_syntax")
			return
		}
		generics = p.getGenerics(argsToks)
	} else {
		generics = p.getGenerics(genericsToks)
		args = p.getArgs(argsToks)
	}
	return p.parseFuncCall(f, generics, args, m, argsToks[0])
}

func (p *Parser) parseTargetedArgs(f *Func, args *models.Args, errTok Tok) {
	tap := targetedArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	tap.parse()
}

func (p *Parser) parsePureArgs(f *Func, args *models.Args, m *exprModel, errTok Tok) {
	pap := pureArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
		m:      m,
	}
	pap.parse()
}

func (p *Parser) parseArgs(f *Func, args *models.Args, m *exprModel, errTok Tok) {
	if args.Targeted {
		p.parseTargetedArgs(f, args, errTok)
		return
	}
	p.parsePureArgs(f, args, m, errTok)
}

func hasExpr(expr Expr) bool {
	return len(expr.Processes) > 0 || expr.Model != nil
}

func paramHasDefaultArg(param *Param) bool {
	return hasExpr(param.Default)
}

//             [identifier]
type paramMap map[string]*paramMapPair
type paramMapPair struct {
	param *Param
	arg   *Arg
}

func getParamMap(params []Param) *paramMap {
	pmap := new(paramMap)
	*pmap = make(paramMap, len(params))
	for i := range params {
		p := &params[i]
		(*pmap)[p.Id] = &paramMapPair{p, nil}
	}
	return pmap
}

func (p *Parser) parseArg(param Param, arg *Arg, variadiced *bool) {
	value, model := p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	p.wg.Add(1)
	go p.checkArgType(param, value, arg.Tok)
}

func (p *Parser) checkArgType(param Param, val value, errTok Tok) {
	defer p.wg.Done()
	if param.Reference && !val.lvalue {
		p.pusherrtok(errTok, "not_lvalue_for_reference_param")
	}
	p.wg.Add(1)
	go assignChecker{
		p:      p,
		t:      param.Type,
		v:      val,
		errtok: errTok,
	}.checkAssignType()
}

// Returns between of brackets.
//
// Special case is:
//  getRange(open, close, tokens) = nil, false if first token is not brace.
func (p *Parser) getRange(open, close string, toks Toks) (_ Toks, ok bool) {
	braceCount := 0
	start := 1
	if toks[0].Id != tokens.Brace {
		return nil, false
	}
	for i, tok := range toks {
		if tok.Id != tokens.Brace {
			continue
		}
		if tok.Kind == open {
			braceCount++
		} else if tok.Kind == close {
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		return toks[start:i], true
	}
	return nil, false
}

func (p *Parser) checkSolidFuncSpecialCases(f *Func) {
	if len(f.Params) > 0 {
		p.pusherrtok(f.Tok, "func_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != xtype.Void {
		p.pusherrtok(f.RetType.Type.Tok, "func_have_return", f.Id)
	}
	if f.Attributes != nil {
		p.pusherrtok(f.Tok, "func_have_attributes", f.Id)
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
		p.nodeBlock = b
		defer func() { p.nodeBlock = oldNode }()
	}
	blockTypes := p.blockTypes
	p.checkBlock(b)

	vars := p.blockVars[len(oldBlockVars):]
	types := p.blockTypes[len(blockTypes):]
	for _, v := range vars {
		if !v.Used {
			p.pusherrtok(v.IdTok, "declared_but_not_used", v.Id)
		}
	}
	for _, t := range types {
		if !t.Used {
			p.pusherrtok(t.Tok, "declared_but_not_used", t.Id)
		}
	}

	p.blockVars = oldBlockVars
	p.blockTypes = blockTypes
}

func (p *Parser) checkNewBlock(b *models.Block) {
	p.checkNewBlockCustom(b, p.blockVars)
}

func (p *Parser) statement(s *models.Statement, recover bool) bool {
	switch t := s.Data.(type) {
	case models.ExprStatement:
		p.exprStatement(&t, recover)
		s.Data = t
	case Var:
		p.varStatement(&t, false)
		s.Data = t
	case models.Assign:
		p.assign(&t)
		s.Data = t
	case models.Iter:
		p.iter(&t)
		s.Data = t
	case models.Break:
		p.breakStatement(&t)
	case models.Continue:
		p.continueStatement(&t)
	case Type:
		if def, _ := p.blockDefById(t.Id); def != nil {
			p.pusherrtok(t.Tok, "exist_id", t.Id)
			break
		} else if xapi.IsIgnoreId(t.Id) {
			p.pusherrtok(t.Tok, "ignore_id")
			break
		}
		t.Type, _ = p.realType(t.Type, true)
		p.blockTypes = append(p.blockTypes, &t)
	case *models.Block:
		p.checkNewBlock(t)
		s.Data = t
	case models.Defer:
		p.deferredCall(&t)
		s.Data = t
	case models.ConcurrentCall:
		p.concurrentCall(&t)
		s.Data = t
	case models.Match:
		p.matchcase(&t)
		s.Data = t
	case models.CxxEmbed:
		p.cxxEmbed(&t)
		s.Data = t
	case models.Comment:
	default:
		return false
	}
	return true
}

func (p *Parser) checkStatement(b *models.Block, i *int) {
	s := b.Tree[*i]
	defer func(i int) { b.Tree[i] = s }(*i)
	if p.statement(&s, true) {
		return
	}
	switch t := s.Data.(type) {
	case models.If:
		p.ifExpr(&t, i, b.Tree)
		s.Data = t
	case models.Goto:
		t.Index = *i
		t.Block = b
		*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, &t)
	case models.Ret:
		rc := retChecker{p: p, retAST: &t, f: b.Func}
		rc.check()
		s.Data = t
	case models.Label:
		t.Block = b
		t.Index = *i
		*p.rootBlock.Labels = append(*b.Labels, &t)
	default:
		p.pusherrtok(s.Tok, "invalid_syntax")
	}
}

func (p *Parser) checkBlock(b *models.Block) {
	for i := 0; i < len(b.Tree); i++ {
		p.checkStatement(b, &i)
	}
}

func (p *Parser) recoverFuncExprStatement(s *models.ExprStatement) {
	errtok := s.Expr.Processes[0][0]
	callToks := s.Expr.Processes[0][1:]
	args := p.getArgs(callToks)
	handleParam := recoverFunc.Ast.Params[0]
	if len(args.Src) == 0 {
		p.pusherrtok(errtok, "missing_argument_for", handleParam.Id)
		return
	} else if len(args.Src) > 1 {
		p.pusherrtok(errtok, "argument_overflow")
	}
	v, _ := p.evalExpr(args.Src[0].Expr)
	if v.data.Type.Kind != handleParam.Type.Kind {
		p.eval.pusherrtok(errtok, "incompatible_datatype", handleParam.Type.Kind, v.data.Type.Kind)
		return
	}
	handler := v.data.Type.Tag.(*Func)
	s.Expr.Model = exprNode{"try{\n"}
	var catcher serieExpr
	catcher.exprs = append(catcher.exprs, "} catch(XID(Error) ")
	catcher.exprs = append(catcher.exprs, handler.Params[0].OutId())
	catcher.exprs = append(catcher.exprs, ") ")
	r, _ := utf8.DecodeRuneInString(v.data.Value)
	if r == '_' || unicode.IsLetter(r) { // Function source
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

func (p *Parser) exprStatement(s *models.ExprStatement, recover bool) {
	if s.Expr.Processes != nil && !isOperator(s.Expr.Processes[0]) {
		process := s.Expr.Processes[0]
		tok := process[0]
		if tok.Id == tokens.Id && tok.Kind == recoverFunc.Ast.Id {
			if ast.IsFuncCall(s.Expr.Toks) != nil {
				if !recover {
					p.pusherrtok(tok, "invalid_syntax")
				}
				def, _, _, _ := p.defById(tok.Kind)
				if def == recoverFunc {
					p.recoverFuncExprStatement(s)
					return
				}
			}
		}
	}
	if s.Expr.Model == nil {
		_, s.Expr.Model = p.evalExpr(s.Expr)
	}
}

func (p *Parser) parseCase(c *models.Case, t DataType) {
	for i := range c.Exprs {
		expr := &c.Exprs[i]
		value, model := p.evalExpr(*expr)
		expr.Model = model
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      t,
			v:      value,
			errtok: expr.Toks[0],
		}.checkAssignType()
	}
	p.caseCount++
	defer func() { p.caseCount-- }()
	p.checkNewBlock(c.Block)
}

func (p *Parser) cases(cases []models.Case, t DataType) {
	for i := range cases {
		p.parseCase(&cases[i], t)
	}
}

func (p *Parser) matchcase(t *models.Match) {
	if len(t.Expr.Processes) > 0 {
		value, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		t.ExprType = value.data.Type
	} else {
		t.ExprType.Id = xtype.Bool
		t.ExprType.Kind = xtype.TypeMap[t.ExprType.Id]
	}
	p.cases(t.Cases, t.ExprType)
	if t.Default != nil {
		p.parseCase(t.Default, t.ExprType)
	}
}

func isCxxReturn(s string) bool {
	return strings.HasPrefix(s, "return")
}

func (p *Parser) cxxEmbed(cxx *models.CxxEmbed) {
	rexpr := regexp.MustCompile(`@[\p{L}|_]([\p{L}0-9_]+)?`)
	match := rexpr.FindStringIndex(cxx.Content)
	for match != nil {
		start := match[0]
		end := match[1]
		// +1 for skip "@" mark.
		id := cxx.Content[start+1 : end]
		def, tok := p.blockDefById(id)
		if def != nil {
			switch t := def.(type) {
			case *Var:
				t.Used = true
			case *Type:
				t.Used = true
			}
		}
		cxx.Content = cxx.Content[:start] + xapi.OutId(id, tok.File) + cxx.Content[end:]
		match = rexpr.FindStringIndex(cxx.Content)
	}
	cxxcode := strings.TrimLeftFunc(cxx.Content, unicode.IsSpace)
	if isCxxReturn(cxxcode) {
		p.embedReturn(cxxcode, cxx.Tok)
	}
}

func (p *Parser) embedReturn(cxx string, errTok Tok) {
	returnKwLen := 6
	cxx = cxx[returnKwLen:]
	cxx = strings.TrimLeftFunc(cxx, unicode.IsSpace)
	if cxx[len(cxx)-1] == ';' {
		cxx = cxx[:len(cxx)-1]
	}
	f := p.rootBlock.Func
	if len(cxx) == 0 && !typeIsVoid(f.RetType.Type) {
		p.pusherrtok(errTok, "require_return_value")
	} else if len(cxx) > 0 && typeIsVoid(f.RetType.Type) {
		p.pusherrtok(errTok, "void_function_return_value")
	}
}

func (p *Parser) findLabel(id string) *models.Label {
	for _, label := range *p.rootBlock.Labels {
		if label.Label == id {
			return label
		}
	}
	return nil
}

func (p *Parser) checkLabels() {
	labels := p.rootBlock.Labels
	for _, label := range *labels {
		for _, checkLabel := range *labels {
			if label.Index == checkLabel.Index {
				break
			} else if label.Label == checkLabel.Label {
				p.pusherrtok(label.Tok, "label_exist", label.Label)
			}
		}
		if !label.Used {
			p.pusherrtok(label.Tok, "declared_but_not_used", label.Label+":")
		}
	}
}

func statementIsDef(s *models.Statement) bool {
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
		if statementIsDef(s) {
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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
			case s.Tok.Row >= label.Tok.Row:
				return true
			case statementIsDef(s):
				p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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
		case s.Tok.Row >= label.Tok.Row:
			return
		case statementIsDef(s):
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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
			if s.Tok.Row <= gt.Tok.Row {
				return
			}
		}
		if statementIsDef(s) {
			p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
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
		label := p.findLabel(gt.Label)
		if label == nil {
			p.pusherrtok(gt.Tok, "label_not_exist", gt.Label)
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

func (p *Parser) checkRets(f *Func) {
	if f.Block != nil {
		for _, s := range f.Block.Tree {
			switch t := s.Data.(type) {
			case models.Ret:
				return
			case models.CxxEmbed:
				cxx := strings.TrimLeftFunc(t.Content, unicode.IsSpace)
				if isCxxReturn(cxx) {
					return
				}
			}
		}
	}
	if !typeIsVoid(f.RetType.Type) {
		p.pusherrtok(f.Tok, "missing_ret")
	}
}

func (p *Parser) checkFunc(f *Func) {
	if f.Block == nil || f.Block.Tree == nil {
		goto always
	}
	f.Block.Func = f
	p.checkNewBlock(f.Block)
always:
	p.checkRets(f)
}

func (p *Parser) varStatement(v *Var, noParse bool) {
	if _, tok := p.blockDefById(v.Id); tok.Id != tokens.NA {
		p.pusherrtok(v.IdTok, "exist_id", v.Id)
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	p.blockVars = append(p.blockVars, v)
}

func (p *Parser) deferredCall(d *models.Defer) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, d.Expr.Model = p.evalExpr(d.Expr)
}

func (p *Parser) concurrentCall(cc *models.ConcurrentCall) {
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	_, cc.Expr.Model = p.evalExpr(cc.Expr)
}

func (p *Parser) assignment(selected value, errtok Tok) bool {
	state := true
	if !selected.lvalue {
		p.eval.pusherrtok(errtok, "assign_nonlvalue")
		state = false
	}
	if selected.constant {
		p.pusherrtok(errtok, "assign_const")
		state = false
	}
	switch selected.data.Type.Tag.(type) {
	case Func:
		if f, _, _ := p.FuncById(selected.data.Tok.Kind); f != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) singleAssign(assign *models.Assign, exprs []value) {
	right := &assign.Right[0]
	val := exprs[0]
	left := &assign.Left[0].Expr
	if len(left.Toks) == 1 && xapi.IsIgnoreId(left.Toks[0].Kind) {
		return
	}
	leftExpr, model := p.evalExpr(*left)
	left.Model = model
	if !p.assignment(leftExpr, assign.Setter) {
		return
	}
	if assign.Setter.Kind != tokens.EQUAL && !isConstExpression(val.data.Value) {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     left.Toks,
			leftVal:  leftExpr,
			right:    right.Toks,
			rightVal: val,
			operator: assign.Setter,
		}
		val = solver.solve()
		assign.Setter.Kind += tokens.EQUAL
	}
	p.wg.Add(1)
	go assignChecker{
		p:        p,
		constant: leftExpr.constant,
		t:        leftExpr.data.Type,
		v:        val,
		errtok:   assign.Setter,
	}.checkAssignType()
}

func (p *Parser) assignExprs(vsAST *models.Assign) []value {
	vals := make([]value, len(vsAST.Right))
	for i, expr := range vsAST.Right {
		val, model := p.evalExpr(expr)
		vsAST.Right[i].Model = model
		vals[i] = val
	}
	return vals
}

func (p *Parser) funcMultiAssign(vsAST *models.Assign, funcVal value) {
	types := funcVal.data.Type.Tag.([]DataType)
	if len(types) != len(vsAST.Left) {
		p.pusherrtok(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	vals := make([]value, len(types))
	for i, t := range types {
		vals[i] = value{data: models.Data{Tok: t.Tok, Type: t}}
	}
	p.multiAssign(vsAST, vals)
}

func (p *Parser) multiAssign(assign *models.Assign, right []value) {
	for i := range assign.Left {
		left := &assign.Left[i]
		left.Ignore = xapi.IsIgnoreId(left.Var.Id)
		right := right[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr, model := p.evalExpr(left.Expr)
			left.Expr.Model = model
			if !p.assignment(leftExpr, assign.Setter) {
				return
			}
			p.wg.Add(1)
			go assignChecker{
				p:        p,
				constant: leftExpr.constant,
				t:        leftExpr.data.Type,
				v:        right,
				errtok:   assign.Setter,
			}.checkAssignType()
			continue
		}
		left.Var.Tag = right
		p.varStatement(&left.Var, false)
	}
}

func (p *Parser) suffix(assign *models.Assign, exprs []value) {
	if len(exprs) > 0 {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	left := &assign.Left[0]
	val, model := p.evalExpr(left.Expr)
	left.Expr.Model = model
	_ = p.assignment(val, assign.Setter)
	if typeIsPtr(val.data.Type) {
		return
	}
	if typeIsPure(val.data.Type) && xtype.IsNumeric(val.data.Type.Id) {
		return
	}
	p.pusherrtok(assign.Setter, "operator_notfor_xtype", assign.Setter.Kind, val.data.Type.Kind)
}

func (p *Parser) assign(assign *models.Assign) {
	leftLength := len(assign.Left)
	rightLength := len(assign.Right)
	exprs := p.assignExprs(assign)
	if rightLength == 0 && ast.IsSuffixOperator(assign.Setter.Kind) { // Suffix
		p.suffix(assign, exprs)
		return
	} else if leftLength == 1 && !assign.Left[0].Var.New {
		p.singleAssign(assign, exprs)
		return
	} else if assign.Setter.Kind != tokens.EQUAL {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	} else if rightLength == 1 {
		expr := exprs[0]
		if expr.data.Type.MultiTyped {
			assign.MultipleRet = true
			p.funcMultiAssign(assign, expr)
			return
		}
	}
	switch {
	case leftLength > rightLength:
		p.pusherrtok(assign.Setter, "overflow_multiassign_identifiers")
		return
	case leftLength < rightLength:
		p.pusherrtok(assign.Setter, "missing_multiassign_identifiers")
		return
	}
	p.multiAssign(assign, exprs)
}

func (p *Parser) whileProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterWhile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	iter.Profile = profile
	if !isBoolExpr(val) {
		p.pusherrtok(iter.Tok, "iter_while_notbool_expr")
	}
	p.checkNewBlock(iter.Block)
}

func (p *Parser) foreachProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterForeach)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	profile.ExprType = val.data.Type
	if !isForeachIterExpr(val) {
		p.pusherrtok(iter.Tok, "iter_foreach_nonenumerable_expr")
	} else {
		fc := foreachChecker{p, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := p.blockVars
	if profile.KeyA.New {
		if xapi.IsIgnoreId(profile.KeyA.Id) {
			p.pusherrtok(profile.KeyA.IdTok, "ignore_id")
		}
		p.varStatement(&profile.KeyA, true)
	}
	if profile.KeyB.New {
		if xapi.IsIgnoreId(profile.KeyB.Id) {
			p.pusherrtok(profile.KeyB.IdTok, "ignore_id")
		}
		p.varStatement(&profile.KeyB, true)
	}
	p.checkNewBlockCustom(iter.Block, blockVars)
}

func (p *Parser) forProfile(iter *models.Iter) {
	profile := iter.Profile.(models.IterFor)
	blockVars := p.blockVars
	if profile.Once.Data != nil {
		_ = p.statement(&profile.Once, false)
	}
	if len(profile.Condition.Processes) > 0 {
		val, model := p.evalExpr(profile.Condition)
		profile.Condition.Model = model
		p.wg.Add(1)
		go assignChecker{
			p:      p,
			t:      DataType{Id: xtype.Bool, Kind: xtype.TypeMap[xtype.Bool]},
			v:      val,
			errtok: profile.Condition.Toks[0],
		}.checkAssignType()
	}
	if profile.Next.Data != nil {
		_ = p.statement(&profile.Next, false)
	}
	iter.Profile = profile
	p.checkNewBlock(iter.Block)
	p.blockVars = blockVars
}

func (p *Parser) iter(iter *models.Iter) {
	p.iterCount++
	defer func() { p.iterCount-- }()
	if iter.Profile != nil {
		switch iter.Profile.(type) {
		case models.IterWhile:
			p.whileProfile(iter)
		case models.IterForeach:
			p.foreachProfile(iter)
		case models.IterFor:
			p.forProfile(iter)
		}
	}
}

func (p *Parser) ifExpr(ifast *models.If, i *int, statements []models.Statement) {
	val, model := p.evalExpr(ifast.Expr)
	ifast.Expr.Model = model
	statement := statements[*i]
	if !isBoolExpr(val) {
		p.pusherrtok(ifast.Tok, "if_notbool_expr")
	}
	p.checkNewBlock(ifast.Block)
node:
	if statement.WithTerminator {
		return
	}
	*i++
	if *i >= len(statements) {
		*i--
		return
	}
	statement = statements[*i]
	switch t := statement.Data.(type) {
	case models.ElseIf:
		val, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		if !isBoolExpr(val) {
			p.pusherrtok(t.Tok, "if_notbool_expr")
		}
		p.checkNewBlock(t.Block)
		statements[*i].Data = t
		goto node
	case models.Else:
		p.elseBlock(&t)
		statement.Data = t
	default:
		*i--
	}
}

func (p *Parser) elseBlock(elseast *models.Else) {
	p.checkNewBlock(elseast.Block)
}

func (p *Parser) breakStatement(breakAST *models.Break) {
	if p.iterCount == 0 && p.caseCount == 0 {
		p.pusherrtok(breakAST.Tok, "break_at_outiter")
	}
}

func (p *Parser) continueStatement(continueAST *models.Continue) {
	if p.iterCount == 0 {
		p.pusherrtok(continueAST.Tok, "continue_at_outiter")
	}
}

func (p *Parser) checkValidityForAutoType(t DataType, errtok Tok) {
	if p.eval.hasError {
		return
	}
	switch t.Id {
	case xtype.Nil:
		p.pusherrtok(errtok, "nil_for_autotype")
	case xtype.Void:
		p.pusherrtok(errtok, "void_for_autotype")
	}
}

func (p *Parser) typeSourceOfMultiTyped(dt DataType, err bool) (DataType, bool) {
	types := dt.Tag.([]DataType)
	ok := false
	for i, t := range types {
		t, okr := p.typeSource(t, err)
		types[i] = t
		if ok {
			ok = okr
		}
	}
	dt.Tag = types
	return dt, ok
}

func (p *Parser) typeSourceIsType(dt DataType, t *Type, err bool) (DataType, bool) {
	original := dt.Original
	old := dt
	dt = t.Type
	dt.Tok = t.Tok
	dt.Original = original
	dt.Kind = t.Type.Kind
	dt, ok := p.typeSource(dt, err)
	if ok && typeIsArray(t.Type) && typeIsSlice(old) {
		p.pusherrtok(dt.Tok, "invalid_type_source")
	}
	return dt, ok
}

func (p *Parser) typeSourceIsEnum(e *Enum, tag any) (dt DataType, _ bool) {
	dt.Id = xtype.Enum
	dt.Kind = e.Id
	dt.Tag = e
	dt.Tok = e.Tok
	if tag != nil {
		p.pusherrtok(dt.Tok, "invalid_type_source")
	}
	return dt, true
}

func (p *Parser) typeSourceIsFunc(dt DataType, err bool) (DataType, bool) {
	f := dt.Tag.(*Func)
	p.reloadFuncTypes(f)
	dt.Kind = f.DataTypeString()
	return dt, true
}

func (p *Parser) typeSourceIsMap(dt DataType, err bool) (DataType, bool) {
	types := dt.Tag.([]DataType)
	key := &types[0]
	*key, _ = p.realType(*key, err)
	value := &types[1]
	*value, _ = p.realType(*value, err)
	dt.Kind = dt.MapKind()
	return dt, true
}

func (p *Parser) typeSourceIsStruct(s *xstruct, tag any, errTok Tok) (dt DataType, _ bool) {
	var generics []DataType
	// Has generics?
	if tag != nil {
		generics = tag.([]DataType)
		_ = p.checkGenericsQuantity(len(s.Ast.Generics), generics, errTok)
		blockTypes := p.blockTypes
		defer func() { p.blockTypes = blockTypes }()
		p.pushGenerics(s.Ast.Generics, generics)
		for i, generic := range generics {
			generics[i], _ = p.typeSource(generic, true)
		}
	} else if len(s.Ast.Generics) > 0 {
		p.pusherrtok(errTok, "has_generics")
	}
	s = p.structConstructorInstance(s)
	s.SetGenerics(generics)
	dt.Id = xtype.Struct
	dt.Kind = s.dataTypeString()
	dt.Tag = s
	dt.Tok = s.Ast.Tok
	return dt, true
}

func (p *Parser) typeSourceIsTrait(t *trait, tag any, errTok Tok) (dt DataType, _ bool) {
	if tag != nil {
		p.pusherrtok(errTok, "invalid_type_source")
	}
	dt.Id = xtype.Trait
	dt.Kind = t.Ast.Id
	dt.Tag = t
	dt.Tok = t.Ast.Tok
	dt.DontUseOriginal = true
	return dt, true
}

func (p *Parser) tokenizeDataType(id string) []Tok {
	parts := strings.SplitN(id, tokens.DOUBLE_COLON, -1)
	var toks []Tok
	for i, part := range parts {
		toks = append(toks, Tok{
			Id:   tokens.Id,
			Kind: part,
			File: p.File,
		})
		if i%2 == 0 {
			toks = append(toks, Tok{
				Id:   tokens.DoubleColon,
				Kind: tokens.DOUBLE_COLON,
				File: p.File,
			})
		}
	}
	return toks
}

func (p *Parser) typeSource(dt DataType, err bool) (ret DataType, ok bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	if dt.Kind == "" {
		return dt, true
	}
	if dt.MultiTyped {
		return p.typeSourceOfMultiTyped(dt, err)
	} else if typeIsMap(dt) {
		return p.typeSourceIsMap(dt, err)
	}
	if typeIsArray(dt) {
		p.checkArrayType(&dt)
	}
	switch dt.Id {
	case xtype.Id:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		var def any
		if strings.Contains(id, tokens.DOUBLE_COLON) { // Has namespace?
			toks := p.tokenizeDataType(id)
			defs := p.eval.getNsDefs(&toks, nil)
			if defs == nil {
				return
			}
			i, defs, t := defs.defById(toks[0].Kind, p.File)
			switch t {
			case 't':
				def = defs.Types[i]
			case 's':
				def = defs.Structs[i]
			case 'e':
				def = defs.Enums[i]
			case 'i':
				def = defs.Traits[i]
			}
		} else {
			def, _, _, _ = p.defById(id)
		}
		switch t := def.(type) {
		case *Type:
			t.Used = true
			return p.typeSourceIsType(dt, t, err)
		case *Enum:
			t.Used = true
			return p.typeSourceIsEnum(t, dt.Tag)
		case *xstruct:
			t.Used = true
			return p.typeSourceIsStruct(t, dt.Tag, dt.Tok)
		case *trait:
			t.Used = true
			return p.typeSourceIsTrait(t, dt.Tag, dt.Tok)
		default:
			if err {
				p.pusherrtok(dt.Tok, "invalid_type_source")
			}
			return dt, false
		}
	case xtype.Func:
		return p.typeSourceIsFunc(dt, err)
	}
	return dt, true
}

func (p *Parser) realType(dt DataType, err bool) (ret DataType, _ bool) {
	original := dt.Original
	defer func() { ret.Original = original }()
	dt.SetToOriginal()
	return p.typeSource(dt, err)
}

func (p *Parser) checkMultiType(real, check DataType, ignoreAny bool, errTok Tok) {
	defer p.wg.Done()
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
		return
	}
	realTypes := real.Tag.([]DataType)
	checkTypes := real.Tag.([]DataType)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		p.checkType(realType, checkType, ignoreAny, errTok)
	}
}

func (p *Parser) checkAssignConst(constant bool, t DataType, val value, errTok Tok) {
	if typeIsMut(t) && val.constant && !constant {
		p.pusherrtok(errTok, "constant_assignto_nonconstant")
	}
}

func (p *Parser) checkType(real, check DataType, ignoreAny bool, errTok Tok) {
	defer p.wg.Done()
	if typeIsVoid(check) {
		p.eval.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
		return
	}
	if !ignoreAny && real.Id == xtype.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.wg.Add(1)
		go p.checkMultiType(real, check, ignoreAny, errTok)
		return
	}
	switch {
	case typesAreCompatible(real, check, ignoreAny),
		typeIsNilCompatible(real) && check.Id == xtype.Nil,
		typeIsSinglePtr(real) && !typeIsPtr(check):
		return
	}
	if real.Kind != check.Kind {
		p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
	} else if typeIsArray(real) || typeIsArray(check) {
		if typeIsArray(real) != typeIsArray(check) {
			p.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
			return
		}
		i := real.Tag.([][]any)[0][0].(uint64)
		j := check.Tag.([][]any)[0][0].(uint64)
		realKind := strings.Replace(real.Kind, x.Mark_Array, strconv.FormatUint(i, 10), 1)
		checkKind := strings.Replace(check.Kind, x.Mark_Array, strconv.FormatUint(j, 10), 1)
		p.pusherrtok(errTok, "incompatible_datatype", realKind, checkKind)
	}
}

func (p *Parser) evalExpr(expr Expr) (value, iExpr) {
	p.eval.hasError = false
	return p.eval.expr(expr)
}

func (p *Parser) evalToks(toks Toks) (value, iExpr) {
	p.eval.hasError = false
	return p.eval.toks(toks)
}
