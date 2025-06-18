package greedy

import "sort"

func divideArray(nums []int, k int) [][]int {
	sort.Ints(nums)
	n := len(nums)
	if n == 0 {
		return nil
	}
	ans := [][]int{}
	for i := 0; i < n-2; i += 3 {
		one := nums[i]
		two := nums[i+1]
		three := nums[i+2]
		if three-one <= k {
			ans = append(ans, []int{one, two, three})
			continue
		}
		return nil
	}

	return ans
}
