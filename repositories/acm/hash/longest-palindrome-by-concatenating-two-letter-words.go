package hash

// https://leetcode.cn/problems/longest-palindrome-by-concatenating-two-letter-words/solutions/1202034/lian-jie-liang-zi-mu-dan-ci-de-dao-de-zu-vs99/?envType=daily-question&envId=2025-05-25

func longestPalindrome(words []string) int {
	ans := 0
	m := map[string]int{}
	for _, word := range words {
		reverse := word[1:] + word[:1]
		if m[reverse] > 0 {
			m[reverse]--
			ans += 4
			continue
		}
		m[word]++
	}
	for k, v := range m {
		if v == 0 {
			continue
		}
		if k[1:] == k[:1] {
			ans += 2
			break
		}
	}
	return ans
}
