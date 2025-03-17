package greedy

func breakPalindrome(palindrome string) string {
	arr := []byte(palindrome)
	n := len(arr)
	for i := 0; i < n; i++ {
		if n%2 != 0 && i == n/2 {
			continue
		}
		if arr[i] != 'a' {
			arr[i] = 'a'
			return string(arr)
		}
	}
	if n == 1 {
		return ""
	}
	arr[n-1] = 'b'
	return string(arr)
}
