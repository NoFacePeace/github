package dp

//  https://leetcode.cn/problems/largest-divisible-subset/?envType=daily-question&envId=2025-04-06

import "sort"

func largestDivisibleSubset(nums []int) []int {
	sort.Ints(nums)
	arr := [][]int{}
	n := len(nums)
	for i := 0; i < n; i++ {
		m := len(arr)
		mx := []int{}
		for j := 0; j < m; j++ {
			l := len(arr[j])
			if nums[i]%arr[j][l-1] != 0 {
				continue
			}
			if len(arr[j]) > len(mx) {
				mx = arr[j]
			}
		}
		tmp := append([]int{}, mx...)
		tmp = append(tmp, nums[i])
		arr = append(arr, tmp)
	}
	ans := []int{}
	for i := range arr {
		if len(arr[i]) > len(ans) {
			ans = arr[i]
		}
	}
	return ans
}
