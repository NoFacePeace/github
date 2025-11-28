package str

func hasSameDigits(s string) bool {
	n := len(s)
	arr := []int{}
	for i := 0; i < n; i++ {
		arr = append(arr, int(s[i]-'0'))
	}
	for len(arr) > 2 {
		tmp := []int{}
		n := len(arr)
		for i := 0; i < n-1; i++ {
			tmp = append(tmp, (arr[i]+arr[i+1])%10)
		}
		arr = tmp
	}
	if len(arr) == 1 {
		return false
	}
	if arr[0] == arr[1] {
		return true
	}
	return false
}
