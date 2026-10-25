package array

func closedIsland(grid [][]int) int {
	ans := 0
	m := len(grid)
	if m == 0 {
		return ans
	}
	n := len(grid[0])
	if n == 0 {
		return ans
	}
	var dfs func(int, int) bool
	dfs = func(i, j int) bool {
		if i < 0 {
			return false
		}
		if j < 0 {
			return false
		}
		if i == m {
			return false
		}
		if j == n {
			return false
		}
		if grid[i][j] == 1 {
			return true
		}
		if grid[i][j] == 2 {
			return true
		}
		if i == 0 {
			return false
		}
		if j == 0 {
			return false
		}
		if i == m-1 {
			return false
		}
		if j == n-1 {
			return false
		}
		grid[i][j] = 2
		ok := dfs(i+1, j)
		ok = dfs(i-1, j) && ok
		ok = dfs(i, j-1) && ok
		ok = dfs(i, j+1) && ok
		return ok
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if grid[i][j] == 1 {
				continue
			}
			if grid[i][j] == 2 {
				continue
			}
			if dfs(i, j) {
				ans++
			}
		}
	}
	return ans
}
