package array

func isValidSudoku(board [][]byte) bool {
	rows := [10][10]bool{}
	columns := [10][10]bool{}
	for i := 0; i < 9; i += 3 {
		for j := 0; j < 9; j += 3 {
			arr := make([]bool, 10)
			for x := i; x < i+3; x++ {
				for y := j; y < j+3; y++ {
					if board[x][y] == '.' {
						continue
					}
					b := board[x][y]
					num := int(b - '0')
					if rows[x][num] {
						return false
					}
					if columns[y][num] {
						return false
					}
					if arr[num] {
						return false
					}
					rows[x][num] = true
					columns[y][num] = true
					arr[num] = true
				}
			}
		}
	}
	return true
}
