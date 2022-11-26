package juleapi

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/julelang/jule/pkg/juleio"
)

// IGNORE operator.
const IGNORE = "_"

// INIT_CALLER identifier.
const INIT_CALLER = "__julec_call_package_initializers"

const typeExtension = "_jt"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool { return id == IGNORE }

// Returns specified identifer as JuleC identifer.
// Equavalents: "JULEC_ID(" + id + ")"
func AsId(id string) string { return "_" + id }

func getPtrAsId(ptr unsafe.Pointer) string {
	address := fmt.Sprintf("%p", ptr)
	address = address[3:] // skip 0xc
	for i, r := range address {
		if r != '0' {
			address = address[i:]
			break
		}
	}
	return address
}

// OutId returns cpp output identifier form of given identifier.
func OutId(id string, f *juleio.File) string {
	if f != nil {
		var out strings.Builder
		out.WriteByte('f')
		out.WriteString(getPtrAsId(unsafe.Pointer(f)))
		out.WriteByte('_')
		out.WriteString(id)
		return out.String()
	}
	return AsId(id)
}

// AsTypeId returns given identifier as output type identifier.
func AsTypeId(id string) string { return id + typeExtension }
