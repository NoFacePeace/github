# 无重复字符的最长子串

## 1. 题目描述

[LeetCode 3：无重复字符的最长子串](https://leetcode.cn/problems/longest-substring-without-repeating-characters/)。

给定一个字符串 `s`，找出其中不含重复字符的最长子串的长度。

子串必须是原字符串中连续的一段字符。`s` 可以包含英文字母、数字、符号和空格，也可以为空字符串。

## 2. 示例

```text
输入：s = "abcabcbb"
输出：3
解释：最长的无重复字符子串之一是 "abc"，长度为 3。
```

```text
输入：s = "bbbbb"
输出：1
解释：最长的无重复字符子串是 "b"，长度为 1。
```

```text
输入：s = "pwwkew"
输出：3
解释：最长的无重复字符子串之一是 "wke"，长度为 3。
"pwke" 是子序列，不是连续的子串。
```

## 3. Go 实现

使用[滑动窗口](../../algorithms/two-pointers/SlidingWindow.md)维护不含重复字符的连续区间。右边界遇到重复字符时，左边界移过窗口中原有的同一字符，并更新字符记录；遍历过程中记录合法窗口的最大长度。

```go
func lengthOfLongestSubstring(s string) int {
    l, r := 0, 0
    n := len(s)
    m := map[string]bool{}
    ans := 0
    for r < n {
        str := s[r:r+1]
        if !m[str] {
            m[str] = true
            ans = max(ans, r - l + 1)
            r++
            continue
        }
        for l < r {
            tmp := s[l:l+1]
            if tmp == str {
                l++
                break
            }
            m[tmp] = false
            l++
        }
        r++
    }

    return ans
}
```
