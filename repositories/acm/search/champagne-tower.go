package search

func champagneTower(poured int, query_row int, query_glass int) float64 {
	var dfs func(i, j int) float64
	m := map[int]float64{}
	dfs = func(i, j int) float64 {
		if i == 0 {
			return float64(poured)
		}
		if val, ok := m[i*100+j]; ok {
			return val
		}
		val := 0.0
		if j-1 >= 0 {
			top := dfs(i-1, j-1)
			if top > 1 {
				val += (top - 1) / 2
			}
		}
		if j < i {
			top := dfs(i-1, j)
			if top > 1 {
				val += (top - 1) / 2
			}
		}
		m[i*100+j] = val
		return val
	}
	val := dfs(query_row, query_glass)
	if val >= 1 {
		return 1
	}
	return val
}
