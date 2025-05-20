package search

import "bytes"

func pushDominoes(dominoes string) string {
	n := len(dominoes)
	q := []int{}
	time := make([]int, n)
	for i := range time {
		time[i] = -1
	}
	force := make([][]byte, n)
	for i, c := range dominoes {
		if c != '.' {
			q = append(q, i)
			time[i] = 0
			force[i] = append(force[i], byte(c))
		}
	}
	ans := bytes.Repeat([]byte{'.'}, n)
	for len(q) > 0 {
		i := q[0]
		q = q[1:]
		if len(force[i]) > 1 {
			continue
		}
		f := force[i][0]
		ans[i] = f
		ni := i - 1
		if f == 'R' {
			ni = i + 1
		}
		if 0 <= ni && ni < n {
			t := time[i]
			if time[ni] == -1 {
				q = append(q, ni)
				time[ni] = t + 1
				force[ni] = append(force[ni], f)
			} else if time[ni] == t+1 {
				force[ni] = append(force[ni], f)
			}
		}
	}
	return string(ans)
}
