package math

func countPermutations(complexity []int) int {
	n := len(complexity)
	if n == 0 {
		return 0
	}
	ans := 1
	mn := complexity[0]
	mod := int(1e9) + 7
	for i := 1; i < n; i++ {
		val := complexity[i]
		if val <= mn {
			return 0
		}
		ans = ans * i
		ans %= mod
	}
	return ans
}
