package prefixsum

func canPartitionGrid(grid [][]int) bool {
	m := len(grid)
	if m == 0 {
		return true
	}
	n := len(grid[0])
	if n == 0 {
		return true
	}
	rows := make([]int, m)
	columns := make([]int, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			rows[i] += grid[i][j]
			columns[j] += grid[i][j]
			if j != 0 && i == m-1 {
				columns[j] += columns[j-1]
			}
		}
		if i != 0 {
			rows[i] += rows[i-1]
		}
	}
	for i := 0; i < m; i++ {
		if rows[m-1]-rows[i] == rows[i] {
			return true
		}
	}
	for j := 0; j < n; j++ {
		if columns[n-1]-columns[j] == columns[j] {
			return true
		}
	}
	return false
}
