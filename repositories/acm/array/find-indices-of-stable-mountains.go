package array

func stableMountains(height []int, threshold int) []int {
	ans := []int{}
	for k := range height {
		if k == len(height)-1 {
			continue
		}
		if height[k] > threshold {
			ans = append(ans, k+1)
		}
	}
	return ans
}
