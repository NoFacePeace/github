package array

func checkTwoChessboards(coordinate1 string, coordinate2 string) bool {
	board := make([][]int, 8)
	for i := 0; i < 8; i++ {
		board[i] = make([]int, 8)
		for j := 0; j < 8; j++ {
			if i%2 == 0 {
				if j%2 == 0 {
					board[i][j] = 1
				}
			} else {
				if j%2 != 0 {
					board[i][j] = 1
				}
			}
		}
	}
	x1 := int(coordinate1[0] - 'a')
	y1 := int(coordinate1[1] - '1')
	x2 := int(coordinate2[0] - 'a')
	y2 := int(coordinate2[1] - '1')
	return board[x1][y1] == board[x2][y2]
}
