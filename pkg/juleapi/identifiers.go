package juleapi

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/jule-lang/jule/pkg/juleio"
)

// Ignore operator.
const Ignore = "_"

// InitializerCaller identifier.
const InitializerCaller = "_julec___call_initializers"

const typeExtension = "_julet"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool {
	return id == Ignore
}

// Returns specified identifer as X identifer.
func AsId(id string) string {
	return "XID(" + id + ")"
}

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
func AsTypeId(id string) string {
	return id + typeExtension
}
