package sort

func minimumTime(time []int, totalTrips int) int64 {
	l := 1
	r := time[0] * totalTrips
	for l < r {
		mid := (l + r) / 2
		cnt := 0
		for _, v := range time {
			cnt += mid / v
		}
		if cnt < totalTrips {
			l = mid + 1
		} else {
			r = mid
		}
	}
	return int64(l)
}
