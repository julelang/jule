// Copyright 2021 The X Authors.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/the-xlang/x/parser"
	"github.com/the-xlang/x/pkg/io"
	"github.com/the-xlang/x/pkg/x"
)

func help(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	println(`help        Show help.
version     Show version.`)
}

func version(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	println("The X Programming Language\n" + x.Version)
}

func processCommand(namespace, cmd string) bool {
	switch namespace {
	case "help":
		help(cmd)
	case "version":
		version(cmd)
	default:
		return false
	}
	return true
}

func init() {
	x.ExecutablePath = filepath.Dir(os.Args[0])
	// Not started with arguments.
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	var sb strings.Builder
	for _, arg := range os.Args[1:] {
		sb.WriteString(" " + arg)
	}
	os.Args[0] = sb.String()[1:]
	arg := os.Args[0]
	index := strings.Index(arg, " ")
	if index == -1 {
		index = len(arg)
	}
	if processCommand(arg[:index], arg[index:]) {
		os.Exit(0)
	}
}

func printErrors(errors []string) {
	defer os.Exit(0)
	for _, message := range errors {
		fmt.Println(message)
	}
}

var routines *sync.WaitGroup

func main() {
	f, err := io.GetX(os.Args[0])
	if err != nil {
		println(err.Error())
		return
	}
	routines = new(sync.WaitGroup)
	info := new(parser.ParseFileInfo)
	info.File = f
	info.Routines = routines
	routines.Add(1)
	go parser.ParseFile(info)
	routines.Wait()
	if info.Errors != nil {
		printErrors(info.Errors)
	}
	os.WriteFile("x.cxx", []byte(info.X_CXX), 0606)
}
