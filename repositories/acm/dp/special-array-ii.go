package dp

func isArraySpecial(nums []int, queries [][]int) []bool {
	n := len(nums)
	arr := make([]int, n)
	start := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			continue
		}
		val := nums[i] + nums[i-1]
		if val%2 == 0 {
			start = i
		}
		arr[i] = start
	}
	ans := []bool{}
	for _, v := range queries {
		l, r := v[0], v[1]
		idx := arr[r]
		if l >= idx {
			ans = append(ans, true)
		} else {
			ans = append(ans, false)
		}
	}
	return ans
}
