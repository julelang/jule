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
	return out.String()
}
