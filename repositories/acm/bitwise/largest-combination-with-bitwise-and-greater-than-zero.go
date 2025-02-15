package bitwise

func largestCombination(candidates []int) int {
	ans := 0
	bit := 1
	for i := 0; i < 32; i++ {
		cnt := 0
		for _, v := range candidates {
			if v&bit > 0 {
				cnt++
			}
		}
		ans = max(ans, cnt)
		bit *= 2
	}
	return ans
}
