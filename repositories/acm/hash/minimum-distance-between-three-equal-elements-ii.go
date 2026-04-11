package hash

import "math"

func minimumDistance(nums []int) int {
	m := map[int][]int{}
	n := len(nums)
	for i := 0; i < n; i++ {
		num := nums[i]
		if m[num] == nil {
			m[num] = []int{}
		}
		m[num] = append(m[num], i)
	}
	ans := math.MaxInt
	for _, arr := range m {
		for i := 0; i+2 < len(arr); i++ {
			ans = min(ans, (arr[i+2]-arr[i])*2)
		}
	}
	if ans == math.MaxInt {
		return -1
	}
	return ans
}
