# 接雨水

## 1. 题目描述

[LeetCode 42：接雨水](https://leetcode.cn/problems/trapping-rain-water/)。

给定一个由非负整数组成的数组 `height`，其中每个元素表示对应位置柱子的高度，每根柱子的宽度为 1。计算按此排列的柱子在下雨后能够接住多少雨水。

雨水只能积存在左右两侧柱子围成的凹槽中，没有两侧阻挡的位置无法积水。返回能够接住的雨水总量。

## 2. 示例

```text
输入：height = [0, 1, 0, 2, 1, 0, 1, 3, 2, 1, 2, 1]
输出：6
解释：各处凹槽接住的雨水总量为 6。
```

```text
输入：height = [4, 2, 0, 3, 2, 5]
输出：9
```

## 3. Go 实现

下面的实现通过从左向右、再从右向左扫描，分别累计具有足够高边界的凹槽水量。

```go
func trap(height []int) int {
    n := len(height)
    if n == 0 {
        return 0
    }
    ans := 0
    mx := height[0]
    sum := 0
    for _, v := range height {
        if v < mx {
            sum += mx - v
            continue
        }
        ans += sum
        sum = 0
        mx = v
    }
    left := mx
    mx = height[n-1]
    sum = 0
    for i := n - 1; i >= 0; i-- {
        v := height[i]
        if mx >= left {
            break
        }
        if v <= mx {
            sum += mx - v
            continue
        }
        mx = v
        ans += sum
        sum = 0
    }
    ans += sum
    return ans
}
```

## 4. 其他解法

- [动态规划（DP）](../../algorithms/dp/DP.md)：预先计算每个位置左侧和右侧的最高柱子高度（包含当前位置），该位置水量为两者较小值减去当前高度。时间复杂度为 O(n)，辅助空间复杂度为 O(n)。
- [双指针](../../algorithms/two-pointers/TwoPointers.md)：从两端向中间扫描，维护两侧已扫描部分的最高柱子，每次根据较低的一侧确定当前位置的水量。时间复杂度为 O(n)，辅助空间复杂度为 O(1)。
- [单调栈](../../data-structures/stack/MonotonicStack.md)：维护柱子高度非递增的下标栈，遇到更高柱子时弹出凹槽底部，结合新的栈顶和当前柱子计算新增水量。时间复杂度为 O(n)，辅助空间复杂度为 O(n)。
