package bitwise

func longestSubsequence(nums []int) int {
	ans := 0
	m := map[int]int{}
	for _, v := range nums {
		cnt := 0
		for v != 0 {
			if v&1 == 1 {
				m[cnt]++
			}
			cnt++
			v = v >> 1
		}
	}
	n := len(nums)
	for _, v := range m {
		if v == 0 {
			continue
		}
		if v%2 == 0 {
			ans = max(ans, n-1)
		} else {
			ans = max(ans, n)
		}
	}
	return ans
}
