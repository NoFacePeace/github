package slidingwindow

// https://leetcode.cn/problems/count-of-substrings-containing-every-vowel-and-k-consonants-i/description/?envType=daily-question&envId=2025-03-12

func countOfSubstrings(word string, k int) int {
	vowels := map[byte]bool{
		'a': true,
		'e': true,
		'i': true,
		'o': true,
		'u': true,
	}
	count := func(m int) int {
		n := len(word)
		res := 0
		consonants := 0
		occur := map[byte]int{}
		for i, j := 0, 0; i < n; i++ {
			for j < n && (consonants < m || len(occur) < 5) {
				if vowels[word[j]] {
					occur[word[j]]++
				} else {
					consonants++
				}
				j++
			}
			if consonants >= m && len(occur) == 5 {
				res += n - j + 1
			}
			if vowels[word[i]] {
				occur[word[i]]--
				if occur[word[i]] == 0 {
					delete(occur, word[i])
				}
			} else {
				consonants--
			}
		}
		return res
	}
	return count(k) - count(k+1)
}
