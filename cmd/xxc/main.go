// Copyright 2021 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/the-xlang/xxc/documenter"
	"github.com/the-xlang/xxc/parser"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xset"
)

type Parser = parser.Parser

const commandHelp = "help"
const commandVersion = "version"
const commandInit = "init"
const commandDoc = "doc"

const localizationErrors = "error.json"
const localizationWarnings = "warning.json"

var helpmap = [...][2]string{
	0: {commandHelp, "Show help."},
	1: {commandVersion, "Show version."},
	2: {commandInit, "Initialize new project here."},
	3: {commandDoc, "Documentize X source code."},
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
	println("xxc version", x.Version)
}

func initProject(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	bytes, err := json.MarshalIndent(*xset.Default, "", "\t")
	if err != nil {
		println(err)
		os.Exit(0)
	}
	err = ioutil.WriteFile(x.SettingsFile, bytes, 0666)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	println("Initialized project.")
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
		if printlogs(p) {
			fmt.Println(x.GetError("doc_couldnt_generated", path))
			continue
		}
		docjson, err := documenter.Doc(p)
		if err != nil {
			fmt.Println(x.GetError("error", err.Error()))
			continue
		}
		// Remove SrcExt from path
		path = path[:len(path)-len(x.SrcExt)]
		path = filepath.Join(x.Set.CppOutDir, path+x.DocExt)
		writeOutput(path, docjson)
	}
}

func processCommand(namespace, cmd string) bool {
	switch namespace {
	case commandHelp:
		help(cmd)
	case commandVersion:
		version(cmd)
	case commandInit:
		initProject(cmd)
	case commandDoc:
		doc(cmd)
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
	execp = filepath.Dir(execp)
	x.ExecPath = execp
	x.StdlibPath = filepath.Join(x.ExecPath, x.Stdlib)
	xapi.XXCHeader = filepath.Join(x.ExecPath, "api")
	xapi.XXCHeader = filepath.Join(xapi.XXCHeader, "xxc.hpp")
	x.LangsPath = filepath.Join(x.ExecPath, x.Localizations)

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
	if processCommand(arg[:i], arg[i:]) {
		os.Exit(0)
	}
}

func loadLangWarns(path string, infos []fs.FileInfo) {
	i := -1
	for j, f := range infos {
		if f.IsDir() || f.Name() != localizationWarnings {
			continue
		}
		i = j
		path = filepath.Join(path, f.Name())
		break
	}
	if i == -1 {
		return
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		println("Language's warnings couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &x.Warnings)
	if err != nil {
		println("Language's warnings couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func loadLangErrs(path string, infos []fs.FileInfo) {
	i := -1
	for j, f := range infos {
		if f.IsDir() || f.Name() != localizationErrors {
			continue
		}
		i = j
		path = filepath.Join(path, f.Name())
		break
	}
	if i == -1 {
		return
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &x.Errors)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func loadLang() {
	lang := strings.TrimSpace(x.Set.Language)
	if lang == "" || lang == "default" {
		return
	}
	path := filepath.Join(x.LangsPath, lang)
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		println("Language couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	loadLangWarns(path, infos)
	loadLangErrs(path, infos)
}

func checkMode() {
	lower := strings.ToLower(x.Set.Mode)
	if lower != xset.ModeTranspile &&
		lower != xset.ModeCompile {
		key, _ := reflect.TypeOf(x.Set).Elem().FieldByName("Mode")
		tag := string(key.Tag)
		// 6 for skip "json:
		tag = tag[6 : len(tag)-1]
		println(x.GetError("invalid_value_for_key", x.Set.Mode, tag))
		os.Exit(0)
	}
	x.Set.Mode = lower
}

func loadXSet() {
	// File check.
	info, err := os.Stat(x.SettingsFile)
	if err != nil || info.IsDir() {
		println(`X settings file ("` + x.SettingsFile + `") is not found!`)
		os.Exit(0)
	}
	bytes, err := os.ReadFile(x.SettingsFile)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	x.Set, err = xset.Load(bytes)
	if err != nil {
		println("X settings has errors;")
		println(err.Error())
		os.Exit(0)
	}
	loadLang()
	checkMode()
}

// printlogs prints logs and returns true
// if logs has error, false if not.
func printlogs(p *Parser) bool {
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

func appendStandard(code *string) {
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)",
		day, month, year, hour, min)
	var sb strings.Builder
	sb.WriteString("// Auto generated by XXC compiler.\n")
	sb.WriteString("// X compiler version:")
	sb.WriteString(x.Version)
	sb.WriteByte('\n')
	sb.WriteString("// Date: ")
	sb.WriteString(timeStr)
	sb.WriteString("\n\n#include \"")
	sb.WriteString(xapi.XXCHeader)
	sb.WriteString("\"\n")
	sb.WriteString(*code)
	*code = sb.String()
}

func writeOutput(path, content string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0o777)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	bytes := []byte(content)
	err = ioutil.WriteFile(path, bytes, 0o666)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
}

func compile(path string, main, nolocal, justDefs bool) *Parser {
	loadXSet()
	p := parser.New(nil)
	f, err := xio.Openfx(path)
	if err != nil {
		println(err.Error())
		return nil
	}
	if !xio.IsUseable(path) {
		p.PushErr("file_not_useable")
		return p
	}
	// Check standard library.
	inf, err := os.Stat(x.StdlibPath)
	if err != nil || !inf.IsDir() {
		p.PushErr("no_stdlib")
		return p
	}
	p.File = f
	p.NoLocalPkg = nolocal
	p.Parsef(main, justDefs)
	return p
}

func execPostCommands() {
	for _, cmd := range x.Set.PostCommands {
		fmt.Println(">", cmd)
		parts := strings.SplitN(cmd, " ", -1)
		err := exec.Command(parts[0], parts[1:]...).Run()
		if err != nil {
			println(err.Error())
		}
	}
}

func doSpell(path, cxx string) {
	defer execPostCommands()
	writeOutput(path, cxx)
	switch x.Set.Mode {
	case xset.ModeCompile:
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
	if printlogs(p) {
		os.Exit(0)
	}
	cxx := p.Cpp()
	appendStandard(&cxx)
	path := filepath.Join(x.Set.CppOutDir, x.Set.CppOutName)
	doSpell(path, cxx)
}
