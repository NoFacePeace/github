package array

func countUnguarded(m int, n int, guards [][]int, walls [][]int) int {
	grid := make([][]int, m)
	for i := range grid {
		grid[i] = make([]int, n)
	}
	for _, guard := range guards {
		x, y := guard[0], guard[1]
		grid[x][y] = 2
	}
	for _, wall := range walls {
		x, y := wall[0], wall[1]
		grid[x][y] = 3
	}
	for _, guard := range guards {
		x, y := guard[0], guard[1]
		for i := y + 1; i < n; i++ {
			if grid[x][i] >= 2 {
				break
			}
			grid[x][i] = 1
		}
		for i := y - 1; i >= 0; i-- {
			if grid[x][i] >= 2 {
				break
			}
			grid[x][i] = 1
		}
		for i := x + 1; i < m; i++ {
			if grid[i][y] >= 2 {
				break
			}
			grid[i][y] = 1
		}
		for i := x - 1; i >= 0; i-- {
			if grid[i][y] >= 2 {
				break
			}
			grid[i][y] = 1
		}
	}
	ans := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if grid[i][j] == 0 {
				ans++
			}
		}
	}
	return ans
}
