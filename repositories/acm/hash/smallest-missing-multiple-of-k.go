package hash

func missingMultiple(nums []int, k int) int {
	m := map[int]bool{}
	for _, v := range nums {
		m[v] = true
	}
	num := k
	for m[num] {
		num += k
	}
	return num
}
