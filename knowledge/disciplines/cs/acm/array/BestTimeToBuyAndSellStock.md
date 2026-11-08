# 买卖股票的最佳时机

## 1. 题目描述

[LeetCode 121：买卖股票的最佳时机](https://leetcode.cn/problems/best-time-to-buy-and-sell-stock/)。

给定一个数组 `prices`，其中 `prices[i]` 表示一支股票在第 `i` 天的价格。

最多选择一天买入，并在之后的某一天卖出这支股票，计算能够获得的最大利润。如果无法获得利润，则返回 `0`。

买入必须发生在卖出之前，最多完成一次买卖交易，也可以不进行交易。

## 2. 示例

```text
输入：prices = [7, 1, 5, 3, 6, 4]
输出：5
解释：以价格 1 买入，再以之后的价格 6 卖出，利润为 6 - 1 = 5。
```

```text
输入：prices = [7, 6, 4, 3, 1]
输出：0
解释：价格持续下降，不进行交易，最大利润为 0。
```

## 3. Go 实现

使用[贪心](../../algorithms/greedy/Greedy.md)思想：固定卖出日时，选择此前的最低价格买入最优。代码用 `mn` 记录当前最低价格，`mx` 记录从该最低价出现后遍历到的最高价格；遇到更低价格时，先更新已有区间的最大利润，再重置两者，遍历结束后补算最后一个区间。

```go
func maxProfit(prices []int) int {
    n := len(prices)
    if n == 0 {
        return 0
    }
    ans := 0
    mn := prices[0]
    mx := prices[0]
    for _, v := range prices {
        if v < mn {
            ans = max(ans, mx-mn)
            mx = v
            mn = v
            continue
        }
        mx = max(v, mx)
    }
    ans = max(ans, mx-mn)
    return ans
}
```
