package array

func stoneGameIX(stones []int) bool {
	cnt0, cnt1, cnt2 := 0, 0, 0
	for _, val := range stones {
		val %= 3
		if val == 0 {
			cnt0++
			continue
		}
		if val == 1 {
			cnt1++
			continue
		}
		cnt2++
	}
	if cnt0%2 == 0 {
		return cnt1 >= 1 && cnt2 >= 1
	}
	return cnt1-cnt2 > 2 || cnt2-cnt1 > 2
}
