# 两数之和

## 1. 题目描述

[LeetCode 1：两数之和](https://leetcode.cn/problems/two-sum/)。

给定一个整数数组 `nums` 和一个整数目标值 `target`，找出数组中和为 `target` 的两个整数，并返回它们的数组下标。

题目保证每种输入恰好只有一个有效答案，不能重复使用同一个元素。答案可以按任意顺序返回。

## 2. 示例

```text
输入：nums = [2, 7, 11, 15], target = 9
输出：[0, 1]
解释：nums[0] + nums[1] = 2 + 7 = 9。
```

```text
输入：nums = [3, 2, 4], target = 6
输出：[1, 2]
```

```text
输入：nums = [3, 3], target = 6
输出：[0, 1]
```

## 3. Go 实现

使用[哈希表](../../data-structures/hash-table/HashTable.md)保存已遍历元素的值与下标；先查找 `target-v`，再记录当前元素，避免重复使用同一个元素。

```go
func twoSum(nums []int, target int) []int {
    m := map[int]int{}
    for k, v := range nums {
        idx, ok := m[target-v]
        if ok {
            return []int{idx, k}
        }
        m[v] = k
    }
    return nil
}
```
