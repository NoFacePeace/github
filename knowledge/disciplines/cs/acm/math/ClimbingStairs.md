# 爬楼梯

相关知识：[动态规划（DP）](../../algorithms/dp/DP.md)。

## 1. 题目描述

[LeetCode 70：爬楼梯](https://leetcode.cn/problems/climbing-stairs/)。

假设正在爬楼梯，需要 `n` 阶才能到达楼顶。每次可以爬 1 阶或 2 阶，求到达楼顶的不同方法数量。

不同的步长顺序视为不同的方法。题目保证 `1 <= n <= 45`。

## 2. 示例

```text
输入：n = 2
输出：2
解释：有两种方法，分别为 1 + 1 和 2。
```

```text
输入：n = 3
输出：3
解释：有三种方法，分别为 1 + 1 + 1、1 + 2 和 2 + 1。
```

## 3. Go 实现

```go
func climbStairs(n int) int {
    one := 1
    two := 1
    for i := 2; i <= n; i++ {
        cur := one + two
        one = two
        two = cur
    }
    return two
}
```
