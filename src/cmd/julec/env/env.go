package env

import (
	"os"
	"path/filepath"

	"github.com/julelang/jule"
)

// Environment Variables.
var STDLIB_PATH string
var EXEC_PATH string
var WORKING_PATH string

// JULE_HEADER is the header path of "jule.hpp"
var JULE_HEADER string

func exit_err(msg string) {
	const ERROR_EXIT_CODE = 0

	println(msg)
	os.Exit(ERROR_EXIT_CODE)
}

func init() {
	path, err := os.Executable()
	if err != nil {
		exit_err(err.Error())
	}
	WORKING_PATH, err = os.Getwd()
	if err != nil {
		exit_err(err.Error())
	}
	EXEC_PATH = filepath.Dir(path)
	path = filepath.Join(EXEC_PATH, "..") // Go to parent directory
	STDLIB_PATH = filepath.Join(path, jule.STDLIB)

	JULE_HEADER = filepath.Join(EXEC_PATH, "..")
	JULE_HEADER = filepath.Join(JULE_HEADER, "api")
	JULE_HEADER = filepath.Join(JULE_HEADER, "jule.hpp")
}
