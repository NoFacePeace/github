package linkedlist

type KthLargest struct {
}

type KthLargestNode struct {
	Left  *KthLargestNode
	Right *KthLargestNode
	Value int
}

func Constructor(k int, nums []int) KthLargest {
	return KthLargest{}
}

func (this *KthLargest) Add(val int) int {
	return 0
}

/**
 * Your KthLargest object will be instantiated and called as such:
 * obj := Constructor(k, nums);
 * param_1 := obj.Add(val);
 */
