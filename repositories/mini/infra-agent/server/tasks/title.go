package tasks

import "unicode/utf8"

// truncateRunes 按 Unicode 码点数截断字符串。
// 非正数上限表示不限制；上限不足以容纳省略号时直接截取。
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	const ellipsis = "..."
	runes := []rune(s)
	if maxRunes <= len(ellipsis) {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-len(ellipsis)]) + ellipsis
}
