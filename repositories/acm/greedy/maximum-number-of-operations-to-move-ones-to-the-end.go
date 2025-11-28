package greedy

func maxOperations(s string) int {
	ans := 0
	var dfs func(s string) int
	dfs = func(s string) int {
		if len(s) == 0 {
			return 0
		}
		if len(s) == 1 {
			return 0
		}
		ret := dfs(s[1:])
		if s[0] == '0' {
			return ret
		}
		if s[1] == '1' {
			ans += ret
			return ret
		}
		ans += ret + 1
		return ret + 1
	}
	dfs(s)
	return ans
}
