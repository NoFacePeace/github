package math

func minTimeToVisitAllPoints(points [][]int) int {
	ans := 0
	n := len(points)
	abs := func(num int) int {
		if num < 0 {
			return -num
		}
		return num
	}
	for i := 0; i < n; i++ {
		if i == 0 {
			continue
		}
		x, y := points[i-1][0], points[i-1][1]
		x1, y1 := points[i][0], points[i][1]
		mn := min(abs(x1-x), abs(y1-y))
		ans += mn
		ans += abs(x1-x) - mn
		ans += abs(y1-y) - mn
	}
	return ans
}
