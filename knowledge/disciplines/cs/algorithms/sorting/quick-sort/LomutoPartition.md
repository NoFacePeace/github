# Lomuto 划分

Lomuto 划分（Lomuto partition scheme）是一种以 Nico Lomuto 命名的数组分区方法，可用于快速排序和快速选择。常见实现以区间末尾元素为基准，从左向右扫描，最后将基准放到最终排序位置。

## 1. 核心思路

以升序排列、闭区间 `[left, right]` 为例：

1. 取 `nums[right]` 作为基准 `pivot`。
2. 扫描指针 `i` 和分界指针 `pos` 都从 `left` 开始。
3. 从左向右扫描到 `right-1`；遇到小于基准的元素，将其交换到 `pos`，再递增 `pos`。
4. 扫描结束后，交换 `nums[pos]` 与末尾的基准，返回 `pos`。

`pos` 指向小于基准区域之后的第一个位置，而不是该区域最后一个元素的位置。扫描范围不包含末尾的基准。

## 2. Go 实现

```go
func lomutoPartition(nums []int, left, right int) int {
    pivot := nums[right]
    pos := left

    for i := left; i < right; i++ {
        if nums[i] < pivot {
            nums[pos], nums[i] = nums[i], nums[pos]
            pos++
        }
    }

    nums[pos], nums[right] = nums[right], nums[pos]
    return pos
}
```

调用前要求 `0 <= left <= right < len(nums)`。函数直接修改数组，返回基准在当前区间内的最终排序下标。

## 3. 划分过程与结果

每次扫描开始时维护：

```text
[left, pos)  小于 pivot
[pos, i)     大于等于 pivot
[i, right)   尚未扫描
right        保存 pivot
```

例如对 `[3, 1, 4, 2]` 划分：

```text
基准：2
扫描 3：不交换
扫描 1：[1, 3, 4, 2]，pos 移到下标 1
扫描 4：不交换
基准归位：[1, 2, 4, 3]，返回下标 1
```

结束后，设返回值为 `p`：

```text
[left, p)   小于 nums[p]
p           基准的最终位置
(p, right]  大于等于 nums[p]
```

两侧内部仍可能无序。存在重复值时，基准占据的是一个合法的最终位置，相等元素的相对顺序不保证保留。

## 4. 如何使用返回值

- 快速排序：继续排序 `[left, p-1]` 和 `[p+1, right]`，跳过已经归位的基准。
- 快速选择：目标升序下标等于 `p` 时直接返回；小于 `p` 时继续左侧，大于 `p` 时继续右侧。

如果要从中间或随机位置选基准，可以先将选中的元素交换到 `right`，再调用上面的函数。移动到末尾是这份实现的约定，不是所有划分方法的要求。

## 5. 复杂度与注意事项

一次划分的时间复杂度为 O(m)，额外空间为 O(1)，其中 m 为当前区间长度。

固定选择末尾基准时，有序输入容易产生极端不均衡的划分。由于使用严格小于判断，全相等数组每次都返回 `left`，仅随机选择基准也无法改善这种退化；重复值较多时可以考虑三路划分。
