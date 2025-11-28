package prefixsum

func countValidSelections(nums []int) int {
	n := len(nums)
	right := make([]int, n)
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			right[i] = nums[i]
			continue
		}
		right[i] = right[i+1] + nums[i]
	}
	cnt := 0
	ans := 0
	for i := 0; i < n; i++ {
		if nums[i] != 0 {
			cnt += nums[i]
			continue
		}
		if cnt == right[i] {
			ans += 2
			continue
		}
		if cnt+1 == right[i] || cnt-1 == right[i] {
			ans += 1
		}
	}
	return ans
}
