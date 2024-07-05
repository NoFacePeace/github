package tree

func maximumBeauty(nums []int, k int) int {
	m := 0
	for _, x := range nums {
		m = max(m, x)
	}
	diff := make([]int, m+2)
	for _, x := range nums {
		diff[max(x-k, 0)]++
		diff[min(x+k+1, m+1)]--
	}
	res, count := 0, 0
	for _, x := range diff {
		count += x
		res = max(res, count)
	}
	return res
}
