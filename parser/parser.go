package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/the-xlang/xxc/ast"
	"github.com/the-xlang/xxc/lex"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xbits"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/preprocessor"
)

type use struct {
	Path string
	defs *defmap
}

var used []*use

// Parser is parser of X code.
type Parser struct {
	attributes []ast.Attribute
	docText    strings.Builder
	iterCount  int
	wg         sync.WaitGroup
	justDefs   bool
	main       bool
	isLocalPkg bool
	rootBlock  *ast.Block
	nodeBlock  *ast.Block

	Embeds         strings.Builder
	Uses           []*use
	Defs           *defmap
	waitingGlobals []ast.Var
	BlockVars      []*ast.Var
	BlockTypes     []*ast.Type
	Errs           []xlog.CompilerLog
	Warns          []xlog.CompilerLog
	File           *xio.File
}

// New returns new instance of Parser.
func New(f *xio.File) *Parser {
	p := new(Parser)
	p.File = f
	p.isLocalPkg = false
	p.Defs = new(defmap)
	return p
}

// Parses object tree and returns parser.
func Parset(tree []ast.Obj, main, justDefs bool) *Parser {
	p := New(nil)
	p.Parset(tree, main, justDefs)
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok lex.Tok, key string, args ...interface{}) {
	p.pusherrmsgtok(tok, x.GetErr(key, args...))
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok lex.Tok, msg string) {
	p.Errs = append(p.Errs, xlog.CompilerLog{
		Type:   xlog.Err,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path,
		Msg:    msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok lex.Tok, key string, args ...interface{}) {
	p.Warns = append(p.Warns, xlog.CompilerLog{
		Type:   xlog.Warn,
		Row:    tok.Row,
		Column: tok.Column,
		Path:   tok.File.Path,
		Msg:    x.GetWarn(key, args...),
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...xlog.CompilerLog) { p.Errs = append(p.Errs, errs...) }

// pusherr appends new error.
func (p *Parser) pusherr(key string, args ...interface{}) {
	p.pusherrmsg(x.GetErr(key, args...))
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errs = append(p.Errs, xlog.CompilerLog{
		Type: xlog.FlatErr,
		Msg:  msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string, args ...interface{}) {
	p.Warns = append(p.Warns, xlog.CompilerLog{
		Type: xlog.FlatWarn,
		Msg:  x.GetWarn(key, args...),
	})
}

// CxxEmbeds return C++ code of cxx embeds.
func (p *Parser) CxxEmbeds() string {
	var cxx strings.Builder
	cxx.WriteString("// region EMBEDS\n")
	cxx.WriteString(p.Embeds.String())
	cxx.WriteString("// endregion EMBEDS")
	return cxx.String()
}

// outableFunc returns true if function is available for
// cxx output, returns false if not.
func outableFunc(f *function) bool {
	return f.Ast.Id == x.EntryPoint || f.used
}

// CxxPrototypes returns C++ code of prototypes of C++ code.
func (p *Parser) CxxPrototypes() string {
	var cxx strings.Builder
	cxx.WriteString("// region PROTOTYPES\n")
	for _, use := range used {
		for _, f := range use.defs.Funcs {
			if outableFunc(f) {
				cxx.WriteString(f.Prototype())
				cxx.WriteByte('\n')
			}
		}
	}
	for _, f := range p.Defs.Funcs {
		if outableFunc(f) {
			cxx.WriteString(f.Prototype())
			cxx.WriteByte('\n')
		}
	}
	cxx.WriteString("// endregion PROTOTYPES")
	return cxx.String()
}

// CxxGlobals returns C++ code of global variables.
func (p *Parser) CxxGlobals() string {
	var cxx strings.Builder
	cxx.WriteString("// region GLOBALS\n")
	for _, use := range used {
		for _, v := range use.defs.Globals {
			if v.Used {
				cxx.WriteString(v.String())
				cxx.WriteByte('\n')
			}
		}
	}
	for _, v := range p.Defs.Globals {
		if v.Used {
			cxx.WriteString(v.String())
			cxx.WriteByte('\n')
		}
	}
	cxx.WriteString("// endregion GLOBALS")
	return cxx.String()
}

// CxxFuncs returns C++ code of functions.
func (p *Parser) CxxFuncs() string {
	var cxx strings.Builder
	cxx.WriteString("// region FUNCTIONS\n")
	for _, use := range used {
		for _, f := range use.defs.Funcs {
			if outableFunc(f) {
				cxx.WriteString(f.String())
				cxx.WriteString("\n\n")
			}
		}
	}
	for _, f := range p.Defs.Funcs {
		if outableFunc(f) {
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
	}
	cxx.WriteString("// endregion FUNCTIONS")
	return cxx.String()
}

// Cxx returns full C++ code of parsed objects.
func (p *Parser) Cxx() string {
	var cxx strings.Builder
	cxx.WriteString(p.CxxEmbeds())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxPrototypes())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxGlobals())
	cxx.WriteString("\n\n")
	cxx.WriteString(p.CxxFuncs())
	return cxx.String()
}

func getTree(toks []lex.Tok, errs *[]xlog.CompilerLog) []ast.Obj {
	b := ast.NewBuilder(toks)
	b.Build()
	if len(b.Errs) > 0 {
		if errs != nil {
			*errs = append(*errs, b.Errs...)
		}
		return nil
	}
	return b.Tree
}

func (p *Parser) checkUsePath(use *ast.Use) bool {
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

func (p *Parser) compileUse(useAST *ast.Use) *use {
	infos, err := ioutil.ReadDir(useAST.Path)
	if err != nil {
		p.pusherrmsg(err.Error())
		return nil
	}
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() || !strings.HasSuffix(name, x.SrcExt) {
			continue
		}
		f, err := xio.Openfx(filepath.Join(useAST.Path, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			continue
		}
		psub := New(f)
		psub.Parsef(false, false)
		if psub.Errs != nil {
			p.pusherrtok(useAST.Tok, "use_has_errors")
		}
		use := new(use)
		use.defs = new(defmap)
		use.Path = useAST.Path
		p.pusherrs(psub.Errs...)
		p.Warns = append(p.Warns, psub.Warns...)
		p.pushUseDefs(use, psub.Defs)
		return use
	}
	return nil
}

func (p *Parser) pushUseTypes(use *use, dm *defmap) {
	for _, t := range dm.Types {
		def := p.typeById(t.Id)
		if def != nil {
			p.pusherrmsgtok(def.Tok,
				fmt.Sprintf(`"%s" identifier is already defined in this source`, t.Id))
		} else {
			use.defs.Types = append(use.defs.Types, t)
		}
	}
}

func (p *Parser) pushUseGlobals(use *use, dm *defmap) {
	for _, g := range dm.Globals {
		def := p.Defs.globalById(g.Id)
		if def != nil {
			p.pusherrmsgtok(def.IdTok,
				fmt.Sprintf(`"%s" identifier is already defined in this source`, g.Id))
		} else {
			use.defs.Globals = append(use.defs.Globals, g)
		}
	}
}

func (p *Parser) pushUseFuncs(use *use, dm *defmap) {
	for _, f := range dm.Funcs {
		def := p.Defs.funcById(f.Ast.Id)
		if def != nil {
			p.pusherrmsgtok(def.Ast.Tok,
				fmt.Sprintf(`"%s" identifier is already defined in this source`, f.Ast.Id))
		} else {
			use.defs.Funcs = append(use.defs.Funcs, f)
		}
	}
}

func (p *Parser) pushUseDefs(use *use, dm *defmap) {
	p.pushUseTypes(use, dm)
	p.pushUseGlobals(use, dm)
	p.pushUseFuncs(use, dm)
}

func (p *Parser) use(useAST *ast.Use) {
	if !p.checkUsePath(useAST) {
		return
	}
	// Already parsed?
	for _, use := range used {
		if useAST.Path == use.Path {
			p.Uses = append(p.Uses, use)
			return
		}
	}
	use := p.compileUse(useAST)
	if use == nil {
		return
	}
	exist := false
	for _, guse := range used {
		if guse.Path == use.Path {
			exist = true
			break
		}
	}
	if !exist {
		used = append(used, use)
	}
	p.Uses = append(p.Uses, use)
}

func (p *Parser) parseUses(tree *[]ast.Obj) {
	for i, obj := range *tree {
		switch t := obj.Value.(type) {
		case ast.Use:
			p.use(&t)
		default:
			*tree = (*tree)[i:]
			return
		}
	}
	*tree = nil
}

func (p *Parser) parseSrcTreeObj(obj ast.Obj) {
	switch t := obj.Value.(type) {
	case ast.Attribute:
		p.PushAttribute(t)
	case ast.Statement:
		p.Statement(t)
	case ast.Type:
		p.Type(t)
	case ast.CxxEmbed:
		p.Embeds.WriteString(t.String())
		p.Embeds.WriteByte('\n')
	case ast.Comment:
		p.Comment(t)
	case ast.Use:
		p.pusherrtok(obj.Tok, "use_at_content")
	case ast.Preprocessor:
	default:
		p.pusherrtok(obj.Tok, "invalid_syntax")
	}
}

func (p *Parser) parseSrcTree(tree []ast.Obj) {
	for _, obj := range tree {
		p.parseSrcTreeObj(obj)
		p.checkDoc(obj)
		p.checkAttribute(obj)
	}
}

func (p *Parser) parseTree(tree []ast.Obj) {
	p.parseUses(&tree)
	p.parseSrcTree(tree)
}

func (p *Parser) checkParse() {
	if p.docText.Len() > 0 {
		p.pushwarn("exist_undefined_doc")
	}
	p.wg.Add(1)
	go p.checkAsync()
}

// Special case is;
//  p.useLocalPackage() -> nothing if p.File is nil
func (p *Parser) useLocalPakcage(tree *[]ast.Obj) {
	if p.File == nil {
		return
	}
	dir := filepath.Dir(p.File.Path)
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		p.pusherrmsg(err.Error())
		return
	}
	_, mainName := filepath.Split(p.File.Path)
	for _, info := range infos {
		name := info.Name()
		// Skip directories.
		if info.IsDir() ||
			!strings.HasSuffix(name, x.SrcExt) ||
			name == mainName {
			continue
		}
		f, err := xio.Openfx(filepath.Join(dir, name))
		if err != nil {
			p.pusherrmsg(err.Error())
			continue
		}
		lexer := lex.NewLex(f)
		toks := lexer.Lex()
		if lexer.Logs != nil {
			p.pusherrs(lexer.Logs...)
			continue
		}
		subtree := getTree(toks, &p.Errs)
		if subtree == nil {
			continue
		}
		preprocessor.TrimEnofi(&subtree)
		p.parseUses(&subtree)
		*tree = append(*tree, subtree...)
	}
}

// Parses X code from object tree.
func (p *Parser) Parset(tree []ast.Obj, main, justDefs bool) {
	p.main = main
	p.justDefs = justDefs
	if !p.isLocalPkg {
		p.useLocalPakcage(&tree)
	}
	if !main {
		preprocessor.TrimEnofi(&tree)
	}
	p.parseTree(tree)
	p.checkParse()
	p.wg.Wait()
}

// Parses X code from tokens.
func (p *Parser) Parse(toks []lex.Tok, main, justDefs bool) {
	tree := getTree(toks, &p.Errs)
	if tree == nil {
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

func (p *Parser) checkDoc(obj ast.Obj) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Value.(type) {
	case ast.Comment, ast.Attribute:
		return
	}
	p.pushwarntok(obj.Tok, "doc_ignored")
	p.docText.Reset()
}

func (p *Parser) checkAttribute(obj ast.Obj) {
	if p.attributes == nil {
		return
	}
	switch obj.Value.(type) {
	case ast.Attribute, ast.Comment:
		return
	}
	p.pusherrtok(obj.Tok, "attribute_not_supports")
	p.attributes = nil
}

func (p *Parser) checkTypeAST(t ast.Type) bool {
	if p.existid(t.Id).Id != lex.NA {
		p.pusherrtok(t.Tok, "exist_id", t.Id)
		return false
	} else if xapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Tok, "ignore_id")
		return false
	}
	return true
}

// Type parses X type define statement.
func (p *Parser) Type(t ast.Type) {
	if !p.checkTypeAST(t) {
		return
	}
	t.Desc = p.docText.String()
	p.docText.Reset()
	p.Defs.Types = append(p.Defs.Types, &t)
}

// Comment parses X documentation comments line.
func (p *Parser) Comment(c ast.Comment) {
	c.Content = strings.TrimSpace(c.Content)
	if p.docText.Len() == 0 {
		if strings.HasPrefix(c.Content, "doc:") {
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
func (p *Parser) PushAttribute(attribute ast.Attribute) {
	switch attribute.Tag.Kind {
	case "inline":
	default:
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

// Statement parse X statement.
func (p *Parser) Statement(s ast.Statement) {
	switch t := s.Val.(type) {
	case ast.Func:
		p.Func(t)
	case ast.Var:
		p.Global(t)
	default:
		p.pusherrtok(s.Tok, "invalid_syntax")
	}
}

func (p *Parser) param(param *ast.Param) {
	param.Type, _ = p.readyType(param.Type, true)
	if paramHasDefaultArg(param) {
		dt := param.Type
		if param.Variadic {
			dt.Val = "[]" + dt.Val // For array.
		}
		v, model := p.evalExpr(param.Default)
		param.Default.Model = model
		p.wg.Add(1)
		go assignChecker{
			p:         p,
			constant:  param.Const,
			t:         dt,
			v:         v,
			ignoreAny: false,
			errtok:    param.Tok,
		}.checkAssignTypeAsync()
	}
}

func (p *Parser) params(params *[]ast.Param) {
	hasDefaultArg := false
	for i := range *params {
		param := &(*params)[i]
		p.param(param)
		if !hasDefaultArg {
			hasDefaultArg = paramHasDefaultArg(param)
			continue
		} else if !paramHasDefaultArg(param) && !param.Variadic {
			p.pusherrtok(param.Tok, "param_must_have_default_arg", param.Id)
		}
	}
}

// Func parse X function.
func (p *Parser) Func(fast ast.Func) {
	if p.existid(fast.Id).Id != lex.NA {
		p.pusherrtok(fast.Tok, "exist_id", fast.Id)
	} else if xapi.IsIgnoreId(fast.Id) {
		p.pusherrtok(fast.Tok, "ignore_id")
	}
	fast.RetType, _ = p.readyType(fast.RetType, true)
	p.params(&fast.Params)
	f := new(function)
	f.Ast = fast
	f.Attributes = p.attributes
	f.Desc = p.docText.String()
	p.attributes = nil
	p.docText.Reset()
	p.checkFuncAttributes(f.Attributes)
	p.Defs.Funcs = append(p.Defs.Funcs, f)
}

// ParseVariable parse X global variable.
func (p *Parser) Global(vast ast.Var) {
	if p.existid(vast.Id).Id != lex.NA {
		p.pusherrtok(vast.IdTok, "exist_id", vast.Id)
		return
	}
	vast.Desc = p.docText.String()
	p.docText.Reset()
	p.waitingGlobals = append(p.waitingGlobals, vast)
}

// Var parse X variable.
func (p *Parser) Var(v ast.Var) *ast.Var {
	if xapi.IsIgnoreId(v.Id) {
		p.pusherrtok(v.IdTok, "ignore_id")
	}
	var val value
	switch t := v.Tag.(type) {
	case value:
		val = t
	default:
		if v.SetterTok.Id != lex.NA {
			val, v.Val.Model = p.evalExpr(v.Val)
		}
	}
	if v.Type.Id != x.Void {
		v.Type, _ = p.readyType(v.Type, true)
		if v.SetterTok.Id != lex.NA {
			p.wg.Add(1)
			go assignChecker{
				p,
				v.Const,
				v.Type,
				val,
				false,
				v.IdTok,
			}.checkAssignTypeAsync()
		}
	} else {
		if v.SetterTok.Id == lex.NA {
			p.pusherrtok(v.IdTok, "missing_autotype_value")
		} else {
			v.Type = val.ast.Type
			p.checkValidityForAutoType(v.Type, v.SetterTok)
			p.checkAssignConst(v.Const, v.Type, val, v.SetterTok)
		}
	}
	if v.Const {
		if v.SetterTok.Id == lex.NA {
			p.pusherrtok(v.IdTok, "missing_const_value")
		}
	}
	return &v
}

func (p *Parser) checkFuncAttributes(attributes []ast.Attribute) {
	for _, attribute := range attributes {
		switch attribute.Tag.Kind {
		case "inline":
		default:
			p.pusherrtok(attribute.Tok, "invalid_attribute")
		}
	}
}

func (p *Parser) varsFromParams(params []ast.Param) []*ast.Var {
	length := len(params)
	vars := make([]*ast.Var, length)
	for i, param := range params {
		v := new(ast.Var)
		v.Id = param.Id
		v.IdTok = param.Tok
		v.Type = param.Type
		v.Const = param.Const
		v.Volatile = param.Volatile
		if param.Variadic {
			if length-i > 1 {
				p.pusherrtok(param.Tok, "variadic_parameter_notlast")
			}
			v.Type.Val = "[]" + v.Type.Val
		}
		vars = append(vars, v)
	}
	return vars
}

// FuncById returns function by specified id.
//
// Special case:
//  FuncById(id) -> nil: if function is not exist.
func (p *Parser) FuncById(id string) *function {
	for _, f := range builtinFuncs {
		if f.Ast.Id == id {
			return f
		}
	}
	for _, use := range p.Uses {
		f := use.defs.funcById(id)
		if f != nil && f.Ast.Pub {
			return f
		}
	}
	return p.Defs.funcById(id)
}

func (p *Parser) varById(id string) *ast.Var {
	for _, v := range p.BlockVars {
		if v != nil && v.Id == id {
			return v
		}
	}
	return p.globalById(id)
}

func (p *Parser) globalById(id string) *ast.Var {
	for _, use := range p.Uses {
		g := use.defs.globalById(id)
		if g != nil && g.Pub {
			return g
		}
	}
	return p.Defs.globalById(id)
}

func (p *Parser) typeById(id string) *ast.Type {
	for _, t := range p.BlockTypes {
		if t != nil && t.Id == id {
			return t
		}
	}
	for _, use := range p.Uses {
		t := use.defs.typeById(id)
		if t != nil && t.Pub {
			return t
		}
	}
	return p.Defs.typeById(id)
}

func (p *Parser) existIdf(id string, exceptGlobals bool) lex.Tok {
	t := p.typeById(id)
	if t != nil {
		return t.Tok
	}
	f := p.FuncById(id)
	if f != nil {
		return f.Ast.Tok
	}
	for _, v := range p.BlockVars {
		if v != nil && v.Id == id {
			return v.IdTok
		}
	}
	if !exceptGlobals {
		v := p.globalById(id)
		if v != nil {
			return v.IdTok
		}
		for _, v := range p.waitingGlobals {
			if v.Id == id {
				return v.IdTok
			}
		}
	}
	return lex.Tok{}
}

func (p *Parser) existid(id string) lex.Tok { return p.existIdf(id, false) }

func (p *Parser) checkAsync() {
	defer func() { p.wg.Done() }()
	if p.main && !p.justDefs {
		if p.FuncById(x.EntryPoint) == nil {
			p.pusherr("no_entry_point")
		}
	}
	p.wg.Add(1)
	go p.checkTypesAsync()
	p.WaitingGlobals()
	p.waitingGlobals = nil
	if !p.justDefs {
		p.wg.Add(1)
		go p.checkFuncsAsync()
	}
}

func (p *Parser) checkTypesAsync() {
	defer func() { p.wg.Done() }()
	for i, t := range p.Defs.Types {
		p.Defs.Types[i].Type, _ = p.readyType(t.Type, true)
	}
}

// WaitingGlobals parse X global variables for waiting parsing.
func (p *Parser) WaitingGlobals() {
	for _, vast := range p.waitingGlobals {
		v := p.Var(vast)
		p.Defs.Globals = append(p.Defs.Globals, v)
	}
}

func (p *Parser) checkFuncsAsync() {
	defer func() { p.wg.Done() }()
	for _, f := range p.Defs.Funcs {
		p.BlockTypes = nil
		p.BlockVars = p.varsFromParams(f.Ast.Params)
		p.wg.Add(1)
		go p.checkFuncSpecialCasesAsync(f)
		p.checkFunc(&f.Ast)
	}
}

func (p *Parser) checkFuncSpecialCasesAsync(fun *function) {
	defer func() { p.wg.Done() }()
	switch fun.Ast.Id {
	case x.EntryPoint:
		p.checkEntryPointSpecialCases(fun)
	}
}

type value struct {
	ast      ast.Value
	constant bool
	volatile bool
	lvalue   bool
	variadic bool
}

func eliminateProcesses(processes *[][]lex.Tok, i, to int) {
	for i < to {
		(*processes)[i] = nil
		i++
	}
}

func (p *Parser) evalProcesses(processes [][]lex.Tok) (v value, e iExpr) {
	if processes == nil {
		return
	}
	m := newExprModel(processes)
	e = m
	if len(processes) == 1 {
		v = p.evalExprPart(processes[0], m)
		return
	}
	process := solver{p: p, model: m}
	boolean := false
	for i := p.nextOperator(processes); i != -1 && !noData(processes); i = p.nextOperator(processes) {
		if !boolean {
			boolean = v.ast.Type.Id == x.Bool
		}
		if boolean {
			v.ast.Type.Id = x.Bool
		}
		m.index = i
		process.operator = processes[m.index][0]
		m.appendSubNode(exprNode{process.operator.Kind})
		if processes[i-1] == nil {
			process.leftVal = v.ast
			m.index = i + 1
			process.right = processes[m.index]
			process.rightVal = p.evalExprPart(process.right, m).ast
			v.ast = process.Solve()
			eliminateProcesses(&processes, i, i+2)
			continue
		} else if processes[i+1] == nil {
			m.index = i - 1
			process.left = processes[m.index]
			process.leftVal = p.evalExprPart(process.left, m).ast
			process.rightVal = v.ast
			v.ast = process.Solve()
			eliminateProcesses(&processes, i-1, i+1)
			continue
		} else if isOperator(processes[i-1]) {
			process.leftVal = v.ast
			m.index = i + 1
			process.right = processes[m.index]
			process.rightVal = p.evalExprPart(process.right, m).ast
			v.ast = process.Solve()
			eliminateProcesses(&processes, i, i+1)
			continue
		}
		m.index = i - 1
		process.left = processes[m.index]
		process.leftVal = p.evalExprPart(process.left, m).ast
		m.index = i + 1
		process.right = processes[m.index]
		process.rightVal = p.evalExprPart(process.right, m).ast
		solvedv := process.Solve()
		if v.ast.Type.Id != x.Void {
			process.operator.Kind = "+"
			process.leftVal = v.ast
			process.right = processes[i+1]
			process.rightVal = solvedv
			solvedv = process.Solve()
		}
		v.ast = solvedv
		eliminateProcesses(&processes, i-1, i+2)
	}
	return
}

func noData(processes [][]lex.Tok) bool {
	for _, p := range processes {
		if !isOperator(p) && p != nil {
			return false
		}
	}
	return true
}

func isOperator(process []lex.Tok) bool {
	return len(process) == 1 && process[0].Id == lex.Operator
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) nextOperator(processes [][]lex.Tok) int {
	precedence5 := -1
	precedence4 := -1
	precedence3 := -1
	precedence2 := -1
	precedence1 := -1
	for i, process := range processes {
		if !isOperator(process) {
			continue
		}
		if processes[i-1] == nil && processes[i+1] == nil {
			continue
		}
		switch process[0].Kind {
		case "*", "/", "%", "<<", ">>", "&":
			precedence5 = i
		case "+", "-", "|", "^":
			precedence4 = i
		case "==", "!=", "<", "<=", ">", ">=":
			precedence3 = i
		case "&&":
			precedence2 = i
		case "||":
			precedence1 = i
		default:
			p.pusherrtok(process[0], "invalid_operator")
		}
	}
	switch {
	case precedence5 != -1:
		return precedence5
	case precedence4 != -1:
		return precedence4
	case precedence3 != -1:
		return precedence3
	case precedence2 != -1:
		return precedence2
	default:
		return precedence1
	}
}

func (p *Parser) evalToks(toks []lex.Tok) (value, iExpr) {
	return p.evalExpr(new(ast.Builder).Expr(toks))
}

func (p *Parser) evalExpr(ex ast.Expr) (value, iExpr) {
	processes := make([][]lex.Tok, len(ex.Processes))
	copy(processes, ex.Processes)
	return p.evalProcesses(processes)
}

func toRawStrLiteral(literal string) string {
	literal = literal[1 : len(literal)-1] // Remove bounds
	literal = `"(` + literal + `)"`
	literal = xapi.ToRawStr(literal)
	return literal
}

type valueEvaluator struct {
	tok    lex.Tok
	model  *exprModel
	parser *Parser
}

func (p *valueEvaluator) str() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = x.Str
	v.ast.Type.Val = "str"
	if israwstr(p.tok.Kind) {
		p.model.appendSubNode(exprNode{toRawStrLiteral(p.tok.Kind)})
	} else {
		p.model.appendSubNode(exprNode{xapi.ToStr(p.tok.Kind)})
	}
	return v
}

func toRuneLiteral(kind string) string {
	kind = "\"" + kind[1:len(kind)-1] + "\""
	return xapi.ToRune(kind)
}

func (p *valueEvaluator) rune() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = x.Rune
	v.ast.Type.Val = "rune"
	p.model.appendSubNode(exprNode{toRuneLiteral(p.tok.Kind)})
	return v
}

func (p *valueEvaluator) bool() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = x.Bool
	v.ast.Type.Val = "bool"
	p.model.appendSubNode(exprNode{p.tok.Kind})
	return v
}

func (p *valueEvaluator) nil() value {
	var v value
	v.ast.Data = p.tok.Kind
	v.ast.Type.Id = x.Nil
	v.ast.Type.Val = x.NilTypeStr
	p.model.appendSubNode(exprNode{p.tok.Kind})
	return v
}

func (p *valueEvaluator) num() value {
	var v value
	v.ast.Data = p.tok.Kind
	p.model.appendSubNode(exprNode{p.tok.Kind})
	if strings.Contains(p.tok.Kind, ".") ||
		strings.ContainsAny(p.tok.Kind, "eE") {
		v.ast.Type.Id = x.F64
		v.ast.Type.Val = "f64"
	} else {
		v.ast.Type.Id = x.I32
		v.ast.Type.Val = "i32"
		ok := xbits.CheckBitInt(p.tok.Kind, 32)
		if !ok {
			v.ast.Type.Id = x.I64
			v.ast.Type.Val = "i64"
		}
	}
	return v
}

func (p *valueEvaluator) id() (v value, ok bool) {
	id := p.tok.Kind
	if variable := p.parser.varById(id); variable != nil {
		variable.Used = true
		v.ast.Data = id
		v.ast.Type = variable.Type
		v.constant = variable.Const
		v.volatile = variable.Volatile
		v.ast.Tok = variable.IdTok
		v.lvalue = true
		p.model.appendSubNode(exprNode{xapi.AsId(id)})
		ok = true
	} else if f := p.parser.FuncById(id); f != nil {
		f.used = true
		v.ast.Data = id
		v.ast.Type.Id = x.Func
		v.ast.Type.Tag = f.Ast
		v.ast.Type.Val = f.Ast.DataTypeString()
		v.ast.Tok = f.Ast.Tok
		p.model.appendSubNode(exprNode{xapi.AsId(id)})
		ok = true
	} else {
		v.ast.Type.Id = x.Void
		v.ast.Type.Val = x.VoidTypeStr
		p.parser.pusherrtok(p.tok, "id_noexist", id)
	}
	return
}

type solver struct {
	p        *Parser
	left     []lex.Tok
	leftVal  ast.Value
	right    []lex.Tok
	rightVal ast.Value
	operator lex.Tok
	model    *exprModel
}

func (s solver) ptr() (v ast.Value) {
	ok := false
	switch {
	case s.leftVal.Type.Val == s.rightVal.Type.Val:
		ok = true
	case typeIsSingle(s.leftVal.Type):
		switch {
		case s.leftVal.Type.Id == x.Nil,
			x.IsIntegerType(s.leftVal.Type.Id):
			ok = true
		}
	case typeIsSingle(s.rightVal.Type):
		switch {
		case s.rightVal.Type.Id == x.Nil,
			x.IsIntegerType(s.rightVal.Type.Id):
			ok = true
		}
	}
	if !ok {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "+", "-":
		if typeIsPtr(s.leftVal.Type) && typeIsPtr(s.rightVal.Type) {
			s.p.pusherrtok(s.operator, "incompatible_datatype",
				s.rightVal.Type.Val, s.leftVal.Type.Val)
			return
		}
		if typeIsPtr(s.leftVal.Type) {
			v.Type = s.leftVal.Type
		} else {
			v.Type = s.rightVal.Type
		}
	case "!=", "==":
		v.Type.Id = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "pointer")
	}
	return
}

func (s solver) str() (v ast.Value) {
	// Not both string?
	if s.leftVal.Type.Id != s.rightVal.Type.Id {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.leftVal.Type.Val, s.rightVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "+":
		v.Type.Id = x.Str
		v.Type.Val = "str"
	case "==", "!=":
		v.Type.Id = x.Bool
		v.Type.Val = "bool"
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "str")
	}
	return
}

func (s solver) any() (v ast.Value) {
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Id = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "any")
	}
	return
}

func (s solver) bool() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Id = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "bool")
	}
	return
}

