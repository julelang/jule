package jule

// Jule constants.
const VERSION       = `@development_channel`
const SRC_EXT       = `.jule`
const SETTINGS_FILE = "jule.set"
const API           = "api"
const STDLIB        = "std"
const LOCALIZATIONS = "localization"
const ENTRY_POINT = "main"
const INIT_FN     = "init"

const ANONYMOUS = "<anonymous>"

const COMMENT_PRAGMA_SEP    = ":"
const PRAGMA_COMMENT_PREFIX = "jule" + COMMENT_PRAGMA_SEP

// This attributes should be added to the attribute map.
const ATTR_CDEF    = "cdef"
const ATTR_TYPEDEF = "typedef"

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
