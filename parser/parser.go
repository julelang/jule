package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/the-xlang/x/ast"
	"github.com/the-xlang/x/lex"
	"github.com/the-xlang/x/pkg/x"
	"github.com/the-xlang/x/pkg/xapi"
	"github.com/the-xlang/x/pkg/xbits"
	"github.com/the-xlang/x/pkg/xio"
	"github.com/the-xlang/x/pkg/xlog"
	"github.com/the-xlang/x/preprocessor"
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

	Embeds         strings.Builder
	Uses           []*use
	Defs           *defmap
	waitingGlobals []ast.Var
	BlockVars      []ast.Var
	Errors         []xlog.CompilerLog
	Warnings       []xlog.CompilerLog
	File           *xio.File
}

// NewParser returns new instance of Parser.
func NewParser(f *xio.File) *Parser {
	parser := new(Parser)
	parser.File = f
	parser.isLocalPkg = false
	parser.Defs = new(defmap)
	return parser
}

// Parses object tree and returns parser.
func Parset(tree []ast.Obj, main, justDefs bool) *Parser {
	p := NewParser(nil)
	p.Parset(tree, main, justDefs)
	return p
}

// pusherrtok appends new error by token.
func (p *Parser) pusherrtok(tok lex.Token, key string) {
	p.pusherrmsgtok(tok, x.Errors[key])
}

// pusherrtok appends new error message by token.
func (p *Parser) pusherrmsgtok(tok lex.Token, msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.Error,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path,
		Message: msg,
	})
}

// pushwarntok appends new warning by token.
func (p *Parser) pushwarntok(tok lex.Token, key string) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.Warning,
		Row:     tok.Row,
		Column:  tok.Column,
		Path:    tok.File.Path,
		Message: x.Warns[key],
	})
}

// pusherrs appends specified errors.
func (p *Parser) pusherrs(errs ...xlog.CompilerLog) {
	p.Errors = append(p.Errors, errs...)
}

// pusherr appends new error.
func (p *Parser) pusherr(key string) {
	p.pusherrmsg(x.Errors[key])
}

// pusherrmsh appends new flat error message
func (p *Parser) pusherrmsg(msg string) {
	p.Errors = append(p.Errors, xlog.CompilerLog{
		Type:    xlog.FlatError,
		Message: msg,
	})
}

// pusherr appends new warning.
func (p *Parser) pushwarn(key string) {
	p.Warnings = append(p.Warnings, xlog.CompilerLog{
		Type:    xlog.FlatWarning,
		Message: x.Warns[key],
	})
}

// String returns full C++ code of parsed objects.
func (p Parser) String() string { return p.Cxx() }

// CxxEmbeds return C++ code of cxx embeds.
func (p *Parser) CxxEmbeds() string {
	var cxx strings.Builder
	cxx.WriteString("// region EMBEDS\n")
	cxx.WriteString(p.Embeds.String())
	cxx.WriteString("// endregion EMBEDS")
	return cxx.String()
}

