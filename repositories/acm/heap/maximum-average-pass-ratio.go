package heap

import "container/heap"

func maxAverageRatio(classes [][]int, extraStudents int) float64 {
	h := maxAverageRatioHeap(classes)
	heap.Init(&h)
	for i := 0; i < extraStudents; i++ {
		class := heap.Pop(&h).([]int)
		class[0]++
		class[1]++
		heap.Push(&h, class)
	}
	sum := 0.0
	for i := 0; i < len(h); i++ {
		sum += float64(h[i][0]) / float64(h[i][1])
	}
	return sum / float64(len(classes))
}

type maxAverageRatioHeap [][]int

func (h maxAverageRatioHeap) Len() int { return len(h) }
func (h maxAverageRatioHeap) Less(i, j int) bool {
	return maxAverageRatioCompute(h[i][0]+1, h[i][1]+1)-maxAverageRatioCompute(h[i][0], h[i][1]) > maxAverageRatioCompute(h[j][0]+1, h[j][1]+1)-maxAverageRatioCompute(h[j][0], h[j][1])
}
func (h maxAverageRatioHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *maxAverageRatioHeap) Push(x any) {
	*h = append(*h, x.([]int))
}

func (h *maxAverageRatioHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func maxAverageRatioCompute(i, j int) float64 {
	return float64(i) / float64(j)
}
