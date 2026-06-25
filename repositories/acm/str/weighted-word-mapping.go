package str

func mapWordWeights(words []string, weights []int) string {
	n := len(words)
	nums := []int{}
	for _, word := range words {
		num := 0
		for i := 0; i < len(word); i++ {
			idx := word[i] - 'a'
			weight := weights[idx]
			num += weight
		}
		num %= 26
		nums = append(nums, num)
	}
	arr := []byte{}
	for i := 0; i < n; i++ {
		arr = append(arr, byte('z'-nums[i]))
	}
	return string(arr)
}
