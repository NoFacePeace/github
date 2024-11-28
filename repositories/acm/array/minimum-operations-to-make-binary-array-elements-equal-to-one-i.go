package array

func minOperations(nums []int) int {
	n := len(nums)
	i := 0
	for i < n {
		if nums[i] != 1 {
			break
		}
		i++
	}
	if i == n {
		return 0
	}
	if i+3 > n {
		return -1
	}
	nums[i] = 1
	if nums[i+1] == 0 {
		nums[i+1] = 1
	} else {
		nums[i+1] = 0
	}
	if nums[i+2] == 0 {
		nums[i+2] = 1
	} else {
		nums[i+2] = 0
	}
	cnt := minOperations(nums[i+1:])
	if cnt == -1 {
		return cnt
	}
	return cnt + 1
}
