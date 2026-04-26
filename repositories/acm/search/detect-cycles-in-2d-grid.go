package search

func containsCycle(grid [][]byte) bool {
	m := len(grid)
	if m == 0 {
		return false
	}
	n := len(grid[0])
	if n == 0 {
		return false
	}
	visited := make([][]int, m)
	for i := range visited {
		visited[i] = make([]int, n)
	}
	var dfs func(i, j, cnt int) bool
	dirs := [][]int{
		{1, 0},
		{0, 1},
		{-1, 0},
		{0, -1},
	}
	abs := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	dfs = func(i, j, cnt int) bool {
		visited[i][j] = cnt
		for _, dir := range dirs {
			x, y := dir[0], dir[1]
			if i+x < 0 {
				continue
			}
			if i+x >= m {
				continue
			}
			if j+y < 0 {
				continue
			}
			if j+y >= n {
				continue
			}
			if grid[i][j] != grid[i+x][j+y] {
				continue
			}
			if visited[i+x][j+y] == 0 {
				if dfs(i+x, j+y, cnt+1) {
					return true
				}
				continue
			}
			if abs(visited[i][j], visited[i+x][j+y]) >= 3 {
				return true
			}
		}
		return false
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if visited[i][j] != 0 {
				continue
			}
			cnt := 1
			if dfs(i, j, cnt) {
				return true
			}
		}
	}
	return false
}
