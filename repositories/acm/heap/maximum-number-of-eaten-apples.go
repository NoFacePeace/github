package heap

import "container/heap"

func eatenApples(apples []int, days []int) int {
	h := &IntSliceHeap{}
	heap.Init(h)
	n := len(apples)
	sum := 0
	for i := 0; i < n; i++ {
		if apples[i] != 0 {
			heap.Push(h, []int{apples[i], days[i] + i})
		}
		for h.Len() > 0 {
			arr := heap.Pop(h).([]int)
			if arr[1] <= i {
				continue
			}
			arr[0]--
			sum++
			if arr[0] != 0 {
				heap.Push(h, arr)
			}
			break
		}
	}
	i := n
	for h.Len() > 0 {
		arr := heap.Pop(h).([]int)
		if arr[1] <= i {
			continue
		}
		arr[0]--
		sum++
		if arr[0] != 0 {
			heap.Push(h, arr)
		}
		i++
	}
	return sum
}

type IntSliceHeap [][]int

func (h IntSliceHeap) Len() int           { return len(h) }
func (h IntSliceHeap) Less(i, j int) bool { return h[i][1] < h[j][1] }
func (h IntSliceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntSliceHeap) Push(x any) {
	*h = append(*h, x.([]int))
}

func (h *IntSliceHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