// CxxPrototypes returns C++ code of prototypes of C++ code.
func (p *Parser) CxxPrototypes() string {
	var cxx strings.Builder
	cxx.WriteString("// region PROTOTYPES\n")
	for _, use := range used {
		for _, f := range use.defs.Funcs {
			cxx.WriteString(f.Prototype())
			cxx.WriteByte('\n')
		}
	}
	for _, f := range p.Defs.Funcs {
		cxx.WriteString(f.Prototype())
		cxx.WriteByte('\n')
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
			cxx.WriteString(v.String())
			cxx.WriteByte('\n')
		}
	}
	for _, v := range p.Defs.Globals {
		cxx.WriteString(v.String())
		cxx.WriteByte('\n')
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
			cxx.WriteString(f.String())
			cxx.WriteString("\n\n")
		}
	}
	for _, f := range p.Defs.Funcs {
		cxx.WriteString(f.String())
		cxx.WriteString("\n\n")
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

func getTree(tokens []lex.Token, errs *[]xlog.CompilerLog) []ast.Obj {
	b := ast.NewBuilder(tokens)
	b.Build()
	if len(b.Errors) > 0 {
		if errs != nil {
			*errs = append(*errs, b.Errors...)
		}
		return nil
	}
	return b.Tree
}

func (p *Parser) checkUsePath(use *ast.Use) bool {
	info, err := os.Stat(use.Path)
	// Exists directory?
	if err != nil || !info.IsDir() {
		p.pusherrtok(use.Token, "use_not_found")
		return false
	}
	// Already uses?
	for _, puse := range p.Uses {
		if use.Path == puse.Path {
			p.pusherrtok(use.Token, "already_uses")
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
		psub := NewParser(f)
		psub.Parsef(false, false)
		if psub.Errors != nil {
			p.pusherrtok(useAST.Token, "use_has_errors")
		}
		use := new(use)
		use.defs = new(defmap)
		use.Path = useAST.Path
		p.pusherrs(psub.Errors...)
		p.Warnings = append(p.Warnings, psub.Warnings...)
		p.pushUseDefs(use, psub.Defs)
		return use
	}
	return nil
}

func (p *Parser) pushUseTypes(use *use, dm *defmap) {
	for _, t := range dm.Types {
		def := p.typeById(t.Id)
		if def != nil {
			p.pusherrmsgtok(def.Token,
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
			p.pusherrmsgtok(def.IdToken,
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
			p.pusherrmsgtok(def.Ast.Token,
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
		p.pusherrtok(obj.Token, "use_at_content")
	case ast.Preprocessor:
	default:
		p.pusherrtok(obj.Token, "invalid_syntax")
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
		tokens := lexer.Lex()
		if lexer.Logs != nil {
			p.pusherrs(lexer.Logs...)
			continue
		}
		subtree := getTree(tokens, &p.Errors)
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
func (p *Parser) Parse(tokens []lex.Token, main, justDefs bool) {
	tree := getTree(tokens, &p.Errors)
	if tree == nil {
		return
	}
	p.Parset(tree, main, justDefs)
}

// Parses X code from file.
func (p *Parser) Parsef(main, justDefs bool) {
	lexer := lex.NewLex(p.File)
	tokens := lexer.Lex()
	if lexer.Logs != nil {
		p.pusherrs(lexer.Logs...)
		return
	}
	p.Parse(tokens, main, justDefs)
}

func (p *Parser) checkDoc(obj ast.Obj) {
	if p.docText.Len() == 0 {
		return
	}
	switch obj.Value.(type) {
	case ast.Comment, ast.Attribute:
		return
	}
	p.pushwarntok(obj.Token, "doc_ignored")
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
	p.pusherrtok(obj.Token, "attribute_not_supports")
	p.attributes = nil
}

// Type parses X type define statement.
func (p *Parser) Type(t ast.Type) {
	if p.existid(t.Id).Id != lex.NA {
		p.pusherrtok(t.Token, "exist_id")
		return
	} else if xapi.IsIgnoreId(t.Id) {
		p.pusherrtok(t.Token, "ignore_id")
		return
	}
	t.Description = p.docText.String()
	p.docText.Reset()
	p.Defs.Types = append(p.Defs.Types, t)
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
		p.pusherrtok(attribute.Tag, "undefined_tag")
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
	switch t := s.Value.(type) {
	case ast.Func:
		p.Func(t)
	case ast.Var:
		p.Global(t)
	default:
		p.pusherrtok(s.Token, "invalid_syntax")
	}
}

// Func parse X function.
func (p *Parser) Func(fast ast.Func) {
	if p.existid(fast.Id).Id != lex.NA {
		p.pusherrtok(fast.Token, "exist_id")
	} else if xapi.IsIgnoreId(fast.Id) {
		p.pusherrtok(fast.Token, "ignore_id")
	}
	fast.RetType, _ = p.readyType(fast.RetType, true)
	for i, param := range fast.Params {
		fast.Params[i].Type, _ = p.readyType(param.Type, true)
	}
	f := new(function)
	f.Ast = fast
	f.Attributes = p.attributes
	f.Description = p.docText.String()
	p.attributes = nil
	p.docText.Reset()
	p.checkFuncAttributes(f.Attributes)
	p.Defs.Funcs = append(p.Defs.Funcs, f)
}

// ParseVariable parse X global variable.
func (p *Parser) Global(vast ast.Var) {
	if p.existid(vast.Id).Id != lex.NA {
		p.pusherrtok(vast.IdToken, "exist_id")
		return
	}
	vast.Description = p.docText.String()
	p.docText.Reset()
	p.waitingGlobals = append(p.waitingGlobals, vast)
}

// Var parse X variable.
func (p *Parser) Var(vast ast.Var) ast.Var {
	if xapi.IsIgnoreId(vast.Id) {
		p.pusherrtok(vast.IdToken, "ignore_id")
	}
	var val value
	switch t := vast.Tag.(type) {
	case value:
		val = t
	default:
		if vast.SetterToken.Id != lex.NA {
			val, vast.Value.Model = p.evalExpr(vast.Value)
		}
	}
	if vast.Type.Code != x.Void {
		if vast.SetterToken.Id != lex.NA {
			p.wg.Add(1)
			go assignChecker{
				p,
				vast.Const,
				vast.Type,
				val,
				false,
				vast.IdToken,
			}.checkAssignTypeAsync()
		} else { // Pass default value.
			dt, ok := p.readyType(vast.Type, true)
			if ok {
				var valueToken lex.Token
				valueToken.Id = lex.Value
				valueToken.Kind = p.defaultValueOfType(dt)
				valueTokens := []lex.Token{valueToken}
				processes := [][]lex.Token{valueTokens}
				vast.Value = ast.Expr{
					Tokens:    valueTokens,
					Processes: processes,
				}
				_, vast.Value.Model = p.evalExpr(vast.Value)
			}
		}
	} else {
		if vast.SetterToken.Id == lex.NA {
			p.pusherrtok(vast.IdToken, "missing_autotype_value")
		} else {
			vast.Type = val.ast.Type
			p.checkValidityForAutoType(vast.Type, vast.SetterToken)
			p.checkAssignConst(vast.Const, vast.Type, val, vast.SetterToken)
		}
	}
	if vast.Const {
		if vast.SetterToken.Id == lex.NA {
			p.pusherrtok(vast.IdToken, "missing_const_value")
		}
	}
	return vast
}

func (p *Parser) checkFuncAttributes(attributes []ast.Attribute) {
	for _, attribute := range attributes {
		switch attribute.Tag.Kind {
		case "inline":
		default:
			p.pusherrtok(attribute.Token, "invalid_attribute")
		}
	}
}

func (p *Parser) varsFromParams(params []ast.Parameter) []ast.Var {
	var vars []ast.Var
	length := len(params)
	for index, param := range params {
		var vast ast.Var
		vast.Id = param.Id
		vast.IdToken = param.Token
		vast.Type = param.Type
		vast.Const = param.Const
		vast.Volatile = param.Volatile
		if param.Variadic {
			if length-index > 1 {
				p.pusherrtok(param.Token, "variadic_parameter_notlast")
			}
			vast.Type.Value = "[]" + vast.Type.Value
		}
		vars = append(vars, vast)
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
		if v.Id == id {
			return &v
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
	for _, use := range p.Uses {
		t := use.defs.typeById(id)
		if t != nil && t.Pub {
			return t
		}
	}
	return p.Defs.typeById(id)
}

func (p *Parser) existIdf(id string, exceptGlobals bool) lex.Token {
	t := p.typeById(id)
	if t != nil {
		return t.Token
	}
	f := p.FuncById(id)
	if f != nil {
		return f.Ast.Token
	}
	for _, v := range p.BlockVars {
		if v.Id == id {
			return v.IdToken
		}
	}
	if !exceptGlobals {
		v := p.globalById(id)
		if v != nil {
			return v.IdToken
		}
		for _, v := range p.waitingGlobals {
			if v.Id == id {
				return v.IdToken
			}
		}
	}
	return lex.Token{}
}

func (p *Parser) existid(id string) lex.Token {
	return p.existIdf(id, false)
}

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
	for _, t := range p.Defs.Types {
		_, _ = p.readyType(t.Type, true)
	}
}

// WaitingGlobals parse X global variables for waiting parsing.
func (p *Parser) WaitingGlobals() {
	for _, varAST := range p.waitingGlobals {
		variable := p.Var(varAST)
		p.Defs.Globals = append(p.Defs.Globals, variable)
	}
}

func (p *Parser) checkFuncsAsync() {
	defer func() { p.wg.Done() }()
	for _, f := range p.Defs.Funcs {
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

func eliminateProcesses(processes *[][]lex.Token, i, to int) {
	for i < to {
		(*processes)[i] = nil
		i++
	}
}

func (p *Parser) evalProcesses(processes [][]lex.Token) (v value, e iExpr) {
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
			boolean = v.ast.Type.Code == x.Bool
		}
		if boolean {
			v.ast.Type.Code = x.Bool
		}
		m.index = i
		process.operator = processes[m.index][0]
		m.appendNodeToSubNodes(exprNode{process.operator.Kind})
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
		if v.ast.Type.Code != x.Void {
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

func noData(processes [][]lex.Token) bool {
	for _, p := range processes {
		if !isOperator(p) && p != nil {
			return false
		}
	}
	return true
}

func isOperator(process []lex.Token) bool {
	return len(process) == 1 && process[0].Id == lex.Operator
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func (p *Parser) nextOperator(processes [][]lex.Token) int {
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

func (p *Parser) evalTokens(tokens []lex.Token) (value, iExpr) {
	return p.evalExpr(new(ast.Builder).Expr(tokens))
}

func (p *Parser) evalExpr(ex ast.Expr) (value, iExpr) {
	processes := make([][]lex.Token, len(ex.Processes))
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
	token  lex.Token
	model  *exprModel
	parser *Parser
}

func (p *valueEvaluator) str() value {
	var v value
	v.ast.Data = p.token.Kind
	v.ast.Type.Code = x.Str
	v.ast.Type.Value = "str"
	if israwstr(p.token.Kind) {
		p.model.appendNodeToSubNodes(exprNode{toRawStrLiteral(p.token.Kind)})
	} else {
		p.model.appendNodeToSubNodes(exprNode{xapi.ToStr(p.token.Kind)})
	}
	return v
}

func (p *valueEvaluator) rune() value {
	var v value
	v.ast.Data = p.token.Kind
	v.ast.Type.Code = x.Rune
	v.ast.Type.Value = "rune"
	p.model.appendNodeToSubNodes(exprNode{xapi.ToRune(p.token.Kind)})
	return v
}

func (p *valueEvaluator) bool() value {
	var v value
	v.ast.Data = p.token.Kind
	v.ast.Type.Code = x.Bool
	v.ast.Type.Value = "bool"
	p.model.appendNodeToSubNodes(exprNode{p.token.Kind})
	return v
}

func (p *valueEvaluator) nil() value {
	var v value
	v.ast.Data = p.token.Kind
	v.ast.Type.Code = x.Nil
	p.model.appendNodeToSubNodes(exprNode{p.token.Kind})
	return v
}

func (p *valueEvaluator) num() value {
	var v value
	v.ast.Data = p.token.Kind
	p.model.appendNodeToSubNodes(exprNode{p.token.Kind})
	if strings.Contains(p.token.Kind, ".") ||
		strings.ContainsAny(p.token.Kind, "eE") {
		v.ast.Type.Code = x.F64
		v.ast.Type.Value = "f64"
	} else {
		v.ast.Type.Code = x.I32
		v.ast.Type.Value = "i32"
		ok := xbits.CheckBitInt(p.token.Kind, 32)
		if !ok {
			v.ast.Type.Code = x.I64
			v.ast.Type.Value = "i64"
		}
	}
	return v
}

func (p *valueEvaluator) id() (v value, ok bool) {
	if variable := p.parser.varById(p.token.Kind); variable != nil {
		v.ast.Data = p.token.Kind
		v.ast.Type = variable.Type
		v.constant = variable.Const
		v.volatile = variable.Volatile
		v.ast.Token = variable.IdToken
		v.lvalue = true
		p.model.appendNodeToSubNodes(exprNode{xapi.AsId(p.token.Kind)})
		ok = true
	} else if fun := p.parser.FuncById(p.token.Kind); fun != nil {
		v.ast.Data = p.token.Kind
		v.ast.Type.Code = x.Func
		v.ast.Type.Tag = fun.Ast
		v.ast.Type.Value = fun.Ast.DataTypeString()
		v.ast.Token = fun.Ast.Token
		p.model.appendNodeToSubNodes(exprNode{xapi.AsId(p.token.Kind)})
		ok = true
	} else {
		p.parser.pusherrtok(p.token, "id_noexist")
	}
	return
}

type solver struct {
	p        *Parser
	left     []lex.Token
	leftVal  ast.Value
	right    []lex.Token
	rightVal ast.Value
	operator lex.Token
	model    *exprModel
}

func (s solver) ptr() (v ast.Value) {
	ok := false
	switch {
	case s.leftVal.Type.Value == s.rightVal.Type.Value:
		ok = true
	case typeIsSingle(s.leftVal.Type):
		switch {
		case s.leftVal.Type.Code == x.Nil,
			x.IsIntegerType(s.leftVal.Type.Code):
			ok = true
		}
	case typeIsSingle(s.rightVal.Type):
		switch {
		case s.rightVal.Type.Code == x.Nil,
			x.IsIntegerType(s.rightVal.Type.Code):
			ok = true
		}
	}
	if !ok {
		s.p.pusherrtok(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "+", "-":
		if typeIsPtr(s.leftVal.Type) && typeIsPtr(s.rightVal.Type) {
			s.p.pusherrtok(s.operator, "incompatible_type")
			return
		}
		if typeIsPtr(s.leftVal.Type) {
			v.Type = s.leftVal.Type
		} else {
			v.Type = s.rightVal.Type
		}
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_pointer")
	}
	return
}

func (s solver) str() (v ast.Value) {
	// Not both string?
	if s.leftVal.Type.Code != s.rightVal.Type.Code {
		s.p.pusherrtok(s.operator, "incompatible_datatype")
		return
	}
	switch s.operator.Kind {
	case "+":
		v.Type.Code = x.Str
	case "==", "!=":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_string")
	}
	return
}

func (s solver) any() (v ast.Value) {
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_any")
	}
	return
}

func (s solver) bool() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_bool")
	}
	return
}

func (s solver) float() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		if !isConstNum(s.leftVal.Data) &&
			!isConstNum(s.rightVal.Data) {
			s.p.pusherrtok(s.operator, "incompatible_type")
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/":
		v.Type.Code = x.F32
		if s.leftVal.Type.Code == x.F64 || s.rightVal.Type.Code == x.F64 {
			v.Type.Code = x.F64
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
			s.p.pusherrtok(s.operator, "incompatible_type")
			return
		}
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Code, v.Type.Code) {
			v.Type = s.rightVal.Type
		}
	case ">>", "<<":
		v.Type = s.leftVal.Type
		if !x.IsUnsignedNumericType(s.rightVal.Type.Code) &&
			!checkIntBit(s.rightVal, xbits.BitsizeType(x.U64)) {
			s.p.pusherrtok(s.rightVal.Token, "bitshift_must_unsigned")
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
			s.p.pusherrtok(s.operator, "incompatible_type")
			return
		}
		return
	}
	switch s.operator.Kind {
	case "!=", "==", "<", ">", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "%", "&", "|", "^":
		v.Type = s.leftVal.Type
		if x.TypeGreaterThan(s.rightVal.Type.Code, v.Type.Code) {
			v.Type = s.rightVal.Type
		}
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_uint")
	}
	return
}

func (s solver) logical() (v ast.Value) {
	v.Type.Code = x.Bool
	if s.leftVal.Type.Code != x.Bool {
		s.p.pusherrtok(s.leftVal.Token, "logical_not_bool")
	}
	if s.rightVal.Type.Code != x.Bool {
		s.p.pusherrtok(s.rightVal.Token, "logical_not_bool")
	}
	return
}

func (s solver) rune() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==", ">", "<", ">=", "<=":
		v.Type.Code = x.Bool
	case "+", "-", "*", "/", "^", "&", "%", "|":
		v.Type.Code = x.Rune
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_rune")
	}
	return
}

func (s solver) array() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, true) {
		s.p.pusherrtok(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_array")
	}
	return
}

func (s solver) nil() (v ast.Value) {
	if !typesAreCompatible(s.leftVal.Type, s.rightVal.Type, false) {
		s.p.pusherrtok(s.operator, "incompatible_type")
		return
	}
	switch s.operator.Kind {
	case "!=", "==":
		v.Type.Code = x.Bool
	default:
		s.p.pusherrtok(s.operator, "operator_notfor_nil")
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
	case s.leftVal.Type.Code == x.Nil || s.rightVal.Type.Code == x.Nil:
		return s.nil()
	case s.leftVal.Type.Code == x.Rune || s.rightVal.Type.Code == x.Rune:
		return s.rune()
	case s.leftVal.Type.Code == x.Any || s.rightVal.Type.Code == x.Any:
		return s.any()
	case s.leftVal.Type.Code == x.Bool || s.rightVal.Type.Code == x.Bool:
		return s.bool()
	case s.leftVal.Type.Code == x.Str || s.rightVal.Type.Code == x.Str:
		return s.str()
	case x.IsFloatType(s.leftVal.Type.Code) ||
		x.IsFloatType(s.rightVal.Type.Code):
		return s.float()
	case x.IsSignedNumericType(s.leftVal.Type.Code) ||
		x.IsSignedNumericType(s.rightVal.Type.Code):
		return s.signed()
	case x.IsUnsignedNumericType(s.leftVal.Type.Code) ||
		x.IsUnsignedNumericType(s.rightVal.Type.Code):
		return s.unsigned()
	}
	return
}

func (p *Parser) evalSingleExpr(token lex.Token, m *exprModel) (v value, ok bool) {
	eval := valueEvaluator{token, m, p}
	v.ast.Type.Code = x.Void
	v.ast.Token = token
	switch token.Id {
	case lex.Value:
		ok = true
		switch {
		case isstr(token.Kind):
			v = eval.str()
		case isrune(token.Kind):
			v = eval.rune()
		case isbool(token.Kind):
			v = eval.bool()
		case isnil(token.Kind):
			v = eval.nil()
		default:
			v = eval.num()
		}
	case lex.Id:
		v, ok = eval.id()
	default:
		p.pusherrtok(token, "invalid_syntax")
	}
	return
}

type operatorProcessor struct {
	token  lex.Token
	tokens []lex.Token
	model  *exprModel
	parser *Parser
}

func (p *operatorProcessor) unary() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.token, "invalid_data_unary")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.pusherrtok(p.token, "invalid_data_unary")
	}
	if isConstNum(v.ast.Data) {
		v.ast.Data = "-" + v.ast.Data
	}
	return v
}

func (p *operatorProcessor) plus() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.token, "invalid_data_plus")
	} else if !x.IsNumericType(v.ast.Type.Code) {
		p.parser.pusherrtok(p.token, "invalid_data_plus")
	}
	return v
}

func (p *operatorProcessor) tilde() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	if !typeIsSingle(v.ast.Type) {
		p.parser.pusherrtok(p.token, "invalid_data_tilde")
	} else if !x.IsIntegerType(v.ast.Type.Code) {
		p.parser.pusherrtok(p.token, "invalid_data_tilde")
	}
	return v
}

