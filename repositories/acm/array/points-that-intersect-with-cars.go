package array

func numberOfPoints(nums [][]int) int {
	C := 0
	for _, interval := range nums {
		if interval[1] > C {
			C = interval[1]
		}
	}
	diff := make([]int, C+2)
	for _, interval := range nums {
		diff[interval[0]]++
		diff[interval[1]+1]--
	}
	ans, count := 0, 0
	for i := 1; i <= C; i++ {
		count += diff[i]
		if count > 0 {
			ans++
		}
	}
	return ans
}
