package constanttime

func squareIsWhite(coordinates string) bool {
	x := int(coordinates[0] - 'a')
	y := int(coordinates[1] - '1')
	if x%2 == 0 {
		if y%2 == 0 {
			return false
		}
		return true
	}
	if y%2 == 0 {
		return true
	}
	return false
}
