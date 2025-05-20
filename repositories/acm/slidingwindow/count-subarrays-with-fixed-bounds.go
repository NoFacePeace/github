package slidingwindow

// https://leetcode.cn/problems/count-subarrays-with-fixed-bounds/?envType=daily-question&envId=2025-04-26

func countSubarrays(nums []int, minK int, maxK int) int64 {
	ans := 0
	border := -1
	mn := -1
	mx := -1
	n := len(nums)
	for i := 0; i < n; i++ {
		num := nums[i]
		if num < minK || num > maxK {
			mx = -1
			mn = -1
			border = i
		}
		if num == minK {
			mn = i
		}
		if num == maxK {
			mx = i
		}
		if mn != -1 && mx != -1 {
			ans += min(mn, mx) - border
		}
	}
	return int64(ans)
}
