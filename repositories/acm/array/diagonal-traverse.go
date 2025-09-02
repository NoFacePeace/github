package array

func findDiagonalOrder(mat [][]int) []int {
	ans := []int{}
	m := len(mat)
	if m == 0 {
		return ans
	}
	n := len(mat[0])
	if n == 0 {
		return ans
	}
	i, j := 0, 0
	x, y := -1, 1
	for i < m && j < n {
		ans = append(ans, mat[i][j])
		if i+x < m && j+y < n && i+x >= 0 && j+y >= 0 {
			i += x
			j += y
			continue
		}
		if i == 0 && j < n-1 {
			j++
			x = 1
			y = -1
			continue
		}
		if j == n-1 {
			i++
			x = 1
			y = -1
			continue
		}
		if j == 0 && i < m-1 {
			i++
			x = -1
			y = 1
			continue
		}
		if i == m-1 {
			j++
			x = -1
			y = 1
			continue
		}
	}
	return ans
}
