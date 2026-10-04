package stack

func checkValidString(s string) bool {
	star := []int{}
	left := []int{}
	n := len(s)
	for i := 0; i < n; i++ {
		str := s[i : i+1]
		if str == "(" {
			left = append(left, i)
			continue
		}
		if str == "*" {
			star = append(star, i)
			continue
		}
		if len(star) == 0 && len(left) == 0 {
			return false
		}
		if len(left) != 0 {
			left = left[:len(left)-1]
			continue
		}
		star = star[:len(star)-1]
	}
	i := 0
	j := 0
	for i < len(left) && j < len(star) {
		if left[i] <= star[j] {
			i++
			j++
			continue
		}
		j++
	}
	if i == len(left) {
		return true
	}
	return false
}
