package array

func finalPositionOfSnake(n int, commands []string) int {
	x, y := 0, 0
	for _, command := range commands {
		switch command {
		case "RIGHT":
			y++
		case "LEFT":
			y--
		case "UP":
			x--
		case "DOWN":
			x++
		}
	}
	return x*n + y
}
