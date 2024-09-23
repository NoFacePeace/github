package stack

func removeStars(s string) string {
	stack := []byte{}
	arr := []byte(s)
	for _, v := range arr {
		if v != '*' {
			stack = append(stack, v)
			continue
		}
		if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
	}
	return string(stack)
}
