package array

func maxDistanceI(colors []int) int {
	n := len(colors)
	ans := 0
	for i := 0; i < n; i++ {
		if colors[i] != colors[0] {
			ans = max(ans, i)
		}
		if colors[i] != colors[n-1] {
			ans = max(ans, n-i-1)
		}
	}
	return ans
}
