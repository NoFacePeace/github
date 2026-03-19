package hash

func findDifferentBinaryString(nums []string) string {
	n := len(nums)
	m := map[string]bool{}
	for _, v := range nums {
		m[v] = true
	}
	var dfs func(str string, idx int) string
	dfs = func(str string, idx int) string {
		if !m[str] {
			return str
		}
		if idx == n {
			return ""
		}
		ans := dfs(str, idx+1)
		if ans != "" {
			return ans
		}
		arr := []byte(str)
		arr[idx] = '1'
		str = string(arr)
		return dfs(str, idx+1)
	}
	str := ""
	for i := 0; i < n; i++ {
		str += "0"
	}
	return dfs(str, 0)
}
