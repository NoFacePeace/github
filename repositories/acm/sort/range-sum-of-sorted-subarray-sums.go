package sort

import "sort"

func rangeSum(nums []int, n int, left int, right int) int {
	arr := []int{}
	for i := 1; i <= n; i++ {
		sum := 0
		for j := 0; j < i; j++ {
			sum += nums[j]
		}
		arr = append(arr, sum)
		for j := i; j < n; j++ {
			sum += nums[j]
			sum -= nums[j-i]
			arr = append(arr, sum)
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	sum := 0
	mod := int(1e9 + 7)
	for i := left - 1; i < right; i++ {
		sum += arr[i]
		sum %= mod
	}
	return sum
}
