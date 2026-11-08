# 全排列

## 1. 题目描述

[LeetCode 46：全排列](https://leetcode.cn/problems/permutations/)。

给定一个不含重复元素的整数数组 `nums`，返回其所有可能的全排列。答案可以按任意顺序返回。

每个排列必须包含原数组中的全部元素，且每个元素恰好使用一次。题目保证 `1 <= nums.length <= 6`。

## 2. 示例

```text
输入：nums = [1, 2, 3]
输出：[[1, 2, 3], [1, 3, 2], [2, 1, 3], [2, 3, 1], [3, 1, 2], [3, 2, 1]]
解释：三个不同元素共有 6 种排列。
```

```text
输入：nums = [0, 1]
输出：[[0, 1], [1, 0]]
```

```text
输入：nums = [1]
输出：[[1]]
```

## 3. Go 实现

使用[回溯](../../algorithms/search/Backtracking.md)逐步构造排列：每次选择一个尚未使用的元素加入路径，递归处理下一位，再撤销访问标记和路径中的选择。路径长度达到数组长度时，复制当前路径并保存，避免后续修改影响已记录的答案。

```go
func permute(nums []int) [][]int {
    visited := map[int]bool{}
    var f func(arr []int)
    ans := [][]int{}
    f = func(arr []int) {
        if len(arr) == len(nums) {
            tmp := append([]int{}, arr...)
            ans = append(ans, tmp)
            return
        }
        for _, v := range nums {
            if visited[v] {
                continue
            }
            visited[v] = true
            arr = append(arr, v)
            f(arr)
            visited[v] = false
            arr = arr[:len(arr)-1]
        }
    }
    f([]int{})
    return ans
}
```