func (p *operatorProcessor) logicalNot() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	if !isBoolExpr(v) {
		p.parser.pusherrtok(p.token, "invalid_data_logical_not")
	}
	v.ast.Type.Value = "bool"
	v.ast.Type.Code = x.Bool
	return v
}

func (p *operatorProcessor) star() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	v.lvalue = true
	if !typeIsPtr(v.ast.Type) {
		p.parser.pusherrtok(p.token, "invalid_data_star")
	} else {
		v.ast.Type.Value = v.ast.Type.Value[1:]
	}
	return v
}

func (p *operatorProcessor) amper() value {
	v := p.parser.evalExprPart(p.tokens, p.model)
	v.lvalue = true
	if !canGetPointer(v) {
		p.parser.pusherrtok(p.token, "invalid_data_amper")
	}
	v.ast.Type.Value = "*" + v.ast.Type.Value
	return v
}

func (p *Parser) evalOperatorExprPart(tokens []lex.Token, m *exprModel) value {
	var v value
	//? Length is 1 cause all length of operator tokens is 1.
	//? Change "1" with length of token's value
	//? if all operators length is not 1.
	exprTokens := tokens[1:]
	processor := operatorProcessor{tokens[0], exprTokens, m, p}
	m.appendNodeToSubNodes(exprNode{processor.token.Kind})
	if processor.tokens == nil {
		p.pusherrtok(processor.token, "invalid_syntax")
		return v
	}
	switch processor.token.Kind {
	case "-":
		v = processor.unary()
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
		p.pusherrtok(processor.token, "invalid_syntax")
	}
	v.ast.Token = processor.token
	return v
}

