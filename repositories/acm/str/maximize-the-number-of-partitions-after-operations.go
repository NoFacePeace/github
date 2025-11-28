package str

import "fmt"

func maxPartitionsAfterOperations(s string, k int) int {
	ans := 0
	idx := 0
	var dfs func(s string) int
	dfs = func(s string) int {
		if len(s) == 0 {
			return 0
		}
		m := map[string]bool{}
		ret := 1
		n := len(s)
		for i := 0; i < n; i++ {
			sub := s[i : i+1]
			m[sub] = true
			if len(m) == k+1 {
				if idx == 277 {
					fmt.Println(s[:i])
				}
				ret = dfs(s[i:]) + 1
				break
			}
		}
		return ret
	}
	for i := 0; i < len(s); i++ {
		idx = i
		bs := []byte(s)
		bs[i] = '.'
		ans = max(ans, dfs(string(bs)))
	}
	return ans
}
