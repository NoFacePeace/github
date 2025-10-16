package array

func pacificAtlantic(heights [][]int) [][]int {
	m := len(heights)
	ans := [][]int{}
	if m == 0 {
		return ans
	}
	n := len(heights[0])
	if n == 0 {
		return ans
	}
	pacific := make([][]bool, m)
	atlantic := make([][]bool, m)
	for i := range pacific {
		pacific[i] = make([]bool, n)
		atlantic[i] = make([]bool, n)
	}
	var dfs func(i, j, h, x, y int, arr [][]bool) bool
	mp := map[int]bool{}
	dfs = func(i, j, h, x, y int, arr [][]bool) bool {
		if i < 0 {
			return false
		}
		if j < 0 {
			return false
		}
		if i >= m {
			return false
		}
		if j >= n {
			return false
		}
		if heights[i][j] > h {
			return false
		}
		if arr[i][j] {
			return true
		}
		if i == x {
			return true
		}
		if j == y {
			return true
		}
		if mp[i*n+j] {
			return false
		}
		mp[i*n+j] = true
		if dfs(i+1, j, h, x, y, arr) {
			return true
		}
		if dfs(i-1, j, h, x, y, arr) {
			return true
		}
		if dfs(i, j+1, h, x, y, arr) {
			return true
		}
		if dfs(i, j-1, h, x, y, arr) {
			return true
		}
		return false
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			mp = map[int]bool{}
			pacific[i][j] = dfs(i, j, heights[i][j], 0, 0, pacific)
			atlantic[i][j] = dfs(i, j, heights[i][j], m-1, n-1, atlantic)
			if pacific[i][j] && atlantic[i][j] {
				ans = append(ans, []int{i, j})
			}
		}
	}
	return ans
}
