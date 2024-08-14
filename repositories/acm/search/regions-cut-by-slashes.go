package search

func regionsBySlashes(grid []string) int {
	n := len(grid)
	arr := make([][]int, n*3)
	for i := 0; i < 3*n; i++ {
		arr[i] = make([]int, n*3)
	}
	left := func(x, y int) {
		for i := x; i < x+3; i++ {
			for j := y; j < y+3; j++ {
				if (i-x)+(j-y) == 2 {
					arr[i][j] = 1
				}
			}
		}
	}
	right := func(x, y int) {
		for i := x; i < x+3; i++ {
			for j := y; j < y+3; j++ {
				if (i-x)+(j-y) == 0 {
					arr[i][j] = 1
				}
				if (i-x)+(j-y) == 4 {
					arr[i][j] = 1
				}
				if (i-x) == 1 && (j-y) == 1 {
					arr[i][j] = 1
				}
			}
		}
	}
	var zero func(x, y int)
	zero = func(x, y int) {
		if x < 0 {
			return
		}
		if y < 0 {
			return
		}
		if x >= 3*n {
			return
		}
		if y >= 3*n {
			return
		}
		if arr[x][y] == 1 {
			return
		}
		arr[x][y] = 1
		zero(x, y-1)
		zero(x, y+1)
		zero(x-1, y)
		zero(x+1, y)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			switch grid[i][j] {
			case '/':
				left(i*3, j*3)
			case '\\':
				right(i*3, j*3)
			}
		}
	}
	cnt := 0
	for i := 0; i < 3*n; i++ {
		for j := 0; j < 3*n; j++ {
			if arr[i][j] == 0 {
				zero(i, j)
				cnt++
			}
		}
	}
	return cnt
}
