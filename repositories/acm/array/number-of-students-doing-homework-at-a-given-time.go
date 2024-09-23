package array

func busyStudent(startTime []int, endTime []int, queryTime int) int {
	cnt := 0
	n := len(startTime)
	for i := 0; i < n; i++ {
		start := startTime[i]
		end := endTime[i]
		if start <= queryTime && end >= queryTime {
			cnt++
		}
	}
	return cnt
}
