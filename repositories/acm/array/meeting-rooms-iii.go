package array

import (
	"container/heap"
	"sort"
)

func mostBooked(n int, meetings [][]int) int {
	sort.Slice(meetings, func(a, b int) bool {
		return meetings[a][0] < meetings[b][0]
	})
	queue := mostBookedQueue{}
	for i := 0; i < n; i++ {
		queue = append(queue, &mostBookedItem{
			index: i,
			end:   0,
		})
	}
	heap.Init(&queue)
	m := len(meetings)
	used := make([]int, n)
	i := 0
	for i < m {
		meeting := meetings[i]
		start, end := meeting[0], meeting[1]
		item := heap.Pop(&queue).(*mostBookedItem)
		idx, t := item.index, item.end
		if t < start {
			item.end = start
		} else {
			item.end = item.end + end - start
			used[idx]++
			i++
		}
		heap.Push(&queue, item)
	}
	ans := 0
	mx := used[0]
	for i := 0; i < n; i++ {
		if used[i] > mx {
			ans = i
			mx = used[i]
		}
	}
	return ans
}

type mostBookedItem struct {
	index int
	end   int
}

type mostBookedQueue []*mostBookedItem

func (m mostBookedQueue) Len() int {
	return len(m)
}

func (m mostBookedQueue) Less(i, j int) bool {
	if m[i].end == m[j].end {
		return m[i].index < m[j].index
	}
	return m[i].end < m[j].end
}

func (m mostBookedQueue) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *mostBookedQueue) Push(x any) {
	item := x.(*mostBookedItem)
	*m = append(*m, item)
}

func (m *mostBookedQueue) Pop() any {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[:n-1]
	return x
}
