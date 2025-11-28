package math

func minimumOperations(nums []int) int {
	ans := 0
	for _, v := range nums {
		mod := v % 3
		mod = min(mod, 3-mod)
		ans += mod
	}
	return ans
}
