package sort

func hIndex(citations []int) int {
	l := len(citations)
	counter := make([]int, l+1)
	for _, v := range citations {
		if v >= l {
			v = l
		}
		counter[v]++
	}
	for i := l; i > 0; i-- {
		if counter[i] >= i {
			return i
		}
		counter[i-1] += counter[i]
	}
	return 0
}
