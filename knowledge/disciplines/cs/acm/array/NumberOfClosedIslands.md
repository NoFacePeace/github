# 统计封闭岛屿的数目

## 1. 题目描述

[LeetCode 1254：统计封闭岛屿的数目](https://leetcode.cn/problems/number-of-closed-islands/)。

给定一个二维网格 `grid`，其中 `0` 表示陆地，`1` 表示水域。上下左右相连的陆地组成一座岛屿，求完全被水包围的岛屿数量。

一座岛屿是封闭岛屿，当且仅当它不包含任何位于网格边界上的陆地。对角相邻的陆地不视为连通。

```text
grid = [
  [1, 1, 1, 1, 1],
  [1, 0, 0, 1, 1],
  [1, 1, 1, 0, 1],
  [1, 1, 1, 1, 1]
]
输出：2
```

## 2. 解题思路

使用[深度优先搜索](../../algorithms/search/DFS.md)遍历每个陆地连通块，同时判断它是否接触网格边界。

1. 遍历网格，遇到尚未访问的陆地时发起一次 DFS。
2. 将当前陆地标记为已访问，防止重复访问。
3. 判断当前格子是否位于边界，并继续访问四个方向的陆地。
4. 只有当前格子和所有相连陆地都不接触边界，该连通块才是封闭岛屿，答案加一。

DFS 的返回值表示当前搜索分支是否满足封闭条件：越界返回 `false`；遇到水域或已访问格子返回 `true`，表示没有发现新的不封闭证据。已访问标记只负责去重，整个连通块是否封闭由本次搜索汇总判断。

## 3. Go 实现

下面假设网格为规则矩形，使用 `2` 标记已访问的陆地，因此会修改输入网格。

```go
func closedIsland(grid [][]int) int {
    m := len(grid)
    if m == 0 {
        return 0
    }
    n := len(grid[0])
    if n == 0 {
        return 0
    }

    var dfs func(int, int) bool
    dfs = func(i, j int) bool {
        if i < 0 || i >= m || j < 0 || j >= n {
            return false
        }
        if grid[i][j] != 0 {
            return true
        }

        grid[i][j] = 2
        closed := i > 0 && i < m-1 && j > 0 && j < n-1
        directions := [][2]int{{1, 0}, {-1, 0}, {0, -1}, {0, 1}}
        for _, d := range directions {
            branchClosed := dfs(i+d[0], j+d[1])
            closed = closed && branchClosed
        }
        return closed
    }

    answer := 0
    for i := 0; i < m; i++ {
        for j := 0; j < n; j++ {
            if grid[i][j] == 0 && dfs(i, j) {
                answer++
            }
        }
    }
    return answer
}
```

## 4. 关键细节

- 入栈后先标记陆地，再搜索邻居，避免在相邻格子之间无限递归。
- 即使已经发现边界陆地，也要继续遍历整个连通块，否则剩余陆地可能在后续扫描时被误当作新的岛屿。
- 不要写成 `closed = closed && dfs(...)`：当 `closed` 为 `false` 时，Go 的短路求值会跳过 DFS。先计算 `branchClosed` 可以确保所有分支都被访问。
- 如果不允许修改输入，可以使用独立的访问记录。

## 5. 复杂度

设网格有 M 行、N 列：

- 时间复杂度：O(MN)，每个陆地格子最多访问一次，每次只检查四个方向。
- 辅助空间复杂度：O(MN)，最坏情况下递归深度可达到陆地格子总数。

## 6. 其他解法

- 边界排除法：先从所有边界陆地出发，将与边界连通的陆地全部标记，再统计剩余陆地连通块的数量。遍历可以使用 DFS 或[广度优先搜索（BFS）](../../algorithms/search/BFS.md)。
- [并查集](../../data-structures/UnionFind.md)：为边界外设置一个虚拟节点，将边界陆地与它合并，同时合并相邻陆地，最后统计不属于虚拟节点所在集合的陆地集合数量。
