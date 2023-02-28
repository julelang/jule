package build

import (
	"strconv"
	"strings"
)

// CPP_IGNORE is the ignoring of cpp.
const CPP_IGNORE = "std::ignore"

// CPP_SELF is the self keyword equavalent of cpp.
const CPP_SELF = "this"

// CPP_DEFAULT_EXPR represents default expression for type.
const CPP_DEFAULT_EXPR = "{}"

// TYPE_EXTENSION is extension of data types.
const TYPE_EXTENSION = "_jt"

// Returns specified identifer as JuleC identifer.
// Equavalents: "JULEC_ID(" + id + ")"
func AsId(id string) string { return "_" + id }

func get_ptr_as_id(ptr uintptr) string {
	address := "_" + strconv.FormatUint(uint64(ptr), 16)
	for i, r := range address {
		if r != '0' {
			address = address[i:]
			break
		}
	}
	return address
}

// OutId returns cpp output identifier form of given identifier.
func OutId(id string, ptr uintptr) string {
	if ptr != 0 {
		var out strings.Builder
		out.WriteString(get_ptr_as_id(ptr))
		out.WriteByte('_')
		out.WriteString(id)
		return out.String()
	}
	return AsId(id)
}

// AsTypeId returns given identifier as output type identifier.
func AsTypeId(id string) string { return id + TYPE_EXTENSION }
