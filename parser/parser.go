package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/jule-lang/jule/ast"
	"github.com/jule-lang/jule/ast/models"
	"github.com/jule-lang/jule/lex"
	"github.com/jule-lang/jule/lex/tokens"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juleio"
	"github.com/jule-lang/jule/pkg/julelog"
	"github.com/jule-lang/jule/pkg/juletype"
	"github.com/jule-lang/jule/preprocessor"
)

type File = juleio.File
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

// Parser is parser of Jule code.
type Parser struct {
	attributes     []Attribute
	docText        strings.Builder
	isNowIntoIter  bool
	currentCase    *models.Case
	wg             sync.WaitGroup
	rootBlock      *models.Block
	nodeBlock      *models.Block
	generics       []*GenericType
	blockTypes     []*Type
	blockVars      []*Var
	waitingGlobals []waitingGlobal
	eval           *eval
	cppLinks       []*models.CppLink
	allowBuiltin   bool

	NoLocalPkg bool
	JustDefs   bool
	NoCheck    bool
	IsMain     bool
	Uses       []*use
	Defs       *Defmap
	Errors     []julelog.CompilerLog
	Warnings   []julelog.CompilerLog
	File       *File
}

// New returns new instance of Parser.
func New(f *File) *Parser {
	p := new(Parser)
	p.File = f
	p.allowBuiltin = true
	p.Defs = new(Defmap)
	p.eval = new(eval)
	p.eval.p = p
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok Tok, key string, args ...any) {
	p.pusherrmsgtok(tok, jule.GetError(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok Tok, msg string) {
	p.Errors = append(p.Errors, julelog.CompilerLog{
		Type:    julelog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok Tok, key string, args ...any) {
	p.Warnings = append(p.Warnings, julelog.CompilerLog{
		Type:    julelog.Warning,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path(),
		Message: jule.GetWarning(key, args...),
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
		Type:    julelog.FlatError,
		Message: msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string, args ...any) {
	p.Warnings = append(p.Warnings, julelog.CompilerLog{
		Type:    julelog.FlatWarning,
		Message: jule.GetWarning(key, args...),
	})
}

// CppLinks returns cpp code of cpp links.
func (p *Parser) CppLinks() string {
	var cpp strings.Builder
	for _, use := range used {
		if use.cppLink {
			cpp.WriteString(`#include "`)
			cpp.WriteString(use.Path)
			cpp.WriteString("\"\n")
		}
	}
	return cpp.String()
}

func cppTypes(dm *Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Types {
		if t.Used && t.Tok.Id != tokens.NA {
			cpp.WriteString(t.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppTypes returns cpp code of types.
func (p *Parser) CppTypes() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTypes(use.defs))
		}
	}
	cpp.WriteString(cppTypes(p.Defs))
	return cpp.String()
}

func cppTraits(dm *Defmap) string {
	var cpp strings.Builder
	for _, t := range dm.Traits {
		if t.Used && t.Ast.Tok.Id != tokens.NA {
			cpp.WriteString(t.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppTraits returns cpp code of traits.
func (p *Parser) CppTraits() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppTraits(use.defs))
		}
	}
	cpp.WriteString(cppTraits(p.Defs))
	return cpp.String()
}

func cppStructs(dm *Defmap) string {
	var cpp strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Tok.Id != tokens.NA {
			cpp.WriteString(s.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppStructs returns cpp code of structures.
func (p *Parser) CppStructs() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructs(use.defs))
		}
	}
	cpp.WriteString(cppStructs(p.Defs))
	return cpp.String()
}

func cppStructPrototypes(dm *Defmap) string {
	var cpp strings.Builder
	for _, s := range dm.Structs {
		if s.Used && s.Ast.Tok.Id != tokens.NA {
			cpp.WriteString(s.prototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

func cppFuncPrototypes(dm *Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cpp.WriteString(f.Prototype())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppPrototypes returns cpp code of prototypes.
func (p *Parser) CppPrototypes() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppStructPrototypes(use.defs))
		}
	}
	cpp.WriteString(cppStructPrototypes(p.Defs))
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncPrototypes(use.defs))
		}
	}
	cpp.WriteString(cppFuncPrototypes(p.Defs))
	return cpp.String()
}

func cppGlobals(dm *Defmap) string {
	var cpp strings.Builder
	for _, g := range dm.Globals {
		if !g.Const && g.Used && g.Token.Id != tokens.NA {
			cpp.WriteString(g.String())
			cpp.WriteByte('\n')
		}
	}
	return cpp.String()
}

// CppGlobals returns cpp code of global variables.
func (p *Parser) CppGlobals() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppGlobals(use.defs))
		}
	}
	cpp.WriteString(cppGlobals(p.Defs))
	return cpp.String()
}

func cppFuncs(dm *Defmap) string {
	var cpp strings.Builder
	for _, f := range dm.Funcs {
		if f.used && f.Ast.Tok.Id != tokens.NA {
			cpp.WriteString(f.String())
			cpp.WriteString("\n\n")
		}
	}
	return cpp.String()
}

// CppFuncs returns cpp code of functions.
func (p *Parser) CppFuncs() string {
	var cpp strings.Builder
	for _, use := range used {
		if !use.cppLink {
			cpp.WriteString(cppFuncs(use.defs))
		}
	}
	cpp.WriteString(cppFuncs(p.Defs))
	return cpp.String()
}

// CppInitializerCaller returns cpp code of initializer caller.
func (p *Parser) CppInitializerCaller() string {
	var cpp strings.Builder
	cpp.WriteString("void ")
	cpp.WriteString(juleapi.InitializerCaller)
	cpp.WriteString("(void) {")
	models.AddIndent()
	indent := models.IndentString()
	models.DoneIndent()
	pushInit := func(defs *Defmap) {
		f, _, _ := defs.funcById(jule.InitializerFunction, nil)
		if f == nil {
			return
		}
		cpp.WriteByte('\n')
		cpp.WriteString(indent)
		cpp.WriteString(f.outId())
		cpp.WriteString("();")
	}
	for _, use := range used {
		if !use.cppLink {
			pushInit(use.defs)
		}
	}
	pushInit(p.Defs)
	cpp.WriteString("\n}")
	return cpp.String()
}

// Cpp returns full cpp code of parsed objects.
func (p *Parser) Cpp() string {
	var cpp strings.Builder
	cpp.WriteString(p.CppLinks())
	cpp.WriteByte('\n')
	cpp.WriteString(p.CppTypes())
	cpp.WriteByte('\n')
	cpp.WriteString(p.CppTraits())
	cpp.WriteString(p.CppPrototypes())
	cpp.WriteString(p.CppStructs())
	cpp.WriteString("\n\n")
	cpp.WriteString(p.CppGlobals())
	cpp.WriteString("\n\n")
	cpp.WriteString(p.CppFuncs())
	cpp.WriteString(p.CppInitializerCaller())
	return cpp.String()
}

func getTree(toks Toks) ([]models.Object, []julelog.CompilerLog) {
	b := ast.NewBuilder(toks)
	b.Build()
	return b.Tree, b.Errors
}

func (p *Parser) checkCppUsePath(use *models.Use) bool {
	ext := filepath.Ext(use.Path)
	if !juleapi.IsValidHeader(ext) {
		p.pusherrtok(use.Tok, "invalid_header_ext", ext)
		return false
	}
	err := os.Chdir(use.Tok.File.Dir)
	if err != nil {
		p.pusherrtok(use.Tok, "use_not_found", use.Path)
		return false
	}
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || info.IsDir() {
		p.pusherrtok(use.Tok, "use_not_found", use.Path)
		return false
	}
	// Set to absolute path for correct include path
	use.Path, _ = filepath.Abs(use.Path)
	_ = os.Chdir(jule.ExecPath)
	return true
}

