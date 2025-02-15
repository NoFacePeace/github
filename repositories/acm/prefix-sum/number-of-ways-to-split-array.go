package prefixsum

func waysToSplitArray(nums []int) int {
	sum := 0
	for _, v := range nums {
		sum += v
	}
	cnt := 0
	val := 0
	for k, v := range nums {
		if k == len(nums)-1 {
			break
		}
		val += v
		sum -= v
		if val >= sum {
			cnt++
		}
	}
	return cnt
}
