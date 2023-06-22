// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package build

import (
	"path/filepath"
)

const EXT = `.jule`
const API = "api"
const STDLIB = "std"
const ENTRY_POINT = "main"
const INIT_FN = "init"


// Reports whether file path is Jule source code.
func Is_jule(path string) bool { return filepath.Ext(path) == EXT }
