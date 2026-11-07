# 快速选择

快速选择（Quickselect）是一种查找顺序统计量的算法，例如第 k 小或第 k 大的元素。它通过基准划分区间，每次只继续处理目标所在的一侧，不需要将整个数组排好序。

## 1. 目标与基本步骤

将目标统一表示为最终升序排列后的下标 `target`，下标从 0 开始：

- 第 k 小：`target = k - 1`。
- 第 k 大：`target = len(nums) - k`。

这里的 k 从 1 开始，重复元素分别占据排名。例如 `[5, 5, 3]` 的第 2 大是 `5`，不是第 2 个不同的值 `3`。

基本步骤：

1. 在当前区间选取基准并划分。
2. 根据划分结果，判断目标位置是否已经确定。
3. 如果尚未确定，只保留目标所在的一侧，重复划分。

快速排序需要分别处理两侧；快速选择每次只继续处理其中一侧。

## 2. 使用 Lomuto 划分

[Lomuto 划分](LomutoPartition.md)返回基准的最终位置 `p`，因此可以直接比较目标下标与 `p`。

下面调用该文档中的 `lomutoPartition` 函数。所有实现都要求数组非空，且 `0 <= target < len(nums)`，并直接修改原数组。

```go
func QuickSelectLomuto(nums []int, target int) int {
    left, right := 0, len(nums)-1

    for left < right {
        p := lomutoPartition(nums, left, right)
        if target == p {
            return nums[p]
        }
        if target < p {
            right = p - 1
        } else {
            left = p + 1
        }
    }
    return nums[left]
}
```

例如，在 `[3, 2, 1, 5, 4]` 中找第 2 大，`target = 3`。以末尾的 `4` 划分后得到 `[3, 2, 1, 4, 5]`，基准位置 `p = 3`，直接返回 `4`。左侧无须继续排序。

## 3. 使用 Hoare 划分

[Hoare 划分](HoarePartition.md)返回分区边界 `b`，不保证基准归位。下面调用该文档中返回单个边界的 `hoarePartition` 函数。

```go
func QuickSelectHoare(nums []int, target int) int {
    left, right := 0, len(nums)-1

    for left < right {
        boundary := hoarePartition(nums, left, right)
        if target <= boundary {
            right = boundary
        } else {
            left = boundary + 1
        }
    }
    return nums[left]
}
```

每轮保留的区间都包含 `target`。当 `left == right` 时，它们必然也等于 `target`，该位置的值就是答案。

不能因为 `target == boundary` 就直接返回，因为边界元素仍可能需要在左区间内继续调整。收缩区间时保留原数组下标，因此 `target` 不变；如果改用切片递归，则需要同步换算目标在子切片中的下标或排名。

## 4. 何时算找到

| 划分方式 | 划分结果 | 可以返回的条件 |
| --- | --- | --- |
| Lomuto | 基准最终位置 `p` | `target == p`，或区间只剩一个元素 |
| 经典 Hoare | 左右分区边界 `b` | 持续缩小区间，直到只剩一个元素 |
| 三路划分 | 小于、等于、大于基准的三个区间 | `target` 落入等于基准的区间 |

快速选择不要求专门交换基准。是否将基准归位由划分方案决定；判断依据是目标位置是否已确定，而不是两个扫描指针是否相等。

## 5. 复杂度与优化

- 平均时间复杂度：在互异元素的随机排列等常见分析条件下为 O(n)。
- 最坏时间复杂度：O(n²)，例如每次都只排除极少元素。
- 额外空间复杂度：上面的迭代实现为 O(1)；递归实现还需要计算调用栈空间。

划分较均衡时，工作量约为 `n + n/2 + n/4 + ...`，总和为 O(n)。快速排序则需要处理每层的全部区间，因此平均为 O(n log n)。

上面的示例使用固定的基准选择规则，不保证每次划分均衡。随机选择基准可以降低特定输入顺序导致持续退化的概率；对于 Lomuto 示例，需要先把随机选中的元素交换到区间末尾。

大量重复元素可能导致上述 Lomuto 实现退化，仅随机化基准不能解决全相等输入。可以使用三路划分，一次排除整个等值区间；随机基准配合三路划分，可在包含重复值的一般输入上获得期望 O(n) 的时间复杂度。若要求最坏 O(n)，需要使用中位数的中位数等确定性选择策略。
