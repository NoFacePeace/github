package stack

func minOperations(nums []int) int {
	stack := []int{}
	n := len(nums)
	i := 0
	ans := 0
	for i < n {
		num := nums[i]
		if len(stack) == 0 {
			stack = append(stack, num)
			i++
			continue
		}
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			if top < num {
				break
			}
			if top > num {
				ans++
			}
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, num)
		i++
	}
	for i := 0; i < len(stack); i++ {
		num := stack[i]
		if num != 0 {
			ans++
		}
	}
	return ans
}
