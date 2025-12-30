package str

func pyramidTransition(bottom string, allowed []string) bool {
	m := map[string][]string{}
	for _, v := range allowed {
		prefix := v[:2]
		suffix := v[2:]
		if _, ok := m[prefix]; !ok {
			m[prefix] = []string{}
		}
		m[prefix] = append(m[prefix], suffix)
	}
	var f func(bottom string, top string) bool
	f = func(bottom, top string) bool {
		if len(bottom) == 1 && len(top) == 0 {
			return true
		}
		if len(bottom) == 1 {
			return f(top, "")
		}
		prefix := bottom[:2]
		if _, ok := m[prefix]; !ok {
			return false
		}
		for _, v := range m[prefix] {
			if f(bottom[1:], top+v) {
				return true
			}
		}
		return false
	}
	return f(bottom, "")
}
