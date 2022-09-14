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
	"reflect"
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

const command_help = "help"
const command_version = "version"
const command_init = "init"
const command_doc = "doc"
const command_bug = "bug"
const command_tool = "tool"

var helpmap = [...][2]string{
	0: {command_help, "Show help"},
	1: {command_version, "Show version"},
	2: {command_init, "Initialize new project here"},
	3: {command_doc, "Documentize Jule source code"},
	4: {command_bug, "Start a new bug report"},
	5: {command_tool, "Tools for effective Jule"},
}

func help(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	max := len(helpmap[0][0])
	for _, key := range helpmap {
		len := len(key[0])
		if len > max {
			max = len
		}
	}
	var sb strings.Builder
	const space = 5 // Space of between command name and description.
	for _, part := range helpmap {
		sb.WriteString(part[0])
		sb.WriteString(strings.Repeat(" ", (max-len(part[0]))+space))
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
	println("julec version", jule.Version)
}

func init_project(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	bytes, err := json.MarshalIndent(*juleset.Default, "", "\t")
	if err != nil {
		println(err)
		os.Exit(0)
	}
	err = os.WriteFile(jule.SettingsFile, bytes, 0666)
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
		path = path[:len(path)-len(jule.SrcExt)]
		path = filepath.Join(jule.Set.CppOutDir, path+jule.DocExt)
		write_output(path, docjson)
	}
}

func open_url(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	command := exec.Command(cmd, args...)
	return command.Start()
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
		println(list_horizontal_slice(jule.Distos))
	case "distarch":
		print("supported architects:\n ")
		println(list_horizontal_slice(jule.Distarch))
	default:
		println("Undefined command: " + cmd)
	}
}

func process_command(namespace, cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	switch namespace {
	case command_help:
		help(cmd)
	case command_version:
		version(cmd)
	case command_init:
		init_project(cmd)
	case command_doc:
		doc(cmd)
	case command_bug:
		bug(cmd)
	case command_tool:
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
	jule.WorkingPath, err = os.Getwd()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	execp = filepath.Dir(execp)
	jule.ExecPath = execp
	jule.StdlibPath = filepath.Join(jule.ExecPath, jule.Stdlib)
	juleapi.JuleCHeader = filepath.Join(jule.ExecPath, "api")
	juleapi.JuleCHeader = filepath.Join(juleapi.JuleCHeader, "julec.hpp")
	jule.LangsPath = filepath.Join(jule.ExecPath, jule.Localizations)

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
	lang := strings.TrimSpace(jule.Set.Language)
	if lang == "" || lang == "default" {
		return
	}
	path := filepath.Join(jule.LangsPath, lang+".json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		println("Language couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &jule.Errors)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func check_mode() {
	lower := strings.ToLower(jule.Set.Mode)
	if lower != juleset.ModeTranspile &&
		lower != juleset.ModeCompile {
		key, _ := reflect.TypeOf(jule.Set).Elem().FieldByName("Mode")
		tag := string(key.Tag)
		// 6 for skip "json:
		tag = tag[6 : len(tag)-1]
		println(jule.GetError("invalid_value_for_key", jule.Set.Mode, tag))
		os.Exit(0)
	}
	jule.Set.Mode = lower
}

func load_juleset() {
	// File check.
	info, err := os.Stat(jule.SettingsFile)
	if err != nil || info.IsDir() {
		jule.Set = new(juleset.Set)
		*jule.Set = *juleset.Default
		return
	}
	bytes, err := os.ReadFile(jule.SettingsFile)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	jule.Set, err = juleset.Load(bytes)
	if err != nil {
		println("Jule settings has errors;")
		println(err.Error())
		os.Exit(0)
	}
	load_localization()
	check_mode()
}

// print_logs prints logs and returns true
// if logs has error, false if not.
func print_logs(p *parser.Parser) bool {
	var str strings.Builder
	for _, log := range p.Warnings {
		str.WriteString(log.String())
		str.WriteByte('\n')
	}
	for _, log := range p.Errors {
		str.WriteString(log.String())
		str.WriteByte('\n')
	}
	print(str.String())
	return len(p.Errors) > 0
}

func append_standard(code *string) {
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)",
		day, month, year, hour, min)
	var sb strings.Builder
	sb.WriteString("// Auto generated by JuleC.\n")
	sb.WriteString("// JuleC version: ")
	sb.WriteString(jule.Version)
	sb.WriteByte('\n')
	sb.WriteString("// Date: ")
	sb.WriteString(timeStr)
	sb.WriteString("\n\n#include \"")
	sb.WriteString(juleapi.JuleCHeader)
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
	inf, err := os.Stat(jule.StdlibPath)
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
	for _, cmd := range jule.Set.PostCommands {
		fmt.Println(">", cmd)
		parts := strings.SplitN(cmd, " ", -1)
		err := exec.Command(parts[0], parts[1:]...).Run()
		if err != nil {
			println(err.Error())
		}
	}
}

func do_spell(path, cpp string) {
	defer exec_post_commands()
	write_output(path, cpp)
	switch jule.Set.Mode {
	case juleset.ModeCompile:
		defer os.Remove(path)
		println("compilation is not supported yet")
	}
}

func main() {
	fpath := os.Args[0]
	p := compile(fpath, true, false, false)
	if p == nil {
		return
	}
	if print_logs(p) {
		os.Exit(0)
	}
	cpp := p.Cpp()
	append_standard(&cpp)
	path := filepath.Join(jule.WorkingPath, jule.Set.CppOutDir)
	path = filepath.Join(path, jule.Set.CppOutName)
	do_spell(path, cpp)
}
