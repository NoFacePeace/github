package array

func findBall(grid [][]int) []int {
	m := len(grid)
	if m == 0 {
		return []int{}
	}
	n := len(grid[0])
	ans := []int{}
	var dfs func(x, y int) int
	dfs = func(x, y int) int {
		if x == m {
			return y
		}
		if y == 0 && grid[x][y] == -1 {
			return -1
		}
		if y == n-1 && grid[x][y] == 1 {
			return -1
		}
		if grid[x][y] == 1 {
			if grid[x][y] != grid[x][y+1] {
				return -1
			}
			return dfs(x+1, y+1)
		}
		if grid[x][y] != grid[x][y-1] {
			return -1
		}
		return dfs(x+1, y-1)

	}
	for i := 0; i < n; i++ {
		ans = append(ans, dfs(0, i))
	}
	return ans
}
