// Copyright 2021 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

// This is the main package of JuleC.
// Naming conventions fully same with Jule.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/julelang/jule"
	"github.com/julelang/jule/build"
	"github.com/julelang/jule/cmd/julec/obj/cxx"
	"github.com/julelang/jule/lex"
)

const CMD_HELP = "help"
const CMD_VERSION = "version"
const CMD_TOOL = "tool"

var HELP_MAP = [...][2]string{
	{CMD_HELP,    "Show help"},
	{CMD_VERSION, "Show version"},
	{CMD_TOOL,    "Tools for effective Jule"},
}

func help() {
	if len(os.Args) > 2 {
		print_error_message("invalid command: " + os.Args[2])
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

func print_error_message(msg string) { println(msg) }

func version() {
	if len(os.Args) > 2 {
		print_error_message("invalid command: " + os.Args[2])
		return
	}
	println("julec version", jule.VERSION)
}

func list_horizontal_slice(s []string) string {
	lst := fmt.Sprint(s)
	return lst[1 : len(lst)-1]
}

func tool() {
	if len(os.Args) == 2 {
		println(`tool commands:
 distos     Lists all supported operating systems
 distarch   Lists all supported architects`)
		return
	} else if len(os.Args) > 3 {
		print_error_message("invalid command: " + os.Args[3])
		return
	}

	cmd := os.Args[2]
	switch cmd {
	case "distos":
		print("supported operating systems:\n ")
		println(list_horizontal_slice(build.DISTOS))

	case "distarch":
		print("supported architects:\n ")
		println(list_horizontal_slice(build.DISTARCH))

	default:
		print_error_message("Undefined command: " + cmd)
	}
}

func process_command() bool {
	switch os.Args[1] {
	case CMD_HELP:
		help()

	case CMD_VERSION:
		version()

	case CMD_TOOL:
		tool()

	default:
		return false
	}

	return true
}

func init() {
	// Not started with arguments.
	// Here is "2" but "os.Args" always have one element for store working directory.
	if len(os.Args) < 2 {
		os.Exit(0)
	}

	if process_command() {
		os.Exit(0)
	}
}

func exit_err(msg string) {
	const ERROR_EXIT_CODE = 0

	println(msg)
	os.Exit(ERROR_EXIT_CODE)
}

func get_option(args []string, i *int) (arg string, content string) {
	for ; *i < len(args); *i++ {
		arg = args[*i]
		j := 0
		runes := []rune(arg)
		r := runes[j]
		if r != '-' {
			content += arg
			arg = "" // Forget argument
			continue
		}
		j++
		if j >= len(runes) {
			exit_err("undefined syntax: " + arg)
		}
		r = runes[j]
		if r == '-' {
			j++
			if j >= len(runes) {
				exit_err("undefined syntax: " + arg)
			}
			r = runes[j]
		}
		if !lex.Is_ident_rune(string(r)) {
			exit_err("undefined syntax: " + arg)
		}
		j++
		for ; j < len(runes); j++ {
			r = runes[j]
			if !lex.Is_space(r) && !lex.Is_letter(r) &&
				!lex.Is_decimal(byte(r)) && r != '_' && r != '-' {
				exit_err("undefined syntax: " + string(runes[j:]))
			}
		}
		break
	}
	return
}

func get_option_value(args []string, i *int) string {
	*i++ // Argument value is the next argument
	if *i < len(args) {
		arg := args[*i]
		return arg
	}
	return ""
}

func parse_out_option(args []string, i *int) {
	value := get_option_value(args, i)
	if value == "" {
		exit_err("missing option value: -o --out")
	}
	cxx.OUT = value
}

func parse_compiler_option(args []string, i *int) {
	value := get_option_value(args, i)
	switch value {
	case "":
		exit_err("missing option value: --compiler")

	case cxx.COMPILER_CLANG:
		cxx.COMPILER_PATH = cxx.COMPILER_PATH_CLANG

	case cxx.COMPILER_GCC:
		cxx.COMPILER_PATH = cxx.COMPILER_PATH_GCC

	default:
		exit_err("invalid option value for --compiler: " + value)
	}

	cxx.COMPILER = value
}

func parse_options(args []string) string {
	cmd := ""
	i := 1 // Start 1 because the index 0 is a path, not an command-line argument
	for ; i < len(args); i++ {
		arg, content := get_option(args, &i)
		cmd += content
		switch arg {
		case "":

		case "-o", "--out":
			parse_out_option(args, &i)

		case "-t", "--transpile":
			cxx.MODE = cxx.MODE_T

		case "-c", "--compile":
			cxx.MODE = cxx.MODE_C

		case "--compiler":
			parse_compiler_option(args, &i)

		default:
			exit_err("undefined option: " + arg)
		}
	}
	cmd = strings.TrimSpace(cmd)
	return cmd
}

func main() {
	path := parse_options(os.Args)
	if path == "" {
		exit_err(build.Errorf("missing_compile_path"))
	}

	cxx.Compile(path)
}
