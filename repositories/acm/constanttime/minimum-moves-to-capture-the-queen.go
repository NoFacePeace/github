package constanttime

func minMovesToCaptureTheQueen(a int, b int, c int, d int, e int, f int) int {
	n := 9
	for i := a + 1; i < n; i++ {
		if i == c && b == d {
			break
		}
		if i == e && b == f {
			return 1
		}
	}
	for i := a - 1; i > 0; i-- {
		if i == c && b == d {
			break
		}
		if i == e && b == f {
			return 1
		}
	}
	for i := b + 1; i < n; i++ {
		if i == d && a == c {
			break
		}
		if i == f && a == e {
			return 1
		}
	}
	for i := b - 1; i > 0; i-- {
		if i == d && a == c {
			break
		}
		if i == f && a == e {
			return 1
		}
	}
	for i := 1; i < n; i++ {
		if c+i < n && d+i < n {
			if c+i == a && d+i == b {
				break
			}
			if c+i == e && d+i == f {
				return 1
			}
		}
	}
	for i := 1; i < n; i++ {
		if c-i >= 0 && d-i > 0 {
			if c-i == a && d-i == b {
				break
			}
			if c-i == e && d-i == f {
				return 1
			}
		}
	}
	for i := 1; i < n; i++ {
		if c+i < n && d-i > 0 {
			if c+i == a && d-i == b {
				break
			}
			if c+i == e && d-i == f {
				return 1
			}
		}
	}
	for i := 1; i < n; i++ {
		if c-i >= 0 && d+i < n {
			if c-i == a && d+i == b {
				break
			}
			if c-i == e && d+i == f {
				return 1
			}
		}
	}
	return 2
}
