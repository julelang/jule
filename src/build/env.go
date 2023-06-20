package build

import (
	"os"
	"path/filepath"
)

// Environment Variables.
var PATH_STDLIB string
var PATH_EXEC string
var PATH_WD string

// PATH_API is the header path of "jule.hpp"
var PATH_API string

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
	PATH_WD, err = os.Getwd()
	if err != nil {
		exit_err(err.Error())
	}
	PATH_EXEC = filepath.Dir(path)
	path = filepath.Join(PATH_EXEC, "..") // Go to parent directory
	PATH_STDLIB = filepath.Join(path, STDLIB)

	PATH_API = filepath.Join(PATH_EXEC, "..")
	PATH_API = filepath.Join(PATH_API, "api")
	PATH_API = filepath.Join(PATH_API, "jule.hpp")
}
