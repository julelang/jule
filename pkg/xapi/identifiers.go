package xapi

import (
	"fmt"
	"strings"

	"github.com/the-xlang/xxc/pkg/xio"
)

// Ignore operator.
const Ignore = "_"

// IsIgnoreId reports identifier is ignore or not.
func IsIgnoreId(id string) bool { return id == Ignore }

// Returns specified identifer as X identifer.
func AsId(id string) string { return "XID(" + id + ")" }

// OutId returns cxx output identifier form of given identifier.
func OutId(id string, f *xio.File) string {
	var out strings.Builder
	/*path = strings.ReplaceAll(path, string(os.PathSeparator), "_")
	pah = strings.ReplaceAll(path, ":", "_")*/
	if f != nil {
		out.WriteString(fmt.Sprintf("%p", f))
		out.WriteByte('_')
	}
	out.WriteString(id)
	return AsId(out.String())
}

// AsTypeId returns given identifier as output type identifier.
func AsTypeId(id string) string { return id + "_xt" }