func (s solver) float() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstNum(s.leftVal.Data) &&
			!isConstNum(s.rightVal.Data) {
			s.p.pusherrtok(s.operator, "incompatible_datatype",
				s.rightVal.Type.Val, s.leftVal.Type.Val)
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Id = x.Bool
	case "+", "-", "*", "/":
		v.Type.Id = x.F32
		if s.leftVal.Type.Id == x.F64 || s.rightVal.Type.Id == x.F64 {
			v.Type.Id = x.F64
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_float")
	}
	return
}

func (s solver) signed() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstNum(s.leftVal.Data) &&
			!isConstNum(s.rightVal.Data) {
			s.p.pusherrtok(s.operator, "incompatible_datatype",
				s.rightVal.Type.Val, s.leftVal.Type.Val)
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Id = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Id, v.Type.Id) {
			v.Type = s.rightVal.Type
		}
	case ">>", "<<":
		v.Type = s.leftVal.Type
		if !x.IsUnsignedNumericType(s.rightVal.Type.Id) &&
			!checkIntBit(s.rightVal, xbits.BitsizeType(x.U64)) {
			s.p.pusherrtok(s.rightVal.Tok, "bitshift_must_unsigned")
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_int")
	}
	return
}

func (s solver) unsigned() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstNum(s.leftVal.Data) &&
			!isConstNum(s.rightVal.Data) {
			s.p.pusherrtok(s.operator, "incompatible_datatype",
				s.rightVal.Type.Val, s.leftVal.Type.Val)
			return
		}
		return
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Id = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Id, v.Type.Id) {
			v.Type = s.rightVal.Type
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_uint")
	}
	return
}

