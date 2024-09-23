package slidingwindow

func minSubArrayLen(target int, nums []int) int {
	n := len(nums)
	for i := 1; i < n; i++ {
		nums[i] += nums[i-1]
	}
	left, right := 0, 0
	min := n + 1
	for right < n {
		sum := 0
		if left == 0 {
			sum = nums[right]
		} else {
			sum = nums[right] - nums[left-1]
		}
		if sum >= target {
			if right-left+1 < min {
				min = right - left + 1
			}
			left++
			if right < left {
				right++
			}
		} else {
			right++
		}
	}
	if min == n+1 {
		return 0
	}
	return min
}
