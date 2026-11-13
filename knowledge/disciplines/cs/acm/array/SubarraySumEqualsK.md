# 和为 K 的子数组

## 1. 题目描述

[LeetCode 560：和为 K 的子数组](https://leetcode.cn/problems/subarray-sum-equals-k/)。

给定一个整数数组 `nums` 和一个整数 `k`，返回数组中和为 `k` 的连续子数组的数量。

子数组必须非空，且是原数组中连续的一段元素。数组中的元素可以为正数、负数或 0；不同起止位置对应的子数组分别计数。

## 2. 示例

```text
输入：nums = [1, 1, 1], k = 2
输出：2
解释：下标区间 [0, 1] 和 [1, 2] 对应的子数组和均为 2。
```

```text
输入：nums = [1, 2, 3], k = 3
输出：2
解释：子数组 [1, 2] 和 [3] 的和均为 3。
```

## 3. Go 实现

使用[前缀和](../../algorithms/prefix-sum/PrefixSum.md)：当前前缀和为 `sum` 时，此前每个值为 `sum-k` 的前缀和都对应一个和为 `k` 的子数组。`m` 记录此前各个前缀和的出现次数，先查询再记录当前值；初始化 `m[0] = 1` 表示空前缀，用于统计从下标 0 开始的子数组。

```go
func subarraySum(nums []int, k int) int {
    m := map[int]int{}
    sum := 0
    ans := 0
    m[0] = 1
    for _, v := range nums {
        sum += v
        ans += m[sum-k]
        m[sum]++
    }
    return ans
}
```
