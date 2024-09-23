package stack

func clearDigits(s string) string {
	arr := []byte(s)
	stack := []byte{}
	n := len(arr)
	for i := 0; i < n; i++ {
		if arr[i] >= 'a' && arr[i] <= 'z' {
			stack = append(stack, arr[i])
			continue
		}
		stack = stack[:len(stack)-1]
	}
	return string(stack)
}
