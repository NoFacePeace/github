package stack

func trap(height []int) int {
	stack := []int{}
	ans := 0
	for i, h := range height {
		for len(stack) > 0 && h > height[stack[len(stack)-1]] {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				break
			}
			left := stack[len(stack)-1]
			curWith := i - left - 1
			curHeight := min(height[left], h) - height[top]
			ans += curWith * curHeight
		}
		stack = append(stack, i)
	}
	return ans
}
