package sort

func nextGreatestLetter(letters []byte, target byte) byte {
	ans := letters[0]
	for _, v := range letters {
		if v <= target {
			continue
		}
		if ans <= target {
			ans = v
			continue
		}
		ans = min(ans, v)
	}
	return ans
}
