package hash

func destCity(paths [][]string) string {
	m := map[string]int{}
	for _, path := range paths {
		n := len(path)
		for k, v := range path {
			m[v] += n - k - 1
		}
	}
	for k, v := range m {
		if v == 0 {
			return k
		}
	}
	return ""
}
