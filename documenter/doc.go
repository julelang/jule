package documenter

import "github.com/jule-lang/jule/transpiler"

// Doc returns documentation of code into JSON format.
func Doc(t *transpiler.Transpiler, json bool) (string, error) {
	if json {
		return doc_fmt_json(t)
	}
	return doc_fmt_meta(t)
}
