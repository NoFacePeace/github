# 数组中的第 K 个最大元素

## 1. 题目描述

[LeetCode 215：数组中的第 K 个最大元素](https://leetcode.cn/problems/kth-largest-element-in-an-array/)。

给定整数数组 `nums` 和整数 `k`，返回数组中第 `k` 个最大的元素。

第 `k` 个最大的元素是数组按非递增顺序排列后位于第 `k` 个位置的元素，而不是第 `k` 个不同的元素；重复元素也参与排名。题目保证 `1 <= k <= nums.length`。

要求在不对整个数组排序的情况下解决问题。

## 2. 示例

```text
输入：nums = [3, 2, 1, 5, 6, 4], k = 2
输出：5
解释：按从大到小排列为 [6, 5, 4, 3, 2, 1]，第 2 个元素为 5。
```

```text
输入：nums = [3, 2, 3, 1, 2, 4, 5, 5, 6], k = 4
输出：4
解释：按从大到小排列为 [6, 5, 5, 4, 3, 3, 2, 2, 1]，第 4 个元素为 4。
```

## 3. Go 实现

使用[快速选择](../../algorithms/selection/QuickSelect.md)，根据目标位置仅继续处理划分后包含目标的区间，无须对整个数组排序。实现会修改输入数组。

```go
func findKthLargest(nums []int, k int) int {
    var dfs func(nums []int, k int) int
    dfs = func(nums []int, k int) int {
        pivot := nums[0]
        n := len(nums)
        l, r := 0, n-1
        for l <= r {
            for l <= r {
                if nums[l] > pivot {
                    break
                }
                l++
            }
            for l <= r {
                if nums[r] < pivot {
                    break
                }
                r--
            }
            if l > r {
                break
            }
            nums[l], nums[r] = nums[r], nums[l]
            l++
            r--
        }
        target := n - k
        if target <= r {
            return dfs(nums[:r+1], k-(n-r-1))
        }
        if target >= l {
            return dfs(nums[l:], k)
        }
        return pivot
    }
    return dfs(nums, k)
}
```
