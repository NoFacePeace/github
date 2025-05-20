package math

// https://leetcode.cn/problems/three-consecutive-odds/?envType=daily-question&envId=2025-05-11

func threeConsecutiveOdds(arr []int) bool {
	cnt := 0
	for _, v := range arr {
		if v%2 == 0 {
			cnt = 0
			continue
		}
		cnt++
		if cnt == 3 {
			return true
		}
	}
	return false
}
