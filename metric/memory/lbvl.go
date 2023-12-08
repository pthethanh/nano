package memory

type lbvl struct {
	kv []string
	m  map[string]string
}

func (l *lbvl) With(vs ...string) *lbvl {
	rs := make([]string, 0, len(l.kv)+len(vs))
	rs = append(rs, l.kv...)
	rs = append(rs, vs...)

	ll := lbvl{
		kv: rs,
	}
	return ll.make()
}

func (l *lbvl) make() *lbvl {
	m := make(map[string]string)
	for i := 0; i < len(l.kv); i += 2 {
		m[l.kv[i]] = l.kv[i+1]
	}
	l.m = m
	return l
}

func (l *lbvl) tags() []string {
	rs := make([]string, 0, len(l.m))
	for i := 0; i < len(l.kv); i += 2 {
		rs = append(rs, l.kv[i]+"="+l.kv[i+1])
	}
	return rs
}

func (l *lbvl) Len() int {
	return len(l.kv)
}
