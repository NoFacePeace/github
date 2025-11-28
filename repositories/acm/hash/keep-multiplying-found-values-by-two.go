package hash

func findFinalValue(nums []int, original int) int {
	m := map[int]bool{}
	for _, v := range nums {
		m[v] = true
	}
	for m[original] {
		m[original] = false
		original *= 2
	}
	return original
}
