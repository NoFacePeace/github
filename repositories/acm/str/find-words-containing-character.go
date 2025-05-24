package str

// https://leetcode.cn/problems/find-words-containing-character/description/?envType=daily-question&envId=2025-05-24

func findWordsContaining(words []string, x byte) []int {
	ans := []int{}
	for k, word := range words {
		n := len(word)
		for i := 0; i < n; i++ {
			if word[i] == x {
				ans = append(ans, k)
				break
			}
		}
	}
	return ans
}
