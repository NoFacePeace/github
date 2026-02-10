package heap

import "container/heap"

func minCost(n int, edges [][]int) int {
	h := &minCostHeap{}
	heap.Init(h)
	adj := make([][][]int, n)
	for _, edge := range edges {
		u, v, w := edge[0], edge[1], edge[2]
		adj[u] = append(adj[u], []int{v, w})
		adj[v] = append(adj[v], []int{u, 2 * w})
	}
	for _, v := range adj[0] {
		heap.Push(h, []int{0, v[0], v[1]})
	}
	visited := make([]bool, n)
	visited[0] = true
	g := make([]int, n)
	for h.Len() > 0 {
		x := heap.Pop(h)
		edge := x.([]int)
		_, v, w := edge[0], edge[1], edge[2]
		if visited[v] {
			continue
		}
		visited[v] = true
		if g[v] == 0 {
			g[v] = w
		} else {
			g[v] = min(g[v], w)
		}
		if v == n-1 {
			break
		}
		for _, u := range adj[v] {
			if visited[u[0]] {
				continue
			}
			heap.Push(h, []int{v, u[0], w + u[1]})
		}
	}
	if g[n-1] == 0 {
		return -1
	}
	return g[n-1]
}

type minCostHeap [][]int

func (h minCostHeap) Len() int {
	return len(h)
}

func (h minCostHeap) Less(i, j int) bool {
	return h[i][2] < h[j][2]
}

func (h minCostHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minCostHeap) Push(x any) {
	*h = append(*h, x.([]int))
}

func (h *minCostHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
