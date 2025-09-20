package backtrack

func makeTheIntegerZero(num1 int, num2 int) int {
	m := map[int]bool{}
	val := 2
	m[0] = true
	for i := 1; i <= 60; i++ {
		m[val] = true
		val *= 2
	}
	ans := -1
	min := func(a, b int) int {
		if a == -1 {
			return b
		}
		if a > b {
			return b
		}
		return a
	}
	visited := map[int]bool{}
	var dfs func(num1, cnt int)
	dfs = func(num1, cnt int) {
		if visited[num1] {
			return
		}
		if num1 < num2 {
			return
		}
		for k := range m {
			if m[num1-num2] {
				ans = min(ans, cnt+1)
				return
			}
			dfs(num1-num2-k, cnt+1)
		}
		visited[num1] = true
	}
	dfs(num1, 0)
	return ans
}
