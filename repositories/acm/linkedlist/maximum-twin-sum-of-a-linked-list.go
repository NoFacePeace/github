package linkedlist

func pairSum(head *ListNode) int {
	arr := []int{}
	for head != nil {
		arr = append(arr, head.Val)
		head = head.Next
	}
	ans := 0
	n := len(arr)
	for i := 0; i < n/2; i++ {
		ans = max(ans, arr[i]+arr[n-i-1])
	}
	return ans
}
