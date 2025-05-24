package dp

func jump(nums []int) int {
	l := len(nums)
	cnt := 0
	end := 0
	m := 0
	for i := 0; i < l-1; i++ {
		m = maxSlice(m, i+nums[i])
		if i == end {
			end = m
			cnt++
		}
	}
	return cnt
}
