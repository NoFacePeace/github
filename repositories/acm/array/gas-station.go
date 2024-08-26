package array

func canCompleteCircuit(gas []int, cost []int) int {
	n := len(gas)
	dist := make([]int, n)
	for i := 0; i < n; i++ {
		dist[i] = gas[i] - cost[i]
	}
	dist = append(dist, dist...)
	l, r := 0, 0
	sum := 0
	for r < 2*n {
		sum += dist[r]
		if sum < 0 {
			sum = 0
			r++
			l = r
			continue
		}
		if r-l+1 == n {
			return l
		}
		r++
	}
	return -1
}
