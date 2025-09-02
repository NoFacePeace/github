package dp

func countSquares(matrix [][]int) int {
	ans := 0
	m := len(matrix)
	if m == 0 {
		return ans
	}
	n := len(matrix[0])
	if n == 0 {
		return ans
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if matrix[i][j] == 0 {
				continue
			}
			if i == 0 {
				ans++
				continue
			}
			if j == 0 {
				ans++
				continue
			}
			mn := min(matrix[i][j-1], matrix[i-1][j])
			mn = min(mn, matrix[i-1][j-1])
			ans += mn + 1
			matrix[i][j] = mn + 1
		}
	}
	return ans
}
