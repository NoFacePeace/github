package str

import "sort"

func removeAnagrams(words []string) []string {
	m := map[string]bool{}
	ans := []string{}
	for _, v := range words {
		bs := []byte(v)
		sort.Slice(bs, func(i, j int) bool {
			return bs[i] < bs[j]
		})
		str := string(bs)
		if _, ok := m[str]; ok {
			continue
		}
		ans = append(ans, v)
		m = map[string]bool{}
		m[str] = true
	}
	return ans
}
