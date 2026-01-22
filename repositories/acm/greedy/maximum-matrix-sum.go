package greedy

import "math"

func maxMatrixSum(matrix [][]int) int64 {
	cnt := 0
	mn := math.MaxInt
	n := len(matrix)
	sum := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if matrix[i][j] < 0 {
				matrix[i][j] = -matrix[i][j]
				cnt++
			}
			sum += matrix[i][j]
			mn = min(mn, matrix[i][j])
		}
	}
	if cnt%2 == 0 {
		return int64(sum)
	}
	return int64(sum - mn - mn)
}
