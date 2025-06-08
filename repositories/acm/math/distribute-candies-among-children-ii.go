package math

func distributeCandies(n int, limit int) int64 {
	cal := func(x int) int64 {
		if x < 0 {
			return 0
		}
		return int64(x) * int64(x-1) / 2
	}

	return cal(n+2) - 3*cal(n-limit+1) + 3*cal(n-(limit+1)*2+2) - cal(n-3*(limit+1)+2)
}
