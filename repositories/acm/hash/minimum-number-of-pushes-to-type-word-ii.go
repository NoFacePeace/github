package hash

import "sort"

func minimumPushes(word string) int {
	m := map[string]int{}
	n := len(word)
	for i := 0; i < n; i++ {
		m[word[i:i+1]]++
	}
	arr := []int{}
	for _, v := range m {
		arr = append(arr, v)
	}
	sort.Slice(arr, func(a, b int) bool {
		return arr[a] > arr[b]
	})
	cnt := 2
	idx := 1
	sum := 0
	for _, v := range arr {
		sum += idx * v
		cnt++
		if cnt == 10 {
			cnt = 2
			idx++
		}
	}
	return sum
}
