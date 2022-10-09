package jule

import "github.com/jule-lang/jule/pkg/juleset"

// Jule constants.
const VERSION       = `@development_channel`
const SRC_EXT       = `.jule`
const DOC_EXT       = SRC_EXT + "doc"
const SETTINGS_FILE = "jule.set"
const STDLIB        = "std"
const LOCALIZATIONS = "localization"

const ENTRY_POINT = "main"
const INIT_FN     = "init"

const ANONYMOUS = "<anonymous>"

const COMMENT_PRAGMA_SEP    = ":"
const PRAGMA_COMMENT_PREFIX = "jule" + COMMENT_PRAGMA_SEP

const OS_WINDOWS = "windows"
const OS_LINUX   = "linux"
const OS_DARWIN  = "darwin"
const OS_UNIX    = "unix"

const ARCH_ARM   = "arm"
const ARCH_ARM64 = "arm64"
const ARCH_AMD64 = "amd64"
const ARCH_I386  = "i386"
const ARCH_64Bit = "64bit"
const ARCH_32Bit = "32bit"

// This attributes should be added to the attribute map.
const ATTR_CDEF    = "cdef"
const ATTR_TYPEDEF = "typedef"

const PREPROCESSOR_DIRECTIVE      = "pragma"
const PREPROCESSOR_DIRECTIVE_ENOFI = "enofi"

const MARK_ARRAY = "..."

const PREFIX_SLICE = "[]"
const PREFIX_ARRAY = "[" + MARK_ARRAY + "]"

const COMPILER_GCC   = "gcc"
const COMPILER_CLANG = "clang"

// Environment Variables.
var LOCALIZATION_PATH string
var STDLIB_PATH string
var EXEC_PATH string
var WORKING_PATH string
var SET *juleset.Set
