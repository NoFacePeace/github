package search

func countGoodNodes(edges [][]int) int {
	ad := map[int][]int{}
	for _, edge := range edges {
		v1, v2 := edge[0], edge[1]
		ad[v1] = append(ad[v1], v2)
	}
	ans := 0
	var dfs func(n int) int
	dfs = func(n int) int {
		arr := []int{}
		for _, v := range ad[n] {
			arr = append(arr, dfs(v))
		}
		if len(arr) == 0 {
			return 1
		}
		cnt := arr[0]
		num := arr[0]
		ans += len(arr) - 1
		for i := 1; i < len(arr); i++ {
			if arr[i] != num {
				ans--
			}
			cnt += arr[i]
		}
		return cnt + 1
	}
	return dfs(0)
}
