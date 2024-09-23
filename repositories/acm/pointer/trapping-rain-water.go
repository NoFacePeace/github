package pointer

func trap(height []int) int {
	l, r := 0, len(height)-1
	lm, rm := 0, 0
	ans := 0
	for l < r {
		lm = max(lm, height[l])
		rm = max(rm, height[r])
		if height[l] < height[r] {
			ans += lm - height[l]
			l++
		} else {
			ans += rm - height[r]
			r--
		}
	}
	return ans
}
