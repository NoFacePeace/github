package stack

import "strconv"

type Operate interface {
	Compute(a, b int) int
}

type Add struct{}

func (Add) Compute(a, b int) int {
	return a + b
}

type Sub struct{}

func (Sub) Compute(a, b int) int {
	return a - b
}

type Mul struct{}

func (Mul) Compute(a, b int) int {
	return a * b
}

type Div struct{}

func (Div) Compute(a, b int) int {
	return a / b
}

func evalRPN(tokens []string) int {
	stack := []int{}
	op := map[string]Operate{
		"+": Add{},
		"-": Sub{},
		"*": Mul{},
		"/": Div{},
	}
	for _, v := range tokens {
		if _, ok := op[v]; !ok {
			num, _ := strconv.Atoi(v)
			stack = append(stack, num)
			continue
		}
		obj := op[v]
		n := len(stack)
		num1, num2 := stack[n-1], stack[n-2]
		num := obj.Compute(num2, num1)
		stack[n-2] = num
		stack = stack[:n-1]
	}
	n := len(stack)
	return stack[n-1]
}
