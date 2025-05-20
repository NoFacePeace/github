package bitwise

// https://leetcode.cn/problems/count-largest-group/?envType=daily-question&envId=2025-04-23

func countLargestGroup(n int) int {
	m := map[int]int{}
	bitsum := func(num int) int {
		sum := 0
		for num != 0 {
			sum += num % 10
			num /= 10
		}
		return sum
	}
	mx := 0
	ans := 0
	for i := 1; i <= n; i++ {
		sum := bitsum(i)
		m[sum]++
		if m[sum] == mx {
			ans++
		}
		if m[sum] > mx {
			mx = m[sum]
			ans = 1
		}
	}
	return ans
}
