package math

func minSubarray(nums []int, p int) int {
	sum := 0
	n := len(nums)
	ans := n
	for k, v := range nums {
		sum += v
		if sum%p == 0 {
			ans = min(ans, n-k-1)
		}
	}
	pre := 0
	m := map[int]int{}
	for i := 0; i < n-1; i++ {
		pre += nums[i]
		mod := pre % p
		m[mod] = i
		sum -= nums[i]
		mod = sum % p
		if mod == 0 {
			ans = min(ans, i+1)
			if _, ok := m[mod]; !ok {
				continue
			}
			ans = min(ans, i-m[mod])
		}
		if _, ok := m[p-mod]; !ok {
			continue
		}
		ans = min(ans, i-m[p-mod])
	}
	if ans == n {
		return -1
	}
	return ans
}
