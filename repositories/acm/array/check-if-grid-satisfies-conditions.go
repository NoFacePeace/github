package array

func satisfiesConditions(grid [][]int) bool {
	m := len(grid)
	if m == 0 {
		return true
	}
	n := len(grid[0])
	if n == 0 {
		return true
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if j != 0 && grid[i][j] == grid[i][j-1] {
				return false
			}
			if i != 0 && grid[i][j] != grid[i-1][j] {
				return false
			}
		}
	}
	return true
}
