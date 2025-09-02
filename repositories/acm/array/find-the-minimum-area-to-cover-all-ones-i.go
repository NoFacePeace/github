package array

func minimumArea(grid [][]int) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	top := 0
	for i := 0; i < m; i++ {
		cnt := 0
		for j := 0; j < n; j++ {
			if grid[i][j] == 1 {
				break
			}
			cnt++
		}
		if cnt != n {
			break
		}
		top++
	}
	left := 0
	for i := 0; i < n; i++ {
		cnt := 0
		for j := 0; j < m; j++ {
			if grid[j][i] == 1 {
				break
			}
			cnt++
		}
		if cnt != m {
			break
		}
		left++
	}
	right := n
	for i := n - 1; i >= 0; i-- {
		cnt := 0
		for j := m - 1; j >= 0; j-- {
			if grid[j][i] == 1 {
				break
			}
			cnt++
		}
		if cnt != m {
			break
		}
		right--
	}
	bottom := m
	for i := m - 1; i >= 0; i-- {
		cnt := 0
		for j := n - 1; j >= 0; j-- {
			if grid[i][j] == 1 {
				break
			}
			cnt++
		}
		if cnt != n {
			break
		}
		bottom--
	}
	return min(min(right*bottom, (n-left)*(bottom)), min((n-top)*right, (n-left)*(n-top)))
}
