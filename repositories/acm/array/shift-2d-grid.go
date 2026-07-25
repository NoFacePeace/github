package array

func shiftGrid(grid [][]int, k int) [][]int {
	m := len(grid)
	if m == 0 {
		return grid
	}
	n := len(grid[0])
	if n == 0 {
		return grid
	}
	cnt := 0
	for cnt < k {
		arr := []int{}
		for i := 0; i < m; i++ {
			arr = append(arr, grid[i][n-1])
			for j := n - 1; j > 0; j-- {
				grid[i][j] = grid[i][j-1]
			}
		}
		grid[0][0] = arr[m-1]
		for i := 1; i < m; i++ {
			grid[i][0] = arr[i-1]
		}
		cnt++
	}
	return grid
}
