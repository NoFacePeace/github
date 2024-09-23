package dp

import "math"

func minimumSubstringsInPartition(s string) int {
	n := len(s)
	d := make([]int, n+1)
	for i := range d {
		d[i] = math.MaxInt
	}
	d[0] = 0
	for i := 1; i <= n; i++ {
		mx := 0
		m := map[byte]int{}
		for j := i; j >= 1; j-- {
			m[s[j-1]]++
			if m[s[j-1]] > mx {
				mx = m[s[j-1]]
			}
			if mx*len(m) == (i-j+1) && d[j-1] != math.MaxInt {
				if d[i] > d[j-1]+1 {
					d[i] = d[j-1] + 1
				}
			}
		}
	}
	return d[n]
}
