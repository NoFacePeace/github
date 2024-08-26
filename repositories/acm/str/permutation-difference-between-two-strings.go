package str

func findPermutationDifference(s string, t string) int {
	arr := make([]int, 26)
	for k, v := range s {
		arr[int(v-'a')] = k
	}
	abs := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	sum := 0
	for k, v := range t {
		sum += abs(k, arr[int(v-'a')])
	}
	return sum
}
