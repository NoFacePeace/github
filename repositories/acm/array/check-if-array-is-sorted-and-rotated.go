package array

func check(nums []int) bool {
	n := len(nums)
	i := 0
	for i < n-1 {
		if nums[i] > nums[i+1] {
			break
		}
		i++
	}
	tmp := nums[i+1:]
	tmp = append(tmp, nums[:i+1]...)
	for i := 0; i < n-1; i++ {
		if tmp[i] > tmp[i+1] {
			return false
		}
	}
	return true
}
