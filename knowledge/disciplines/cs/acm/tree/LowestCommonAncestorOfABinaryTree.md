# 二叉树的最近公共祖先

## 1. 题目描述

[LeetCode 236：二叉树的最近公共祖先](https://leetcode.cn/problems/lowest-common-ancestor-of-a-binary-tree/)。

给定一棵二叉树的根节点 `root`，以及树中的两个不同节点 `p` 和 `q`，返回它们的最近公共祖先节点。

最近公共祖先是同时以 `p` 和 `q` 为后代的最深节点。一个节点也可以是它自己的后代，因此最近公共祖先可以是 `p` 或 `q` 本身。

题目保证所有节点的值互不相同，且 `p` 和 `q` 均存在于树中。

## 2. 示例

以下输入使用二叉树的层序序列化表示，`null` 表示空节点；`p` 和 `q` 以对应节点的值展示，实际参数为节点引用。

```text
输入：root = [3, 5, 1, 6, 2, 0, 8, null, null, 7, 4], p = 5, q = 1
输出：3
解释：节点 5 和节点 1 的最近公共祖先是节点 3。
```

```text
输入：root = [3, 5, 1, 6, 2, 0, 8, null, null, 7, 4], p = 5, q = 4
输出：5
解释：节点 5 是节点 4 的祖先，也是它自己的祖先。
```

```text
输入：root = [1, 2], p = 1, q = 2
输出：1
```

## 3. Go 实现

使用[递归](../../algorithms/basics/Recursion.md)先处理左右子树，再结合当前节点是否为 `p` 或 `q` 判断最近公共祖先。找到答案后将其保存在 `ans` 中，并返回 `false`，避免祖先节点重复更新答案。

```go
/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func lowestCommonAncestor(root, p, q *TreeNode) *TreeNode {
    var ans *TreeNode
    var dfs func(root *TreeNode) bool
    dfs = func(root *TreeNode) bool {
        if root == nil {
            return false
        }
        left := dfs(root.Left)
        right := dfs(root.Right)
        if left && right {
            ans = root
            return false
        }
        cur := root == p || root == q
        if cur && (left || right) {
            ans = root
            return false
        }
        return cur || left || right
    }
    dfs(root)
    return ans
}
```
