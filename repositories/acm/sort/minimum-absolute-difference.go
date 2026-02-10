package sort

import "sort"

func minimumAbsDifference(arr []int) [][]int {
	sort.Ints(arr)
	n := len(arr)
	mn := arr[1] - arr[0]
	for i := 1; i < n; i++ {
		mn = min(mn, arr[i]-arr[i-1])
	}
	ans := [][]int{}
	for i := 1; i < n; i++ {
		if arr[i]-arr[i-1] == mn {
			ans = append(ans, []int{arr[i-1], arr[i]})
		}
	}
	return ans
}
