package str

func scoreOfString(s string) int {
	arr := []byte(s)
	n := len(arr)
	ans := 0
	for i := 0; i < n-1; i++ {
		dist := int(arr[i]) - int(arr[i+1])
		if dist < 0 {
			dist = 0 - dist
		}
		ans += dist
	}
	return ans
}
