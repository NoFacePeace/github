# 最长递增子序列

## 1. 题目描述

[LeetCode 300：最长递增子序列](https://leetcode.cn/problems/longest-increasing-subsequence/)。

给定一个整数数组 `nums`，返回其中最长严格递增子序列的长度。

子序列可以通过删除原数组中的部分元素或不删除任何元素得到，但必须保留剩余元素的相对顺序，不要求元素在原数组中连续。严格递增要求后一个元素大于前一个元素，相等元素不能连续选入该子序列。

进阶要求：设计时间复杂度为 O(n log n) 的算法。

## 2. 示例

```text
输入：nums = [10, 9, 2, 5, 3, 7, 101, 18]
输出：4
解释：最长严格递增子序列之一为 [2, 3, 7, 101]，长度为 4。
```

```text
输入：nums = [0, 1, 0, 3, 2, 3]
输出：4
解释：最长严格递增子序列之一为 [0, 1, 2, 3]，长度为 4。
```

```text
输入：nums = [7, 7, 7, 7, 7, 7, 7]
输出：1
解释：相等元素不能形成严格递增关系，因此最长长度为 1。
```

## 3. Go 实现

使用[贪心](../../algorithms/greedy/Greedy.md)维护不同长度递增子序列的最小末尾值，其中 `arr[i]` 表示长度为 `i+1` 的严格递增子序列能够取得的最小末尾值。末尾值越小，越有利于后续扩展。

对每个元素，通过[二分查找](../../algorithms/search/BinarySearch.md)找到 `arr` 中第一个大于或等于它的位置并替换；若不存在则追加。最终 `arr` 的长度即为答案，但其内容不一定是原数组中的一条实际子序列。时间复杂度为 O(n log n)，辅助空间复杂度为 O(n)。

```go
func lengthOfLIS(nums []int) int {
    arr := []int{}
    search := func(num int) int {
        l, r := 0, len(arr)-1
        for l <= r {
            mid := (l + r) / 2
            if arr[mid] < num {
                l = mid + 1
                continue
            }
            r = mid - 1
        }
        return l
    }
    for _, v := range nums {
        idx := search(v)
        if idx == len(arr) {
            arr = append(arr, v)
            continue
        }
        arr[idx] = v
    }
    return len(arr)
}
```
