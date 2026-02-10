package hash

func longestBalanced(nums []int) int {
	n := len(nums)
	ans := 0
	for i := 0; i < n; i++ {
		even := map[int]int{}
		odd := map[int]int{}
		for j := i; j < n; j++ {
			num := nums[j]
			if nums[j]%2 == 0 {
				even[num]++
			} else {
				odd[num]++
			}
			if len(even) == len(odd) {
				ans = max(ans, j-i+1)
			}
		}
	}
	return ans
}
