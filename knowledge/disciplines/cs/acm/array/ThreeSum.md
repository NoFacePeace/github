# 三数之和

## 1. 题目描述

[LeetCode 15：三数之和](https://leetcode.cn/problems/3sum/)。

给定一个整数数组 `nums`，找出所有满足 `nums[i] + nums[j] + nums[k] == 0` 的三元组，其中 `i`、`j`、`k` 互不相同。

返回的答案中不能包含重复的三元组。三元组内部的顺序和三元组之间的顺序均不影响答案。

## 2. 示例

```text
输入：nums = [-1, 0, 1, 2, -1, -4]
输出：[[-1, -1, 2], [-1, 0, 1]]
解释：两个 -1 来自不同下标；相同数值组合只返回一次。
```

```text
输入：nums = [0, 1, 1]
输出：[]
解释：不存在和为 0 的三元组。
```

```text
输入：nums = [0, 0, 0]
输出：[[0, 0, 0]]
```

## 3. Go 实现

先使用标准库 `sort` 包中的 `sort.Ints` 对数组[排序](../../algorithms/sorting/Sorting.md)，再固定一个元素，通过[双指针](../../algorithms/two-pointers/TwoPointers.md)查找其余两个元素。排序会修改输入数组。

```go
import "sort"

func threeSum(nums []int) [][]int {
    n := len(nums)
    ans := [][]int{}
    sort.Ints(nums)
    for i := 0; i < n; i++ {
        if i != 0 && nums[i] == nums[i-1] {
            continue
        }
        k := n - 1
        for j := i + 1; j < k; j++ {
            if j != i+1 && nums[j] == nums[j-1] {
                continue
            }
            if nums[i]+nums[j]+nums[k] > 0 {
                k--
                j--
                continue
            }
            if j != k && nums[i]+nums[j]+nums[k] == 0 {
                ans = append(ans, []int{nums[i], nums[j], nums[k]})
                continue
            }
        }
    }
    return ans
}
```
