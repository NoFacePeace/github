package heap

import "container/heap"

type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

func New[T any](less func(a, b T) bool) *Heap[T] {
	h := &Heap[T]{
		less: less,
	}
	heap.Init(h)
	return h
}

func (h *Heap[T]) Push(x any) {
	h.data = append(h.data, x.(T))
}

func (h *Heap[T]) Pop() any {
	n := len(h.data)
	x := h.data[n-1]
	h.data = h.data[:n-1]
	return x
}

func (h *Heap[T]) Len() int {
	return len(h.data)
}

func (h *Heap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

func (h *Heap[T]) Swap(i, j int) { h.data[i], h.data[j] = h.data[j], h.data[i] }

func (h *Heap[T]) PushItem(x T) {
	heap.Push(h, x)
}

func (h *Heap[T]) PopItem() T {
	return heap.Pop(h).(T)
}
