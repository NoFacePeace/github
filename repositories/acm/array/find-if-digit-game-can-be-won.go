package array

func canAliceWin(nums []int) bool {
	single, double, other := 0, 0, 0
	for _, v := range nums {
		if v < 10 {
			single += v
			continue
		}
		if v < 100 {
			double += v
			continue
		}
		other += v
	}
	if single > double+other {
		return true
	}
	if double > single+other {
		return true
	}
	return false
}
