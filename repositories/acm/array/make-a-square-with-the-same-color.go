package array

func canMakeSquare(grid [][]byte) bool {
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			m := map[byte]int{
				'B': 0,
				'W': 0,
			}
			m[grid[i][j]]++
			m[grid[i][j+1]]++
			m[grid[i+1][j]]++
			m[grid[i+1][j+1]]++
			if m['B'] >= 3 || m['W'] >= 3 {
				return true
			}
		}
	}
	return false
}
