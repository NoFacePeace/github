package array

func rotateI(matrix [][]int) {
	n := len(matrix)
	for i := 0; i < n/2; i++ {
		x, y := i, i
		for j := 0; y+j < n-i-1; j++ {
			one := matrix[x][y+j]
			two := matrix[x+j][n-1-i]
			three := matrix[n-1-i][n-1-i-j]
			four := matrix[n-1-i-j][y]
			matrix[x][y+j] = four
			matrix[x+j][n-1-i] = one
			matrix[n-1-i][n-1-i-j] = two
			matrix[n-1-i-j][y] = three
		}
	}
}
