package prefixsum

func numberOfSubmatrices(grid [][]byte) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	y := make([]int, n)
	x := make([]int, n)
	ans := 0
	for i := 0; i < m; i++ {
		x1, y1 := 0, 0
		for j := 0; j < n; j++ {
			if grid[i][j] == 'X' {
				x1++
			}
			if grid[i][j] == 'Y' {
				y1++
			}
			x[j] += x1
			y[j] += y1
			if x[j] == y[j] && x[j] != 0 {
				ans++
			}
		}
	}
	return ans
}
