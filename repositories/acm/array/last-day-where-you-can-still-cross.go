package array

func latestDayToCross(row int, col int, cells [][]int) int {
	matrix := make([][]int, row)
	for k := range matrix {
		matrix[k] = make([]int, col)
	}
	check := func(mid int) bool {
		visited := make([][]bool, row)
		for k := range visited {
			visited[k] = make([]bool, col)
		}
		var f func(i, j int) bool
		f = func(i, j int) bool {
			if i < 0 {
				return false
			}
			if j < 0 {
				return false
			}
			if i >= row {
				return false
			}
			if j >= col {
				return false
			}
			if matrix[i][j] <= mid {
				return false
			}
			if visited[i][j] {
				return false
			}
			if i == row-1 {
				return true
			}
			visited[i][j] = true
			return f(i+1, j) || f(i-1, j) || f(i, j+1) || f(i, j-1)
		}
		for i := 0; i < col; i++ {
			if f(0, i) {
				return true
			}
		}
		return false
	}
	for k, cell := range cells {
		r, c := cell[0], cell[1]
		matrix[r-1][c-1] = k + 1
	}
	left := 1
	right := len(cells)
	ans := -1
	for left < right {
		mid := (left + right) / 2
		if check(mid) {
			ans = mid
			left = mid + 1
			continue
		}
		right = mid
	}
	return ans
}
