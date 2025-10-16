package math

func maxBottlesDrunk(numBottles int, numExchange int) int {
	var f func(num, ex int) int
	f = func(num, ex int) int {
		if num < ex {
			return 0
		}
		return f(num-ex+1, ex+1) + 1
	}
	return f(numBottles, numExchange) + numBottles
}
