package search

func canReach(arr []int, start int) bool {
	var dfs func(idx int) bool
	n := len(arr)
	visited := make([]bool, n)
	dfs = func(idx int) bool {
		if idx < 0 {
			return false
		}
		if idx >= n {
			return false
		}
		if visited[idx] {
			return false
		}
		if arr[idx] == 0 {
			return true
		}
		visited[idx] = true
		return dfs(idx-arr[idx]) || dfs(idx+arr[idx])
	}
	return dfs(start)
}
