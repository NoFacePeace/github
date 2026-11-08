# 搜索旋转排序数组

## 1. 题目描述

[LeetCode 33：搜索旋转排序数组](https://leetcode.cn/problems/search-in-rotated-sorted-array/)。

给定一个由互不相同的整数组成、按升序排列后可能经过旋转的数组 `nums`，以及整数目标值 `target`，返回目标值在数组中的下标；如果不存在，则返回 `-1`。

例如，数组 `[0, 1, 2, 4, 5, 6, 7]` 旋转后可以变为 `[4, 5, 6, 7, 0, 1, 2]`。旋转保留各段内部的相对顺序。

要求算法的时间复杂度为 O(log n)。

## 2. 示例

```text
输入：nums = [4, 5, 6, 7, 0, 1, 2], target = 0
输出：4
```

```text
输入：nums = [4, 5, 6, 7, 0, 1, 2], target = 3
输出：-1
```

```text
输入：nums = [1], target = 0
输出：-1
```

## 3. Go 实现

使用[二分查找](../../algorithms/search/BinarySearch.md)。数组元素互不相同，因此每轮至少有一侧区间有序，可以结合目标值、中间值与区间端点的大小关系排除不可能包含目标的一侧。

```go
func search(nums []int, target int) int {
    n := len(nums)
    if n == 0 {
        return -1
    }
    l, r := 0, n-1
    for l <= r {
        mid := (l + r) / 2
        if nums[mid] == target {
            return mid
        }
        left := nums[l]
        right := nums[r]
        if target < nums[mid] {
            if target >= left {
                r = mid - 1
                continue
            }
            if left <= nums[mid] {
                l = mid + 1
                continue
            }
            r = mid - 1
            continue
        }
        if target <= right {
            l = mid + 1
            continue
        }
        if right >= nums[mid] {
            r = mid - 1
            continue
        }
        l = mid + 1
    }
    return -1
}
```
