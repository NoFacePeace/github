package algorithm

func majorityElement(nums []int) int {
	l := len(nums)
	if l == 0 {
		return 0
	}
	val := nums[0]
	cnt := 0
	for _, v := range nums {
		if cnt == 0 {
			val = v
			cnt++
			continue
		}
		if v == val {
			cnt++
		} else {
			cnt--
		}
	}
	return val
}
