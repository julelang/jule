package jule

import (
	"os"
	"path/filepath"
)

// Jule constants.
const VERSION = `@master`
const EXT = `.jule`
const API = "api"
const STDLIB = "std"
const ENTRY_POINT = "main"
const INIT_FN = "init"

// Environment Variables.
var LOCALIZATION_PATH string
var STDLIB_PATH string
var EXEC_PATH string
var WORKING_PATH string

func exit_err(msg string) {
	println(msg)
	const ERROR_EXIT_CODE = 0
	os.Exit(ERROR_EXIT_CODE)
}

func init() {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		exit_err(err.Error())
	}
	WORKING_PATH, err = os.Getwd()
	if err != nil {
		exit_err(err.Error())
	}
	EXEC_PATH = filepath.Dir(path)
	path = filepath.Join(EXEC_PATH, "..") // Go to parent directory
	STDLIB_PATH = filepath.Join(path, STDLIB)
}
