package array

import (
	"math"
	"sort"
)

func minCost(basket1 []int, basket2 []int) int64 {
	freq := map[int]int{}
	m := math.MaxInt
	abs := func(a int) int {
		if a < 0 {
			return -a
		}
		return a
	}
	for _, b := range basket1 {
		freq[b]++
		m = min(m, b)
	}
	for _, b := range basket2 {
		freq[b]--
		m = min(m, b)
	}
	var merge []int
	for k, c := range freq {
		if c%2 != 0 {
			return -1
		}
		for i := 0; i < abs(c)/2; i++ {
			merge = append(merge, k)
		}
	}
	sort.Ints(merge)
	var res int64
	for i := 0; i < len(merge)/2; i++ {
		if 2*m < merge[i] {
			res += int64(2 * m)
		} else {
			res += int64(merge[i])
		}
	}
	return res
}
