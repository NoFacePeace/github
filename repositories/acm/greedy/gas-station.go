package greedy

func canCompleteCircuit(gas []int, cost []int) int {
	n := len(gas)
	if n == 0 {
		return -1
	}
	sum := 0
	i := 0
	start := 0
	dist := func(a, b int) int {
		if a <= b {
			return b - a + 1
		}
		return n - a + b + 1
	}
	for {
		sum += gas[i]
		sum -= cost[i]
		if sum < 0 {
			sum = 0
			i++
			i %= n
			if i <= start {
				break
			}
			start = i
			continue
		}
		if dist(start, i) == n {
			return start
		}
		i++
		i %= n
	}
	return -1
}
