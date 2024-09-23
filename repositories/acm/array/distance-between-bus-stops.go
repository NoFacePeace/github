package array

func distanceBetweenBusStops(distance []int, start int, destination int) int {
	n := len(distance)
	cw := 0
	for i := start; i != destination; {
		cw += distance[i]
		i++
		i %= n
	}
	ccw := 0
	for i := destination; i != start; {
		ccw += distance[i]
		i++
		i %= n
	}
	return min(ccw, cw)
}
