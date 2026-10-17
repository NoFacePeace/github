# 数位和排序需要的最小交换次数

相关知识：[排序](../../algorithms/sorting/Sorting.md)、[置换环](../../data-structures/graph/PermutationCycle.md)。排序确定每个元素的目标位置后，通过置换环分解计算最少交换次数，答案为元素数量减去环的数量。

```go
func minSwaps(nums []int) int {
    n := len(nums)
    digitSum := make([]int, n)
    for i, v := range nums {
        sum := 0
        for v > 0 {
            sum += v % 10
            v /= 10
        }
        digitSum[i] = sum
    }
    indices := make([]int, n)
    for i := 0; i < n; i++ {
        indices[i] = i
    }
    sort.Slice(indices, func(a, b int) bool {
        if digitSum[indices[a]] != digitSum[indices[b]] {
            return digitSum[indices[a]] < digitSum[indices[b]]
        }
        return nums[indices[a]] < nums[indices[b]]
    })
    targetPos := make([]int, n)
    for pos, idx := range indices {
        targetPos[idx] = pos
    }
    visited := make([]bool, n)
    cycles := 0
    for i := 0; i < n; i++ {
        if !visited[i] {
            cycles++
            j := i
            for !visited[j] {
                visited[j] = true
                j = targetPos[j]
            }
        }
    }
    return n - cycles
}
```
