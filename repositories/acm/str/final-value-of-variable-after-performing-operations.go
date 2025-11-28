package str

func finalValueAfterOperations(operations []string) int {
	ans := 0
	for _, v := range operations {
		if v == "--X" || v == "X--" {
			ans--
		} else {
			ans++
		}
	}
	return ans
}
