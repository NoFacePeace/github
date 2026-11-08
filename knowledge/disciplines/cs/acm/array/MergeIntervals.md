# 合并区间

## 1. 题目描述

[LeetCode 56：合并区间](https://leetcode.cn/problems/merge-intervals/)。

给定一个区间数组 `intervals`，其中 `intervals[i] = [start, end]` 表示一个闭区间。合并所有重叠的区间，并返回一个不重叠的区间数组，恰好覆盖输入中的所有区间。

闭区间包含左右端点，因此 `[1, 4]` 和 `[4, 5]` 也视为重叠，可以合并为 `[1, 5]`。

## 2. 示例

```text
输入：intervals = [[1, 3], [2, 6], [8, 10], [15, 18]]
输出：[[1, 6], [8, 10], [15, 18]]
解释：[1, 3] 与 [2, 6] 重叠，合并为 [1, 6]。
```

```text
输入：intervals = [[1, 4], [4, 5]]
输出：[[1, 5]]
解释：两个闭区间在端点 4 处重叠。
```

## 3. Go 实现

先按区间左端点升序[排序](../../algorithms/sorting/Sorting.md)，再依次合并重叠区间；遇到不重叠的区间时，将当前合并结果加入答案。排序会修改输入区间的排列顺序。

```go
import "sort"

func merge(intervals [][]int) [][]int {
    sort.Slice(intervals, func(a, b int) bool {
        return intervals[a][0] < intervals[b][0]
    })
    ans := [][]int{}
    n := len(intervals)
    if n == 0 {
        return ans
    }
    left, right := intervals[0][0], intervals[0][1]
    for _, v := range intervals {
        if right < v[0] {
            ans = append(ans, []int{left, right})
            left, right = v[0], v[1]
            continue
        }
        right = max(right, v[1])
    }
    ans = append(ans, []int{left, right})
    return ans
}
```
