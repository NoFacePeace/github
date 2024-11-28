package array

func minOperationsII(nums []int) int {
	cnt := 0
	for _, v := range nums {
		if cnt%2 == 0 && v == 1 {
			continue
		}
		if cnt%2 == 1 && v == 0 {
			continue
		}
		cnt++
	}
	return cnt
}
