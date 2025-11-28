package array

func kLengthApart(nums []int, k int) bool {
	pos := []int{}
	for k, v := range nums {
		if v == 1 {
			pos = append(pos, k)
		}
	}
	n := len(pos)
	for i := 0; i < n-1; i++ {
		dist := pos[i+1] - pos[i]
		if dist <= k {
			return false
		}
	}
	return true
}
