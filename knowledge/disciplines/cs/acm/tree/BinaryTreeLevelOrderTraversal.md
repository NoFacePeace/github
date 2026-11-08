# 二叉树的层序遍历

## 1. 题目描述

[LeetCode 102：二叉树的层序遍历](https://leetcode.cn/problems/binary-tree-level-order-traversal/)。

给定二叉树的根节点 `root`，返回其节点值的层序遍历结果，即逐层、从左到右访问所有节点。

结果使用二维数组表示，每个子数组保存同一层的节点值。空树返回空数组。

## 2. 示例

以下输入使用二叉树的层序序列化表示，`null` 表示空节点。

```text
输入：root = [3, 9, 20, null, null, 15, 7]
输出：[[3], [9, 20], [15, 7]]
解释：第一层为节点 3，第二层为节点 9、20，第三层为节点 15、7。
```

```text
输入：root = [1]
输出：[[1]]
```

```text
输入：root = []
输出：[]
```

## 3. Go 实现

使用[广度优先搜索（BFS）](../../algorithms/search/BFS.md)逐层遍历二叉树。每轮用 `tmp` 保存当前层的节点，遍历时依次记录节点值，并将非空的左右子节点加入下一层的队列 `q`。

```go
/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func levelOrder(root *TreeNode) [][]int {
    ans := [][]int{}
    if root == nil {
        return ans
    }
    q := []*TreeNode{root}
    cnt := 0
    for len(q) > 0 {
        tmp := q
        ans = append(ans, []int{})
        q = []*TreeNode{}
        for _, v := range tmp {
            if v == nil {
                continue
            }
            ans[cnt] = append(ans[cnt], v.Val)
            if v.Left != nil {
                q = append(q, v.Left)
            }
            if v.Right != nil {
                q = append(q, v.Right)
            }
        }
        cnt++
    }
    return ans
}
```
