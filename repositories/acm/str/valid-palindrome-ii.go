package str

// https://leetcode.cn/problems/valid-palindrome-ii/
func validPalindrome(s string) bool {
	arr := []byte(s)
	l, r := 0, len(arr)-1
	for l <= r {
		if arr[l] != arr[r] {
			break
		}
		l++
		r--
	}
	if l > r {
		return true
	}
	f := func(s string) bool {
		arr := []byte(s)
		n := len(arr)
		for i := 0; i < n/2; i++ {
			if arr[i] != arr[n-i-1] {
				return false
			}
		}
		return true
	}
	arr1 := []byte{}
	arr1 = append(arr1, arr[:l]...)
	arr1 = append(arr1, arr[l+1:]...)
	arr2 := []byte{}
	arr2 = append(arr2, arr[:r]...)
	arr2 = append(arr2, arr[r+1:]...)
	return f(string(arr1)) || f(string(arr2))
}
