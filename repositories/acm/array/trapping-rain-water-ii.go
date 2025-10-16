package array

func trapRainWater(heightMap [][]int) int {
	m := len(heightMap)
	if m == 0 {
		return 0
	}
	n := len(heightMap[0])
	if n == 0 {
		return 0
	}
	arr := make([][][4]int, m)
	for i := range arr {
		arr[i] = make([][4]int, n)
	}
	for i := 0; i < m; i++ {
		mx := 0
		for j := 0; j < n; j++ {
			if j == 0 {
				mx = heightMap[i][j]
				continue
			}
			arr[i][j][0] = mx
			mx = max(mx, heightMap[i][j])
		}
	}
	for i := 0; i < m; i++ {
		mx := 0
		for j := n - 1; j >= 0; j-- {
			if j == n-1 {
				mx = heightMap[i][j]
				continue
			}
			arr[i][j][1] = mx
			mx = max(mx, heightMap[i][j])
		}
	}
	for j := 0; j < n; j++ {
		mx := 0
		for i := 0; i < m; i++ {
			if i == 0 {
				mx = heightMap[i][j]
				continue
			}
			arr[i][j][2] = mx
			mx = max(mx, heightMap[i][j])
		}
	}
	for j := 0; j < n; j++ {
		mx := 0
		for i := m - 1; i >= 0; i-- {
			if i == m-1 {
				mx = heightMap[i][j]
				continue
			}
			arr[i][j][3] = mx
			mx = max(mx, heightMap[i][j])
		}
	}
	bucket := make([][]int, m)
	for i := range bucket {
		bucket[i] = make([]int, n)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			mn := min(arr[i][j][0], arr[i][j][1], arr[i][j][2], arr[i][j][3])
			if mn-heightMap[i][j] > 0 {
				bucket[i][j] = mn
			}
		}
	}
	mn := 0
	mp := map[int]bool{}
	var dfs func(i, j int)
	dfs = func(i, j int) {
		if i < 0 {
			return
		}
		if i >= m {
			return
		}
		if j < 0 {
			return
		}
		if j >= n {
			return
		}
		if bucket[i][j] == 0 {
			return
		}
		if mp[i*n+j] {
			return
		}
		mp[i*n+j] = true
		mn = min(mn, bucket[i][j])
		dfs(i+1, j)
		dfs(i-1, j)
		dfs(i, j+1)
		dfs(i, j-1)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if bucket[i][j] == 0 {
				continue
			}
			mp = map[int]bool{}
			mn = bucket[i][j]
			dfs(i, j)
			bucket[i][j] = mn
		}
	}
	ans := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if bucket[i][j] == 0 {
				continue
			}
			ans += bucket[i][j] - heightMap[i][j]
		}
	}
	return ans
}
