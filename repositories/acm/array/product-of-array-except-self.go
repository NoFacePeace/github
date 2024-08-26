package array

func productExceptSelf(nums []int) []int {
	n := len(nums)
	left := make([]int, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			left[i] = nums[i]
			continue
		}
		left[i] = left[i-1] * nums[i]
	}
	ans := make([]int, n)
	mult := 0
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			ans[i] = left[i-1]
			mult = nums[i]
			continue
		}
		if i == 0 {
			ans[i] = mult
			continue
		}
		ans[i] = left[i-1] * mult
		mult *= nums[i]
	}
	return ans
}
