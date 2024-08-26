package greedy

func canJump(nums []int) bool {
	l := len(nums)
	if l == 0 {
		return true
	}
	max := 0
	for k, v := range nums {
		if k > max {
			return false
		}
		if k+v > max {
			max = k + v
		}
	}
	return true
}
