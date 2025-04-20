package array

func countPairs(nums []int, k int) int {
	m := map[int][]int{}
	for k, v := range nums {
		m[v] = append(m[v], k)
	}
	ans := 0
	for _, arr := range m {
		n := len(arr)
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if arr[i]*arr[j]%k == 0 {
					ans++
				}
			}
		}
	}
	return ans
}
