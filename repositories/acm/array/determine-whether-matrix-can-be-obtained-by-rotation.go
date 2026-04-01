package array

func findRotation(mat [][]int, target [][]int) bool {
	s1 := ""
	s2 := ""
	n := len(mat)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if target[i][j] == 1 {
				s1 += "1"
			} else {
				s1 += "0"
			}
			if mat[i][j] == 1 {
				s2 += "1"
			} else {
				s2 += "0"
			}
		}
	}
	if s1 == s2 {
		return true
	}
	s3 := ""
	for i := len(s2) - 1; i >= 0; i-- {
		s3 += s2[i : i+1]
	}
	if s1 == s3 {
		return true
	}
	s3 = ""
	for j := 0; j < n; j++ {
		for i := n - 1; i >= 0; i-- {
			if mat[i][j] == 1 {
				s3 += "1"
			} else {
				s3 += "0"
			}
		}
	}
	if s1 == s3 {
		return true
	}
	s2 = ""
	for i := len(s3) - 1; i >= 0; i-- {
		s2 += s3[i : i+1]
	}
	if s1 == s2 {
		return true
	}
	return false

}
