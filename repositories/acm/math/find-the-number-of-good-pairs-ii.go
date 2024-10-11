package math

import "math"

func numberOfPairsII(nums1 []int, nums2 []int, k int) int64 {
	mx := math.MinInt
	m := map[int]int{}
	for _, v := range nums1 {
		mx = max(mx, v)
		m[v]++
	}
	for i := range nums2 {
		nums2[i] *= k
	}
	m1 := map[int]int{}
	cnt := 0
	for _, v := range nums2 {
		if val, ok := m1[v]; ok {
			cnt += val
			continue
		}
		i := 1
		c := 0
		for i*v <= mx {
			if val, ok := m[i*v]; ok {
				c += val
			}
			i++
		}
		cnt += c
		m1[v] = c
	}
	return int64(cnt)
}
