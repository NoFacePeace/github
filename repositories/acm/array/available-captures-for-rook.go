package array

func numRookCaptures(board [][]byte) int {
	x, y := -1, -1
	n := 8
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if board[i][j] == 'R' {
				x, y = i, j
				break
			}
		}
		if x != -1 && y != -1 {
			break
		}
	}
	ans := 0
	for i := x + 1; i < n; i++ {
		if board[i][y] == 'B' {
			break
		}
		if board[i][y] == 'p' {
			ans++
			break
		}
	}
	for i := x - 1; i >= 0; i-- {
		if board[i][y] == 'B' {
			break
		}
		if board[i][y] == 'p' {
			ans++
			break
		}
	}
	for i := y + 1; i < n; i++ {
		if board[x][i] == 'B' {
			break
		}
		if board[x][i] == 'p' {
			ans++
			break
		}
	}
	for i := y - 1; i >= 0; i-- {
		if board[x][i] == 'B' {
			break
		}
		if board[x][i] == 'p' {
			ans++
			break
		}
	}
	return ans
}
