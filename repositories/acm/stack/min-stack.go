package stack

import "math"

type MinStack struct {
	stack1 []int
	stack2 []int
	mn     int
}

func NewMinStack() MinStack {
	return MinStack{
		stack1: []int{},
		stack2: []int{},
		mn:     math.MaxInt,
	}
}

func (this *MinStack) Push(val int) {
	this.stack1 = append(this.stack1, val)
	if val < this.mn {
		this.stack2 = append(this.stack2, val)
		this.mn = val
	} else {
		this.stack2 = append(this.stack2, this.mn)
	}
}

func (this *MinStack) Pop() {
	if len(this.stack1) == 0 {
		return
	}
	n := len(this.stack1)
	this.stack1 = this.stack1[:n-1]
	this.stack2 = this.stack2[:n-1]
	if n == 1 {
		this.mn = math.MaxInt
	} else {
		this.mn = this.stack2[n-2]
	}
}

func (this *MinStack) Top() int {
	n := len(this.stack1)
	return this.stack1[n-1]
}

func (this *MinStack) GetMin() int {
	return this.mn
}
