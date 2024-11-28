package constanttime

func losingPlayer(x int, y int) string {
	y = y / 4
	mn := min(x, y)
	if mn%2 == 0 {
		return "Bob"
	}
	return "Alice"
}
