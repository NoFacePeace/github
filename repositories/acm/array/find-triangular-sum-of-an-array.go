package array

func triangularSum(nums []int) int {
	var f func(arr []int) int
	f = func(arr []int) int {
		if len(arr) == 0 {
			return 0
		}
		if len(arr) == 1 {
			return arr[0]
		}
		tmp := []int{}
		n := len(arr)
		for i := 0; i < n-1; i++ {
			tmp = append(tmp, (arr[i]+arr[i+1])%10)
		}
		return f(tmp)
	}
	return f(nums)
}
