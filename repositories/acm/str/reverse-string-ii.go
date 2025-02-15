package str

// https://leetcode.cn/problems/reverse-string-ii/description/

func reverseStr(s string, k int) string {
	arr := []byte(s)
	reverse := func(arr []byte) {
		n := len(arr)
		for i := 0; i < n/2; i++ {
			arr[i], arr[n-i-1] = arr[n-i-1], arr[i]
		}
	}
	n := len(arr)
	for i := 0; i < n; i += 2 * k {
		reverse(arr[i:min(n, i+k)])
	}
	return string(arr)
}
