package parser

type precedencer struct {
	pairs [][]any
}

func (p *precedencer) set(level uint, expr any) {
	for _, pair := range p.pairs {
		pair_level := pair[0].(uint)
		if pair_level == level {
			if pair[1] == nil {
				pair[1] = expr
			}
			return
		}
	}
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

func (p *precedencer) get() any {
	for _, pair := range p.pairs {
		data := pair[1]
		if data != nil {
			return data
		}
	}
	return nil
}
