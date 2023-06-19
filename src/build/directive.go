// Copyright 2023 The Jule Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package build

// Prefix of directive comments.
const DIRECTIVE_PREFIX = "jule:"

// These directives must added to the DIRECTIVES.
const DIRECTIVE_CDEF = "cdef"        // Directive: jule:cdef
const DIRECTIVE_TYPEDEF = "typedef"  // Directive: jule:typedef
const DIRECTIVE_DERIVE = "derive"    // Directive: jule:derive
const DIRECTIVE_PASS = "pass"        // Directive: jule:pass

const DERIVE_CLONE = "Clone"

// List of all directives.
var DIRECTIVES = [...]string{
	DIRECTIVE_CDEF,
	DIRECTIVE_TYPEDEF,
	DIRECTIVE_DERIVE,
	DIRECTIVE_PASS,
}

// Reports whether directive is top-directive.
func Is_top_directive(directive string) bool {
	return directive == DIRECTIVE_PASS
}
