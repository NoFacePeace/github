package str

func findLexSmallestString(s string, a int, b int) string {
	var dfs func(string)
	visited := map[string]bool{}
	ans := s
	dfs = func(s string) {
		if visited[s] {
			return
		}
		visited[s] = true
		if ans > s {
			ans = s
		}
		dfs(s[b:] + s[:b])
		bs := []byte(s)
		n := len(bs)
		for i := 0; i < n; i++ {
			if i%2 == 0 {
				continue
			}
			bs[i] = byte('0' + (int(bs[i]-'0')+a)%10)
		}
		dfs(string(bs))
	}
	dfs(s)
	return ans
}
