package slidingwindow

func maxTotalFruits(fruits [][]int, startPos int, k int) int {
	left := 0
	right := 0
	n := len(fruits)
	sum := 0
	ans := 0
	step := func(left, right int) int {
		if fruits[right][0] <= startPos {
			return startPos - fruits[left][0]
		}
		if fruits[left][0] >= startPos {
			return fruits[right][0] - startPos
		}
		return min(fruits[right][0]-startPos, startPos-fruits[left][0]) + fruits[right][0] - fruits[left][0]
	}
	for right < n {
		sum += fruits[right][1]
		for left <= right && step(left, right) > k {
			sum -= fruits[left][1]
			left++
		}
		ans = max(ans, sum)
		right++
	}
	return ans
}
