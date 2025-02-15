package array

func minOperationsI(nums []int, k int) int {
	cnt := 0
	for _, v := range nums {
		if v < k {
			cnt++
		}
	}
	return cnt
}
