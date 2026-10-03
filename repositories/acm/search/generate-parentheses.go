package search

func generateParenthesis(n int) []string {
	ans := []string{}
	var dfs func(string, int)
	dfs = func(str string, cnt int) {
		if len(str) == n*2 {
			ans = append(ans, str)
			return
		}
		if cnt < n {
			dfs(str+"(", cnt+1)
		}
		if len(str)-cnt < cnt {
			dfs(str+")", cnt)
		}
	}
	dfs("", 0)
	return ans
}