func (s solver) logical() (v ast.Value) {
	v.Type.Id = x.Bool
	if s.leftVal.Type.Id != x.Bool {
		s.p.pusherrtok(s.leftVal.Tok, "logical_not_bool")
	}
	if s.rightVal.Type.Id != x.Bool {
		s.p.pusherrtok(s.rightVal.Tok, "logical_not_bool")
	}
	return
}

func (s solver) rune() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "!=", "==", ">", "<", ">=", "<=":
		v.Type.Id = x.Bool
	case "+", "-", "*", "/", "^", "&", "%", "|":
		v.Type.Id = x.Rune
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "rune")
	}
	return
}

func (s solver) array() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Id = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "array")
	}
	return
}

func (s solver) nil() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, false) {
		s.p.pusherrtok(s.operator, "incompatible_datatype",
			s.rightVal.Type.Val, s.leftVal.Type.Val)
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Id = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_xtype", "nil")
	}
	return
}

func (s solver) Solve() (v ast.Value) {
	switch s.operator.Kind {
	case "+", "-", "*", "/", "%", ">>",
		"<<", "&", "|", "^", "==", "!=", ">", "<", ">=", "<=":
		break
	case "&&", "||":
		return s.logical()
	default:
		s.p.pusherrtok(s.operator, "invalid_operator")
	}
	switch {
	case typeIsArray(s.leftVal.Type) || typeIsArray(s.rightVal.Type):
		return s.array()
	case typeIsPtr(s.leftVal.Type) || typeIsPtr(s.rightVal.Type):
		return s.ptr()
	case s.leftVal.Type.Id == x.Nil || s.rightVal.Type.Id == x.Nil:
		return s.nil()
	case s.leftVal.Type.Id == x.Rune || s.rightVal.Type.Id == x.Rune:
		return s.rune()
	case s.leftVal.Type.Id == x.Any || s.rightVal.Type.Id == x.Any:
		return s.any()
	case s.leftVal.Type.Id == x.Bool || s.rightVal.Type.Id == x.Bool:
		return s.bool()
	case s.leftVal.Type.Id == x.Str || s.rightVal.Type.Id == x.Str:
		return s.str()
	case x.IsFloatType(s.leftVal.Type.Id) ||
		x.IsFloatType(s.rightVal.Type.Id):
		return s.float()
	case x.IsSignedNumericType(s.leftVal.Type.Id) ||
		x.IsSignedNumericType(s.rightVal.Type.Id):
		return s.signed()
	case x.IsUnsignedNumericType(s.leftVal.Type.Id) ||
		x.IsUnsignedNumericType(s.rightVal.Type.Id):
		return s.unsigned()
	}
	return
}

