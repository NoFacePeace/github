package str

import "strings"

func numOfStrings(patterns []string, word string) int {
	ans := 0
	for _, v := range patterns {
		if strings.Contains(word, v) {
			ans++
		}
	}
	return ans
}
