package array

func minFlips(grid [][]int) int {
	m := len(grid)
	if m <= 1 {
		return 0
	}
	n := len(grid[0])
	if n <= 1 {
		return 0
	}
	row := 0
	column := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if j < n/2 && grid[i][j] != grid[i][n-j-1] {
				row++
			}
			if i < m/2 && grid[i][j] != grid[m-i-1][j] {
				column++
			}
		}
	}
	return min(row, column)
}
