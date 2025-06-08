package sort

import "sort"

func candy(ratings []int) int {
	n := len(ratings)
	pos := make([]int, n)
	for i := 0; i < n; i++ {
		pos[i] = i
	}
	sort.Slice(pos, func(a, b int) bool {
		return ratings[pos[a]] < ratings[pos[b]]
	})
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = 1
	}
	for i := 0; i < n; i++ {
		p := pos[i]
		if p != 0 && ratings[p-1] > ratings[p] {
			arr[p-1] = max(arr[p-1], arr[p]+1)
		}
		if p != n-1 && ratings[p+1] > ratings[p] {
			arr[p+1] = max(arr[p+1], arr[p]+1)
		}
	}
	ans := 0
	for _, v := range arr {
		ans += v
	}
	return ans
}
