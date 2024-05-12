package slidingwindow

func decrypt(code []int, k int) []int {
	n := len(code)
	ans := make([]int, n)
	if k == 0 {
		return ans
	}
	mod := k
	if k < 0 {
		for i := 0; i < n/2; i++ {
			code[i], code[n-1-i] = code[n-1-i], code[i]
		}
		mod = -k
	}
	sum := 0
	for i := 0; i < mod; i++ {
		sum += code[i%n]
	}
	for i := 0; i < n; i++ {
		sum -= code[i]
		sum += code[(i+mod)%n]
		ans[i] = sum
	}
	if k < 0 {
		for i := 0; i < n/2; i++ {
			ans[i], ans[n-1-i] = ans[n-1-i], ans[i]
		}
	}
	return ans
}
