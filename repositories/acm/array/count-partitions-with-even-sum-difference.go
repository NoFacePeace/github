package array

func countPartitions(nums []int) int {
	n := len(nums)
	for i := 0; i < n; i++ {
		if i == 0 {
			continue
		}
		nums[i] += nums[i-1]
	}
	ans := 0
	for i := 0; i < n-1; i++ {
		if (nums[n-1]-nums[i]-nums[i])%2 == 0 {
			ans++
		}
	}
	return ans
}
