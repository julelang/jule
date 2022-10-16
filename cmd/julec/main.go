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

	"github.com/jule-lang/jule/documenter"
	"github.com/jule-lang/jule/parser"
	"github.com/jule-lang/jule/pkg/jule"
	"github.com/jule-lang/jule/pkg/juleapi"
	"github.com/jule-lang/jule/pkg/juleio"
	"github.com/jule-lang/jule/pkg/juleset"
)

const CMD_HELP = "help"
const CMD_VERSION = "version"
const CMD_INIT = "init"
const CMD_DOC = "doc"
const CMD_BUG = "bug"
const CMD_TOOL = "tool"

var HELP_MAP = [...][2]string{
	{CMD_HELP, "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_INIT, "Initialize new project here"},
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

func init_project(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	bytes, err := json.MarshalIndent(*juleset.DEFAULT, "", "\t")
	if err != nil {
		println(err)
		os.Exit(0)
	}
	err = os.WriteFile(jule.SETTINGS_FILE, bytes, 0666)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	println("Initialized project.")
}

func doc(cmd string) {
	fmt_json := false
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "--json") {
		cmd = strings.TrimSpace(cmd[len("--json"):])
		fmt_json = true
	}
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
		docjson, err := documenter.Doc(p, fmt_json)
		if err != nil {
			fmt.Println(jule.GetError("error", err.Error()))
			continue
		}
		// Remove SrcExt from path
		path = path[:len(path)-len(jule.SRC_EXT)]
		path = filepath.Join(jule.SET.CppOutDir, path+jule.DOC_EXT)
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
	case CMD_INIT:
		init_project(cmd)
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
	lang := strings.TrimSpace(jule.SET.Language)
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
	mode := jule.SET.Mode
	if mode != juleset.MDOE_TRANSPILE && mode != juleset.MODE_COMPILE {
		println(jule.GetError("invalid_value_for_key", mode, "mode"))
		os.Exit(0)
	}
}

func check_compiler() {
	c := jule.SET.Compiler
	if c != jule.COMPILER_GCC && c != jule.COMPILER_CLANG {
		println(jule.GetError("invalid_value_for_key", c, "compiler"))
		os.Exit(0)
	}
}

func load_juleset() {
	// File check.
	info, err := os.Stat(jule.SETTINGS_FILE)
	if err != nil || info.IsDir() {
		jule.SET = new(juleset.Set)
		*jule.SET = *juleset.DEFAULT
		return
	}
	bytes, err := os.ReadFile(jule.SETTINGS_FILE)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	jule.SET, err = juleset.Load(bytes)
	if err != nil {
		println("Jule settings has errors;")
		println(err.Error())
		os.Exit(0)
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
	load_juleset()
	p := parser.New(nil)
	// Check standard library.
	inf, err := os.Stat(jule.STDLIB_PATH)
	if err != nil || !inf.IsDir() {
		p.PushErr("stdlib_not_exist")
		return p
	}
	f, err := juleio.OpenJuleF(path)
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
	p.Parsef(main, justDefs)
	return p
}

func exec_post_commands() {
	for _, cmd := range jule.SET.PostCommands {
		fmt.Println(">", cmd)
		parts := strings.SplitN(cmd, " ", -1)
		err := exec.Command(parts[0], parts[1:]...).Run()
		if err != nil {
			println(err.Error())
		}
	}
}

func generate_compile_command(source_path string) (c, cmd string) {
	var cpp strings.Builder
	cpp.WriteString("-g -O0 ")
	cpp.WriteString(source_path)
	return jule.SET.CompilerPath, cpp.String()
}

func do_spell(cpp string) {
	defer exec_post_commands()
	path := filepath.Join(jule.WORKING_PATH, jule.SET.CppOutDir)
	path = filepath.Join(path, jule.SET.CppOutName)
	write_output(path, cpp)
	switch jule.SET.Mode {
	case juleset.MODE_COMPILE:
		c, cmd := generate_compile_command(path)
		println(c + " " + cmd)
		command := exec.Command(c, strings.SplitN(cmd, " ", -1)...)
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

func main() {
	fpath := os.Args[0]
	t := compile(fpath, true, false, false)
	if t == nil {
		return
	}
	if print_logs(t) {
		os.Exit(0)
	}
	cpp := t.Cpp()
	append_standard(&cpp)
	do_spell(cpp)
}
