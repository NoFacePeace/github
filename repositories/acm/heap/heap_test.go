package heap

import "testing"

func Test_heap(t *testing.T) {
	heap := heap{}
	for i := 0; i < 10; i++ {
		heap.Push(i)
	}
	for i := 0; i < 10; i++ {
		if heap.Pop() != 10-i-1 {
			t.Error("Pop incorrect")
		}
	}
}
