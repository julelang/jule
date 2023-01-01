package parser

type precedencer struct {
	pairs [][]any
}

func (p *precedencer) set(level uint, expr any) {
	for i, pair := range p.pairs {
		pair_level := pair[0].(uint)
		if level > pair_level {
			first := p.pairs[:i]
			appended := append([][]any{{level, expr}}, p.pairs[i:]...)
			p.pairs = append(first, appended...)
			return
		}
	}
	p.pairs = append(p.pairs, []any{level, expr})
}

func (p *precedencer) get_lower() any {
	for i := len(p.pairs) - 1; i >= 0; i-- {
		data := p.pairs[i][1]
		if data != nil {
			return data
		}
	}
	return nil
}
