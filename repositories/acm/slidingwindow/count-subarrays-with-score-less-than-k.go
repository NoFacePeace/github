package slidingwindow

// https://leetcode.cn/problems/count-subarrays-with-score-less-than-k/?envType=daily-question&envId=2025-04-28

func countSubarraysK(nums []int, k int64) int64 {
	n := len(nums)
	prefix := make([]int, n+1)
	for i := 0; i < n; i++ {
		prefix[i+1] += prefix[i] + nums[i]
	}
	l := 0
	ans := 0
	for i := 1; i <= n; i++ {
		for l <= i {
			if (prefix[i]-prefix[l])*(i-l) < int(k) {
				break
			}
			l++
		}
		if l <= i {
			ans += i - l
		}
	}
	return int64(ans)
}
