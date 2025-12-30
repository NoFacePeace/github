package array

func numMagicSquaresInside(grid [][]int) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	check := func(i, j int) bool {
		m := map[int]bool{}
		for x := i; x <= i+2; x++ {
			for y := j; y <= j+2; y++ {

				val := grid[x][y]
				if val > 9 {
					return false
				}
				if val < 1 {
					return false
				}
				m[val] = true
			}
		}
		if len(m) != 9 {
			return false
		}
		sum := grid[i][j] + grid[i][j+1] + grid[i][j+2]
		for k := i; k < i+3; k++ {
			tmp := grid[k][j] + grid[k][j+1] + grid[k][j+2]
			if tmp != sum {
				return false
			}
		}
		for k := j; k < j+3; k++ {
			tmp := grid[i][k] + grid[i+1][k] + grid[i+2][k]
			if tmp != sum {
				return false
			}
		}
		tmp := grid[i][j] + grid[i+1][j+1] + grid[i+2][j+2]
		if tmp != sum {
			return false
		}
		tmp = grid[i+2][j] + grid[i+1][j+1] + grid[i][j+2]
		if tmp != sum {
			return false
		}
		return true
	}
	ans := 0
	for i := 0; i < m-2; i++ {
		for j := 0; j < n-2; j++ {
			if check(i, j) {
				ans++
			}
		}
	}
	return ans
}
