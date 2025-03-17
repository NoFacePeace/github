package array

// https://leetcode.cn/problems/maximum-distance-in-arrays/description/?envType=daily-question&envId=2025-02-19

func maxDistance(arrays [][]int) int {
	m := len(arrays)
	mn, mx := 0, 0
	ans := 0
	for i := 0; i < m; i++ {
		n := len(arrays[i])
		first := arrays[i][0]
		last := arrays[i][n-1]
		if i == 0 {
			mn = first
			mx = last
			continue
		}
		ans = max(ans, mx-first)
		ans = max(ans, last-mn)
		mx = max(mx, last)
		mn = min(mn, first)
	}
	return ans
}
