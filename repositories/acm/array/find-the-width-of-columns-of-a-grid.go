package array

// https://leetcode.cn/problems/find-the-width-of-columns-of-a-grid/
func findColumnWidth(grid [][]int) []int {
	m := len(grid)
	if m == 0 {
		return nil
	}
	n := len(grid[0])
	if n == 0 {
		return nil
	}
	ans := make([]int, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			num := grid[i][j]
			cnt := 0
			if num <= 0 {
				cnt += 1
			}
			for num != 0 {
				cnt += 1
				num /= 10
			}
			if cnt > ans[j] {
				ans[j] = cnt
			}
		}
	}
	return ans
}
