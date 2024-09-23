package search

func medianOfUniquenessArray(nums []int) int {
	n := len(nums)
	median := (n*(n+1)/2 + 1) / 2
	check := func(t int) bool {
		cnt := map[int]int{}
		tot := 0
		for i, j := 0, 0; i < n; i++ {
			cnt[nums[i]]++
			for len(cnt) > t {
				cnt[nums[j]]--
				if cnt[nums[j]] == 0 {
					delete(cnt, nums[j])
				}
				j++
			}
			tot += i - j + 1
		}
		return tot >= median
	}
	res := 0
	lo, hi := 1, n
	for lo <= hi {
		mid := (lo + hi) / 2
		if check(mid) {
			res = mid
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return res
}