func (p *Parser) evalSingleExpr(tok lex.Tok, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{tok, m, p}
	v.ast.Type.Id = x.Void
	v.ast.Tok = tok
	switch tok.Id {
	case lex.Value:
		ok = true
		switch {
		case isstr(tok.Kind):
			v = eval.str()
		case isRune(tok.Kind):
			v = eval.rune()
		case isbool(tok.Kind):
			v = eval.bool()
		case isnil(tok.Kind):
			v = eval.nil()
		default:
			v = eval.num()
		}
	case lex.Id:
		v, ok = eval.id()
	default:
		p.pusherrtok(tok, "invalid_syntax")
	}
	return
}

type unaryProcessor struct {
	tok    lex.Tok
	toks   []lex.Tok
	model  *exprModel
	parser *Parser
}

func (p *unaryProcessor) minus() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '-')
	} else if !x.IsNumericType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '-')
	}
	if isConstNum(v.ast.Data) {
		v.ast.Data = "-" + v.ast.Data
	}
	return v
}

func (p *unaryProcessor) plus() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '+')
	} else if !x.IsNumericType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '+')
	}
	return v
}

func (p *unaryProcessor) tilde() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '~')
	} else if !x.IsIntegerType(v.ast.Type.Id) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '~')
	}
	return v
}

func (p *unaryProcessor) logicalNot() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	if !isBoolExpr(v) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '!')
	}
	v.ast.Type.Val = "bool"
	v.ast.Type.Id = x.Bool
	return v
}

func (p *unaryProcessor) star() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	v.lvalue = true
	if !typeIsPtr(v.ast.Type) {
		p.parser.pusherrtok(p.tok, "invalid_type_unary_operator", '*')
	} else {
		v.ast.Type.Val = v.ast.Type.Val[1:]
	}
	return v
}

func (p *unaryProcessor) amper() value {
	v := p.parser.evalExprPart(p.toks, p.model)
	switch {
	case typeIsFunc(v.ast.Type):
		mainNode := &p.model.nodes[p.model.index]
		mainNode.nodes = mainNode.nodes[1:] // Remove unary operator from model
		node := &p.model.nodes[p.model.index].nodes[0]
		switch t := (*node).(type) {
		case anonFuncExpr:
			if t.capture == xapi.LambdaByReference {
				p.parser.pusherrmsgtok(p.tok, "invalid_type_unary_amper")
				break
			}
			t.capture = xapi.LambdaByReference
			*node = t
		default:
			p.parser.pusherrtok(p.tok, "invalid_type_unary_amper")
		}
	default:
		v.lvalue = true
		if !canGetPointer(v) {
			p.parser.pusherrtok(p.tok, "invalid_type_unary_amper")
		}
		v.ast.Type.Val = "*" + v.ast.Type.Val
	}
	return v
}

func (p *Parser) evalOperatorExprPart(toks []lex.Tok, m *exprModel) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprToks := toks[1:]
	processor := unaryProcessor{toks[0], exprToks, m, p}
	m.appendSubNode(exprNode{processor.tok.Kind})
	if processor.toks == nil {
		p.pusherrtok(processor.tok, "invalid_syntax")
		return v
	}
	switch processor.tok.Kind {
	case "-":
		v = processor.minus()
	case "+":
		v = processor.plus()
	case "~":
		v = processor.tilde()
	case "!":
		v = processor.logicalNot()
	case "*":
		v = processor.star()
	case "&":
		v = processor.amper()
	default:
		p.pusherrtok(processor.tok, "invalid_syntax")
	}
	v.ast.Tok = processor.tok
	return v
}

func canGetPointer(v value) bool {
	if v.ast.Type.Id == x.Func {
		return false
	}
	return v.ast.Tok.Id == lex.Id
}

func (p *Parser) evalHeapAllocExpr(toks []lex.Tok, m *exprModel) (v value) {
	if len(toks) == 1 {
		p.pusherrtok(toks[0], "invalid_syntax_keyword_new")
		return
	}
	v.lvalue = true
	v.ast.Tok = toks[0]
	toks = toks[1:]
	b := new(ast.Builder)
	i := new(int)
	dt, ok := b.DataType(toks, i, true)
	m.appendSubNode(newHeapAllocExpr{dt})
	dt.Val = "*" + dt.Val
	v.ast.Type = dt
	if !ok {
		p.pusherrtok(toks[0], "fail_build_heap_allocation_type", dt.Val)
		return
	}
	if *i < len(toks)-1 {
		p.pusherrtok(toks[*i+1], "invalid_syntax")
	}
	return
}

