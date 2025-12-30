package hash

import (
	"sort"
)

func findAllPeople(n int, meetings [][]int, firstPerson int) []int {
	time := map[int][][]int{}
	ts := []int{}
	for _, v := range meetings {
		t := v[2]
		if _, ok := time[t]; !ok {
			time[t] = [][]int{}
			ts = append(ts, t)
		}
		time[t] = append(time[t], v)
	}
	s := map[int]bool{}
	s[0] = true
	s[firstPerson] = true
	sort.Ints(ts)
	for _, v := range ts {
		if _, ok := time[v]; !ok {
			continue
		}
		sort.Slice(time[v], func(a, b int) bool {
			return s[time[v][a][0]] || s[time[v][a][1]]
		})
		for _, v := range time[v] {
			if s[v[0]] || s[v[1]] {
				s[v[0]] = true
				s[v[1]] = true
			}
		}
	}
	ans := []int{}
	for k := range s {
		ans = append(ans, k)
	}
	return ans
}
