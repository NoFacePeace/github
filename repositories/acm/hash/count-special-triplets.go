package hash

func specialTriplets(nums []int) int {
	right := map[int]int{}
	for _, v := range nums {
		right[v]++
	}
	left := map[int]int{}
	ans := 0
	mod := int(1e9) + 7
	for _, num := range nums {
		right[num]--
		r := right[num*2]
		l := left[num*2]
		ans += r * l
		ans %= mod
		left[num]++
	}
	return ans
}
