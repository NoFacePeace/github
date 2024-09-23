package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
type BSTIterator struct {
	arr []int
}

func NewBSTIterator(root *TreeNode) BSTIterator {
	arr := []int{}
	var f func(root *TreeNode)
	f = func(root *TreeNode) {
		if root == nil {
			return
		}
		f(root.Left)
		arr = append(arr, root.Val)
		f(root.Right)
	}
	f(root)
	return BSTIterator{
		arr: arr,
	}
}

func (this *BSTIterator) Next() int {
	num := this.arr[0]
	this.arr = this.arr[1:]
	return num
}

func (this *BSTIterator) HasNext() bool {
	return len(this.arr) > 0
}

/**
 * Your BSTIterator object will be instantiated and called as such:
 * obj := Constructor(root);
 * param_1 := obj.Next();
 * param_2 := obj.HasNext();
 */
