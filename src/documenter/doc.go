package documenter

import "github.com/julelang/jule/parser"

// Doc returns documentation of code into JSON format.
func Doc(p *parser.Parser) (string, error) {
	return doc_fmt_json(p)
}
