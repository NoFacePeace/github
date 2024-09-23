package stack

func isValid(s string) bool {
	arr := []byte(s)
	stack := []byte{}
	m := map[byte]byte{
		')': '(',
		']': '[',
		'}': '{',
	}
	for _, v := range arr {
		if _, ok := m[v]; !ok {
			stack = append(stack, v)
			continue
		}
		n := len(stack)
		if n == 0 {
			return false
		}
		c := stack[n-1]
		if m[v] != c {
			return false
		}
		stack = stack[:n-1]
	}
	return len(stack) == 0
}
