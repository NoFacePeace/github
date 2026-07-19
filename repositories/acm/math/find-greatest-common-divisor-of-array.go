package math

func findGCD(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 1
	}
	mn := nums[0]
	mx := nums[0]
	for _, v := range nums {
		mn = min(v, mn)
		mx = max(v, mx)
	}
	var f func(a, b int) int
	f = func(a, b int) int {
		if a < b {
			a, b = b, a
		}
		mod := a % b
		if mod == 0 {
			return b
		}
		return f(mod, b)
	}
	return f(mn, mx)
}
