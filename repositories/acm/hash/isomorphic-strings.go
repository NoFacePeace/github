package hash

func isIsomorphic(s string, t string) bool {
	if len(s) != len(t) {
		return false
	}
	f := func(s string) string {
		arr := []byte(s)
		start := byte('a')
		m := map[byte]byte{}
		ret := []byte{}
		for _, v := range arr {
			if _, ok := m[v]; !ok {
				ret = append(ret, start)
				m[v] = start
				start++
				continue
			}
			ret = append(ret, m[v])
		}
		return string(ret)
	}
	return f(s) == f(t)
}
