package hash

func findLucky(arr []int) int {
	m := map[int]int{}
	for _, v := range arr {
		m[v]++
	}
	ans := 0
	for k, v := range m {
		if k != v {
			continue
		}
		ans = max(ans, k)
	}
	if ans == 0 {
		return -1
	}
	return ans
}