func (p *Parser) checkPureUsePath(use *models.Use) bool {
	info, err := os.Stat(use.Path)
	// Exist?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Tok, "use_not_found", use.Path)
		return false
	}
	return true
}

func (p *Parser) checkUsePath(use *models.Use) bool {
	if use.Cpp {
		if !p.checkCppUsePath(use) {
			return false
		}
	} else {
		if !p.checkPureUsePath(use) {
			return false
		}
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

func (p *Parser) pushSelects(use *use, selectors []Tok) (addNs bool) {
	if len(selectors) > 0 {
		p.Defs.side = new(Defmap)
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
		if id.Id == tokens.Self {
			addNs = true
			continue
		}
		i, m, t := use.defs.findById(id.Kind, p.File)
		if i == -1 {
			p.pusherrtok(id, "id_noexist", id.Kind)
			continue
		}
		switch t {
		case 'i':
			p.Defs.side.Traits = append(p.Defs.side.Traits, m.Traits[i])
		case 'f':
			p.Defs.side.Funcs = append(p.Defs.side.Funcs, m.Funcs[i])
		case 'e':
			p.Defs.side.Enums = append(p.Defs.side.Enums, m.Enums[i])
		case 'g':
			p.Defs.side.Globals = append(p.Defs.side.Globals, m.Globals[i])
		case 't':
			p.Defs.side.Types = append(p.Defs.side.Types, m.Types[i])
		case 's':
			p.Defs.side.Structs = append(p.Defs.side.Structs, m.Structs[i])
		}
	}
	return
}

func (p *Parser) pushUse(use *use, selectors []Tok) {
	if len(selectors) > 0 {
		if !p.pushSelects(use, selectors) {
			return
		}
	} else if selectors != nil {
		return
	} else if use.fullUse {
		if p.Defs.side == nil {
			p.Defs.side = new(Defmap)
		}
		p.pushDefs(p.Defs.side, use.defs)
	}
	ns := new(models.Namespace)
	ns.Ids = strings.SplitN(use.LinkString, tokens.DOUBLE_COLON, -1)
	src := p.pushNs(ns)
	src.defs = use.defs
}

func (p *Parser) compileCppLinkUse(useAST *models.Use) (*use, bool) {
	use := new(use)
	use.cppLink = true
	use.Path = useAST.Path
	use.tok = useAST.Tok
	return use, false
}

func (p *Parser) compilePureUse(useAST *models.Use) (_ *use, hassErr bool) {
	infos, err := ioutil.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil, true
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, jule.SrcExt) ||
			!juleio.IsUseable(name) {
			continue
		}
		f, err := juleio.Openfx(filepath.Join(useAST.Path, name))
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
		use.fullUse = useAST.FullUse
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		p.pushDefs(use.defs, psub.Defs)
		p.pushUse(use, useAST.Selectors)
		if psub.Errors != nil {
			p.pusherrtok(useAST.Tok, "use_has_errors")
			return use, true
		}
		return use, false
	}
	return nil, false
}

func (p *Parser) compileUse(useAST *models.Use) (*use, bool) {
	if useAST.Cpp {
		return p.compileCppLinkUse(useAST)
	}
	return p.compilePureUse(useAST)
}

func (p *Parser) pushDefs(dest, src *Defmap) {
	dest.Types = append(dest.Types, src.Types...)
	dest.Traits = append(dest.Traits, src.Traits...)
	dest.Structs = append(dest.Structs, src.Structs...)
	dest.Enums = append(dest.Enums, src.Enums...)
	dest.Globals = append(dest.Globals, src.Globals...)
	dest.Funcs = append(dest.Funcs, src.Funcs...)
}

