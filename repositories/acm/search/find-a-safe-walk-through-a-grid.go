package search

func findSafeWalk(grid [][]int, health int) bool {
	m := len(grid)
	n := len(grid[0])
	visited := make([][]int, m)
	for i := range visited {
		visited[i] = make([]int, n)
	}
	var dfs func(x, y, h int) bool
	dfs = func(x, y, h int) bool {
		if x < 0 {
			return false
		}
		if x >= m {
			return false
		}
		if y < 0 {
			return false
		}
		if y >= n {
			return false
		}
		h = h - grid[x][y]
		if h <= 0 {
			return false
		}
		if h <= visited[x][y] {
			return false
		}
		visited[x][y] = h
		if x == m-1 && y == n-1 {
			return true
		}
		if dfs(x+1, y, h) || dfs(x-1, y, h) || dfs(x, y+1, h) || dfs(x, y-1, h) {
			return true
		}
		return false
	}
	return dfs(0, 0, health)
}
