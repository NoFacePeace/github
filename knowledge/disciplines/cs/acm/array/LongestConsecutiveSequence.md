# 最长连续序列

## 1. 题目描述

[LeetCode 128：最长连续序列](https://leetcode.cn/problems/longest-consecutive-sequence/)。

给定一个未排序的整数数组 `nums`，找出数值连续的最长序列，并返回其长度。

连续序列中的相邻数值相差 1，元素不要求在原数组中位置连续，也不要求保持原数组中的相对顺序。重复值不会增加序列长度；空数组返回 `0`。

要求算法的时间复杂度为 O(n)。

## 2. 示例

```text
输入：nums = [100, 4, 200, 1, 3, 2]
输出：4
解释：最长连续序列为 [1, 2, 3, 4]，长度为 4。
```

```text
输入：nums = [0, 3, 7, 2, 5, 8, 4, 6, 0, 1]
输出：9
解释：最长连续序列为 [0, 1, 2, 3, 4, 5, 6, 7, 8]，重复的 0 不增加长度。
```

```text
输入：nums = []
输出：0
```

## 3. Go 实现

使用[哈希表](../../data-structures/hash-table/HashTable.md)记录数组中存在的数值，并用 `visited` 标记已处理的数值。每次从尚未访问的数值向两侧扩展，统计整个连续序列的长度，再更新答案，避免重复处理同一序列。

```go
func longestConsecutive(nums []int) int {
    m := map[int]bool{}
    for _, v := range nums {
        m[v] = true
    }
    visited := map[int]bool{}
    ans := 0
    for _, v := range nums {
        if visited[v] {
            continue
        }
        visited[v] = true
        cnt := 1
        i := v - 1
        for m[i] {
            visited[i] = true
            i--
            cnt++
        }
        i = v + 1
        for m[i] {
            visited[i] = true
            i++
            cnt++
        }
        ans = max(ans, cnt)
    }
    return ans
}
```
