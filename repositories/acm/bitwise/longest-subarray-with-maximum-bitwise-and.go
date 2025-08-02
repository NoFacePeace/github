package bitwise

func longestSubarray(nums []int) int {
	mx := 0
	cnt := 0
	ans := 0
	for _, v := range nums {
		if mx > v {
			cnt = 0
			continue
		}
		if mx == v {
			cnt++
			ans = max(ans, cnt)
			continue
		}
		mx = v
		cnt = 1
		ans = cnt
	}
	return ans
}
