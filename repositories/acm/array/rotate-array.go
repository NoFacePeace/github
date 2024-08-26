package array

func rotate(nums []int, k int) {
	if k == 0 {
		return
	}
	l := len(nums)
	if l <= 1 {
		return
	}
	if l < k {
		k %= l
	}
	for i := 0; i < k; i++ {
		nums[i], nums[l-k+i] = nums[l-k+i], nums[i]
	}
	rotate(nums[k:], k)
}
