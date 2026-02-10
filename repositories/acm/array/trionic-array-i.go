package array

func isTrionic(nums []int) bool {
	l := 0
	n := len(nums)
	for l+1 < n {
		if nums[l+1] <= nums[l] {
			break
		}
		l++
	}
	if l == 0 {
		return false
	}
	r := n - 1
	for r-1 >= 0 {
		if nums[r-1] >= nums[r] {
			break
		}
		r--
	}
	if r == n-1 {
		return false
	}
	if l >= r {
		return false
	}
	for i := l; i < r; i++ {
		if nums[i] <= nums[i+1] {
			return false
		}
	}
	return true
}
