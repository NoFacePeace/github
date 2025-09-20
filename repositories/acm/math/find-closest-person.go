package math

func findClosest(x int, y int, z int) int {
	d1 := z - x
	d2 := z - y
	if d1*d2 > 0 {
		if d1 <= 0 {
			d1 = -d1
			d2 = -d2
		}
		if d1 > d2 {
			return 2
		}
		if d1 == d2 {
			return 0
		}
		return 1
	}
	if d1 < 0 {
		d1 = -d1
	}
	if d2 < 0 {
		d2 = -d2
	}
	if d1 < d2 {
		return 1
	}
	if d1 == d2 {
		return 0
	}
	return 2
}
