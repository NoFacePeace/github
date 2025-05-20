package array

// https://leetcode.cn/problems/count-the-hidden-sequences/?envType=daily-question&envId=2025-04-21
func numberOfArrays(differences []int, lower int, upper int) int {
	mx := 0
	mn := 0
	sum := 0
	for _, v := range differences {
		sum += v
		mx = max(mx, sum)
		mn = min(mn, sum)
	}
	if mn < 0 {
		mx = mx - mn
	}
	if upper-mx < lower {
		return 0
	}
	return upper - mx - lower + 1
}
