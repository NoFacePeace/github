package array

func countNegatives(grid [][]int) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	i := 0
	j := n - 1
	ans := 0
	for i < m && j >= 0 {
		for j >= 0 {
			if grid[i][j] >= 0 {
				break
			}
			ans += m - i
			j--
		}
		if j == -1 {
			break
		}
		for i < m {
			if grid[i][j] >= 0 {
				i++
				continue
			}
			ans += m - i
			j--
			break
		}
	}
	return ans
}
