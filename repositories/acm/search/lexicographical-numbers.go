package search

func lexicalOrder(n int) []int {
	ans := []int{}
	var dfs func(num int)
	dfs = func(num int) {
		if num > n {
			return
		}
		ans = append(ans, num)
		for i := 0; i <= 9; i++ {
			dfs(num*10 + i)
		}
	}
	for i := 1; i <= 9; i++ {
		dfs(i)
	}
	return ans
}
