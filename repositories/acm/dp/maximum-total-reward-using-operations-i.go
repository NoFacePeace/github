package dp

import "sort"

func maxTotalReward(rewardValues []int) int {
	sort.Ints(rewardValues)
	m := rewardValues[len(rewardValues)-1]
	dp := make([]int, 2*m)
	dp[0] = 1
	for _, x := range rewardValues {
		for k := 2*x - 1; k >= x; k-- {
			if dp[k-x] == 1 {
				dp[k] = 1
			}
		}
	}
	res := 0
	for i := 0; i < len(dp); i++ {
		if dp[i] == 1 {
			res = i
		}
	}
	return res
}
