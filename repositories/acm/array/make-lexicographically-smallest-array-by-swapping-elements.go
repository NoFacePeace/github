package array

import "sort"

func lexicographicallySmallestArray(nums []int, limit int) []int {
	arr := []node{}
	for k, v := range nums {
		arr = append(arr, node{
			num: v,
			pos: k,
		})
	}
	sort.Slice(arr, func(a, b int) bool {
		return arr[a].num < arr[b].num
	})
	n := len(arr)
	if n == 0 {
		return nums
	}
	tmp := []int{arr[0].pos}
	for i := 1; i < n; i++ {
		if arr[i].num-arr[i-1].num > limit {
			if len(tmp) == 1 {
				tmp = []int{arr[i].pos}
				continue
			}
			sort.Ints(tmp)
			for k, v := range tmp {
				nums[v] = arr[i-len(tmp)+k].num
			}
			continue
		}
		tmp = append(tmp, arr[i].pos)
	}
	if len(tmp) != 1 {
		sort.Ints(tmp)
		for k, v := range tmp {
			nums[v] = arr[n-len(tmp)+k].num
		}
	}
	return nums
}

type node struct {
	num int
	pos int
}
