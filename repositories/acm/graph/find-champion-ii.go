package graph

// https://leetcode.cn/problems/find-champion-ii/
func findChampion(n int, edges [][]int) int {
	win := map[int]bool{}
	for _, v := range edges {
		v2 := v[1]
		win[v2] = true
	}
	for i := 0; i < n; i++ {
		if win[i] {
			continue
		}
		if len(win) == n-1 {
			return i
		}
	}
	return -1
}
