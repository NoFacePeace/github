package graph

import (
	"container/heap"
	"math"
)

func networkDelayTime(times [][]int, n int, k int) int {
	type edge struct {
		to   int
		time int
	}
	g := make([][]edge, n)
	for _, t := range times {
		x, y := t[0]-1, t[1]-1
		g[x] = append(g[x], edge{y, t[2]})
	}
	const inf int = math.MaxInt64 / 2
	dist := make([]int, n)
	for i := range dist {
		dist[i] = inf
	}
	dist[k-1] = 0
	h := &hp{{0, k - 1}}
	for h.Len() > 0 {
		p := heap.Pop(h).(pair)
		x := p.x
		if dist[x] < p.d {
			continue
		}
		for _, e := range g[x] {
			y := e.to
			if d := dist[x] + e.time; d < dist[y] {
				dist[y] = d
				heap.Push(h, pair{d, y})
			}
		}
	}
	ans := 0
	for _, d := range dist {
		if d == inf {
			return -1
		}
		ans = max(ans, d)
	}
	return ans
}

type pair struct {
	d int
	x int
}

type hp []pair

func (h hp) Len() int {
	return len(h)
}

func (h hp) Less(i, j int) bool {
	return h[i].d < h[j].d
}

func (h hp) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *hp) Push(v interface{}) {
	*h = append(*h, v.(pair))
}

func (h *hp) Pop() (v interface{}) {
	a := *h
	*h, v = a[:len(a)-1], a[len(a)-1]
	return
}
