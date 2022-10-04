package documenter

import "github.com/jule-lang/jule/parser"

// Doc returns documentation of code into JSON format.
func Doc(p *parser.Parser, json bool) (string, error) {
	if json {
		return doc_fmt_json(p)
	}
	return doc_fmt_meta(p)
}
