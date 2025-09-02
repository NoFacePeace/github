package math

func solveSudoku(board [][]byte) {
	rows := [9][9]bool{}
	columns := [9][9]bool{}
	b := [9][9]bool{}
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if board[i][j] == '.' {
				continue
			}
			num := int(board[i][j] - '1')
			rows[i][num] = true
			columns[j][num] = true
			b[i/3*3+j/3][num] = true
		}
	}
	var dfs func(x, y int) bool
	dfs = func(x, y int) bool {
		if board[x][y] == '.' {
			for i := 0; i < 9; i++ {
				if rows[x][i] {
					continue
				}
				if columns[y][i] {
					continue
				}
				if b[x/3*3+y/3][i] {
					continue
				}
				board[x][y] = byte('1' + i)
				rows[x][i] = true
				columns[y][i] = true
				b[x/3*3+y/3][i] = true
				if x == 8 && y == 8 {
					return true
				}
				nx, ny := x, y
				if y == 8 {
					nx = x + 1
					ny = 0
				} else {
					nx = x
					ny = y + 1
				}
				if dfs(nx, ny) {
					return true
				}
				board[x][y] = '.'
				rows[x][i] = false
				columns[y][i] = false
				b[x/3*3+y/3][i] = false
			}
			return false
		}
		if x == 8 && y == 8 {
			return true
		}
		if y == 8 {
			return dfs(x+1, 0)
		}
		return dfs(x, y+1)
	}
	dfs(0, 0)
}
