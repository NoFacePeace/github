package array

import "sort"

func latestTimeCatchTheBus(buses []int, passengers []int, capacity int) int {
	sort.Ints(buses)
	sort.Ints(passengers)
	b := 0
	p := 0
	cnt := 0
	m := map[int]bool{}
	last := cnt
	for b < len(buses) && p < len(passengers) {
		if passengers[p] <= buses[b] {
			m[passengers[p]] = true
			p++
			cnt++
		} else {
			b++
			last = cnt
			cnt = 0
		}
		if cnt == capacity {
			b++
			last = cnt
			cnt = 0
		}
	}
	ans := 0
	if p == 0 {
		return buses[len(buses)-1]
	}
	for i := passengers[p-1]; i >= 0; i-- {
		if _, ok := m[i]; !ok {
			ans = i
			break
		}
	}
	if last == capacity && b == len(buses) {
		return ans
	}
	for i := passengers[p-1]; i <= buses[len(buses)-1]; i++ {
		if _, ok := m[i]; !ok {
			ans = i
		}
	}
	return ans
}
