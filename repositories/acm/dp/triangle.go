package dp

import "math"

func minimumTotal(triangle [][]int) int {
	m := len(triangle)
	if m == 0 {
		return 0
	}
	for i := 0; i < m; i++ {
		if i == 0 {
			continue
		}
		for j := 0; j < len(triangle[i]); j++ {
			if j == 0 {
				triangle[i][j] += triangle[i-1][j]
				continue
			}
			if j == len(triangle[i])-1 {
				triangle[i][j] += triangle[i-1][j-1]
				continue
			}
			triangle[i][j] += min(triangle[i-1][j], triangle[i-1][j-1])
		}
	}
	mx := math.MaxInt
	for i := 0; i < len(triangle[m-1]); i++ {
		mx = min(triangle[m-1][i], mx)
	}
	return mx
}
