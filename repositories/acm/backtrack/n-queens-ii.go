package backtrack

func totalNQueens(n int) int {
	ans := [][]string{}
	strs := []string{}
	column := ""
	for i := 0; i < n; i++ {
		column += "."
	}
	check := func(x, y int) bool {
		for i := 0; i < x; i++ {
			if strs[i][y] == 'Q' {
				return true
			}
		}
		for i := 1; i <= min(x, y); i++ {
			if strs[x-i][y-i] == 'Q' {
				return true
			}
		}
		for i := 1; i <= min(x, n-y-1); i++ {
			if strs[x-i][y+i] == 'Q' {
				return true
			}
		}
		return false
	}
	var dfs func(row int)
	dfs = func(row int) {
		if row == n {
			tmp := append([]string{}, strs...)
			ans = append(ans, tmp)
			return
		}
		str := []byte(column)
		for i := 0; i < n; i++ {
			if check(row, i) {
				continue
			}
			str[i] = 'Q'
			strs = append(strs, string(str))
			dfs(row + 1)
			str[i] = '.'
			strs = strs[:len(strs)-1]
		}
	}
	dfs(0)
	return len(ans)
}
