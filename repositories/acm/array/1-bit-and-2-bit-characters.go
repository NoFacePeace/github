package array

func isOneBitCharacter(bits []int) bool {
	n := len(bits)
	i := 0
	for i < n {
		bit := bits[i]
		if bit == 1 {
			i += 2
			continue
		}
		if i == n-1 {
			return true
		}
		i++
	}
	return false
}
