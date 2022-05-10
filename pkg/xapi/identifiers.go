package xapi

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/the-xlang/xxc/pkg/xio"
)

// Ignore operator.
const Ignore = "_"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool { return id == Ignore }

// Returns specified identifer as X identifer.
func AsId(id string) string { return "XID(" + id + ")" }

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

// OutId returns cxx output identifier form of given identifier.
func OutId(id string, f *xio.File) string {
	var out strings.Builder
	/*path = strings.ReplaceAll(path, string(os.PathSeparator), "_")
	pah = strings.ReplaceAll(path, ":", "_")*/
	if f != nil {
		out.WriteString(getPtrAsId(unsafe.Pointer(f)))
		out.WriteByte('_')
	}
	out.WriteString(id)
	return AsId(out.String())
}

// AsTypeId returns given identifier as output type identifier.
func AsTypeId(id string) string { return id + "_xt" }
