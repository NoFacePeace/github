package array

import "math"

func maxScore(grid [][]int) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	left := make([][]int, m)
	right := make([][]int, m)
	for i := 0; i < m; i++ {
		left[i] = make([]int, n)
		right[i] = make([]int, n)
		min := 0
		for j := 0; j < n; j++ {
			if j == 0 {
				left[i][j] = math.MinInt
				min = grid[i][j]
				continue
			}
			if grid[i][j] < min {
				left[i][j] = grid[i][j] - min
				min = grid[i][j]
				continue
			}
			left[i][j] = grid[i][j] - min
		}
	}

	for i := n - 1; i >= 0; i-- {
		max := 0
		for j := m - 1; j >= 0; j-- {

			if j == m-1 {
				right[j][i] = math.MinInt
				max = grid[j][i]
				continue
			}
			if grid[j][i] > max {
				right[j][i] = max - grid[j][i]
				max = grid[j][i]
				continue
			}
			right[j][i] = max - grid[j][i]
		}
	}
	max := math.MinInt
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if left[i][j] > 0 && right[i][j] > 0 {
				score := left[i][j] + right[i][j]
				if score > max {
					max = score
				}
				continue
			}
			if right[i][j] > 0 {
				score := right[i][j]
				if score > max {
					max = score
				}
				continue
			}
			if left[i][j] > 0 {
				score := left[i][j]
				if score > max {
					max = score
				}
				continue
			}
			if right[i][j] > left[i][j] {
				score := right[i][j]
				if score > max {
					max = score
				}
				continue
			}
			score := left[i][j]
			if score > max {
				max = score
			}
		}
	}
	return max
}
