package graph

func findMaxPathScore(edges [][]int, online []bool, k int64) int {

	if len(edges) == 0 {
		return -1
	}
	n := len(online) - 1
	ad := map[int]map[int]int{}
	q := []edge{}
	for _, e := range edges {
		u, v, c := e[0], e[1], e[2]
		if !online[u] {
			continue
		}
		if !online[v] {
			continue
		}
		if _, ok := ad[u]; !ok {
			ad[u] = map[int]int{}
		}
		ad[u][v] = c
		if u == 0 {
			q = append(q, edge{
				u:     u,
				v:     v,
				total: c,
				cost:  c,
			})
		}
	}
	ans := -1
	for len(q) > 0 {
		e := q[0]
		q = q[1:]
		v, total, cost := e.v, e.total, e.cost
		if total > int(k) {
			continue
		}
		if cost < ans {
			continue
		}
		if v == n {
			ans = max(ans, cost)
			continue
		}
		for k, c := range ad[v] {
			q = append(q, edge{
				u:     v,
				v:     k,
				total: total + c,
				cost:  min(cost, c),
			})
		}
	}
	return -1
}

type edge struct {
	u     int
	v     int
	total int
	cost  int
}