func (p *Parser) evalExprPart(toks []lex.Tok, m *exprModel) (v value) {
	if len(toks) == 1 {
		val, ok := p.evalSingleExpr(toks[0], m)
		if ok {
			v = val
			return
		}
	}
	tok := toks[0]
	switch tok.Id {
	case lex.Operator:
		return p.evalOperatorExprPart(toks, m)
	case lex.New:
		return p.evalHeapAllocExpr(toks, m)
	case lex.Brace:
		switch tok.Kind {
		case "(":
			val, ok := p.evalTryCastExpr(toks, m)
			if ok {
				v = val
				return
			}
			val, ok = p.evalTryAssignExpr(toks, m)
			if ok {
				v = val
				return
			}
		}
	}
	tok = toks[len(toks)-1]
	switch tok.Id {
	case lex.Id:
		return p.evalIdExprPart(toks, m)
	case lex.Operator:
		return p.evalOperatorExprPartRight(toks, m)
	case lex.Brace:
		switch tok.Kind {
		case ")":
			return p.evalParenthesesRangeExpr(toks, m)
		case "}":
			return p.evalBraceRangeExpr(toks, m)
		case "]":
			return p.evalBracketRangeExpr(toks, m)
		}
	default:
		p.pusherrtok(toks[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalStrSubId(val value, idTok lex.Tok, m *exprModel) (v value) {
	i, t := strDefs.defById(idTok.Kind)
	if i == -1 {
		p.pusherrtok(idTok, "obj_have_not_id", val.ast.Type.Val)
		return
	}
	v = val
	m.appendSubNode(exprNode{subIdAccessorOfType(val.ast.Type)})
	switch t {
	case 'g':
		g := strDefs.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.ast.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	}
	return
}

func (p *Parser) evalArraySubId(val value, idTok lex.Tok, m *exprModel) (v value) {
	i, t := arrDefs.defById(idTok.Kind)
	if i == -1 {
		p.pusherrtok(idTok, "obj_have_not_id", val.ast.Type.Val)
		return
	}
	v = val
	m.appendSubNode(exprNode{subIdAccessorOfType(val.ast.Type)})
	switch t {
	case 'g':
		g := arrDefs.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.ast.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	}
	return
}

func (p *Parser) evalMapSubId(val value, idTok lex.Tok, m *exprModel) (v value) {
	readyMapDefs(val.ast.Type)
	i, t := mapDefs.defById(idTok.Kind)
	if i == -1 {
		p.pusherrtok(idTok, "obj_have_not_id", val.ast.Type.Val)
		return
	}
	v = val
	v.lvalue = false
	v.ast.Data = idTok.Kind
	m.appendSubNode(exprNode{subIdAccessorOfType(val.ast.Type)})
	switch t {
	case 'g':
		g := mapDefs.Globals[i]
		m.appendSubNode(exprNode{g.Tag.(string)})
		v.ast.Type = g.Type
		v.lvalue = true
		v.constant = g.Const
	case 'f':
		f := mapDefs.Funcs[i]
		v.ast.Type.Id = x.Func
		v.ast.Type.Tag = f.Ast
		v.ast.Type.Val = f.Ast.DataTypeString()
		v.ast.Tok = f.Ast.Tok
		m.appendSubNode(exprNode{f.Ast.Id})
	}
	return
}

func (p *Parser) evalIdExprPart(toks []lex.Tok, m *exprModel) (v value) {
	i := len(toks) - 1
	tok := toks[i]
	if i <= 0 {
		v, _ = p.evalSingleExpr(tok, m)
		return
	}
	i--
	if i == 0 || toks[i].Id != lex.Dot {
		p.pusherrtok(toks[i], "invalid_syntax")
		return
	}
	idTok := toks[i+1]
	valTok := toks[i]
	toks = toks[:i]
	val := p.evalExprPart(toks, m)
	checkType := val.ast.Type
	if typeIsPtr(checkType) {
		// Remove pointer mark
		checkType.Val = checkType.Val[1:]
	}
	switch {
	case typeIsSingle(checkType) && checkType.Id == x.Str:
		return p.evalStrSubId(val, idTok, m)
	case typeIsArray(checkType):
		return p.evalArraySubId(val, idTok, m)
	case typeIsMap(checkType):
		return p.evalMapSubId(val, idTok, m)
	}
	p.pusherrtok(valTok, "obj_not_support_sub_fields", val.ast.Type.Val)
	return
}

func (p *Parser) evalTryCastExpr(toks []lex.Tok, m *exprModel) (v value, _ bool) {
	braceCount := 0
	errTok := toks[0]
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "[", "{":
				braceCount++
				continue
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		astb := ast.NewBuilder(nil)
		dtindex := 0
		typeToks := toks[1:i]
		dt, ok := astb.DataType(typeToks, &dtindex, false)
		if !ok {
			return
		}
		dt, ok = p.readyType(dt, false)
		if !ok {
			return
		}
		if dtindex+1 < len(typeToks) {
			return
		}
		if i+1 >= len(toks) {
			p.pusherrtok(tok, "casting_missing_expr")
			return
		}
		exprToks := toks[i+1:]
		m.appendSubNode(exprNode{"(" + dt.String() + ")"})
		val := p.evalExprPart(exprToks, m)
		val = p.evalCast(val, dt, errTok)
		return val, true
	}
	return
}

func (p *Parser) evalTryAssignExpr(toks []lex.Tok, m *exprModel) (v value, ok bool) {
	b := ast.NewBuilder(nil)
	toks = toks[1 : len(toks)-1] // Remove first-last parentheses
	assign, ok := b.AssignExpr(toks, true)
	if !ok {
		return
	}
	ok = true
	if len(b.Errs) > 0 {
		p.pusherrs(b.Errs...)
		return
	}
	p.checkAssign(&assign)
	m.appendSubNode(assignExpr{assign})
	v, _ = p.evalExpr(assign.SelectExprs[0].Expr)
	return
}

func (p *Parser) evalCast(v value, t ast.DataType, errtok lex.Tok) value {
	switch {
	case typeIsPtr(t):
		p.checkCastPtr(v.ast.Type, errtok)
	case typeIsArray(t):
		p.checkCastArray(t, v.ast.Type, errtok)
	case typeIsSingle(t):
		v.lvalue = false
		p.checkCastSingle(t, v.ast.Type, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting", t.Val)
	}
	v.ast.Type = t
	v.constant = false
	v.volatile = false
	return v
}

func (p *Parser) checkCastSingle(t, vt ast.DataType, errtok lex.Tok) {
	switch t.Id {
	case x.Str:
		p.checkCastStr(vt, errtok)
		return
	}
	switch {
	case x.IsIntegerType(t.Id):
		p.checkCastInteger(t, vt, errtok)
	case x.IsNumericType(t.Id):
		p.checkCastNumeric(t, vt, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting", t.Val)
	}
}

func (p *Parser) checkCastStr(vt ast.DataType, errtok lex.Tok) {
	if !typeIsArray(vt) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
		return
	}
	vt.Val = vt.Val[2:] // Remove array brackets
	if !typeIsSingle(vt) || (vt.Id != x.Rune && vt.Id != x.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
	}
}

func (p *Parser) checkCastInteger(t, vt ast.DataType, errtok lex.Tok) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsSingle(vt) && x.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Val, t.Val)
}

func (p *Parser) checkCastNumeric(t, vt ast.DataType, errtok lex.Tok) {
	if typeIsSingle(vt) && x.IsNumericType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting_to", vt.Val, t.Val)
}

func (p *Parser) checkCastPtr(vt ast.DataType, errtok lex.Tok) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsSingle(vt) && x.IsIntegerType(vt.Id) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
}

func (p *Parser) checkCastArray(t, vt ast.DataType, errtok lex.Tok) {
	if !typeIsSingle(vt) || vt.Id != x.Str {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
		return
	}
	t.Val = t.Val[2:] // Remove array brackets
	if !typeIsSingle(t) || (t.Id != x.Rune && t.Id != x.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting", vt.Val)
	}
}

func (p *Parser) evalOperatorExprPartRight(toks []lex.Tok, m *exprModel) (v value) {
	tok := toks[len(toks)-1]
	switch tok.Kind {
	case "...":
		toks = toks[:len(toks)-1]
		return p.evalVariadicExprPart(toks, m, tok)
	default:
		p.pusherrtok(tok, "invalid_syntax")
	}
	return
}

func (p *Parser) evalVariadicExprPart(toks []lex.Tok, m *exprModel, errtok lex.Tok) (v value) {
	v = p.evalExprPart(toks, m)
	if !typeIsVariadicable(v.ast.Type) {
		p.pusherrtok(errtok, "variadic_with_nonvariadicable", v.ast.Type.Val)
		return
	}
	v.ast.Type.Val = v.ast.Type.Val[2:] // Remove array type.
	v.variadic = true
	return
}

func (p *Parser) evalParenthesesRangeExpr(toks []lex.Tok, m *exprModel) (v value) {
	var valueToks []lex.Tok
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != lex.Brace {
			continue
		}
		switch tok.Kind {
		case ")", "}", "]":
			braceCount++
		case "(", "{", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueToks = toks[:i]
		break
	}
	if len(valueToks) == 0 && braceCount == 0 {
		// Write parentheses.
		m.appendSubNode(exprNode{"("})
		defer m.appendSubNode(exprNode{")"})

		tk := toks[0]
		toks = toks[1 : len(toks)-1]
		if len(toks) == 0 {
			p.pusherrtok(tk, "invalid_syntax")
		}
		val, model := p.evalToks(toks)
		v = val
		m.appendSubNode(model)
		return
	}
	v = p.evalExprPart(valueToks, m)

	// Write parentheses.
	m.appendSubNode(exprNode{"("})
	defer m.appendSubNode(exprNode{")"})

	switch {
	case typeIsFunc(v.ast.Type):
		f := v.ast.Type.Tag.(ast.Func)
		p.parseFuncCallToks(f, toks[len(valueToks):], m)
		v.ast.Type = f.RetType
		v.lvalue = typeIsLvalue(v.ast.Type)
	default:
		p.pusherrtok(toks[len(valueToks)], "invalid_syntax")
	}
	return
}

func (p *Parser) evalBraceRangeExpr(toks []lex.Tok, m *exprModel) (v value) {
	var exprToks []lex.Tok
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != lex.Brace {
			continue
		}
		switch tok.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	valToksLen := len(exprToks)
	if valToksLen == 0 || braceCount > 0 {
		p.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	switch exprToks[0].Id {
	case lex.Brace:
		switch exprToks[0].Kind {
		case "[":
			b := ast.NewBuilder(nil)
			t, ok := b.DataType(exprToks, new(int), true)
			if !ok {
				p.pusherrs(b.Errs...)
				return
			}
			exprToks = toks[len(exprToks):]
			var model iExpr
			switch {
			case typeIsArray(t):
				v, model = p.buildArray(p.buildEnumerableParts(exprToks), t, exprToks[0])
			case typeIsMap(t):
				v, model = p.buildMap(p.buildEnumerableParts(exprToks), t, exprToks[0])
			}
			m.appendSubNode(model)
			return
		case "(":
			b := ast.NewBuilder(toks)
			f := b.Func(b.Toks, true)
			if len(b.Errs) > 0 {
				p.pusherrs(b.Errs...)
				return
			}
			p.checkAnonFunc(&f)
			v.ast.Type.Tag = f
			v.ast.Type.Id = x.Func
			v.ast.Type.Val = f.DataTypeString()
			m.appendSubNode(anonFuncExpr{f, xapi.LambdaByCopy})
			return
		default:
			p.pusherrtok(exprToks[0], "invalid_syntax")
		}
	default:
		p.pusherrtok(exprToks[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalBracketRangeExpr(toks []lex.Tok, m *exprModel) (v value) {
	var exprToks []lex.Tok
	braceCount := 0
	for i := len(toks) - 1; i >= 0; i-- {
		tok := toks[i]
		if tok.Id != lex.Brace {
			continue
		}
		switch tok.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprToks = toks[:i]
		break
	}
	valToksLen := len(exprToks)
	if valToksLen == 0 || braceCount > 0 {
		p.pusherrtok(toks[0], "invalid_syntax")
		return
	}
	var model iExpr
	v, model = p.evalToks(exprToks)
	m.appendSubNode(model)
	toks = toks[len(exprToks)+1 : len(toks)-1] // Removed array syntax "["..."]"
	m.appendSubNode(exprNode{"["})
	selectv, model := p.evalToks(toks)
	m.appendSubNode(model)
	m.appendSubNode(exprNode{"]"})
	return p.evalEnumerableSelect(v, selectv, toks[0])
}

func (p *Parser) evalEnumerableSelect(enumv, selectv value, errtok lex.Tok) (v value) {
	switch {
	case typeIsArray(enumv.ast.Type):
		return p.evalArraySelect(enumv, selectv, errtok)
	case typeIsMap(enumv.ast.Type):
		return p.evalMapSelect(enumv, selectv, errtok)
	case typeIsSingle(enumv.ast.Type):
		return p.evalStrSelect(enumv, selectv, errtok)
	}
	p.pusherrtok(errtok, "not_enumerable")
	return
}

func (p *Parser) evalArraySelect(arrv, selectv value, errtok lex.Tok) value {
	arrv.lvalue = true
	arrv.ast.Type = typeOfArrayElements(arrv.ast.Type)
	if !typeIsSingle(selectv.ast.Type) ||
		!x.IsIntegerType(selectv.ast.Type.Id) {
		p.pusherrtok(errtok, "notint_array_select")
	}
	return arrv
}

func (p *Parser) evalMapSelect(mapv, selectv value, errtok lex.Tok) value {
	mapv.lvalue = true
	types := mapv.ast.Type.Tag.([]ast.DataType)
	keyType := types[0]
	valType := types[1]
	mapv.ast.Type = valType
	p.wg.Add(1)
	go p.checkTypeAsync(keyType, selectv.ast.Type, false, errtok)
	return mapv
}

func (p *Parser) evalStrSelect(strv, selectv value, errtok lex.Tok) value {
	strv.lvalue = true
	strv.ast.Type.Id = x.Rune
	if !typeIsSingle(selectv.ast.Type) ||
		!x.IsIntegerType(selectv.ast.Type.Id) {
		p.pusherrtok(errtok, "notint_string_select")
	}
	return strv
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (p *Parser) buildEnumerableParts(toks []lex.Tok) [][]lex.Tok {
	toks = toks[1 : len(toks)-1]
	braceCount := 0
	lastComma := -1
	var parts [][]lex.Tok
	for i, tok := range toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tok.Id == lex.Comma {
			if i-lastComma-1 == 0 {
				p.pusherrtok(tok, "missing_expr")
				lastComma = i
				continue
			}
			parts = append(parts, toks[lastComma+1:i])
			lastComma = i
		}
	}
	if lastComma+1 < len(toks) {
		parts = append(parts, toks[lastComma+1:])
	}
	return parts
}

func (p *Parser) buildArray(parts [][]lex.Tok, t ast.DataType, errtok lex.Tok) (value, iExpr) {
	var v value
	v.ast.Type = t
	model := arrayExpr{dataType: t}
	elemType := typeOfArrayElements(t)
	for _, part := range parts {
		partVal, expModel := p.evalToks(part)
		model.expr = append(model.expr, expModel)
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			elemType,
			partVal,
			false,
			part[0],
		}.checkAssignTypeAsync()
	}
	return v, model
}

func (p *Parser) buildMap(parts [][]lex.Tok, t ast.DataType, errtok lex.Tok) (value, iExpr) {
	var v value
	v.ast.Type = t
	model := mapExpr{dataType: t}
	types := t.Tag.([]ast.DataType)
	keyType := types[0]
	valType := types[1]
	for _, part := range parts {
		braceCount := 0
		colon := -1
		for i, tok := range part {
			if tok.Id == lex.Brace {
				switch tok.Kind {
				case "(", "[", "{":
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount != 0 {
				continue
			}
			if tok.Id == lex.Colon {
				colon = i
				break
			}
		}
		if colon < 1 || colon+1 >= len(part) {
			p.pusherrtok(errtok, "missing_expr")
			continue
		}
		colonTok := part[colon]
		keyToks := part[:colon]
		valToks := part[colon+1:]
		key, keyModel := p.evalToks(keyToks)
		model.keyExprs = append(model.keyExprs, keyModel)
		val, valModel := p.evalToks(valToks)
		model.valExprs = append(model.valExprs, valModel)
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			keyType,
			key,
			false,
			colonTok,
		}.checkAssignTypeAsync()
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			valType,
			val,
			false,
			colonTok,
		}.checkAssignTypeAsync()
	}
	return v, model
}

func (p *Parser) checkAnonFunc(f *ast.Func) {
	globals := p.Defs.Globals
	blockVariables := p.BlockVars
	p.Defs.Globals = append(blockVariables, p.Defs.Globals...)
	p.BlockVars = p.varsFromParams(f.Params)
	p.checkFunc(f)
	p.Defs.Globals = globals
	p.BlockVars = blockVariables
}

func (p *Parser) getArgs(toks []lex.Tok) *ast.Args {
	toks, _ = p.getRange("(", ")", toks)
	if toks == nil {
		toks = make([]lex.Tok, 0)
	}
	b := new(ast.Builder)
	args := b.Args(toks)
	if len(b.Errs) > 0 {
		p.pusherrs(b.Errs...)
		args = nil
	}
	return args
}

func (p *Parser) parseFuncCall(f ast.Func, args *ast.Args, m *exprModel, errTok lex.Tok) {
	if args == nil {
		return
	}
	p.parseArgs(&f, args, m, errTok)
	if m != nil {
		m.appendSubNode(argsExpr{args.Src})
	}
}

func (p *Parser) parseFuncCallToks(f ast.Func, argsToks []lex.Tok, m *exprModel) {
	p.parseFuncCall(f, p.getArgs(argsToks), m, argsToks[0])
}

func (p *Parser) parseArgs(f *ast.Func, args *ast.Args, m *exprModel, errTok lex.Tok) {
	if args.Targetted {
		tap := targettedArgParser{
			p:      p,
			f:      f,
			args:   args,
			errTok: errTok,
		}
		tap.parse()
	} else {
		pap := pureArgParser{
			p:      p,
			f:      f,
			args:   args,
			errTok: errTok,
			m:      m,
		}
		pap.parse()
	}
}

func paramHasDefaultArg(param *ast.Param) bool { return param.Default.Toks != nil }

//             [identifier]
type paramMap map[string]*paramMapPair
type paramMapPair struct {
	param *ast.Param
	arg   *ast.Arg
}

func getParamMap(params []ast.Param) *paramMap {
	pmap := new(paramMap)
	*pmap = make(paramMap, len(params))
	for i := range params {
		p := &params[i]
		(*pmap)[p.Id] = &paramMapPair{p, nil}
	}
	return pmap
}

type targettedArgParser struct {
	p      *Parser
	pmap   *paramMap
	f      *ast.Func
	args   *ast.Args
	i      int
	arg    ast.Arg
	errTok lex.Tok
}

func (tap *targettedArgParser) buildArgs() {
	tap.args.Src = make([]ast.Arg, 0)
	for _, p := range tap.f.Params {
		pair := (*tap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			tap.args.Src = append(tap.args.Src, *pair.arg)
		case paramHasDefaultArg(pair.param):
			arg := ast.Arg{Expr: pair.param.Default}
			tap.args.Src = append(tap.args.Src, arg)
		case pair.param.Variadic:
			model := arrayExpr{pair.param.Type, nil}
			model.dataType.Val = "[]" + model.dataType.Val // For array.
			arg := ast.Arg{Expr: ast.Expr{Model: model}}
			tap.args.Src = append(tap.args.Src, arg)
		}
	}
}

func (tap *targettedArgParser) pushVariadicArgs(pair *paramMapPair) {
	model := arrayExpr{pair.param.Type, nil}
	model.dataType.Val = "[]" + model.dataType.Val // For array.
	variadiced := false
	tap.p.parseArg(*pair.param, pair.arg, &variadiced)
	model.expr = append(model.expr, pair.arg.Expr.Model.(iExpr))
	once := false
	for tap.i++; tap.i < len(tap.args.Src); tap.i++ {
		arg := tap.args.Src[tap.i]
		if arg.TargetId != "" {
			tap.i--
			break
		}
		once = true
		tap.p.parseArg(*pair.param, &arg, &variadiced)
		model.expr = append(model.expr, arg.Expr.Model.(iExpr))
	}
	if !once {
		return
	}
	// Variadic argument have one more variadiced expressions.
	if variadiced {
		tap.p.pusherrtok(tap.errTok, "more_args_with_variadiced")
	}
	pair.arg.Expr.Model = model
}

func (tap *targettedArgParser) pushArg() {
	defer func() { tap.i++ }()
	if tap.arg.TargetId == "" {
		tap.p.pusherrtok(tap.arg.Tok, "argument_must_target_to_parameter")
		return
	}
	pair, ok := (*tap.pmap)[tap.arg.TargetId]
	if !ok {
		tap.p.pusherrtok(tap.arg.Tok, "function_not_has_parameter", tap.arg.TargetId)
		return
	} else if pair.arg != nil {
		tap.p.pusherrtok(tap.arg.Tok, "parameter_already_has_argument", tap.arg.TargetId)
		return
	}
	arg := tap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		tap.pushVariadicArgs(pair)
	} else {
		tap.p.parseArg(*pair.param, pair.arg, nil)
	}
}

func (tap *targettedArgParser) checkPasses() {
	for _, pair := range *tap.pmap {
		if pair.arg == nil &&
			!pair.param.Variadic &&
			!paramHasDefaultArg(pair.param) {
			tap.p.pusherrtok(tap.errTok, "missing_argument_for", pair.param.Id)
		}
	}
}

func (tap *targettedArgParser) parse() {
	tap.pmap = getParamMap(tap.f.Params)
	// Check non targetteds
	argCount := 0
	for tap.i, tap.arg = range tap.args.Src {
		if tap.arg.TargetId != "" { // Targetted?
			break
		}
		if argCount >= len(tap.f.Params) {
			tap.p.pusherrtok(tap.errTok, "argument_overflow")
			return
		}
		argCount++
		param := tap.f.Params[tap.i]
		arg := tap.arg
		(*tap.pmap)[param.Id].arg = &arg
		tap.p.parseArg(param, &arg, nil)
	}
	for tap.i < len(tap.args.Src) {
		tap.arg = tap.args.Src[tap.i]
		tap.pushArg()
	}
	tap.checkPasses()
	tap.buildArgs()
}

type pureArgParser struct {
	p       *Parser
	pmap    *paramMap
	f       *ast.Func
	args    *ast.Args
	i       int
	arg     ast.Arg
	errTok  lex.Tok
	m       *exprModel
	paramId string
}

func (pap *pureArgParser) buildArgs() {
	pap.args.Src = make([]ast.Arg, 0)
	for _, p := range pap.f.Params {
		pair := (*pap.pmap)[p.Id]
		switch {
		case pair.arg != nil:
			pap.args.Src = append(pap.args.Src, *pair.arg)
		case paramHasDefaultArg(pair.param):
			arg := ast.Arg{Expr: pair.param.Default}
			pap.args.Src = append(pap.args.Src, arg)
		case pair.param.Variadic:
			model := arrayExpr{pair.param.Type, nil}
			model.dataType.Val = "[]" + model.dataType.Val // For array.
			arg := ast.Arg{Expr: ast.Expr{Model: model}}
			pap.args.Src = append(pap.args.Src, arg)
		}
	}
}

func (pap *pureArgParser) pushVariadicArgs(pair *paramMapPair) {
	model := arrayExpr{pair.param.Type, nil}
	model.dataType.Val = "[]" + model.dataType.Val // For array.
	variadiced := false
	pap.p.parseArg(*pair.param, pair.arg, &variadiced)
	model.expr = append(model.expr, pair.arg.Expr.Model.(iExpr))
	once := false
	for pap.i++; pap.i < len(pap.args.Src); pap.i++ {
		arg := pap.args.Src[pap.i]
		if arg.TargetId != "" {
			pap.i--
			break
		}
		once = true
		pap.p.parseArg(*pair.param, &arg, &variadiced)
		model.expr = append(model.expr, arg.Expr.Model.(iExpr))
	}
	if !once {
		return
	}
	// Variadic argument have one more variadiced expressions.
	if variadiced {
		pap.p.pusherrtok(pap.errTok, "more_args_with_variadiced")
	}
	pair.arg.Expr.Model = model
}

func (pap *pureArgParser) checkPasses() {
	for _, pair := range *pap.pmap {
		if pair.arg == nil &&
			!pair.param.Variadic &&
			!paramHasDefaultArg(pair.param) {
			pap.p.pusherrtok(pap.errTok, "missing_argument_for", pair.param.Id)
		}
	}
}

func (pap *pureArgParser) pushArg() {
	defer func() { pap.i++ }()
	pair, _ := (*pap.pmap)[pap.paramId]
	arg := pap.arg
	pair.arg = &arg
	if pair.param.Variadic {
		pap.pushVariadicArgs(pair)
	} else {
		pap.p.parseArg(*pair.param, pair.arg, nil)
	}
}

func (pap *pureArgParser) parse() {
	if len(pap.args.Src) < len(pap.f.Params) {
		if len(pap.args.Src) == 1 {
			if pap.tryFuncMultiRetAsArgs() {
				return
			}
		}
	}
	pap.pmap = getParamMap(pap.f.Params)
	argCount := 0
	for pap.i < len(pap.args.Src) {
		if argCount >= len(pap.f.Params) {
			pap.p.pusherrtok(pap.errTok, "argument_overflow")
			return
		}
		argCount++
		pap.arg = pap.args.Src[pap.i]
		pap.paramId = pap.f.Params[pap.i].Id
		pap.pushArg()
	}
	pap.checkPasses()
	pap.buildArgs()
}

func (pap *pureArgParser) tryFuncMultiRetAsArgs() bool {
	arg := pap.args.Src[0]
	val, model := pap.p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if !val.ast.Type.MultiTyped {
		return false
	}
	types := val.ast.Type.Tag.([]ast.DataType)
	if len(types) < len(pap.f.Params) {
		return false
	} else if len(types) > len(pap.f.Params) {
		return false
	}
	if pap.m != nil {
		fname := pap.m.nodes[pap.m.index].nodes[0]
		pap.m.nodes[pap.m.index].nodes[0] = exprNode{"tuple_as_args"}
		pap.args.Src = make([]ast.Arg, 2)
		pap.args.Src[0] = ast.Arg{Expr: ast.Expr{Model: fname}}
		pap.args.Src[1] = arg
	}
	for i, param := range pap.f.Params {
		rt := types[i]
		pap.p.wg.Add(1)
		val := value{ast: ast.Value{Type: rt}}
		go pap.p.checkArgTypeAsync(param, val, false, arg.Tok)
	}
	return true
}

func (p *Parser) parseArg(param ast.Param, arg *ast.Arg, variadiced *bool) {
	value, model := p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	p.wg.Add(1)
	go p.checkArgTypeAsync(param, value, false, arg.Tok)
}

func (p *Parser) checkArgTypeAsync(param ast.Param, val value, ignoreAny bool, errTok lex.Tok) {
	defer func() { p.wg.Done() }()
	p.wg.Add(1)
	go assignChecker{
		p,
		param.Const,
		param.Type,
		val,
		false,
		errTok,
	}.checkAssignTypeAsync()
}

// Returns between of brackets.
//
// Special case is:
//  getRange(open, close, tokens) = nil, false if first token is not brace.
func (p *Parser) getRange(open, close string, toks []lex.Tok) (_ []lex.Tok, ok bool) {
	braceCount := 0
	start := 1
	if toks[0].Id != lex.Brace {
		return nil, false
	}
	for i, tok := range toks {
		if tok.Id != lex.Brace {
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

func (p *Parser) checkEntryPointSpecialCases(fun *function) {
	if len(fun.Ast.Params) > 0 {
		p.pusherrtok(fun.Ast.Tok, "entrypoint_have_parameters")
	}
	if fun.Ast.RetType.Id != x.Void {
		p.pusherrtok(fun.Ast.RetType.Tok, "entrypoint_have_return")
	}
	if fun.Attributes != nil {
		p.pusherrtok(fun.Ast.Tok, "entrypoint_have_attributes")
	}
}

func (p *Parser) checkNewBlockCustom(b *ast.Block, oldBlockVars []*ast.Var) {
	b.Gotos = new(ast.Gotos)
	b.Labels = new(ast.Labels)
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
	blockTypes := p.BlockTypes
	p.checkBlock(b)

	vars := p.BlockVars[len(oldBlockVars):]
	types := p.BlockTypes[len(blockTypes):]
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

	p.BlockVars = oldBlockVars
	p.BlockTypes = blockTypes
}

func (p *Parser) checkNewBlock(b *ast.Block) { p.checkNewBlockCustom(b, p.BlockVars) }

func (p *Parser) checkBlock(b *ast.Block) {
	for i := 0; i < len(b.Tree); i++ {
		model := &b.Tree[i]
		switch t := model.Val.(type) {
		case ast.ExprStatement:
			_, t.Expr.Model = p.evalExpr(t.Expr)
			model.Val = t
		case ast.Var:
			p.checkVarStatement(&t, false)
			model.Val = t
		case ast.Assign:
			p.checkAssign(&t)
			model.Val = t
		case ast.Free:
			p.checkFreeStatement(&t)
			model.Val = t
		case ast.Iter:
			p.checkIterExpr(&t)
			model.Val = t
		case ast.Break:
			p.checkBreakStatement(&t)
		case ast.Continue:
			p.checkContinueStatement(&t)
		case ast.If:
			p.checkIfExpr(&t, &i, b.Tree)
			model.Val = t
		case ast.Type:
			if p.checkTypeAST(t) {
				t.Type, _ = p.readyType(t.Type, true)
			}
			p.BlockTypes = append(p.BlockTypes, &t)
			model.Val = nil
		case ast.Block:
			p.checkNewBlock(&t)
			model.Val = t
		case ast.Defer:
			p.checkDeferStatement(&t)
			model.Val = t
		case ast.Label:
			t.Index = i
			t.Block = b
			*p.rootBlock.Labels = append(*p.rootBlock.Labels, &t)
		case ast.Ret:
			rc := retChecker{p: p, retAST: &t, fun: b.Func}
			rc.check()
			model.Val = t
		case ast.Goto:
			t.Index = i
			t.Block = b
			*p.rootBlock.Gotos = append(*p.rootBlock.Gotos, &t)
		case ast.CxxEmbed:
		case ast.Comment:
		default:
			p.pusherrtok(model.Tok, "invalid_syntax")
		}
	}
}

func (p *Parser) findLabel(id string) *ast.Label {
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

func statementIsDef(s *ast.Statement) bool {
	switch t := s.Val.(type) {
	case ast.Var:
		return true
	case ast.Assign:
		for _, selector := range t.SelectExprs {
			if selector.Var.New {
				return true
			}
		}
	}
	return false
}

func (p *Parser) checkSameScopeGoto(gt *ast.Goto, label *ast.Label) {
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

func (p *Parser) checkLabelParents(gt *ast.Goto, label *ast.Label) bool {
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

func (p *Parser) checkDiffScopeGoto(gt *ast.Goto, label *ast.Label) {
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
		switch s.Val.(type) {
		case ast.Block:
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
parent_scopes:
	if block.Parent != nil && block.Parent != gt.Block {
		block = block.Parent
		for i := 0; i < len(block.Tree); i++ {
			s := &block.Tree[i]
			switch {
			case s.Tok.Row >= label.Tok.Row:
				return
			case statementIsDef(s):
				p.pusherrtok(gt.Tok, "goto_jumps_declarations", gt.Label)
				return
			}
		}
		goto parent_scopes
	} else { // goto Scope
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
}

func (p *Parser) checkGoto(gt *ast.Goto, label *ast.Label) {
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

type retChecker struct {
	p        *Parser
	retAST   *ast.Ret
	fun      *ast.Func
	expModel multiRetExpr
	values   []value
}

func (rc *retChecker) pushval(last, current int, errTk lex.Tok) {
	if current-last == 0 {
		rc.p.pusherrtok(errTk, "missing_expr")
		return
	}
	toks := rc.retAST.Expr.Toks[last:current]
	val, model := rc.p.evalToks(toks)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.values = append(rc.values, val)
}

func (rc *retChecker) checkepxrs() {
	braceCount := 0
	last := 0
	for i, tok := range rc.retAST.Expr.Toks {
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case "(", "{", "[":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || tok.Id != lex.Comma {
			continue
		}
		rc.pushval(last, i, tok)
		last = i + 1
	}
	length := len(rc.retAST.Expr.Toks)
	if last < length {
		if last == 0 {
			rc.pushval(0, length, rc.retAST.Tok)
		} else {
			rc.pushval(last, length, rc.retAST.Expr.Toks[last-1])
		}
	}
	if !typeIsVoidRet(rc.fun.RetType) {
		rc.checkExprTypes()
	}
}

func (rc *retChecker) checkExprTypes() {
	valLength := len(rc.values)
	if !rc.fun.RetType.MultiTyped {
		rc.retAST.Expr.Model = rc.expModel.models[0]
		if valLength > 1 {
			rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         rc.fun.RetType,
			v:         rc.values[0],
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
		return
	}
	// Multi return
	rc.retAST.Expr.Model = rc.expModel
	types := rc.fun.RetType.Tag.([]ast.DataType)
	if valLength == 1 {
		rc.checkMultiRetAsMutliRet()
		return
	} else if valLength > len(types) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
	}
	for i, t := range types {
		if i >= valLength {
			break
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         t,
			v:         rc.values[i],
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
	}
}

func (rc *retChecker) checkMultiRetAsMutliRet() {
	val := rc.values[0]
	if !val.ast.Type.MultiTyped {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	}
	valTypes := val.ast.Type.Tag.([]ast.DataType)
	retTypes := rc.fun.RetType.Tag.([]ast.DataType)
	if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "missing_multi_return")
		return
	} else if len(valTypes) < len(retTypes) {
		rc.p.pusherrtok(rc.retAST.Tok, "overflow_return")
		return
	}
	// Set model for just signle return
	rc.retAST.Expr.Model = rc.expModel.models[0]
	for i, rt := range retTypes {
		vt := valTypes[i]
		val := value{ast: ast.Value{Type: vt}}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         rt,
			v:         val,
			ignoreAny: false,
			errtok:    rc.retAST.Tok,
		}.checkAssignTypeAsync()
	}
}

func (rc *retChecker) check() {
	exprToksLen := len(rc.retAST.Expr.Toks)
	if exprToksLen == 0 && !typeIsVoidRet(rc.fun.RetType) {
		rc.p.pusherrtok(rc.retAST.Tok, "require_return_value")
		return
	}
	if exprToksLen > 0 && typeIsVoidRet(rc.fun.RetType) {
		rc.p.pusherrtok(rc.retAST.Tok, "void_function_return_value")
	}
	rc.checkepxrs()
}

func (p *Parser) checkRets(fun *ast.Func) {
	for _, s := range fun.Block.Tree {
		switch s.Val.(type) {
		case ast.Ret:
			return
		}
	}
	if !typeIsVoidRet(fun.RetType) {
		p.pusherrtok(fun.Tok, "missing_ret")
	}
}

func (p *Parser) checkFunc(f *ast.Func) {
	f.Block.Func = f
	p.checkNewBlock(&f.Block)
	p.checkRets(f)
}

func (p *Parser) checkVarStatement(v *ast.Var, noParse bool) {
	if p.existIdf(v.Id, true).Id != lex.NA {
		p.pusherrtok(v.IdTok, "exist_id", v.Id)
	}
	if !noParse {
		*v = *p.Var(*v)
	}
	p.BlockVars = append(p.BlockVars, v)
}

func (p *Parser) checkDeferStatement(d *ast.Defer) {
	tokens := d.Expr.Toks
	if t := tokens[len(tokens)-1]; t.Id != lex.Brace && t.Kind != ")" {
		p.pusherrtok(d.Tok, "defer_expr_not_func_call")
		return
	}
	var exprToks []lex.Tok
	braceCount := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		tok := tokens[i]
		if tok.Id == lex.Brace {
			switch tok.Kind {
			case ")":
				braceCount++
			case "(":
				braceCount--
			}
			if braceCount == 0 {
				exprToks = tokens[:i]
				break
			}
		}
	}
	if len(exprToks) == 0 {
		p.pusherrtok(d.Tok, "defer_expr_not_func_call")
		return
	}
	m := new(exprModel)
	m.nodes = make([]exprBuildNode, 1)
	if !typeIsFunc(p.evalExprPart(exprToks, m).ast.Type) {
		p.pusherrtok(d.Tok, "defer_expr_not_func_call")
		return
	}
	m.nodes[0].nodes = nil
	_ = p.evalExprPart(tokens, m)
	d.Expr.Model = m
}

func (p *Parser) checkAssignment(selected value, errtok lex.Tok) bool {
	state := true
	if !selected.lvalue {
		p.pusherrtok(errtok, "assign_nonlvalue")
		state = false
	}
	if selected.constant {
		p.pusherrtok(errtok, "assign_const")
		state = false
	}
	switch selected.ast.Type.Tag.(type) {
	case ast.Func:
		if p.FuncById(selected.ast.Tok.Kind) != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) checkSingleAssign(assign *ast.Assign) {
	vexpr := &assign.ValueExprs[0]
	val, model := p.evalExpr(*vexpr)
	*vexpr = model.(*exprModel).Expr()
	sexpr := &assign.SelectExprs[0].Expr
	if len(sexpr.Toks) == 1 && xapi.IsIgnoreId(sexpr.Toks[0].Kind) {
		return
	}
	selected, _ := p.evalExpr(*sexpr)
	if !p.checkAssignment(selected, assign.Setter) {
		return
	}
	if assign.Setter.Kind != "=" {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     sexpr.Toks,
			leftVal:  selected.ast,
			right:    vexpr.Toks,
			rightVal: val.ast,
			operator: assign.Setter,
		}
		val.ast = solver.Solve()
		assign.Setter.Kind += "="
	}
	p.wg.Add(1)
	go assignChecker{
		p,
		selected.constant,
		selected.ast.Type,
		val,
		false,
		assign.Setter,
	}.checkAssignTypeAsync()
}

func (p *Parser) assignExprs(vsAST *ast.Assign) []value {
	vals := make([]value, len(vsAST.ValueExprs))
	for i, expr := range vsAST.ValueExprs {
		val, model := p.evalExpr(expr)
		vsAST.ValueExprs[i].Model = model
		vals[i] = val
	}
	return vals
}

func (p *Parser) processFuncMultiAssign(vsAST *ast.Assign, funcVal value) {
	types := funcVal.ast.Type.Tag.([]ast.DataType)
	if len(types) != len(vsAST.SelectExprs) {
		p.pusherrtok(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	vals := make([]value, len(types))
	for i, t := range types {
		vals[i] = value{
			ast: ast.Value{
				Tok:  t.Tok,
				Type: t,
			},
		}
	}
	p.processMultiAssign(vsAST, vals)
}

func (p *Parser) processMultiAssign(assign *ast.Assign, vals []value) {
	for i := range assign.SelectExprs {
		selector := &assign.SelectExprs[i]
		selector.Ignore = xapi.IsIgnoreId(selector.Var.Id)
		val := vals[i]
		if !selector.Var.New {
			if selector.Ignore {
				continue
			}
			selected, _ := p.evalExpr(selector.Expr)
			if !p.checkAssignment(selected, assign.Setter) {
				return
			}
			p.wg.Add(1)
			go assignChecker{
				p,
				selected.constant,
				selected.ast.Type,
				val,
				false,
				assign.Setter,
			}.checkAssignTypeAsync()
			continue
		}
		selector.Var.Tag = val
		p.checkVarStatement(&selector.Var, false)
	}
}

func (p *Parser) checkAssign(assign *ast.Assign) {
	selectLength := len(assign.SelectExprs)
	valueLength := len(assign.ValueExprs)
	if selectLength == 1 && !assign.SelectExprs[0].Var.New {
		p.checkSingleAssign(assign)
		return
	} else if assign.Setter.Kind != "=" {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	if valueLength == 1 {
		firstVal, _ := p.evalExpr(assign.ValueExprs[0])
		if firstVal.ast.Type.MultiTyped {
			assign.MultipleRet = true
			p.processFuncMultiAssign(assign, firstVal)
			return
		}
	}
	switch {
	case selectLength > valueLength:
		p.pusherrtok(assign.Setter, "overflow_multiassign_identifiers")
		return
	case selectLength < valueLength:
		p.pusherrtok(assign.Setter, "missing_multiassign_identifiers")
		return
	}
	p.processMultiAssign(assign, p.assignExprs(assign))
}

func (p *Parser) checkFreeStatement(freeAST *ast.Free) {
	val, model := p.evalExpr(freeAST.Expr)
	freeAST.Expr.Model = model
	if !typeIsPtr(val.ast.Type) {
		p.pusherrtok(freeAST.Tok, "free_nonpointer")
	}
}

func (p *Parser) checkWhileProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.WhileProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	iter.Profile = profile
	if !isBoolExpr(val) {
		p.pusherrtok(iter.Tok, "iter_while_notbool_expr")
	}
	p.checkNewBlock(&iter.Block)
}

type foreachChecker struct {
	p       *Parser
	profile *ast.ForeachProfile
	val     value
}

func (fc *foreachChecker) array() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	elementType := fc.profile.ExprType
	elementType.Val = elementType.Val[2:]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == x.Void {
		keyB.Type = elementType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(elementType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) xmap() {
	fc.checkKeyAMapKey()
	fc.checkKeyBMapVal()
}

func (fc *foreachChecker) checkKeyASize() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == x.Void {
		keyA.Type.Id = x.Size
		keyA.Type.Val = x.CxxTypeIdFromType(keyA.Type.Id)
		return
	}
	var ok bool
	keyA.Type, ok = fc.p.readyType(keyA.Type, true)
	if ok {
		if !typeIsSingle(keyA.Type) || !x.IsNumericType(keyA.Type.Id) {
			fc.p.pusherrtok(keyA.IdTok, "incompatible_datatype",
				keyA.Type.Val, x.NumericTypeStr)
		}
	}
}

func (fc *foreachChecker) checkKeyAMapKey() {
	if xapi.IsIgnoreId(fc.profile.KeyA.Id) {
		return
	}
	keyType := fc.val.ast.Type.Tag.([]ast.DataType)[0]
	keyA := &fc.profile.KeyA
	if keyA.Type.Id == x.Void {
		keyA.Type = keyType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(keyType, keyA.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) checkKeyBMapVal() {
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	valType := fc.val.ast.Type.Tag.([]ast.DataType)[1]
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == x.Void {
		keyB.Type = valType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(valType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) str() {
	fc.checkKeyASize()
	if xapi.IsIgnoreId(fc.profile.KeyB.Id) {
		return
	}
	runeType := ast.DataType{
		Id:  x.Rune,
		Val: x.CxxTypeIdFromType(x.Rune),
	}
	keyB := &fc.profile.KeyB
	if keyB.Type.Id == x.Void {
		keyB.Type = runeType
		return
	}
	fc.p.wg.Add(1)
	go fc.p.checkTypeAsync(runeType, keyB.Type, true, fc.profile.InTok)
}

func (fc *foreachChecker) check() {
	switch {
	case typeIsArray(fc.val.ast.Type):
		fc.array()
	case typeIsMap(fc.val.ast.Type):
		fc.xmap()
	case fc.val.ast.Type.Id == x.Str:
		fc.str()
	}
}

func (p *Parser) checkForeachProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.ForeachProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	profile.ExprType = val.ast.Type
	if !isForeachIterExpr(val) {
		p.pusherrtok(iter.Tok, "iter_foreach_nonenumerable_expr")
	} else {
		fc := foreachChecker{p, &profile, val}
		fc.check()
	}
	iter.Profile = profile
	blockVars := p.BlockVars
	if profile.KeyA.New {
		if xapi.IsIgnoreId(profile.KeyA.Id) {
			p.pusherrtok(profile.KeyA.IdTok, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyA, true)
	}
	if profile.KeyB.New {
		if xapi.IsIgnoreId(profile.KeyB.Id) {
			p.pusherrtok(profile.KeyB.IdTok, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyB, true)
	}
	p.checkNewBlockCustom(&iter.Block, blockVars)
}

func (p *Parser) checkIterExpr(iter *ast.Iter) {
	p.iterCount++
	if iter.Profile != nil {
		switch iter.Profile.(type) {
		case ast.WhileProfile:
			p.checkWhileProfile(iter)
		case ast.ForeachProfile:
			p.checkForeachProfile(iter)
		}
	}
	p.iterCount--
}

func (p *Parser) checkIfExpr(ifast *ast.If, i *int, statements []ast.Statement) {
	val, model := p.evalExpr(ifast.Expr)
	ifast.Expr.Model = model
	statement := statements[*i]
	if !isBoolExpr(val) {
		p.pusherrtok(ifast.Tok, "if_notbool_expr")
	}
	p.checkNewBlock(&ifast.Block)
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
	switch t := statement.Val.(type) {
	case ast.ElseIf:
		val, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		if !isBoolExpr(val) {
			p.pusherrtok(t.Tok, "if_notbool_expr")
		}
		p.checkNewBlock(&t.Block)
		goto node
	case ast.Else:
		p.checkElseBlock(&t)
		statement.Val = t
	default:
		*i--
	}
}

func (p *Parser) checkElseBlock(elseast *ast.Else) { p.checkNewBlock(&elseast.Block) }

func (p *Parser) checkBreakStatement(breakAST *ast.Break) {
	if p.iterCount == 0 {
		p.pusherrtok(breakAST.Tok, "break_at_outiter")
	}
}

func (p *Parser) checkContinueStatement(continueAST *ast.Continue) {
	if p.iterCount == 0 {
		p.pusherrtok(continueAST.Tok, "continue_at_outiter")
	}
}

func (p *Parser) checkValidityForAutoType(t ast.DataType, err lex.Tok) {
	switch t.Id {
	case x.Nil:
		p.pusherrtok(err, "nil_for_autotype")
	case x.Void:
		p.pusherrtok(err, "void_for_autotype")
	}
}

func (p *Parser) defaultValueOfType(t ast.DataType) string {
	if typeIsPtr(t) || typeIsArray(t) {
		return "nil"
	}
	return x.DefaultValOfType(t.Id)
}

func (p *Parser) readyType(dt ast.DataType, err bool) (_ ast.DataType, ok bool) {
	if dt.Val == "" {
		return dt, true
	}
	if dt.MultiTyped {
		types := dt.Tag.([]ast.DataType)
		for i, t := range types {
			t, okr := p.readyType(t, err)
			types[i] = t
			if ok {
				ok = okr
			}
		}
		dt.Tag = types
		return dt, ok
	}
	switch dt.Id {
	case x.Id:
		t := p.typeById(dt.Tok.Kind)
		if t == nil {
			if err {
				p.pusherrtok(dt.Tok, "invalid_type_source")
			}
			return dt, false
		}
		t.Used = true
		dt = t.Type
		dt.Val = dt.Val[:len(dt.Val)-len(dt.Tok.Kind)] + t.Type.Val
		return p.readyType(dt, err)
	case x.Func:
		f := dt.Tag.(ast.Func)
		for i, param := range f.Params {
			f.Params[i].Type, _ = p.readyType(param.Type, err)
		}
		f.RetType, _ = p.readyType(f.RetType, err)
		dt.Val = dt.Tag.(ast.Func).DataTypeString()
	}
	return dt, true
}

func (p *Parser) checkMultiTypeAsync(real, check ast.DataType, ignoreAny bool, errTok lex.Tok) {
	defer func() { p.wg.Done() }()
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		return
	}
	realTypes := real.Tag.([]ast.DataType)
	checkTypes := real.Tag.([]ast.DataType)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		return
	}
	for i := 0; i < len(realTypes); i++ {
		realType := realTypes[i]
		checkType := checkTypes[i]
		p.checkTypeAsync(realType, checkType, ignoreAny, errTok)
	}
}

func (p *Parser) checkAssignConst(constant bool, t ast.DataType, val value, errTok lex.Tok) {
	if typeIsMut(t) && val.constant && !constant {
		p.pusherrtok(errTok, "constant_assignto_nonconstant")
	}
}

type assignChecker struct {
	p         *Parser
	constant  bool
	t         ast.DataType
	v         value
	ignoreAny bool
	errtok    lex.Tok
}

func (ac assignChecker) checkAssignTypeAsync() {
	defer func() { ac.p.wg.Done() }()
	ac.p.checkAssignConst(ac.constant, ac.t, ac.v, ac.errtok)
	if typeIsSingle(ac.t) && isConstNum(ac.v.ast.Data) {
		switch {
		case x.IsSignedIntegerType(ac.t.Id):
			if xbits.CheckBitInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Id)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype", ac.t, ac.v.ast.Type)
			return
		case x.IsFloatType(ac.t.Id):
			if checkFloatBit(ac.v.ast, xbits.BitsizeType(ac.t.Id)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype", ac.t, ac.v.ast.Type)
			return
		case x.IsUnsignedNumericType(ac.t.Id):
			if xbits.CheckBitUInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Id)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype", ac.t, ac.v.ast.Type)
			return
		}
	}
	ac.p.wg.Add(1)
	go ac.p.checkTypeAsync(ac.t, ac.v.ast.Type, ac.ignoreAny, ac.errtok)
}

func (p *Parser) checkTypeAsync(real, check ast.DataType, ignoreAny bool, errTok lex.Tok) {
	defer func() { p.wg.Done() }()
	if !ignoreAny && real.Id == x.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.wg.Add(1)
		go p.checkMultiTypeAsync(real, check, ignoreAny, errTok)
		return
	}
	if typeIsSingle(real) && typeIsSingle(check) {
		if !typesAreCompatible(real, check, ignoreAny) {
			p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
		}
		return
	}
	if typeIsNilCompatible(real) && check.Id == x.Nil {
		return
	}
	if real.Val != check.Val {
		p.pusherrtok(errTok, "incompatible_datatype", real.Val, check.Val)
	}
}
