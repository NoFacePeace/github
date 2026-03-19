package array

import "container/heap"

func minNumberOfSeconds(mountainHeight int, workerTimes []int) int64 {
	n := len(workerTimes)
	h := &workerHeap{}
	for i := 0; i < n; i++ {
		heap.Push(h, worker{time: workerTimes[i], height: 1, total: workerTimes[i]})
	}
	cnt := 0
	ans := 0
	for cnt < mountainHeight {
		worker := heap.Pop(h).(worker)
		ans = worker.total
		cnt++
		worker.height++
		worker.total += worker.time * worker.height
		heap.Push(h, worker)
	}
	return int64(ans)
}

type worker struct {
	time   int
	height int
	total  int
}

type workerHeap []worker

func (h workerHeap) Len() int {
	return len(h)
}

func (h workerHeap) Less(i, j int) bool {
	return h[i].total < h[j].total
}

func (h workerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *workerHeap) Push(x any) {
	*h = append(*h, x.(worker))
}

func (h *workerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
