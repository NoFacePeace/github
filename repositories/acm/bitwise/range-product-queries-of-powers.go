package bitwise

func productQueries(n int, queries [][]int) []int {
	arr := []int{}
	cnt := 0
	for n > 0 {
		if n&1 != 0 {
			arr = append(arr, cnt)
		}
		n = n >> 1
		cnt++
	}
	prefix := make([]int, len(arr))
	for i := 0; i < len(arr); i++ {
		if i == 0 {
			prefix[i] = arr[i]
			continue
		}
		prefix[i] = arr[i] + prefix[i-1]
	}
	ans := []int{}
	for _, query := range queries {
		left, right := query[0], query[1]
		if left == 0 {
			ans = append(ans, prefix[right])
			continue
		}
		ans = append(ans, prefix[right]-prefix[left-1])
	}
	mod := int(1e9) + 7
	for k, v := range ans {
		val := 1
		for i := 0; i < v; i++ {
			val = val * 2 % mod
		}
		ans[k] = val
	}
	return ans
}
