package sort

import "sort"

func maximumBeauty(items [][]int, queries []int) []int {
	sort.Slice(items, func(a, b int) bool {
		return items[a][0] < items[b][0]
	})
	arr := [][]int{}
	for _, v := range items {
		if len(arr) == 0 {
			arr = append(arr, v)
			continue
		}
		n := len(arr)
		item := arr[n-1]
		if item[1] >= v[1] {
			continue
		}
		if item[0] == v[0] {
			arr = arr[:n-1]
		}
		arr = append(arr, v)
	}
	ans := []int{}
	for _, q := range queries {
		a := 0
		for _, v := range arr {
			if v[0] > q {
				break
			}
			a = v[1]
		}
		ans = append(ans, a)
	}
	return ans
}
