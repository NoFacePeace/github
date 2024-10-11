package heap

func minRefuelStops(target int, startFuel int, stations [][]int) int {
	heap := []int{}
	up := func() {
		n := len(heap)
		i := n - 1
		for i > 0 {
			if heap[i] <= heap[(i-1)/2] {
				break
			}
			heap[(i-1)/2], heap[i] = heap[i], heap[(i-1)/2]
			i = (i - 1) / 2
		}
	}
	down := func() {
		i := 0
		n := len(heap)
		for i*2+1 < n {
			mx := i*2 + 1
			if i*2+2 < n && heap[mx] < heap[i*2+2] {
				mx = i*2 + 2
			}
			if heap[i] > heap[mx] {
				break
			}
			heap[mx], heap[i] = heap[i], heap[mx]
			i = mx
		}
	}
	push := func(num int) {
		heap = append(heap, num)
		up()
	}
	pop := func() int {
		num := heap[0]
		n := len(heap)
		heap[0] = heap[n-1]
		heap = heap[:n-1]
		down()
		return num
	}
	cnt := 0
	i := 0
	for startFuel < target {
		for i < len(stations) {
			if stations[i][0] > startFuel {
				break
			}
			push(stations[i][1])
			i++
		}
		if len(heap) == 0 {
			return -1
		}
		num := pop()
		startFuel += num
		cnt++
	}
	return cnt
}
