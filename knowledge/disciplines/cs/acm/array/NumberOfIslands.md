# 岛屿数量

## 1. 题目描述

[LeetCode 200：岛屿数量](https://leetcode.cn/problems/number-of-islands/)。

给定一个由字符 `'1'`（陆地）和 `'0'`（水域）组成的二维网格 `grid`，计算其中岛屿的数量。

一座岛屿由水平方向或竖直方向相邻的陆地连接形成，对角相邻的陆地不视为连通。可以假设网格的四条边均被水包围。

## 2. 示例

```text
输入：grid = [
  ["1", "1", "1", "1", "0"],
  ["1", "1", "0", "1", "0"],
  ["1", "1", "0", "0", "0"],
  ["0", "0", "0", "0", "0"]
]
输出：1
解释：所有陆地通过上下左右相邻关系连接成一座岛屿。
```

```text
输入：grid = [
  ["1", "1", "0", "0", "0"],
  ["1", "1", "0", "0", "0"],
  ["0", "0", "1", "0", "0"],
  ["0", "0", "0", "1", "1"]
]
输出：3
解释：左上角、中间和右下角的陆地分别组成一座岛屿。
```

## 3. Go 实现

使用[深度优先搜索（DFS）](../../algorithms/search/DFS.md)遍历陆地连通块。每遇到一块尚未访问的陆地，岛屿数量加一，再通过 DFS 标记与它上下左右相连的所有陆地，避免重复计数。

也可以使用[广度优先搜索（BFS）](../../algorithms/search/BFS.md)：用队列遍历每个陆地连通块，并在入队时标记已访问，避免同一格子重复入队。

```go
func numIslands(grid [][]byte) int {
    ans := 0
    m := len(grid)
    if m == 0 {
        return ans
    }
    n := len(grid[0])
    if n == 0 {
        return ans
    }
    visited := make([][]bool, m)
    for k := range visited {
        visited[k] = make([]bool, n)
    }
    var dfs func(i, j int)
    dfs = func(i, j int) {
        if i < 0 {
            return
        }
        if j < 0 {
            return
        }
        if i >= m {
            return
        }
        if j >= n {
            return
        }
        if grid[i][j] == '0' {
            return
        }
        if visited[i][j] {
            return
        }
        visited[i][j] = true
        dfs(i, j+1)
        dfs(i, j-1)
        dfs(i+1, j)
        dfs(i-1, j)
    }
    for i := 0; i < m; i++ {
        for j := 0; j < n; j++ {
            if grid[i][j] == '0' {
                continue
            }
            if visited[i][j] {
                continue
            }
            ans++
            dfs(i, j)
        }
    }
    return ans
}
```
