package str

func isPalindrome(s string) bool {
	arr := []byte{}
	for i := 0; i < len(s); i++ {
		if s[i] >= 'a' && s[i] <= 'z' {
			arr = append(arr, s[i])
			continue
		}
		if s[i] >= 'A' && s[i] <= 'Z' {
			arr = append(arr, s[i]-'A'+'a')
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			arr = append(arr, s[i])
			continue
		}
	}
	n := len(arr)
	for i := 0; i < n/2; i++ {
		if arr[i] != arr[n-i-1] {
			return false
		}
	}
	return true
}
