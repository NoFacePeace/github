package str

func countBinarySubstrings(s string) int {
	arr := []int{}
	n := len(s)
	var last byte
	cnt := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			last = s[i]
			cnt = 1
			continue
		}
		if s[i] == last {
			cnt++
			continue
		}
		arr = append(arr, cnt)
		cnt = 1
		last = s[i]
	}
	arr = append(arr, cnt)
	ans := 0
	for i := 0; i < len(arr)-1; i++ {
		ans += min(arr[i], arr[i+1])
	}
	return ans
}
