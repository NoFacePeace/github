package stack

func minAddToMakeValid(s string) int {
	stack := []string{}
	ans := 0
	n := len(s)
	for i := 0; i < n; i++ {
		str := s[i : i+1]
		if str == "(" {
			stack = append(stack, str)
			continue
		}
		if len(stack) == 0 {
			ans++
			continue
		}
		stack = stack[:len(stack)-1]
	}
	ans += len(stack)
	return ans
}
