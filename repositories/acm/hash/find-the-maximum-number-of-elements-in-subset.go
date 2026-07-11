package hash

import "sort"

func maximumLength(nums []int) int {
	sort.Ints(nums)
	m := map[int]int{}
	arr := []int{}
	for _, num := range nums {
		m[num]++
		if m[num] == 1 {
			arr = append(arr, num)
		}
	}
	cnt := map[int]int{}
	ans := 1
	for _, num := range arr {
		if num == 1 {
			if m[num]%2 == 0 {
				ans = m[num] - 1
			} else {
				ans = m[num]
			}
			continue
		}
		if cnt[num] == 0 {
			cnt[num] = 1
		}
		if m[num*num] > 0 && m[num] >= 2 {
			cnt[num*num] = cnt[num] + 2
			ans = max(ans, cnt[num*num])
		}
	}
	return ans
}
