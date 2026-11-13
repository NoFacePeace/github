# 滑动窗口最大值

## 1. 题目描述

[LeetCode 239：滑动窗口最大值](https://leetcode.cn/problems/sliding-window-maximum/)。

给定一个整数数组 `nums` 和一个窗口长度 `k`。窗口从数组最左侧开始，每次向右移动一个位置，返回每个窗口中的最大值。

每个窗口恰好包含 `k` 个连续元素，题目保证 `1 <= k <= nums.length`。返回结果的长度为 `nums.length - k + 1`。

## 2. 示例

```text
输入：nums = [1, 3, -1, -3, 5, 3, 6, 7], k = 3
输出：[3, 3, 5, 5, 6, 7]
```

各个窗口对应的最大值为：

| 窗口 | 最大值 |
| --- | --- |
| `[1, 3, -1]` | 3 |
| `[3, -1, -3]` | 3 |
| `[-1, -3, 5]` | 5 |
| `[-3, 5, 3]` | 5 |
| `[5, 3, 6]` | 6 |
| `[3, 6, 7]` | 7 |

```text
输入：nums = [1], k = 1
输出：[1]
```

## 3. Go 实现

使用[单调队列](../../data-structures/queue/MonotonicQueue.md)保存候选元素的下标，使对应值从队头到队尾非递增。窗口右移时，从队头移除过期下标，再从队尾移除比新元素小的候选，加入当前下标后，队头对应值就是窗口最大值。

```go
func maxSlidingWindow(nums []int, k int) []int {
    ans := []int{}
    n := len(nums)
    if n == 0 {
        return ans
    }
    q := []int{}
    for i := 0; i < min(k, n); i++ {
        for len(q) > 0 {
            if nums[i] <= nums[q[len(q)-1]] {
                break
            }
            q = q[:len(q)-1]
        }
        q = append(q, i)
    }
    ans = append(ans, nums[q[0]])
    for i := k; i < n; i++ {
        if q[0] <= i-k {
            q = q[1:]
        }
        for len(q) > 0 {
            if nums[i] <= nums[q[len(q)-1]] {
                break
            }
            q = q[:len(q)-1]
        }
        q = append(q, i)
        ans = append(ans, nums[q[0]])
    }
    return ans
}
```
