package math

func divisorSubstrings(num int, k int) int {
	mod := 1
	for i := 0; i < k; i++ {
		mod *= 10
	}
	tmp := num
	ans := 0
	for tmp >= mod/10 {
		div := tmp % mod
		if div != 0 && num%div == 0 {
			ans++
		}
		tmp /= 10
	}
	return ans
}
