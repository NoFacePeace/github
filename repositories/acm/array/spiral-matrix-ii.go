package array

// https://leetcode.cn/problems/spiral-matrix-ii/

func generateMatrix(n int) [][]int {
	ans := make([][]int, n)
	for i := 0; i < n; i++ {
		ans[i] = make([]int, n)
	}
	cnt := 1
	sx, sy := 0, 0
	ex, ey := n-1, n-1
	for sx <= ex && sy <= ey {
		for i := sy; i <= ey; i++ {
			ans[sx][i] = cnt
			cnt++
		}
		for i := sx + 1; i <= ex; i++ {
			ans[i][ey] = cnt
			cnt++
		}
		for i := ey - 1; i >= sy; i-- {
			ans[ex][i] = cnt
			cnt++
		}
		for i := ex - 1; i > sx; i-- {
			ans[i][sy] = cnt
			cnt++
		}
		sx++
		sy++
		ex--
		ey--
	}
	return ans
}
