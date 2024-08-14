package slidingwindow

func minSwaps(nums []int) int {
	n := len(nums)
	cnt := 0
	for _, v := range nums {
		if v == 1 {
			cnt++
		}
	}
	if cnt == 0 {
		return 0
	}
	left := 0
	right := 0
	zero := 0
	for right < cnt {
		if nums[right] == 0 {
			zero++
		}
		right++
	}
	min := n
	if zero < min {
		min = zero
	}
	nums = append(nums, nums...)
	for right < 2*n {
		if nums[right] == 0 {
			zero++
		}
		if nums[left] == 0 {
			zero--
		}
		left++
		right++
		if zero < min {
			min = zero
		}
	}
	return min
}
