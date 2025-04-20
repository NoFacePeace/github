package hash

func findMatrix(nums []int) [][]int {
	cnt := map[int]int{}
	for _, x := range nums {
		cnt[x]++
	}
	ans := [][]int{}
	for len(cnt) > 0 {
		arr := []int{}
		for k, v := range cnt {
			cnt[k] = v - 1
			arr = append(arr, k)
			if cnt[k] == 0 {
				delete(cnt, k)
			}
		}
		ans = append(ans, arr)
	}
	return ans
}
