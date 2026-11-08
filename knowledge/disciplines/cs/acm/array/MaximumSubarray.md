# 最大子数组和

## 1. 题目描述

[LeetCode 53：最大子数组和](https://leetcode.cn/problems/maximum-subarray/)。

给定一个整数数组 `nums`，找出一个具有最大和的连续子数组，并返回其最大和。

子数组至少包含一个元素，且必须是原数组中连续的一段元素。题目保证 `nums` 非空。

## 2. 示例

```text
输入：nums = [-2, 1, -3, 4, -1, 2, 1, -5, 4]
输出：6
解释：连续子数组 [4, -1, 2, 1] 的和最大，为 6。
```

```text
输入：nums = [1]
输出：1
```

```text
输入：nums = [5, 4, -1, 7, 8]
输出：23
解释：整个数组的和最大，为 23。
```

## 3. Go 实现

使用[动态规划（DP）](../../algorithms/dp/DP.md)：以当前元素结尾的最大子数组和，等于当前元素与“前一个位置的状态加上当前元素”两者中的较大值。代码用 `sum` 维护累加和，每轮先更新答案，再将负和清零，使后续元素可以重新开始一个子数组。

```go
import "math"

func maxSubArray(nums []int) int {
    ans := math.MinInt
    n := len(nums)
    sum := 0
    for i := 0; i < n; i++ {
        sum += nums[i]
        ans = max(ans, sum)
        if sum < 0 {
            sum = 0
        }
    }
    return ans
}
```