func (p *Parser) use(useAST *models.Use) (err bool) {
	if !p.checkUsePath(useAST) {
		return true
	}
	// Already parsed?
	for _, use := range used {
		if useAST.Path == use.Path {
			p.pushUse(use, nil)
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
	case models.CppLink:
		p.CppLink(t)
	case models.Comment:
		p.Comment(t)
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
			!strings.HasSuffix(name, jule.SrcExt) ||
			!juleio.IsUseable(name) ||
			name == p.File.Name {
			continue
		}
		f, err := juleio.Openfx(filepath.Join(p.File.Dir, name))
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

// Parses Jule code from object tree.
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

// Parses Jule code from tokens.
func (p *Parser) Parse(toks Toks, main, justDefs bool) {
	tree, errors := getTree(toks)
	if len(errors) > 0 {
		p.pusherrs(errors...)
		return
	}
	p.Parset(tree, main, justDefs)
}

// Parses Jule code from file.
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
		if juleapi.IsIgnoreId(generic.Id) {
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

// Type parses Jule type define statement.
func (p *Parser) Type(t Type) {
	_, tok, canshadow := p.defById(t.Id)
	if tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return
	} else if juleapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return
	}
	t.Desc = p.docText.String()
	p.docText.Reset()
	p.Defs.Types = append(p.Defs.Types, &t)
}

// Enum parses Jule enumerator statement.
func (p *Parser) Enum(e Enum) {
	if juleapi.IsIgnoreId(e.Id) {
		p.pusherrtok(e.Tok, "ignore_id")
		return
	} else if _, tok, _ := p.defById(e.Id); tok.Id != tokens.NA {
		p.pusherrtok(e.Tok, "exist_id", e.Id)
		return
	}
	e.Desc = p.docText.String()
	p.docText.Reset()
	e.Type, _ = p.realType(e.Type, true)
	if !typeIsPure(e.Type) || !juletype.IsInteger(e.Type.Id) {
		p.pusherrtok(e.Type.Tok, "invalid_type_source")
		return
	}
	pdefs := p.Defs
	puses := p.Uses
	p.Defs = new(Defmap)
	defer func() {
		p.Defs = pdefs
		p.Uses = puses
		p.Defs.Enums = append(p.Defs.Enums, &e)
	}()
	max := juletype.MaxOfType(e.Type.Id)
	for i, item := range e.Items {
		if max == 0 {
			p.pusherrtok(item.Tok, "overflow_limits")
		} else {
			max--
		}
		if juleapi.IsIgnoreId(item.Id) {
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
			if !val.constExpr && !p.eval.hasError {
				p.pusherrtok(item.Expr.Toks[0], "expr_not_const")
			}
			item.ExprTag = val.expr
			item.Expr.Model = model
			assignChecker{
				p:         p,
				t:         e.Type,
				v:         val,
				ignoreAny: true,
				errtok:    item.Tok,
			}.checkAssignType()
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
		itemVar.Token = e.Tok
		p.Defs.Globals = append(p.Defs.Globals, itemVar)
	}
}

func (p *Parser) pushField(s *structure, f **Var, i int) {
	for _, cf := range s.Ast.Fields {
		if *f == cf {
			break
		}
		if (*f).Id == cf.Id {
			p.pusherrtok((*f).Token, "exist_id", (*f).Id)
			break
		}
	}
	if len(s.Ast.Generics) == 0 {
		p.parseField(s, f, i)
	} else {
		p.parseNonGenericType(s.Ast.Generics, &(*f).Type)
		param := models.Param{Id: (*f).Id, Type: (*f).Type}
		param.Default.Model = exprNode{juleapi.DefaultExpr}
		s.constructor.Params[i] = param
	}
}

func (p *Parser) parseFields(s *structure) {
	s.constructor = new(Func)
	s.constructor.Id = s.Ast.Id
	s.constructor.Tok = s.Ast.Tok
	s.constructor.Params = make([]models.Param, len(s.Ast.Fields))
	s.constructor.RetType.Type = DataType{
		Id:   juletype.Struct,
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
		p.pushField(s, &f, i)
		s.Defs.Globals[i] = f
	}
}

// Struct parses Jule structure.
func (p *Parser) Struct(s Struct) {
	if juleapi.IsIgnoreId(s.Id) {
		p.pusherrtok(s.Tok, "ignore_id")
		return
	} else if _, tok, _ := p.defById(s.Id); tok.Id != tokens.NA {
		p.pusherrtok(s.Tok, "exist_id", s.Id)
		return
	}
	xs := new(structure)
	p.Defs.Structs = append(p.Defs.Structs, xs)
	xs.Desc = p.docText.String()
	p.docText.Reset()
	xs.Ast = s
	xs.Ast.Owner = p
	xs.Ast.Generics = p.generics
	p.generics = nil
	xs.Defs = new(Defmap)
	p.parseFields(xs)
}

func (p *Parser) checkCppLinkAttributes(f *Func) {
	for _, attribute := range f.Attributes {
		switch attribute.Tag {
		case jule.Attribute_CDef:
		default:
			p.pusherrtok(attribute.Tok, "invalid_attribute")
		}
	}
}

// CppLink parses cpp link.
func (p *Parser) CppLink(link models.CppLink) {
	if juleapi.IsIgnoreId(link.Link.Id) {
		p.pusherrtok(link.Tok, "ignore_id")
		return
	} else if p.linkById(link.Link.Id) != nil {
		p.pusherrtok(link.Tok, "exist_id", link.Link.Id)
		return
	}
	linkf := link.Link
	linkf.Owner = p
	setGenerics(linkf, p.generics)
	p.generics = nil
	linkf.Attributes = p.attributes
	p.attributes = nil
	p.checkCppLinkAttributes(linkf)
	p.cppLinks = append(p.cppLinks, &link)
}

// Trait parses Jule trait.
func (p *Parser) Trait(t models.Trait) {
	if juleapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return
	} else if _, tok, _ := p.defById(t.Id); tok.Id != tokens.NA {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return
	}
	trait := new(trait)
	trait.Desc = p.docText.String()
	p.docText.Reset()
	trait.Ast = &t
	trait.Defs = new(Defmap)
	trait.Defs.Funcs = make([]*function, len(t.Funcs))
	for i, f := range trait.Ast.Funcs {
		if juleapi.IsIgnoreId(f.Id) {
			p.pusherrtok(f.Tok, "ignore_id")
		}
		for j, jf := range trait.Ast.Funcs {
			if j >= i {
				break
			} else if f.Id == jf.Id {
				p.pusherrtok(f.Tok, "exist_id", f.Id)
			}
		}
		_ = p.checkParamDup(f.Params)
		p.parseTypesNonGenerics(f)
		tf := new(function)
		tf.Ast = f
		trait.Defs.Funcs[i] = tf
	}
	p.Defs.Traits = append(p.Defs.Traits, trait)
}

func (p *Parser) implTrait(impl models.Impl) {
	trait, _, _ := p.traitById(impl.Trait.Kind)
	if trait == nil {
		p.pusherrtok(impl.Trait, "id_noexist", impl.Trait.Kind)
		return
	}
	trait.Used = true
	sid, _ := impl.Target.KindId()
	xs, _, _ := p.Defs.structById(sid, nil)
	if xs == nil {
		p.pusherrtok(impl.Target.Tok, "id_noexist", sid)
		return
	}
	impl.Target.Tag = xs
	xs.traits = append(xs.traits, trait)
	for _, tf := range trait.Defs.Funcs {
		ok := false
		ds := tf.Ast.DefString()
		for _, obj := range impl.Tree {
			switch t := obj.Data.(type) {
			case *Func:
				if tf.Ast.Pub == t.Pub && ds == t.DefString() {
					ok = true
					break
				}
			}
		}
		if !ok {
			p.pusherrtok(impl.Target.Tok, "notimpl_trait_def", trait.Ast.Id, ds)
		}
	}
	for _, obj := range impl.Tree {
		switch t := obj.Data.(type) {
		case models.Attribute:
			p.PushAttribute(t)
		case models.Comment:
			p.Comment(t)
		case *Func:
			if trait.FindFunc(t.Id) == nil {
				p.pusherrtok(impl.Target.Tok, "trait_hasnt_id", trait.Ast.Id, t.Id)
				break
			}
			i, _, _ := xs.Defs.findById(t.Id, nil)
			if i != -1 {
				p.pusherrtok(t.Tok, "exist_id", t.Id)
				continue
			}
			sf := new(function)
			sf.Ast = t
			sf.Ast.Receiver.Tok = xs.Ast.Tok
			sf.Ast.Receiver.Tag = xs
			sf.Ast.Attributes = p.attributes
			sf.Ast.Owner = p
			p.attributes = nil
			sf.Desc = p.docText.String()
			p.docText.Reset()
			sf.used = true
			if len(xs.Ast.Generics) == 0 {
				p.parseTypesNonGenerics(sf.Ast)
			}
			xs.Defs.Funcs = append(xs.Defs.Funcs, sf)
		}
	}
}

func (p *Parser) implStruct(impl models.Impl) {
	xs, _, _ := p.Defs.structById(impl.Trait.Kind, nil)
	if xs == nil {
		p.pusherrtok(impl.Trait, "id_noexist", impl.Trait.Kind)
		return
	}
	for _, obj := range impl.Tree {
		switch t := obj.Data.(type) {
		case models.Attribute:
			p.PushAttribute(t)
		case []GenericType:
			p.Generics(t)
		case models.Comment:
			p.Comment(t)
		case *Func:
			i, _, _ := xs.Defs.findById(t.Id, nil)
			if i != -1 {
				p.pusherrtok(t.Tok, "exist_id", t.Id)
				continue
			}
			sf := new(function)
			sf.Ast = t
			sf.Ast.Receiver.Tok = xs.Ast.Tok
			sf.Ast.Receiver.Tag = xs
			sf.Ast.Attributes = p.attributes
			sf.Desc = p.docText.String()
			sf.Ast.Owner = p
			p.docText.Reset()
			p.attributes = nil
			setGenerics(sf.Ast, p.generics)
			p.generics = nil
			for _, generic := range t.Generics {
				if findGeneric(generic.Id, xs.Ast.Generics) != nil {
					p.pusherrtok(generic.Tok, "exist_id", generic.Id)
				}
			}
			if len(xs.Ast.Generics) == 0 {
				p.parseTypesNonGenerics(sf.Ast)
			}
			xs.Defs.Funcs = append(xs.Defs.Funcs, sf)
		}
	}
}

// Impl parses Jule impl.
func (p *Parser) Impl(impl models.Impl) {
	if !typeIsVoid(impl.Target) {
		p.implTrait(impl)
		return
	}
	p.implStruct(impl)
}

// pushNS pushes namespace to defmap and returns leaf namespace.
func (p *Parser) pushNs(ns *models.Namespace) *namespace {
	var src *namespace
	prev := p.Defs
	for _, id := range ns.Ids {
		src = prev.nsById(id)
		if src == nil {
			src = new(namespace)
			src.Id = id
			src.Tok = ns.Tok
			src.defs = new(Defmap)
			prev.Namespaces = append(prev.Namespaces, src)
		}
		prev = src.defs
	}
	return src
}

// Comment parses Jule documentation comments line.
func (p *Parser) Comment(c models.Comment) {
	c.Content = strings.TrimSpace(c.Content)
	if p.docText.Len() == 0 {
		if strings.HasPrefix(c.Content, jule.DocPrefix) {
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
	for _, kind := range jule.Attributes {
		if attribute.Tag == kind {
			ok = true
			break
		}
	}
	if !ok {
		p.pusherrtok(attribute.Tok, "undefined_attribute")
	}
	for _, attr := range p.attributes {
		if attr.Tag == attribute.Tag {
			p.pusherrtok(attribute.Tok, "attribute_repeat")
			return
		}
	}
	p.attributes = append(p.attributes, attribute)
}

func genericsToCpp(generics []*GenericType) string {
	if len(generics) == 0 {
		return ""
	}
	var cpp strings.Builder
	cpp.WriteString("template<")
	for _, generic := range generics {
		cpp.WriteString(generic.String())
		cpp.WriteByte(',')
	}
	return cpp.String()[:cpp.Len()-1] + ">"
}

// Statement parse Jule statement.
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
	if t.Id == juletype.Id {
		id, prefix := t.KindId()
		def, _, _ := p.defById(id)
		switch deft := def.(type) {
		case *structure:
			deft = p.structConstructorInstance(deft)
			if t.Tag != nil {
				deft.SetGenerics(t.Tag.([]DataType))
			}
			t.Kind = prefix + deft.dataTypeString()
			t.Id = juletype.Struct
			t.Tag = deft
			t.Pure = true
			t.Original = nil
			goto tagcheck
		}
	}
	if typeIsGeneric(generics, *t) {
		return
	}
tagcheck:
	if t.Tag != nil {
		switch t := t.Tag.(type) {
		case *structure:
			for _, ct := range t.Generics() {
				if typeIsGeneric(generics, ct) {
					return
				}
			}
		case []DataType:
			for _, ct := range t {
				if typeIsGeneric(generics, ct) {
					return
				}
			}
		}
	}
	*t, _ = p.realType(*t, true)
}

func (p *Parser) parseNonGenericType(generics []*GenericType, t *DataType) {
	switch {
	case t.MultiTyped:
		p.parseMultiNonGenericType(generics, t)
	case typeIsFunc(*t):
		p.parseFuncNonGenericType(generics, t)
	case typeIsMap(*t):
		p.parseMapNonGenericType(generics, t)
	case typeIsArray(*t):
		p.parseNonGenericType(generics, t.ComponentType)
		t.Kind = jule.Prefix_Array + t.ComponentType.Kind
	case typeIsSlice(*t):
		p.parseNonGenericType(generics, t.ComponentType)
		t.Kind = jule.Prefix_Slice + t.ComponentType.Kind
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
		if juleapi.IsIgnoreId(v.Kind) {
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

func setGenerics(f *Func, generics []*models.GenericType) {
	f.Generics = generics
	if len(f.Generics) > 0 {
		f.Combines = new([][]models.DataType)
	}
}

// Func parse Jule function.
func (p *Parser) Func(fast Func) {
	_, tok, canshadow := p.defById(fast.Id)
	if tok.Id != tokens.NA && !canshadow {
		p.pusherrtok(fast.Tok, "exist_id", fast.Id)
	} else if juleapi.IsIgnoreId(fast.Id) {
		p.pusherrtok(fast.Tok, "ignore_id")
	}
	f := new(function)
	f.Ast = new(Func)
	*f.Ast = fast
	f.Ast.Attributes = p.attributes
	p.attributes = nil
	f.Ast.Owner = p
	f.Desc = p.docText.String()
	p.docText.Reset()
	setGenerics(f.Ast, p.generics)
	p.generics = nil
	p.checkRetVars(f)
	p.checkFuncAttributes(f)
	f.used = f.Ast.Id == jule.InitializerFunction
	p.parseTypesNonGenerics(f.Ast)
	p.Defs.Funcs = append(p.Defs.Funcs, f)
}

// ParseVariable parse Jule global variable.
func (p *Parser) Global(vast Var) {
	def, _, _ := p.defById(vast.Id)
	if def != nil {
		p.pusherrtok(vast.Token, "exist_id", vast.Id)
		return
	} else {
		for _, g := range p.waitingGlobals {
			if vast.Id == g.Var.Id {
				p.pusherrtok(vast.Token, "exist_id", vast.Id)
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

// Var parse Jule variable.
func (p *Parser) Var(v Var) *Var {
	if juleapi.IsIgnoreId(v.Id) {
		p.pusherrtok(v.Token, "ignore_id")
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
	if v.Type.Id != juletype.Void {
		t, ok := p.realType(v.Type, true)
		if ok {
			v.Type = t
			if v.SetterTok.Id != tokens.NA {
				assignChecker{
					p:      p,
					t:      v.Type,
					v:      val,
					errtok: v.Token,
				}.checkAssignType()
			}
		}
	} else {
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.Token, "missing_autotype_value")
		} else {
			p.eval.hasError = p.eval.hasError || val.data.Value == ""
			v.Type = val.data.Type
			if val.constExpr && typeIsPure(v.Type) {
				switch val.expr.(type) {
				case int64:
					dt := DataType{
						Id:   juletype.Int,
						Kind: juletype.TypeMap[juletype.Int],
					}
					if integerAssignable(dt, val) {
						v.Type = dt
					}
				case uint64:
					dt := DataType{
						Id:   juletype.UInt,
						Kind: juletype.TypeMap[juletype.UInt],
					}
					if integerAssignable(dt, val) {
						v.Type = dt
					}
				}
			}
			p.checkValidityForAutoType(v.Type, v.SetterTok)
		}
	}
	if v.Const {
		v.ExprTag = val.expr
		if !typeIsAllowForConst(v.Type) {
			p.pusherrtok(v.Token, "invalid_type_for_const", v.Type.Kind)
		}
		if v.SetterTok.Id == tokens.NA {
			p.pusherrtok(v.Token, "missing_const_value")
		} else {
			if !validExprForConst(val) {
				p.eval.pusherrtok(v.Token, "expr_not_const")
			}
		}
	}
	return &v
}

func (p *Parser) checkTypeParam(f *function) {
	if len(f.Ast.Generics) == 0 {
		p.pusherrtok(f.Ast.Tok, "func_must_have_generics_if_has_attribute", jule.Attribute_TypeArg)
	}
	if len(f.Ast.Params) != 0 {
		p.pusherrtok(f.Ast.Tok, "func_cant_have_params_if_has_attribute", jule.Attribute_TypeArg)
	}
}

func (p *Parser) checkFuncAttributes(f *function) {
	for _, attribute := range f.Ast.Attributes {
		switch attribute.Tag {
		case jule.Attribute_Inline:
		case jule.Attribute_TypeArg:
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
		v.IsField = true
		v.Id = param.Id
		v.Token = param.Tok
		v.Type = param.Type
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Tok, "variadic_parameter_notlast")
			}
			v.Type.Original = nil
			v.Type.ComponentType = new(models.DataType)
			*v.Type.ComponentType = param.Type
			v.Type.Id = juletype.Slice
			v.Type.Kind = jule.Prefix_Slice + v.Type.Kind
		}
		vars[i] = v
	}
	return vars
}

func (p *Parser) linkById(id string) *models.CppLink {
	for _, link := range p.cppLinks {
		if link.Link.Id == id {
			return link
		}
	}
	return nil
}

// FuncById returns function by specified id.
//
// Special case:
//  FuncById(id) -> nil: if function is not exist.
func (p *Parser) FuncById(id string) (*function, *Defmap, bool) {
	if p.allowBuiltin {
		f, _, _ := Builtin.funcById(id, nil)
		if f != nil {
			return f, nil, false
		}
	}
	return p.Defs.funcById(id, p.File)
}

func (p *Parser) globalById(id string) (*Var, *Defmap, bool) {
	g, m, _ := p.Defs.globalById(id, p.File)
	return g, m, true
}

func (p *Parser) nsById(id string) *namespace {
	return p.Defs.nsById(id)
}

func (p *Parser) typeById(id string) (*Type, *Defmap, bool) {
	t := p.blockTypeById(id)
	if t != nil {
		return t, nil, false
	}
	if p.allowBuiltin {
		t, _, _ = Builtin.typeById(id, nil)
		if t != nil {
			return t, nil, false
		}
	}
	return p.Defs.typeById(id, p.File)
}

func (p *Parser) enumById(id string) (*Enum, *Defmap, bool) {
	if p.allowBuiltin {
		s, _, _ := Builtin.enumById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return p.Defs.enumById(id, p.File)
}

func (p *Parser) structById(id string) (*structure, *Defmap, bool) {
	if p.allowBuiltin {
		s, _, _ := Builtin.structById(id, nil)
		if s != nil {
			return s, nil, false
		}
	}
	return p.Defs.structById(id, p.File)
}

func (p *Parser) traitById(id string) (*trait, *Defmap, bool) {
	if p.allowBuiltin {
		t, _, _ := Builtin.traitById(id, nil)
		if t != nil {
			return t, nil, false
		}
	}
	return p.Defs.traitById(id, p.File)
}

func (p *Parser) blockTypeById(id string) *Type {
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

func (p *Parser) defById(id string) (def any, tok Tok, canshadow bool) {
	var t *Type
	t, _, canshadow = p.typeById(id)
	if t != nil {
		return t, t.Tok, canshadow
	}
	var e *Enum
	e, _, canshadow = p.enumById(id)
	if e != nil {
		return e, e.Tok, canshadow
	}
	var s *structure
	s, _, canshadow = p.structById(id)
	if s != nil {
		return s, s.Ast.Tok, canshadow
	}
	var trait *trait
	trait, _, canshadow = p.traitById(id)
	if trait != nil {
		return trait, trait.Ast.Tok, canshadow
	}
	var f *function
	f, _, canshadow = p.FuncById(id)
	if f != nil {
		return f, f.Ast.Tok, canshadow
	}
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, bv.Token, false
	}
	g, _, _ := p.globalById(id)
	if g != nil {
		return g, g.Token, true
	}
	return
}

func (p *Parser) blockDefById(id string) (def any, tok Tok) {
	bv := p.blockVarById(id)
	if bv != nil {
		return bv, bv.Token
	}
	t := p.blockTypeById(id)
	if t != nil {
		return t, t.Tok
	}
	return
}

func (p *Parser) check() {
	defer p.wg.Done()
	if p.IsMain && !p.JustDefs {
		f, _, _ := p.Defs.funcById(jule.EntryPoint, nil)
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
	if !p.JustDefs {
		p.checkFuncs()
		p.checkStructs()
	}
}

func (p *Parser) checkTypes() {
	for i, t := range p.Defs.Types {
		p.Defs.Types[i].Type, _ = p.realType(t.Type, true)
	}
}

// WaitingGlobals parses Jule global variables for waiting to parsing.
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
		if param.Default.Model.String() == juleapi.DefaultExpr {
			p.checkParamDefaultExprWithDefault(param)
			return
		}
	}
	dt := param.Type
	if param.Variadic {
		dt.Id = juletype.Slice
		dt.Kind = jule.Prefix_Slice + dt.Kind
		dt.ComponentType = new(models.DataType)
		*dt.ComponentType = param.Type
		dt.Original = nil
		dt.Pure = true
	}
	v, model := p.evalExpr(param.Default)
	param.Default.Model = model
	p.checkArgType(param, v, param.Tok)
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
	}
	p.checkParamDefaultExpr(f, param)
	return
}

func (p *Parser) checkParamDup(params []models.Param) (err bool) {
	for i, param := range params {
		for j, jparam := range params {
			if j >= i {
				break
			} else if param.Id == jparam.Id {
				err = true
				p.pusherrtok(param.Tok, "exist_id", param.Id)
			}
		}
	}
	return
}

func (p *Parser) params(f *Func) (err bool) {
	hasDefaultArg := false
	err = p.checkParamDup(f.Params)
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
		s := f.Receiver.Tag.(*structure)
		vars = append(vars, s.selfVar(*f.Receiver))
	}
	return vars
}

func (p *Parser) parsePureFunc(f *Func) (err bool) {
	hasError := p.eval.hasError
	defer func() { p.eval.hasError = hasError }()
	owner := f.Owner.(*Parser)
	err = owner.params(f)
	if err {
		return
	}
	owner.blockVars = owner.blockVarsOfFunc(f)
	owner.checkFunc(f)
	if owner != p {
		owner.wg.Wait()
		p.pusherrs(owner.Errors...)
		owner.Errors = nil
	}
	owner.blockTypes = nil
	owner.blockVars = nil
	return
}

func (p *Parser) parseFunc(f *function) (err bool) {
	if f.checked || len(f.Ast.Generics) > 0 {
		return false
	}
	return p.parsePureFunc(f.Ast)
}

func (p *Parser) checkFuncs() {
	err := false
	check := func(f *function) {
		if len(f.Ast.Generics) > 0 {
			return
		}
		p.parseTypesNonGenerics(f.Ast)
		p.wg.Add(1)
		go p.checkFuncSpecialCases(f.Ast)
		if err {
			return
		}
		p.blockTypes = nil
		err = p.parseFunc(f)
		f.checked = true
	}
	for _, f := range p.Defs.Funcs {
		check(f)
	}
}

func (p *Parser) parseStructFunc(s *structure, f *function) (err bool) {
	if len(f.Ast.Generics) > 0 {
		return
	}
	if len(s.Ast.Generics) == 0 {
		p.parseTypesNonGenerics(f.Ast)
		return p.parseFunc(f)
	}
	return
}

func (p *Parser) checkStruct(xs *structure) (err bool) {
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
	check := func(xs *structure) {
		if err {
			return
		}
		p.checkStruct(xs)
	}
	for _, s := range p.Defs.Structs {
		check(s)
	}
}

func (p *Parser) checkFuncSpecialCases(f *Func) {
	defer p.wg.Done()
	switch f.Id {
	case jule.EntryPoint, jule.InitializerFunction:
		p.checkSolidFuncSpecialCases(f)
	}
}

func (p *Parser) callFunc(f *Func, data callData, m *exprModel) value {
	v := p.parseFuncCallToks(f, data.generics, data.args, m)
	v.lvalue = typeIsLvalue(v.data.Type)
	return v
}

func (p *Parser) callStructConstructor(s *structure, argsToks Toks, m *exprModel) (v value) {
	f := s.constructor
	s = f.RetType.Type.Tag.(*structure)
	v.data.Type = f.RetType.Type.Copy()
	v.data.Type.Kind = s.dataTypeString()
	v.isType = false
	v.lvalue = false
	v.constExpr = false
	v.data.Value = s.Ast.Id
	// Set braces to parentheses
	argsToks[0].Kind = tokens.LPARENTHESES
	argsToks[len(argsToks)-1].Kind = tokens.RPARENTHESES
	args := p.getArgs(argsToks, true)
	m.appendSubNode(exprNode{f.RetType.String()})
	m.appendSubNode(exprNode{tokens.LPARENTHESES})
	p.parseArgs(f, args, m, f.Tok)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
	m.appendSubNode(exprNode{tokens.RPARENTHESES})
	return v
}

func (p *Parser) parseField(s *structure, f **Var, i int) {
	*f = p.Var(**f)
	v := *f
	param := models.Param{Id: v.Id, Type: v.Type}
	if v.Type.Id == juletype.Struct && v.Type.Tag == s && typeIsPure(v.Type) {
		p.pusherrtok(v.Type.Tok, "invalid_type_source")
	}
	if hasExpr(v.Expr) {
		param.Default = v.Expr
	} else {
		param.Default.Model = exprNode{juleapi.DefaultExpr}
	}
	s.constructor.Params[i] = param
}

func (p *Parser) structConstructorInstance(as *structure) *structure {
	s := new(structure)
	s.Ast = as.Ast
	s.traits = as.traits
	s.constructor = new(Func)
	*s.constructor = *as.constructor
	s.constructor.RetType.Type.Tag = s
	s.Defs = as.Defs
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
	p.checkFunc(f)
	p.rootBlock = rootBlock
	p.nodeBlock = nodeBlock
	p.Defs.Globals = globals
	p.blockVars = blockVariables
}

// Returns nil if has error.
func (p *Parser) getArgs(toks Toks, targeting bool) *models.Args {
	toks, _ = p.getrange(tokens.LPARENTHESES, tokens.RPARENTHESES, toks)
	if toks == nil {
		toks = make(Toks, 0)
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
func (p *Parser) getGenerics(toks Toks) (_ []DataType, err bool) {
	if len(toks) == 0 {
		return nil, false
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
		j := 0
		generic, _ := b.DataType(part, &j, false, true)
		b.Wait()
		if j+1 < len(part) {
			p.pusherrtok(part[j+1], "invalid_syntax")
		}
		p.pusherrs(b.Errors...)
		var ok bool
		generics[i], ok = p.realType(generic, true)
		if !ok {
			err = true
		}
	}
	return generics, err
}

func (p *Parser) checkGenericsQuantity(required, given int, errTok Tok) bool {
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

func (p *Parser) pushGeneric(generic *GenericType, source DataType) {
	t := &Type{
		Id:      generic.Id,
		Tok:     generic.Tok,
		Type:    source,
		Used:    true,
		Generic: true,
	}
	p.blockTypes = append(p.blockTypes, t)
}

func (p *Parser) pushGenerics(generics []*GenericType, sources []DataType) {
	for i, generic := range generics {
		p.pushGeneric(generic, sources[i])
	}
}

func (p *Parser) reloadFuncTypes(f *Func) {
	for i, param := range f.Params {
		f.Params[i].Type, _ = p.realType(param.Type, true)
	}
	f.RetType.Type, _ = p.realType(f.RetType.Type, true)
}

func itsCombined(f *Func, generics []DataType) bool {
	if f.Combines == nil { // Built-in
		return true
	}
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

func (p *Parser) parseGenericFunc(f *Func, generics []DataType, errtok Tok) {
	owner := f.Owner.(*Parser)
	if f.Receiver != nil {
		s := f.Receiver.Tag.(*structure)
		owner.pushGenerics(s.Ast.Generics, s.Generics())
	}
	owner.reloadFuncTypes(f)
	if f.Block == nil {
		return
	} else if itsCombined(f, generics) {
		return
	}
	*f.Combines = append(*f.Combines, generics)
	p.parsePureFunc(f)
}

func (p *Parser) parseGenerics(f *Func, args *models.Args, errTok Tok) bool {
	if len(f.Generics) > 0 && len(args.Generics) == 0 {
		for _, generic := range f.Generics {
			ok := false
			for _, param := range f.Params {
				if typeHasThisGeneric(generic, param.Type) {
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
	}
	f.Owner.(*Parser).pushGenerics(f.Generics, args.Generics)
	f.Owner.(*Parser).reloadFuncTypes(f)
ok:
	return true
}

func (p *Parser) parseFuncCall(f *Func, args *models.Args, m *exprModel, errTok Tok) (v value) {
	args.NeedsPureType = p.rootBlock == nil || len(p.rootBlock.Func.Generics) == 0
	if len(f.Generics) > 0 {
		params := make([]Param, len(f.Params))
		copy(params, f.Params)
		retType := f.RetType.Type
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
			owner := f.Owner.(*Parser)
			s := f.Receiver.Tag.(*structure)
			generics := s.Generics()
			if len(generics) > 0 {
				owner.pushGenerics(s.Ast.Generics, generics)
				owner.reloadFuncTypes(f)
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
		m.appendSubNode(callExpr{
			generics: genericsExpr{args.Generics},
			args:     argsExpr{args.Src},
			f:        f,
		})
	}
end:
	v.data.Value = f.Id
	v.data.Type = f.RetType.Type.Copy()
	if args.NeedsPureType {
		v.data.Type.Pure = true
		v.data.Type.Original = nil
	}
	return
}

func (p *Parser) parseFuncCallToks(f *Func, genericsToks, argsToks Toks, m *exprModel) (v value) {
	var generics []DataType
	var args *models.Args
	if f.FindAttribute(jule.Attribute_TypeArg) != nil {
		if len(genericsToks) > 0 {
			p.pusherrtok(genericsToks[0], "invalid_syntax")
			return
		}
		var err bool
		generics, err = p.getGenerics(argsToks)
		if err {
			p.eval.hasError = true
			return
		}
		args = new(models.Args)
		args.Generics = generics
	} else {
		var err bool
		generics, err = p.getGenerics(genericsToks)
		if err {
			p.eval.hasError = true
			return
		}
		args = p.getArgs(argsToks, false)
		args.Generics = generics
	}
	return p.parseFuncCall(f, args, m, argsToks[0])
}

func (p *Parser) parseStructArgs(f *Func, args *models.Args, errTok Tok) {
	sap := structArgParser{
		p:      p,
		f:      f,
		args:   args,
		errTok: errTok,
	}
	sap.parse()
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
		p.parseStructArgs(f, args, errTok)
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

func (p *Parser) pushGenericByFunc(f *Func, pair *paramMapPair, args *models.Args, t DataType) bool {
	tf := t.Tag.(*Func)
	cf := pair.param.Type.Tag.(*Func)
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

func (p *Parser) pushGenericByMultiTyped(f *Func, pair *paramMapPair, args *models.Args, t DataType) bool {
	types := t.Tag.([]DataType)
	for _, t := range types {
		for _, generic := range f.Generics {
			if typeHasThisGeneric(generic, pair.param.Type) {
				p.pushGenericByType(f, generic, args, t)
				break
			}
		}
	}
	return true
}

func (p *Parser) pushGenericByCommonArg(f *Func, pair *paramMapPair, args *models.Args, t DataType) bool {
	for _, generic := range f.Generics {
		if typeIsThisGeneric(generic, pair.param.Type) {
			p.pushGenericByType(f, generic, args, t)
			return true
		}
	}
	return false
}

func (p *Parser) pushGenericByType(f *Func, generic *GenericType, args *models.Args, t DataType) {
	owner := f.Owner.(*Parser)
	// Already added
	if owner.blockTypeById(generic.Id) != nil {
		return
	}
	id, _ := t.KindId()
	t.Kind = id
	f.Owner.(*Parser).pushGeneric(generic, t)
	args.Generics = append(args.Generics, t)
}

func (p *Parser) pushGenericByComponent(f *Func, pair *paramMapPair, args *models.Args, argType DataType) bool {
	for argType.ComponentType != nil {
		argType = *argType.ComponentType
	}
	return p.pushGenericByCommonArg(f, pair, args, argType)
}

func (p *Parser) pushGenericByArg(f *Func, pair *paramMapPair, args *models.Args, argType DataType) bool {
	_, prefix := pair.param.Type.KindId()
	_, tprefix := argType.KindId()
	if prefix != tprefix {
		return false
	}
	switch {
	case typeIsFunc(argType):
		return p.pushGenericByFunc(f, pair, args, argType)
	case argType.MultiTyped, typeIsMap(argType):
		return p.pushGenericByMultiTyped(f, pair, args, argType)
	case typeIsArray(argType), typeIsSlice(argType):
		return p.pushGenericByComponent(f, pair, args, argType)
	default:
		return p.pushGenericByCommonArg(f, pair, args, argType)
	}
}

func (p *Parser) parseArg(f *Func, pair *paramMapPair, args *models.Args, variadiced *bool) {
	value, model := p.evalExpr(pair.arg.Expr)
	pair.arg.Expr.Model = model
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	if args.DynamicGenericAnnotation && typeHasGenerics(f.Generics, pair.param.Type) {
		ok := p.pushGenericByArg(f, pair, args, value.data.Type)
		if !ok {
			p.pusherrtok(pair.arg.Tok, "dynamic_generic_annotation_failed")
		}
		return
	}
	p.checkArgType(pair.param, value, pair.arg.Tok)
}

func (p *Parser) checkArgType(param *Param, val value, errTok Tok) {
	if param.Reference && !val.lvalue {
		p.pusherrtok(errTok, "not_lvalue_for_reference_param")
	}
	assignChecker{
		p:      p,
		t:      param.Type,
		v:      val,
		errtok: errTok,
	}.checkAssignType()
}

// getrange returns between of brackets.
//
// Special case is:
//  getrange(open, close, tokens) = nil, false if fail
func (p *Parser) getrange(open, close string, toks Toks) (_ Toks, ok bool) {
	i := 0
	toks = ast.Range(&i, open, close, toks)
	return toks, toks != nil
}

func (p *Parser) checkSolidFuncSpecialCases(f *Func) {
	if len(f.Params) > 0 {
		p.pusherrtok(f.Tok, "func_have_parameters", f.Id)
	}
	if f.RetType.Type.Id != juletype.Void {
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
			p.pusherrtok(v.Token, "declared_but_not_used", v.Id)
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
		s.Data = t
	case models.Continue:
		p.continueStatement(&t)
	case Type:
		if def, _ := p.blockDefById(t.Id); def != nil {
			p.pusherrtok(t.Tok, "exist_id", t.Id)
			break
		} else if juleapi.IsIgnoreId(t.Id) {
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
	case models.Comment:
	default:
		return false
	}
	return true
}

func (p *Parser) fallthroughStatement(f *models.Fallthrough, b *models.Block, i *int) {
	switch {
	case p.currentCase == nil || *i+1 < len(b.Tree):
		p.pusherrtok(f.Tok, "fallthrough_wrong_use")
		return
	case p.currentCase.Next == nil:
		p.pusherrtok(f.Tok, "fallthrough_into_final_case")
		return
	}
	f.Case = p.currentCase
}

func (p *Parser) checkStatement(b *models.Block, i *int) {
	s := b.Tree[*i]
	defer func(i int) { b.Tree[i] = s }(*i)
	if p.statement(&s, true) {
		return
	}
	switch t := s.Data.(type) {
	case models.Fallthrough:
		p.fallthroughStatement(&t, b, i)
		s.Data = t
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
	args := p.getArgs(callToks, false)
	handleParam := recoverFunc.Ast.Params[0]
	if len(args.Src) == 0 {
		p.pusherrtok(errtok, "missing_expr_for", handleParam.Id)
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
	catcher.exprs = append(catcher.exprs, "} catch(trait<JULEC_ID(Error)> ")
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
				def, _, _ := p.defById(tok.Kind)
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
		assignChecker{
			p:      p,
			t:      t,
			v:      value,
			errtok: expr.Toks[0],
		}.checkAssignType()
	}
	oldCase := p.currentCase
	oldIter := p.isNowIntoIter
	p.currentCase = c
	p.isNowIntoIter = false
	p.checkNewBlock(c.Block)
	p.currentCase = oldCase
	p.isNowIntoIter = oldIter
}

func (p *Parser) cases(m *models.Match, t DataType) {
	for i := range m.Cases {
		p.parseCase(&m.Cases[i], t)
	}
}

func (p *Parser) matchcase(t *models.Match) {
	if len(t.Expr.Processes) > 0 {
		value, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		t.ExprType = value.data.Type
	} else {
		t.ExprType.Id = juletype.Bool
		t.ExprType.Kind = juletype.TypeMap[t.ExprType.Id]
	}
	p.cases(t, t.ExprType)
	if t.Default != nil {
		p.parseCase(t.Default, t.ExprType)
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
			switch s.Data.(type) {
			case models.Ret:
				return
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

func (p *Parser) varStatement(v *Var, noParse bool) {
	if _, tok := p.blockDefById(v.Id); tok.Id != tokens.NA {
		p.pusherrtok(v.Token, "exist_id", v.Id)
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	v.IsField = true
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
	if selected.constExpr {
		p.pusherrtok(errtok, "assign_const")
		state = false
	}
	switch selected.data.Type.Tag.(type) {
	case Func:
		f, _, _ := p.FuncById(selected.data.Tok.Kind)
		if f != nil {
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
	if len(left.Toks) == 1 && juleapi.IsIgnoreId(left.Toks[0].Kind) {
		return
	}
	leftExpr, model := p.evalExpr(*left)
	left.Model = model
	if leftExpr.isField {
		right.Model = exprNode{exprMustHeap(right.Model.String())}
	}
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
	assignChecker{
		p:      p,
		t:      leftExpr.data.Type,
		v:      val,
		errtok: assign.Setter,
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
		left.Ignore = juleapi.IsIgnoreId(left.Var.Id)
		right := right[i]
		if !left.Var.New {
			if left.Ignore {
				continue
			}
			leftExpr, model := p.evalExpr(left.Expr)
			left.Expr.Model = model
			if leftExpr.isField {
				right := &assign.Right[i]
				right.Model = exprNode{exprMustHeap(right.Model.String())}
			}
			if !p.assignment(leftExpr, assign.Setter) {
				return
			}
			assignChecker{
				p:      p,
				t:      leftExpr.data.Type,
				v:      right,
				errtok: assign.Setter,
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
	if typeIsExplicitPtr(val.data.Type) {
		return
	}
	if typeIsPure(val.data.Type) && juletype.IsNumeric(val.data.Type.Id) {
		return
	}
	p.pusherrtok(assign.Setter, "operator_notfor_juletype", assign.Setter.Kind, val.data.Type.Kind)
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
	profile.KeyA.IsField = true
	profile.KeyB.IsField = true
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
		if juleapi.IsIgnoreId(profile.KeyA.Id) {
			p.pusherrtok(profile.KeyA.Token, "ignore_id")
		}
		p.varStatement(&profile.KeyA, true)
	}
	if profile.KeyB.New {
		if juleapi.IsIgnoreId(profile.KeyB.Id) {
			p.pusherrtok(profile.KeyB.Token, "ignore_id")
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
		assignChecker{
			p:      p,
			t:      DataType{Id: juletype.Bool, Kind: juletype.TypeMap[juletype.Bool]},
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
	oldCase := p.currentCase
	oldIter := p.isNowIntoIter
	p.currentCase = nil
	p.isNowIntoIter = true
	switch iter.Profile.(type) {
	case models.IterWhile:
		p.whileProfile(iter)
	case models.IterForeach:
		p.foreachProfile(iter)
	case models.IterFor:
		p.forProfile(iter)
	default:
		p.checkNewBlock(iter.Block)
	}
	p.currentCase = oldCase
	p.isNowIntoIter = oldIter
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
	switch {
	case p.isNowIntoIter:
	case p.currentCase != nil:
		breakAST.Case = p.currentCase
	default:
		p.pusherrtok(breakAST.Tok, "break_at_outiter")
	}
}

func (p *Parser) continueStatement(continueAST *models.Continue) {
	if !p.isNowIntoIter {
		p.pusherrtok(continueAST.Tok, "continue_at_outiter")
	}
}

func (p *Parser) checkValidityForAutoType(t DataType, errtok Tok) {
	if p.eval.hasError {
		return
	}
	switch t.Id {
	case juletype.Nil:
		p.pusherrtok(errtok, "nil_for_autotype")
	case juletype.Void:
		p.pusherrtok(errtok, "void_for_autotype")
	}
}

func (p *Parser) typeSourceOfMultiTyped(dt DataType, err bool) (DataType, bool) {
	types := dt.Tag.([]DataType)
	ok := false
	for i, t := range types {
		t, ok = p.typeSource(t, err)
		types[i] = t
	}
	dt.Tag = types
	return dt, ok
}

func (p *Parser) typeSourceIsType(dt DataType, t *Type, err bool) (DataType, bool) {
	original := dt.Original
	old := dt
	dt = t.Type
	dt.Tok = t.Tok
	dt.Generic = t.Generic
	dt.Original = original
	dt, ok := p.typeSource(dt, err)
	dt.Pure = false
	if ok && old.Tag != nil && !typeIsStruct(t.Type) { // Has generics
		p.pusherrtok(dt.Tok, "invalid_type_source")
	}
	return dt, ok
}

func (p *Parser) typeSourceIsEnum(e *Enum, tag any) (dt DataType, _ bool) {
	dt.Id = juletype.Enum
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

func (p *Parser) typeSourceIsStruct(s *structure, t DataType) (dt DataType, _ bool) {
	generics := s.Generics()
	if len(generics) > 0 {
		if !p.checkGenericsQuantity(len(s.Ast.Generics), len(generics), t.Tok) {
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
		*s.constructor.Combines = append(*s.constructor.Combines, generics)
		owner := s.Ast.Owner.(*Parser)
		blockTypes := owner.blockTypes
		owner.blockTypes = nil
		defer func() { owner.blockTypes = blockTypes }()
		owner.pushGenerics(s.Ast.Generics, generics)
		for i, f := range s.Ast.Fields {
			owner.parseField(s, &f, i)
		}
		if len(s.Defs.Funcs) > 0 {
			for _, f := range s.Defs.Funcs {
				if len(f.Ast.Generics) == 0 {
					blockVars := owner.blockVars
					blockTypes := owner.blockTypes
					owner.reloadFuncTypes(f.Ast)
					_ = p.parsePureFunc(f.Ast)
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
	} else if len(s.Ast.Generics) > 0 {
		p.pusherrtok(t.Tok, "has_generics")
	}
end:
	dt.Id = juletype.Struct
	dt.Kind = s.dataTypeString()
	dt.Tag = s
	dt.Tok = s.Ast.Tok
	return dt, true
}

func (p *Parser) typeSourceIsTrait(t *trait, tag any, errTok Tok) (dt DataType, _ bool) {
	if tag != nil {
		p.pusherrtok(errTok, "invalid_type_source")
	}
	t.Used = true
	dt.Id = juletype.Trait
	dt.Kind = t.Ast.Id
	dt.Tag = t
	dt.Tok = t.Ast.Tok
	dt.Pure = true
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
		if i < len(parts)-1 {
			toks = append(toks, Tok{
				Id:   tokens.DoubleColon,
				Kind: tokens.DOUBLE_COLON,
				File: p.File,
			})
		}
	}
	return toks
}

func (p *Parser) typeSourceIsArrayType(t *DataType) (ok bool) {
	ok = true
	t.Original = nil
	t.Pure = true
	*t.ComponentType, ok = p.realType(*t.ComponentType, true)
	if !ok {
		return
	}
	ptrs := t.Pointers()
	t.Kind = ptrs + jule.Prefix_Array + t.ComponentType.Kind
	if t.Size.AutoSized || t.Size.Expr.Model != nil {
		return
	}
	val, model := p.evalExpr(t.Size.Expr)
	t.Size.Expr.Model = model
	if val.constExpr {
		t.Size.N = models.Size(tonumu(val.expr))
	} else {
		p.eval.pusherrtok(t.Tok, "expr_not_const")
	}
	assignChecker{
		p:      p,
		t:      DataType{Id: juletype.UInt, Kind: juletype.TypeMap[juletype.UInt]},
		v:      val,
		errtok: t.Size.Expr.Toks[0],
	}.checkAssignType()
	return
}

func (p *Parser) typeSourceIsSliceType(t *DataType) (ok bool) {
	*t.ComponentType, ok = p.realType(*t.ComponentType, true)
	ptrs := t.Pointers()
	t.Kind = ptrs + jule.Prefix_Slice + t.ComponentType.Kind
	if ok && typeIsArray(*t.ComponentType) { // Array into slice
		p.pusherrtok(t.Tok, "invalid_type_source")
	}
	return
}

func (p *Parser) typeSource(dt DataType, err bool) (ret DataType, ok bool) {
	if dt.Kind == "" {
		return dt, true
	}
	original := dt.Original
	defer func() { ret.Original = original }()
	dt.SetToOriginal()
	switch {
	case dt.MultiTyped:
		return p.typeSourceOfMultiTyped(dt, err)
	case typeIsMap(dt):
		return p.typeSourceIsMap(dt, err)
	case typeIsArray(dt):
		ok = p.typeSourceIsArrayType(&dt)
		return dt, ok
	case typeIsSlice(dt):
		ok = p.typeSourceIsSliceType(&dt)
		return dt, ok
	}
	switch dt.Id {
	case juletype.Struct:
		_, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		return p.typeSourceIsStruct(dt.Tag.(*structure), dt)
	case juletype.Id:
		id, prefix := dt.KindId()
		defer func() { ret.Kind = prefix + ret.Kind }()
		var def any
		if strings.Contains(id, tokens.DOUBLE_COLON) { // Has namespace?
			toks := p.tokenizeDataType(id)
			defs := p.eval.getNs(&toks)
			if defs == nil {
				return
			}
			i, m, t := defs.findById(toks[0].Kind, p.File)
			switch t {
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
			def, _, _ = p.defById(id)
		}
		switch t := def.(type) {
		case *Type:
			t.Used = true
			return p.typeSourceIsType(dt, t, err)
		case *Enum:
			t.Used = true
			return p.typeSourceIsEnum(t, dt.Tag)
		case *structure:
			t.Used = true
			t = p.structConstructorInstance(t)
			switch tagt := dt.Tag.(type) {
			case []models.DataType:
				t.SetGenerics(tagt)
			}
			return p.typeSourceIsStruct(t, dt)
		case *trait:
			t.Used = true
			return p.typeSourceIsTrait(t, dt.Tag, dt.Tok)
		default:
			if err {
				p.pusherrtok(dt.Tok, "invalid_type_source")
			}
			return dt, false
		}
	case juletype.Func:
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

func (p *Parser) checkType(real, check DataType, ignoreAny bool, errTok Tok) {
	if typeIsVoid(check) {
		p.eval.pusherrtok(errTok, "incompatible_datatype", real.Kind, check.Kind)
		return
	}
	if !ignoreAny && real.Id == juletype.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.checkMultiType(real, check, ignoreAny, errTok)
		return
	}
	if typesAreCompatible(real, check, ignoreAny) {
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
		realKind := strings.Replace(real.Kind, jule.Mark_Array, strconv.FormatUint(i, 10), 1)
		checkKind := strings.Replace(check.Kind, jule.Mark_Array, strconv.FormatUint(j, 10), 1)
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