func canGetPointer(v value) bool {
	if v.ast.Type.Code == x.Func {
		return false
	}
	return v.ast.Token.Id == lex.Id
}

func (p *Parser) evalHeapAllocExpr(tokens []lex.Token, m *exprModel) (v value) {
	if len(tokens) == 1 {
		p.pusherrtok(tokens[0], "invalid_syntax_keyword_new")
		return
	}
	v.lvalue = true
	v.ast.Token = tokens[0]
	tokens = tokens[1:]
	astb := new(ast.Builder)
	index := new(int)
	dt, ok := astb.DataType(tokens, index, true)
	m.appendNodeToSubNodes(newHeapAllocExpr{dt})
	dt.Value = "*" + dt.Value
	v.ast.Type = dt
	if !ok {
		p.pusherrtok(tokens[0], "fail_build_heap_allocation_type")
		return
	}
	if *index < len(tokens)-1 {
		p.pusherrtok(tokens[*index+1], "invalid_syntax")
	}
	return
}

func (p *Parser) evalExprPart(tokens []lex.Token, m *exprModel) (v value) {
	if len(tokens) == 1 {
		val, ok := p.evalSingleExpr(tokens[0], m)
		if ok {
			v = val
			return
		}
	}
	token := tokens[0]
	switch token.Id {
	case lex.Operator:
		return p.evalOperatorExprPart(tokens, m)
	case lex.New:
		return p.evalHeapAllocExpr(tokens, m)
	case lex.Brace:
		switch token.Kind {
		case "(":
			val, ok := p.evalTryCastExpr(tokens, m)
			if ok {
				v = val
				return
			}
			val, ok = p.evalTryAssignExpr(tokens, m)
			if ok {
				v = val
				return
			}
		}
	}
	token = tokens[len(tokens)-1]
	switch token.Id {
	case lex.Operator:
		return p.evalOperatorExprPartRight(tokens, m)
	case lex.Brace:
		switch token.Kind {
		case ")":
			return p.evalParenthesesRangeExpr(tokens, m)
		case "}":
			return p.evalBraceRangeExpr(tokens, m)
		case "]":
			return p.evalBracketRangeExpr(tokens, m)
		}
	default:
		p.pusherrtok(tokens[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalTryCastExpr(tokens []lex.Token, m *exprModel) (v value, _ bool) {
	braceCount := 0
	errToken := tokens[0]
	for index, token := range tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
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
		typeTokens := tokens[1:index]
		dt, ok := astb.DataType(typeTokens, &dtindex, false)
		if !ok {
			return
		}
		dt, ok = p.readyType(dt, false)
		if !ok {
			return
		}
		if dtindex+1 < len(typeTokens) {
			return
		}
		if index+1 >= len(tokens) {
			p.pusherrtok(token, "casting_missing_expr")
			return
		}
		exprTokens := tokens[index+1:]
		m.appendNodeToSubNodes(exprNode{"(" + dt.String() + ")"})
		val := p.evalExprPart(exprTokens, m)
		val = p.evalCast(val, dt, errToken)
		return val, true
	}
	return
}

func (p *Parser) evalTryAssignExpr(tokens []lex.Token, m *exprModel) (v value, ok bool) {
	astb := ast.NewBuilder(nil)
	tokens = tokens[1 : len(tokens)-1] // Remove first-last parentheses
	assign, ok := astb.AssignExpr(tokens, true)
	if !ok {
		return
	}
	ok = true
	if len(astb.Errors) > 0 {
		p.pusherrs(astb.Errors...)
		return
	}
	p.checkAssign(&assign)
	m.appendNodeToSubNodes(assignExpr{assign})
	v, _ = p.evalExpr(assign.SelectExprs[0].Expr)
	return
}

func (p *Parser) evalCast(v value, t ast.DataType, errtok lex.Token) value {
	switch {
	case typeIsPtr(t):
		p.checkCastPtr(v.ast.Type, errtok)
	case typeIsArray(t):
		p.checkCastArray(t, v.ast.Type, errtok)
	case typeIsSingle(t):
		p.checkCastSingle(v.ast.Type, t.Code, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting")
	}
	v.ast.Type = t
	v.constant = false
	v.volatile = false
	return v
}

func (p *Parser) checkCastSingle(vt ast.DataType, t uint8, errtok lex.Token) {
	switch t {
	case x.Str:
		p.checkCastStr(vt, errtok)
		return
	}
	switch {
	case x.IsIntegerType(t):
		p.checkCastInteger(vt, errtok)
	case x.IsNumericType(t):
		p.checkCastNumeric(vt, errtok)
	default:
		p.pusherrtok(errtok, "type_notsupports_casting")
	}
}

func (p *Parser) checkCastStr(vt ast.DataType, errtok lex.Token) {
	if !typeIsArray(vt) {
		p.pusherrtok(errtok, "type_notsupports_casting")
		return
	}
	vt.Value = vt.Value[2:] // Remove array brackets
	if !typeIsSingle(vt) || (vt.Code != x.Rune && vt.Code != x.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting")
	}
}

func (p *Parser) checkCastInteger(vt ast.DataType, errtok lex.Token) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsSingle(vt) && x.IsNumericType(vt.Code) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting")
}

func (p *Parser) checkCastNumeric(vt ast.DataType, errtok lex.Token) {
	if typeIsSingle(vt) && x.IsNumericType(vt.Code) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting")
}

func (p *Parser) checkCastPtr(vt ast.DataType, errtok lex.Token) {
	if typeIsPtr(vt) {
		return
	}
	if typeIsSingle(vt) && x.IsIntegerType(vt.Code) {
		return
	}
	p.pusherrtok(errtok, "type_notsupports_casting")
}

func (p *Parser) checkCastArray(t, vt ast.DataType, errtok lex.Token) {
	if !typeIsSingle(vt) || vt.Code != x.Str {
		p.pusherrtok(errtok, "type_notsupports_casting")
		return
	}
	t.Value = t.Value[2:] // Remove array brackets
	if !typeIsSingle(t) || (t.Code != x.Rune && t.Code != x.U8) {
		p.pusherrtok(errtok, "type_notsupports_casting")
	}
}

func (p *Parser) evalOperatorExprPartRight(tokens []lex.Token, m *exprModel) (v value) {
	token := tokens[len(tokens)-1]
	switch token.Kind {
	case "...":
		tokens = tokens[:len(tokens)-1]
		return p.evalVariadicExprPart(tokens, m, token)
	default:
		p.pusherrtok(token, "invalid_syntax")
	}
	return
}

func (p *Parser) evalVariadicExprPart(tokens []lex.Token, m *exprModel, errtok lex.Token) (v value) {
	v = p.evalExprPart(tokens, m)
	if !typeIsVariadicable(v.ast.Type) {
		p.pusherrtok(errtok, "variadic_with_nonvariadicable")
		return
	}
	v.ast.Type.Value = v.ast.Type.Value[2:] // Remove array type.
	v.variadic = true
	return
}

func (p *Parser) evalParenthesesRangeExpr(tokens []lex.Token, m *exprModel) (v value) {
	var valueTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case ")", "}", "]":
			braceCount++
		case "(", "{", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		valueTokens = tokens[:j]
		break
	}
	if len(valueTokens) == 0 && braceCount == 0 {
		// Write parentheses.
		m.appendNodeToSubNodes(exprNode{"("})
		defer m.appendNodeToSubNodes(exprNode{")"})

		tk := tokens[0]
		tokens = tokens[1 : len(tokens)-1]
		if len(tokens) == 0 {
			p.pusherrtok(tk, "invalid_syntax")
		}
		value, model := p.evalTokens(tokens)
		v = value
		m.appendNodeToSubNodes(model)
		return
	}
	v = p.evalExprPart(valueTokens, m)

	// Write parentheses.
	m.appendNodeToSubNodes(exprNode{"("})
	defer m.appendNodeToSubNodes(exprNode{")"})

	switch v.ast.Type.Code {
	case x.Func:
		fun := v.ast.Type.Tag.(ast.Func)
		p.parseFuncCall(fun, tokens[len(valueTokens):], m)
		v.ast.Type = fun.RetType
		v.lvalue = typeIsLvalue(v.ast.Type)
	default:
		p.pusherrtok(tokens[len(valueTokens)], "invalid_syntax")
	}
	return
}

func (p *Parser) evalBraceRangeExpr(tokens []lex.Token, m *exprModel) (v value) {
	var exprTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprTokens = tokens[:j]
		break
	}
	valTokensLen := len(exprTokens)
	if valTokensLen == 0 || braceCount > 0 {
		p.pusherrtok(tokens[0], "invalid_syntax")
		return
	}
	switch exprTokens[0].Id {
	case lex.Brace:
		switch exprTokens[0].Kind {
		case "[":
			ast := ast.NewBuilder(nil)
			dt, ok := ast.DataType(exprTokens, new(int), true)
			if !ok {
				p.pusherrs(ast.Errors...)
				return
			}
			exprTokens = tokens[len(exprTokens):]
			var model iExpr
			v, model = p.buildArray(p.buildEnumerableParts(exprTokens),
				dt, exprTokens[0])
			m.appendNodeToSubNodes(model)
			return
		case "(":
			astBuilder := ast.NewBuilder(tokens)
			funAST := astBuilder.Func(astBuilder.Tokens, true)
			if len(astBuilder.Errors) > 0 {
				p.pusherrs(astBuilder.Errors...)
				return
			}
			p.checkAnonFunc(&funAST)
			v.ast.Type.Tag = funAST
			v.ast.Type.Code = x.Func
			v.ast.Type.Value = funAST.DataTypeString()
			m.appendNodeToSubNodes(anonFunc{funAST})
			return
		default:
			p.pusherrtok(exprTokens[0], "invalid_syntax")
		}
	default:
		p.pusherrtok(exprTokens[0], "invalid_syntax")
	}
	return
}

func (p *Parser) evalBracketRangeExpr(tokens []lex.Token, m *exprModel) (v value) {
	var exprTokens []lex.Token
	j := len(tokens) - 1
	braceCount := 0
	for ; j >= 0; j-- {
		token := tokens[j]
		if token.Id != lex.Brace {
			continue
		}
		switch token.Kind {
		case "}", "]", ")":
			braceCount++
		case "{", "(", "[":
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		exprTokens = tokens[:j]
		break
	}
	valTokensLen := len(exprTokens)
	if valTokensLen == 0 || braceCount > 0 {
		p.pusherrtok(tokens[0], "invalid_syntax")
		return
	}
	var model iExpr
	v, model = p.evalTokens(exprTokens)
	m.appendNodeToSubNodes(model)
	tokens = tokens[len(exprTokens)+1 : len(tokens)-1] // Removed array syntax "["..."]"
	m.appendNodeToSubNodes(exprNode{"["})
	selectv, model := p.evalTokens(tokens)
	m.appendNodeToSubNodes(model)
	m.appendNodeToSubNodes(exprNode{"]"})
	return p.evalEnumerableSelect(v, selectv, tokens[0])
}

func (p *Parser) evalEnumerableSelect(enumv, selectv value, errtok lex.Token) (v value) {
	switch {
	case typeIsArray(enumv.ast.Type):
		return p.evalArraySelect(enumv, selectv, errtok)
	case typeIsSingle(enumv.ast.Type):
		return p.evalStrSelect(enumv, selectv, errtok)
	}
	p.pusherrtok(errtok, "not_enumerable")
	return
}

func (p *Parser) evalArraySelect(arrv, selectv value, errtok lex.Token) value {
	arrv.lvalue = true
	arrv.ast.Type = typeOfArrayElements(arrv.ast.Type)
	if !typeIsSingle(selectv.ast.Type) ||
		!x.IsIntegerType(selectv.ast.Type.Code) {
		p.pusherrtok(errtok, "notint_array_select")
	}
	return arrv
}

func (p *Parser) evalStrSelect(strv, selectv value, errtok lex.Token) value {
	strv.lvalue = true
	strv.ast.Type.Code = x.Rune
	if !typeIsSingle(selectv.ast.Type) ||
		!x.IsIntegerType(selectv.ast.Type.Code) {
		p.pusherrtok(errtok, "notint_string_select")
	}
	return strv
}

//! IMPORTANT: Tokens is should be store enumerable parentheses.
func (p *Parser) buildEnumerableParts(tokens []lex.Token) [][]lex.Token {
	tokens = tokens[1 : len(tokens)-1]
	braceCount := 0
	lastComma := -1
	var parts [][]lex.Token
	for index, token := range tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if token.Id == lex.Comma {
			if index-lastComma-1 == 0 {
				p.pusherrtok(token, "missing_expression")
				lastComma = index
				continue
			}
			parts = append(parts, tokens[lastComma+1:index])
			lastComma = index
		}
	}
	if lastComma+1 < len(tokens) {
		parts = append(parts, tokens[lastComma+1:])
	}
	return parts
}

func (p *Parser) buildArray(parts [][]lex.Token, t ast.DataType, errtok lex.Token) (value, iExpr) {
	var v value
	v.ast.Type = t
	model := arrayExpr{dataType: t}
	elementType := typeOfArrayElements(t)
	for _, part := range parts {
		partValue, expModel := p.evalTokens(part)
		model.expr = append(model.expr, expModel)
		p.wg.Add(1)
		go assignChecker{
			p,
			false,
			elementType,
			partValue,
			false,
			part[0],
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

func (p *Parser) parseFuncCall(f ast.Func, tokens []lex.Token, m *exprModel) {
	errToken := tokens[0]
	tokens, _ = p.getRange("(", ")", tokens)
	if tokens == nil {
		tokens = make([]lex.Token, 0)
	}
	ast := new(ast.Builder)
	args := ast.Args(tokens)
	if len(ast.Errors) > 0 {
		p.pusherrs(ast.Errors...)
	}
	p.parseArgs(f.Params, &args, errToken, m)
	if m != nil {
		m.appendNodeToSubNodes(argsExpr{args})
	}
}

func (p *Parser) parseArgs(params []ast.Parameter, args *[]ast.Arg, errTok lex.Token, m *exprModel) {
	parsedArgs := make([]ast.Arg, 0)
	if len(params) > 0 && params[len(params)-1].Variadic {
		if len(*args) == 0 && len(params) == 1 {
			return
		} else if len(*args) < len(params)-1 {
			p.pusherrtok(errTok, "missing_argument")
			goto argParse
		} else if len(*args) <= len(params)-1 {
			goto argParse
		}
		variadicArgs := (*args)[len(params)-1:]
		variadicParam := params[len(params)-1]
		*args = (*args)[:len(params)-1]
		params = params[:len(params)-1]
		defer func() {
			model := arrayExpr{variadicParam.Type, nil}
			model.dataType.Value = "[]" + model.dataType.Value // For array.
			variadiced := false
			for _, arg := range variadicArgs {
				p.parseArg(variadicParam, &arg, &variadiced)
				model.expr = append(model.expr, arg.Expr.Model.(iExpr))
			}
			if variadiced && len(variadicArgs) > 1 {
				p.pusherrtok(errTok, "more_args_with_varidiced")
			}
			arg := ast.Arg{Expr: ast.Expr{Model: model}}
			parsedArgs = append(parsedArgs, arg)
			*args = parsedArgs
		}()
	}
	if len(*args) == 0 && len(params) == 0 {
		return
	} else if len(*args) < len(params) {
		p.pusherrtok(errTok, "missing_argument")
	} else if len(*args) > len(params) {
		p.pusherrtok(errTok, "argument_overflow")
		return
	}
argParse:
	for index, arg := range *args {
		p.parseArg(params[index], &arg, nil)
		parsedArgs = append(parsedArgs, arg)
	}
	*args = parsedArgs
}

func (p *Parser) parseArg(param ast.Parameter, arg *ast.Arg, variadiced *bool) {
	value, model := p.evalExpr(arg.Expr)
	arg.Expr.Model = model
	if variadiced != nil && !*variadiced {
		*variadiced = value.variadic
	}
	p.wg.Add(1)
	go p.checkArgTypeAsync(param, value, false, arg.Token)
}

func (p *Parser) checkArgTypeAsync(param ast.Parameter, val value, ignoreAny bool, errTok lex.Token) {
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
func (p *Parser) getRange(open, close string, tokens []lex.Token) (_ []lex.Token, ok bool) {
	braceCount := 0
	start := 1
	if tokens[0].Id != lex.Brace {
		return nil, false
	}
	for index, token := range tokens {
		if token.Id != lex.Brace {
			continue
		}
		if token.Kind == open {
			braceCount++
		} else if token.Kind == close {
			braceCount--
		}
		if braceCount > 0 {
			continue
		}
		return tokens[start:index], true
	}
	return nil, false
}

func (p *Parser) checkEntryPointSpecialCases(fun *function) {
	if len(fun.Ast.Params) > 0 {
		p.pusherrtok(fun.Ast.Token, "entrypoint_have_parameters")
	}
	if fun.Ast.RetType.Code != x.Void {
		p.pusherrtok(fun.Ast.RetType.Token, "entrypoint_have_return")
	}
	if fun.Attributes != nil {
		p.pusherrtok(fun.Ast.Token, "entrypoint_have_attributes")
	}
}

func (p *Parser) checkBlock(b *ast.BlockAST) {
	for i := 0; i < len(b.Tree); i++ {
		model := &b.Tree[i]
		switch t := model.Value.(type) {
		case ast.ExprStatement:
			_, t.Expr.Model = p.evalExpr(t.Expr)
			model.Value = t
		case ast.Var:
			p.checkVarStatement(&t, false)
			model.Value = t
		case ast.Assign:
			p.checkAssign(&t)
			model.Value = t
		case ast.Free:
			p.checkFreeStatement(&t)
			model.Value = t
		case ast.Iter:
			p.checkIterExpr(&t)
			model.Value = t
		case ast.Break:
			p.checkBreakStatement(&t)
		case ast.Continue:
			p.checkContinueStatement(&t)
		case ast.If:
			p.checkIfExpr(&t, &i, b.Tree)
			model.Value = t
		case ast.CxxEmbed:
		case ast.Comment:
		case ast.Ret:
		default:
			p.pusherrtok(model.Token, "invalid_syntax")
		}
	}
}

type retChecker struct {
	p        *Parser
	retAST   *ast.Ret
	fun      *ast.Func
	expModel multiRetExpr
	values   []value
}

func (rc *retChecker) pushval(last, current int, errTk lex.Token) {
	if current-last == 0 {
		rc.p.pusherrtok(errTk, "missing_value")
		return
	}
	tokens := rc.retAST.Expr.Tokens[last:current]
	value, model := rc.p.evalTokens(tokens)
	rc.expModel.models = append(rc.expModel.models, model)
	rc.values = append(rc.values, value)
}

func (rc *retChecker) checkepxrs() {
	braceCount := 0
	last := 0
	for index, token := range rc.retAST.Expr.Tokens {
		if token.Id == lex.Brace {
			switch token.Kind {
			case "(", "{", "[":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 || token.Id != lex.Comma {
			continue
		}
		rc.pushval(last, index, token)
		last = index + 1
	}
	length := len(rc.retAST.Expr.Tokens)
	if last < length {
		if last == 0 {
			rc.pushval(0, length, rc.retAST.Token)
		} else {
			rc.pushval(last, length, rc.retAST.Expr.Tokens[last-1])
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
			rc.p.pusherrtok(rc.retAST.Token, "overflow_return")
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         rc.fun.RetType,
			v:         rc.values[0],
			ignoreAny: false,
			errtok:    rc.retAST.Token,
		}.checkAssignTypeAsync()
		return
	}
	// Multi return
	rc.retAST.Expr.Model = rc.expModel
	types := rc.fun.RetType.Tag.([]ast.DataType)
	if valLength == 1 {
		rc.p.pusherrtok(rc.retAST.Token, "missing_multi_return")
	} else if valLength > len(types) {
		rc.p.pusherrtok(rc.retAST.Token, "overflow_return")
	}
	for index, t := range types {
		if index >= valLength {
			break
		}
		rc.p.wg.Add(1)
		go assignChecker{
			p:         rc.p,
			constant:  false,
			t:         t,
			v:         rc.values[index],
			ignoreAny: false,
			errtok:    rc.retAST.Token,
		}.checkAssignTypeAsync()
	}
}

func (rc *retChecker) check() {
	exprTokensLen := len(rc.retAST.Expr.Tokens)
	if exprTokensLen == 0 && !typeIsVoidRet(rc.fun.RetType) {
		rc.p.pusherrtok(rc.retAST.Token, "require_return_value")
		return
	}
	if exprTokensLen > 0 && typeIsVoidRet(rc.fun.RetType) {
		rc.p.pusherrtok(rc.retAST.Token, "void_function_return_value")
	}
	rc.checkepxrs()
}

func (p *Parser) checkRets(fun *ast.Func) {
	missed := true
	for index, s := range fun.Block.Tree {
		switch t := s.Value.(type) {
		case ast.Ret:
			rc := retChecker{p: p, retAST: &t, fun: fun}
			rc.check()
			fun.Block.Tree[index].Value = t
			missed = false
		}
	}
	if missed && !typeIsVoidRet(fun.RetType) {
		p.pusherrtok(fun.Token, "missing_return")
	}
}

func (p *Parser) checkFunc(f *ast.Func) {
	p.checkBlock(&f.Block)
	p.checkRets(f)
}

func (p *Parser) checkVarStatement(varAST *ast.Var, noParse bool) {
	if p.existIdf(varAST.Id, true).Id != lex.NA {
		p.pusherrtok(varAST.IdToken, "exist_id")
	}
	if !noParse {
		*varAST = p.Var(*varAST)
	}
	p.BlockVars = append(p.BlockVars, *varAST)
}

func (p *Parser) checkAssignment(selected value, errtok lex.Token) bool {
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
		if p.FuncById(selected.ast.Token.Kind) != nil {
			p.pusherrtok(errtok, "assign_type_not_support_value")
			state = false
		}
	}
	return state
}

func (p *Parser) checkSingleAssign(assign *ast.Assign) {
	sexpr := &assign.SelectExprs[0].Expr
	if len(sexpr.Tokens) == 1 && xapi.IsIgnoreId(sexpr.Tokens[0].Kind) {
		return
	}
	selected, _ := p.evalExpr(*sexpr)
	if !p.checkAssignment(selected, assign.Setter) {
		return
	}
	vexpr := &assign.ValueExprs[0]
	val, model := p.evalExpr(*vexpr)
	*vexpr = model.(*exprModel).Expr()
	if assign.Setter.Kind != "=" {
		assign.Setter.Kind = assign.Setter.Kind[:len(assign.Setter.Kind)-1]
		solver := solver{
			p:        p,
			left:     sexpr.Tokens,
			leftVal:  selected.ast,
			right:    vexpr.Tokens,
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

func (p *Parser) parseAssignSelections(vsAST *ast.Assign) {
	for index, selector := range vsAST.SelectExprs {
		p.checkVarStatement(&selector.Var, false)
		vsAST.SelectExprs[index] = selector
	}
}

func (p *Parser) assignExprs(vsAST *ast.Assign) []value {
	values := make([]value, len(vsAST.ValueExprs))
	for index, expr := range vsAST.ValueExprs {
		val, model := p.evalExpr(expr)
		vsAST.ValueExprs[index].Model = model
		values[index] = val
	}
	return values
}

func (p *Parser) processFuncMultiAssign(vsAST *ast.Assign, funcVal value) {
	types := funcVal.ast.Type.Tag.([]ast.DataType)
	if len(types) != len(vsAST.SelectExprs) {
		p.pusherrtok(vsAST.Setter, "missing_multiassign_identifiers")
		return
	}
	values := make([]value, len(types))
	for index, t := range types {
		values[index] = value{
			ast: ast.Value{
				Token: t.Token,
				Type:  t,
			},
		}
	}
	p.processMultiAssign(vsAST, values)
}

func (p *Parser) processMultiAssign(assign *ast.Assign, vals []value) {
	for index := range assign.SelectExprs {
		selector := &assign.SelectExprs[index]
		selector.Ignore = xapi.IsIgnoreId(selector.Var.Id)
		val := vals[index]
		if !selector.NewVariable {
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
	if assign.JustDeclare {
		p.parseAssignSelections(assign)
		return
	} else if selectLength == 1 && !assign.SelectExprs[0].NewVariable {
		p.checkSingleAssign(assign)
		return
	} else if assign.Setter.Kind != "=" {
		p.pusherrtok(assign.Setter, "invalid_syntax")
		return
	}
	if valueLength == 1 {
		firstVal, _ := p.evalExpr(assign.ValueExprs[0])
		if firstVal.ast.Type.MultiTyped {
			assign.MultipleReturn = true
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
		p.pusherrtok(freeAST.Token, "free_nonpointer")
	}
}

func (p *Parser) checkWhileProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.WhileProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	iter.Profile = profile
	if !isBoolExpr(val) {
		p.pusherrtok(iter.Token, "iter_while_notbool_expr")
	}
	p.checkBlock(&iter.Block)
}

type foreachTypeChecker struct {
	p       *Parser
	profile *ast.ForeachProfile
	value   value
}

func (frc *foreachTypeChecker) array() {
	if !xapi.IsIgnoreId(frc.profile.KeyA.Id) {
		keyA := &frc.profile.KeyA
		if keyA.Type.Code == x.Void {
			keyA.Type.Code = x.Size
			keyA.Type.Value = x.CxxTypeIdFromType(keyA.Type.Code)
		} else {
			var ok bool
			keyA.Type, ok = frc.p.readyType(keyA.Type, true)
			if ok {
				if !typeIsSingle(keyA.Type) || !x.IsNumericType(keyA.Type.Code) {
					frc.p.pusherrtok(keyA.IdToken, "incompatible_datatype")
				}
			}
		}
	}
	if !xapi.IsIgnoreId(frc.profile.KeyB.Id) {
		elementType := frc.profile.ExprType
		elementType.Value = elementType.Value[2:]
		keyB := &frc.profile.KeyB
		if keyB.Type.Code == x.Void {
			keyB.Type = elementType
		} else {
			frc.p.wg.Add(1)
			go frc.p.checkTypeAsync(elementType, frc.profile.KeyB.Type, true, frc.profile.InToken)
		}
	}
}

func (frc *foreachTypeChecker) str() {
	if !xapi.IsIgnoreId(frc.profile.KeyA.Id) {
		keyA := &frc.profile.KeyA
		if keyA.Type.Code == x.Void {
			keyA.Type.Code = x.Size
			keyA.Type.Value = x.CxxTypeIdFromType(keyA.Type.Code)
		} else {
			var ok bool
			keyA.Type, ok = frc.p.readyType(keyA.Type, true)
			if ok {
				if !typeIsSingle(keyA.Type) || !x.IsNumericType(keyA.Type.Code) {
					frc.p.pusherrtok(keyA.IdToken, "incompatible_datatype")
				}
			}
		}
	}
	if !xapi.IsIgnoreId(frc.profile.KeyB.Id) {
		runeType := ast.DataType{
			Code:  x.Rune,
			Value: x.CxxTypeIdFromType(x.Rune),
		}
		keyB := &frc.profile.KeyB
		if keyB.Type.Code == x.Void {
			keyB.Type = runeType
		} else {
			frc.p.wg.Add(1)
			go frc.p.checkTypeAsync(runeType, frc.profile.KeyB.Type, true, frc.profile.InToken)
		}
	}
}

func (ftc *foreachTypeChecker) check() {
	switch {
	case typeIsArray(ftc.value.ast.Type):
		ftc.array()
	case ftc.value.ast.Type.Code == x.Str:
		ftc.str()
	}
}

func (p *Parser) checkForeachProfile(iter *ast.Iter) {
	profile := iter.Profile.(ast.ForeachProfile)
	val, model := p.evalExpr(profile.Expr)
	profile.Expr.Model = model
	profile.ExprType = val.ast.Type
	if !isForeachIterExpr(val) {
		p.pusherrtok(iter.Token, "iter_foreach_nonenumerable_expr")
	} else {
		checker := foreachTypeChecker{p, &profile, val}
		checker.check()
	}
	iter.Profile = profile
	blockVariables := p.BlockVars
	if profile.KeyA.New {
		if xapi.IsIgnoreId(profile.KeyA.Id) {
			p.pusherrtok(profile.KeyA.IdToken, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyA, true)
	}
	if profile.KeyB.New {
		if xapi.IsIgnoreId(profile.KeyB.Id) {
			p.pusherrtok(profile.KeyB.IdToken, "ignore_id")
		}
		p.checkVarStatement(&profile.KeyB, true)
	}
	p.checkBlock(&iter.Block)
	p.BlockVars = blockVariables
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

func (p *Parser) checkIfExpr(ifast *ast.If, index *int, statements []ast.Statement) {
	val, model := p.evalExpr(ifast.Expr)
	ifast.Expr.Model = model
	statement := statements[*index]
	if !isBoolExpr(val) {
		p.pusherrtok(ifast.Token, "if_notbool_expr")
	}
	p.checkBlock(&ifast.Block)
node:
	if statement.WithTerminator {
		return
	}
	*index++
	if *index >= len(statements) {
		*index--
		return
	}
	statement = statements[*index]
	switch t := statement.Value.(type) {
	case ast.ElseIf:
		val, model := p.evalExpr(t.Expr)
		t.Expr.Model = model
		if !isBoolExpr(val) {
			p.pusherrtok(t.Token, "if_notbool_expr")
		}
		p.checkBlock(&t.Block)
		goto node
	case ast.Else:
		p.checkElseBlock(&t)
		statement.Value = t
	default:
		*index--
	}
}

func (p *Parser) checkElseBlock(elseast *ast.Else) {
	p.checkBlock(&elseast.Block)
}

func (p *Parser) checkBreakStatement(breakAST *ast.Break) {
	if p.iterCount == 0 {
		p.pusherrtok(breakAST.Token, "break_at_outiter")
	}
}

func (p *Parser) checkContinueStatement(continueAST *ast.Continue) {
	if p.iterCount == 0 {
		p.pusherrtok(continueAST.Token, "continue_at_outiter")
	}
}

func (p *Parser) checkValidityForAutoType(t ast.DataType, err lex.Token) {
	switch t.Code {
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
	return x.DefaultValueOfType(t.Code)
}

func (p *Parser) readyType(dt ast.DataType, err bool) (_ ast.DataType, ok bool) {
	if dt.Value == "" {
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
	switch dt.Code {
	case x.Id:
		t := p.typeById(dt.Token.Kind)
		if t == nil {
			if err {
				p.pusherrtok(dt.Token, "invalid_type_source")
			}
			return dt, false
		}
		t.Type.Value = dt.Value[:len(dt.Value)-len(dt.Token.Kind)] + t.Type.Value
		return p.readyType(t.Type, err)
	case x.Func:
		f := dt.Tag.(ast.Func)
		for i, param := range f.Params {
			f.Params[i].Type, _ = p.readyType(param.Type, err)
		}
		f.RetType, _ = p.readyType(f.RetType, err)
		dt.Value = dt.Tag.(ast.Func).DataTypeString()
	}
	return dt, true
}

func (p *Parser) checkMultiTypeAsync(real, check ast.DataType, ignoreAny bool, errToken lex.Token) {
	defer func() { p.wg.Done() }()
	if real.MultiTyped != check.MultiTyped {
		p.pusherrtok(errToken, "incompatible_datatype")
		return
	}
	realTypes := real.Tag.([]ast.DataType)
	checkTypes := real.Tag.([]ast.DataType)
	if len(realTypes) != len(checkTypes) {
		p.pusherrtok(errToken, "incompatible_datatype")
		return
	}
	for index := 0; index < len(realTypes); index++ {
		realType := realTypes[index]
		checkType := checkTypes[index]
		p.checkTypeAsync(realType, checkType, ignoreAny, errToken)
	}
}

func (p *Parser) checkAssignConst(constant bool, t ast.DataType, val value, errToken lex.Token) {
	if typeIsMut(t) && val.constant && !constant {
		p.pusherrtok(errToken, "constant_assignto_nonconstant")
	}
}

type assignChecker struct {
	p         *Parser
	constant  bool
	t         ast.DataType
	v         value
	ignoreAny bool
	errtok    lex.Token
}

func (ac assignChecker) checkAssignTypeAsync() {
	defer func() { ac.p.wg.Done() }()
	ac.p.checkAssignConst(ac.constant, ac.t, ac.v, ac.errtok)
	if typeIsSingle(ac.t) && isConstNum(ac.v.ast.Data) {
		switch {
		case x.IsSignedIntegerType(ac.t.Code):
			if xbits.CheckBitInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Code)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype")
			return
		case x.IsFloatType(ac.t.Code):
			if checkFloatBit(ac.v.ast, xbits.BitsizeType(ac.t.Code)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype")
			return
		case x.IsUnsignedNumericType(ac.t.Code):
			if xbits.CheckBitUInt(ac.v.ast.Data, xbits.BitsizeType(ac.t.Code)) {
				return
			}
			ac.p.pusherrtok(ac.errtok, "incompatible_datatype")
			return
		}
	}
	ac.p.wg.Add(1)
	go ac.p.checkTypeAsync(ac.t, ac.v.ast.Type, ac.ignoreAny, ac.errtok)
}

func (p *Parser) checkTypeAsync(real, check ast.DataType, ignoreAny bool, errToken lex.Token) {
	defer func() { p.wg.Done() }()
	if !ignoreAny && real.Code == x.Any {
		return
	}
	if real.MultiTyped || check.MultiTyped {
		p.wg.Add(1)
		go p.checkMultiTypeAsync(real, check, ignoreAny, errToken)
		return
	}
	if typeIsSingle(real) && typeIsSingle(check) {
		if !typesAreCompatible(real, check, ignoreAny) {
			p.pusherrtok(errToken, "incompatible_datatype")
		}
		return
	}
	if (typeIsPtr(real) || typeIsArray(real)) && check.Code == x.Nil {
		return
	}
	if real.Value != check.Value {
		p.pusherrtok(errToken, "incompatible_datatype")
	}
}
