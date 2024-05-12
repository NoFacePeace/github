package constanttime

func average(salary []int) float64 {
	n := len(salary)
	if n == 0 {
		return 0
	}
	min, max := salary[0], salary[0]
	sum := 0
	for _, v := range salary {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
		sum += v
	}
	sum -= max
	sum -= min
	return float64(sum) / float64((n - 2))
}
