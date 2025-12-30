package sort

import "sort"

func maximumHappinessSum(happiness []int, k int) int64 {
	sort.Ints(happiness)
	n := len(happiness)
	cnt := 0
	ans := 0
	for i := n - 1; i >= max(n-k, 0); i-- {
		ans += max(happiness[i]+cnt, 0)
		cnt--
	}
	return int64(ans)
}
