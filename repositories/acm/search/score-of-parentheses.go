package search

func scoreOfParentheses(s string) int {
	var dfs func(int, int) int
	dfs = func(l, r int) int {
		if l+1 == r {
			return 1
		}
		cnt := 0
		i := l
		for i <= r {
			if s[i:i+1] == "(" {
				cnt++
			} else {
				cnt--
			}
			if cnt == 0 {
				break
			}
			i++
		}
		if i >= r {
			return dfs(l+1, r-1) * 2
		}
		return dfs(l, i) + dfs(i+1, r)
	}
	n := len(s)
	return dfs(0, n-1)
}
