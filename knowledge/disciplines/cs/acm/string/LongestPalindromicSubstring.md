# 最长回文子串

## 1. 题目描述

[LeetCode 5：最长回文子串](https://leetcode.cn/problems/longest-palindromic-substring/)。

给定一个字符串 `s`，返回其中最长的回文子串。

回文是正向和反向读取结果相同的字符串。子串必须是原字符串中连续的一段字符，不能跳过中间字符。如果存在多个长度相同的最长回文子串，返回其中任意一个即可。

题目保证 `1 <= s.length <= 1000`，且 `s` 仅由数字和英文字母组成。

## 2. 示例

```text
输入：s = "babad"
输出："bab"
解释："aba" 也是符合要求的答案。
```

```text
输入：s = "cbbd"
输出："bb"
```

## 3. Go 实现

使用[动态规划（DP）](../../algorithms/dp/DP.md)，令 `dp[l][r]` 表示闭区间 `s[l:r+1]` 是否为回文。长度为 1 时为回文，长度为 2 时判断两端字符是否相同；更长的区间还要求内部区间 `dp[l+1][r-1]` 为回文。代码按区间长度递增计算，并记录最长的回文子串。

```go
func longestPalindrome(s string) string {
    n := len(s)
    if n == 0 {
        return ""
    }
    dp := make([][]bool, n)
    for k := range dp {
        dp[k] = make([]bool, n)
    }
    ans := ""
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            if i+j >= n {
                break
            }
            if i+j == j {
                dp[j][j+i] = true
            } else if i == 1 {
                if s[j] == s[j+i] {
                    dp[j][j+1] = true
                }
            } else {
                if dp[j+1][j+i-1] && s[j] == s[j+i] {
                    dp[j][j+i] = true
                }
            }
            if dp[j][j+i] && (i+1) > len(ans) {
                ans = s[j : j+i+1]
            }
        }
    }
    return ans
}
```
