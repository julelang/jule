// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/julelang/jule/documenter"
	"github.com/julelang/jule/lex"
	"github.com/julelang/jule/parser"
	"github.com/julelang/jule/pkg/jule"
	"github.com/julelang/jule/pkg/juleapi"
	"github.com/julelang/jule/pkg/juleio"
)

const mode_transpile = "transpile"
const mode_compile = "compile"

const compiler_gcc = "gcc"
const compiler_clang = "clang"

const compiler_path_gcc = "g++"
const compiler_path_clang = "clang++"

var out_dir = "./dist"
var language = "default"
var mode = mode_compile
var out_name = "ir.cpp"

// Sets by compiler or command-line inputs
var compiler = ""
var compiler_path = ""

const CMD_HELP = "help"
const CMD_VERSION = "version"
const CMD_DOC = "doc"
const CMD_BUG = "bug"
const CMD_TOOL = "tool"

var HELP_MAP = [...][2]string{
	{CMD_HELP, "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_DOC, "Documentize Jule source code"},
	{CMD_BUG, "Start a new bug report"},
	{CMD_TOOL, "Tools for effective Jule"},
}

func help(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	max := len(HELP_MAP[0][0])
	for _, k := range HELP_MAP {
		n := len(k[0])
		if n > max {
			max = n
		}
	}
	var sb strings.Builder
	const SPACE = 5 // Space of between command name and description.
	for _, part := range HELP_MAP {
		sb.WriteString(part[0])
		sb.WriteString(strings.Repeat(" ", (max-len(part[0]))+SPACE))
		sb.WriteString(part[1])
		sb.WriteByte('\n')
	}
	println(sb.String()[:sb.Len()-1])
}

func version(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	println("julec version", jule.VERSION)
}

func doc(cmd string) {
	cmd = strings.TrimSpace(cmd)
	paths := strings.SplitN(cmd, " ", -1)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		p := compile(path, false, true, true)
		if p == nil {
			continue
		}
		if print_logs(p) {
			fmt.Println(jule.GetError("doc_couldnt_generated", path))
			continue
		}
		docjson, err := documenter.Doc(p)
		if err != nil {
			fmt.Println(jule.GetError("error", err.Error()))
			continue
		}
		// Remove SrcExt from path
		path = path[:len(path)-len(jule.SRC_EXT)]
		path = filepath.Join(out_dir, path+jule.DOC_EXT)
		write_output(path, docjson)
	}
}

func open_url(url string) error {
	var name string
	var args []string

	switch runtime.GOOS {
	case "windows":
		name = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		name = "open"
	default:
		name = "xdg-open"
	}
	args = append(args, url)
	cmd := exec.Command(name, args...)
	return cmd.Start()
}

func bug(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	err := open_url("https://github.com/jule-lang/jule/issues/new?assignees=&labels=bug&template=bug-report.md&title=bug%3A+parser+generates+wrong+variable+declaration")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func list_horizontal_slice(s []string) string {
	lst := fmt.Sprint(s)
	return lst[1 : len(lst)-1]
}

func tool(cmd string) {
	if cmd == "" {
		println(`tool commands:
 distos     Lists all supported operating systems
 distarch   Lists all supported architects`)
		return
	}
	switch cmd {
	case "distos":
		print("supported operating systems:\n ")
		println(list_horizontal_slice(jule.DISTOS))
	case "distarch":
		print("supported architects:\n ")
		println(list_horizontal_slice(jule.DISTARCH))
	default:
		println("Undefined command: " + cmd)
	}
}

func process_command(namespace, cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	switch namespace {
	case CMD_HELP:
		help(cmd)
	case CMD_VERSION:
		version(cmd)
	case CMD_DOC:
		doc(cmd)
	case CMD_BUG:
		bug(cmd)
	case CMD_TOOL:
		tool(cmd)
	default:
		return false
	}
	return true
}

func init() {
	execp, err := os.Executable()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	jule.WORKING_PATH, err = os.Getwd()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	execp = filepath.Dir(execp)
	jule.EXEC_PATH = execp
	jule.STDLIB_PATH = filepath.Join(jule.EXEC_PATH, jule.STDLIB)
	juleapi.JULEC_HEADER = filepath.Join(jule.EXEC_PATH, "api")
	juleapi.JULEC_HEADER = filepath.Join(juleapi.JULEC_HEADER, "julec.hpp")
	jule.LOCALIZATION_PATH = filepath.Join(jule.EXEC_PATH, jule.LOCALIZATIONS)

	// Not started with arguments.
	// Here is "2" but "os.Args" always have one element for store working directory.
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	var sb strings.Builder
	for _, arg := range os.Args[1:] {
		sb.WriteString(" " + arg)
	}
	os.Args[0] = sb.String()[1:]
	arg := os.Args[0]
	i := strings.Index(arg, " ")
	if i == -1 {
		i = len(arg)
	}
	if process_command(arg[:i], arg[i:]) {
		os.Exit(0)
	}
}

func load_localization() {
	lang := strings.TrimSpace(language)
	if lang == "" || lang == "default" {
		return
	}
	path := filepath.Join(jule.LOCALIZATION_PATH, lang+".json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		println("Language couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &jule.ERRORS)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func check_mode() {
	if mode != mode_transpile && mode != mode_compile {
		println(jule.GetError("invalid_value_for_key", mode, "mode"))
		os.Exit(0)
	}
}

func check_compiler() {
	if compiler != jule.COMPILER_GCC && compiler != jule.COMPILER_CLANG {
		println(jule.GetError("invalid_value_for_key", compiler, "compiler"))
		os.Exit(0)
	}
}

func set() {
	if runtime.GOOS == "windows" {
		compiler = compiler_gcc
		compiler_path = compiler_path_gcc
	} else {
		compiler = compiler_clang
		compiler_path = compiler_path_clang
	}

	load_localization()
	check_mode()
	check_compiler()
}

// print_logs prints logs and returns true
// if logs has error, false if not.
func print_logs(p *parser.Parser) bool {
	var str strings.Builder
	for _, l := range p.Warnings {
		str.WriteString(l.String())
		str.WriteByte('\n')
	}
	for _, l := range p.Errors {
		str.WriteString(l.String())
		str.WriteByte('\n')
	}
	print(str.String())
	return len(p.Errors) > 0
}

func append_standard(code *string) {
	y, m, d := time.Now().Date()
	h, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)",
		d, m, y, h, min)
	var sb strings.Builder
	sb.WriteString("// Auto generated by JuleC.\n")
	sb.WriteString("// JuleC version: ")
	sb.WriteString(jule.VERSION)
	sb.WriteByte('\n')
	sb.WriteString("// Date: ")
	sb.WriteString(timeStr)
	sb.WriteString("\n\n#include \"")
	sb.WriteString(juleapi.JULEC_HEADER)
	sb.WriteString("\"\n\n")
	sb.WriteString(*code)
	*code = sb.String()
}

