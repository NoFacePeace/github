package bitwise

func doesValidArrayExist(derived []int) bool {
	n := len(derived)
	if n == 0 {
		return true
	}
	arr := make([]int, n)
	arr[0] = 1
	for i := 0; i < n; i++ {
		if i == n-1 {
			if arr[0]^arr[i] == derived[i] {
				return true
			}
			continue
		}
		arr[i+1] = arr[i] ^ derived[i]
	}
	arr[0] = 0
	for i := 0; i < n; i++ {
		if i == n-1 {
			if arr[0]^arr[i] == derived[i] {
				return true
			}
			continue
		}
		arr[i+1] = arr[i] ^ derived[i]
	}
	return false
}
