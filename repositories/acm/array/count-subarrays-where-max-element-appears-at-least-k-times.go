package array

func countSubarraysK(nums []int, k int) int64 {
	mx := 0
	for _, v := range nums {
		mx = max(mx, v)
	}
	n := len(nums)
	prefix := make([]int, n+1)
	for i := 0; i < n; i++ {
		num := nums[i]
		cnt := 0
		if num == mx {
			cnt++
		}
		prefix[i+1] = prefix[i] + cnt
	}
	ans := 0
	j := 0
	for i := 0; i < n; i++ {
		for j < n {
			if prefix[j+1]-prefix[i] >= k {
				break
			}
			j++
		}
		ans += n - j
	}
	return int64(ans)
}
