# 字母异位词分组

## 1. 题目描述

[LeetCode 49：字母异位词分组](https://leetcode.cn/problems/group-anagrams/)。

给定一个字符串数组 `strs`，将字母异位词组合在一起，返回分组结果。答案可以按任意顺序返回。

字母异位词是由相同字符以不同顺序排列形成的字符串，各字符的出现次数必须相同。输入中的字符串仅包含小写英文字母，也可以是空字符串。

## 2. 示例

```text
输入：strs = ["eat", "tea", "tan", "ate", "nat", "bat"]
输出：[["bat"], ["nat", "tan"], ["ate", "eat", "tea"]]
解释："eat"、"tea"、"ate" 的字符及其出现次数相同，属于同一组；"tan" 与 "nat" 属于另一组。
```

```text
输入：strs = [""]
输出：[[""]]
```

```text
输入：strs = ["a"]
输出：[["a"]]
```

## 3. Go 实现

使用[哈希表](../../data-structures/hash-table/HashTable.md)按排序后的字符串分组。字母异位词排序后得到相同的字符串，将其作为键，原字符串追加到对应分组中，最后收集所有分组作为答案。

```go
import "sort"

func groupAnagrams(strs []string) [][]string {
    m := map[string][]string{}
    for _, v := range strs {
        tmp := []byte(v)
        sort.Slice(tmp, func(a, b int) bool {
            return tmp[a] < tmp[b]
        })
        str := string(tmp)
        if m[str] == nil {
            m[str] = []string{}
        }
        m[str] = append(m[str], v)
    }
    ans := [][]string{}
    for _, v := range m {
        ans = append(ans, v)
    }
    return ans
}
```