func write_output(path, content string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0o777)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	f, err := os.Create(path)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	_, err = f.WriteString(content)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
}

func compile(path string, main, nolocal, justDefs bool) *parser.Parser {
	set()
	p := parser.New(nil)
	// Check standard library.
	inf, err := os.Stat(jule.STDLIB_PATH)
	if err != nil || !inf.IsDir() {
		p.PushErr("stdlib_not_exist")
		return p
	}
	f, err := juleio.Jopen(path)
	if err != nil {
		println(err.Error())
		return nil
	}
	if !juleio.IsPassFileAnnotation(path) {
		p.PushErr("file_not_useable")
		return p
	}
	p.File = f
	p.NoLocalPkg = nolocal
	p.SetupPackage()
	p.Parsef(main, justDefs)
	return p
}

func generate_compile_command(source_path string) (c, cmd string) {
	var cpp strings.Builder
	cpp.WriteString("-g -O0 ")
	cpp.WriteString(source_path)
	return compiler_path, cpp.String()
}

func do_spell(cpp string) {
	path := filepath.Join(jule.WORKING_PATH, out_dir)
	path = filepath.Join(path, out_name)
	write_output(path, cpp)
	switch mode {
	case mode_compile:
		c, cmd := generate_compile_command(path)
		println(c + " " + cmd)
		entries := strings.SplitN(cmd, " ", -1)
		command := exec.Command(c, entries...)
		err := command.Start()
		if err != nil {
			println(err.Error())
		}
		err = command.Wait()
		if err != nil {
			println(err.Error())
		}
	}
}

func get_arg(i *int, runes []rune) (arg string, content string) {
	first := *i
	for ; *i < len(runes); *i++ {
		r := runes[*i]
		if r != '-' {
			continue
		}
		j := *i
		*i++
		if *i >= len(runes) {
			println("error: undefined syntax: " + string(runes[j:]))
			os.Exit(0)
		}
		r = runes[*i]
		if r == '-' {
			*i++
			if *i >= len(runes) {
				println("error: undefined syntax: " + string(runes[j:]))
				os.Exit(0)
			}
			r = runes[*i]
		}
		if !lex.IsIdentifierRune(string(r)) {
			println("error: undefined syntax: " + string(runes[j:]))
			os.Exit(0)
		}
		*i++
		for ; *i < len(runes); *i++ {
			r = runes[*i]
			if lex.IsSpace(byte(r)) {
				break
			} else if !lex.IsLetter(r) && !lex.IsDecimal(byte(r)) &&
				r != '_' && r != '-' {
				println("error: undefined syntax: " + string(runes[j:]))
				os.Exit(0)
			}
		}
		arg = string(runes[j:*i])
		content = string(runes[first:j])
		return
	}
	content = string(runes[first:])
	return
}

func get_arg_value(i *int, runes []rune) string {
	first := -1
	for ; *i < len(runes); *i++ {
		r := runes[*i]
		if lex.IsSpace(byte(r)) {
			if first != -1 {
				return string(runes[first:*i])
			}
			continue
		}
		if first == -1 {
			first = *i
		}
	}
	if first == -1 {
		return ""
	}
	return string(runes[first:])
}

func parse_compiler_arg(i *int, runes []rune) {
	value := get_arg_value(i, runes)
	if value == "" {
		println("error: missing argument value: -c --compiler")
		os.Exit(0)
	}
	switch value {
	case compiler_clang:
		compiler = value
		compiler_path = compiler_path_clang
	case compiler_gcc:
		compiler = value
		compiler_path = compiler_path_gcc
	default:
		println("error: invalid argument value: " + value)
		os.Exit(0)
	}
}

func parse_arguments(cmd string) string {
	runes := []rune(cmd)
	cmd = ""
	i := 0
	for ; i < len(runes); i++ {
		arg, content := get_arg(&i, runes)
		cmd += content
		switch arg {
		case "":
		case "-t", "--transpile":
			mode = mode_transpile
		case "-c", "--compile":
			mode = mode_compile
		case "--compiler":
			parse_compiler_arg(&i, runes)
		default:
			println("error: undefined argument: " + arg)
			os.Exit(0)
		}
	}
	cmd = strings.TrimSpace(cmd)
	return cmd
}

func main() {
	cmd := os.Args[0]
	cmd = parse_arguments(cmd)
	if cmd == "" {
		println("error: missing compile path")
		return
	}
	t := compile(cmd, true, false, false)
	if t == nil {
		return
	}
	if print_logs(t) {
		return
	}
	cpp := t.Cpp()
	append_standard(&cpp)
	do_spell(cpp)
}
