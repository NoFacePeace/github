# 课程表 II

## 1. 题目描述

[LeetCode 210：课程表 II](https://leetcode.cn/problems/course-schedule-ii/)。

共有 `numCourses` 门课程，编号为 `0` 到 `numCourses - 1`。给定先修关系数组 `prerequisites`，其中 `[a, b]` 表示学习课程 `a` 之前，必须先完成课程 `b`。

返回一个能够学完所有课程的合法学习顺序。如果存在多个答案，返回任意一个；如果无法完成所有课程，返回空数组。

## 2. 示例

```text
输入：numCourses = 4
      prerequisites = [[1,0], [2,0], [3,1], [3,2]]
输出：[0,1,2,3]
解释：先完成课程 0，再完成课程 1 和 2，最后完成课程 3。
      [0,2,1,3] 也是合法答案。
```

```text
输入：numCourses = 2
      prerequisites = [[0,1], [1,0]]
输出：[]
解释：课程 0 和 1 互相依赖，无法开始学习。
```

## 3. 解题思路

使用 [Kahn 算法](../../algorithms/topological-sort/Kahn.md)。将课程视为节点，先修关系 `[a, b]` 视为有向边 `b → a`。

- `m[a]` 记录课程 `a` 剩余的前置依赖数，即当前入度。
- `arr[b]` 保存直接依赖课程 `b` 的后继课程。
- `q` 保存当前可以学习的课程，即入度为 0 的节点。

先将所有没有前置依赖的课程入队。每完成一门课程，就把它的后继课程的入度减 1；后继课程的入度变为 0 时，将其加入队列。

下面的实现按批次处理队列，但将每批课程依次追加到一维数组 `ans` 中，所以返回的是拓扑顺序。不同批次的课程不能随意交换，同一批中的课程可以按任意顺序输出。

## 4. Go 实现

```go
func findOrder(numCourses int, prerequisites [][]int) []int {
    m := map[int]int{}
    arr := make([][]int, numCourses)
    for _, v := range prerequisites {
        a, b := v[0], v[1]
        m[a]++
        if arr[b] == nil {
            arr[b] = []int{}
        }
        arr[b] = append(arr[b], a)
    }

    q := []int{}
    for i := 0; i < numCourses; i++ {
        if m[i] == 0 {
            q = append(q, i)
        }
    }

    ans := []int{}
    for len(q) > 0 {
        tmp := q
        q = nil
        for _, b := range tmp {
            ans = append(ans, b)
            for _, a := range arr[b] {
                m[a]--
                if m[a] == 0 {
                    q = append(q, a)
                    delete(m, a)
                }
            }
        }
    }

    if len(m) != 0 {
        return []int{}
    }
    return ans
}
```

## 5. 实现细节

### 5.1 批次划分

`tmp := q` 保存当前批次，`q = nil` 让后续入队的课程进入下一批。切片赋值只复制切片描述符，不复制底层数组，因此 `tmp := q` 的时间复杂度为 O(1)。

若要输出可并行执行的批次，可以将返回类型改为 `[][]int`，每轮保存当前的 `tmp`，并单独统计已处理课程数以检测环。

### 5.2 使用 map 判环

初始化时，`m` 只包含有前置依赖的课程；没有对应键时，读取 `m[i]` 得到 0。当某门课程的所有前置依赖都满足后，代码通过 `delete(m, a)` 将其移除。

因此，结束时 `m` 中剩余的键都对应未满足的依赖。如果 `len(m) != 0`，说明存在环；剩余课程可能位于环内，也可能依赖环中的课程。

也可以用 `len(ans) != numCourses` 判断是否有环。采用这个条件后，不再需要 `delete(m, a)`。

### 5.3 可简化之处

- Go 的 nil 切片可以直接 `append`，因此可以删除 `if arr[b] == nil` 这一段初始化。
- 课程编号连续，可以使用 `make([]int, numCourses)` 代替 map 保存入度。
- 本题只要求返回一维顺序，无须特意分批；也可以通过递增队首下标，依次处理队列中的课程。

## 6. 复杂度分析

设 V 为课程数 `numCourses`，E 为先修关系数 `len(prerequisites)`，map 操作按平均 O(1) 计算。

- 时间复杂度：O(V + E)。构建邻接表需要 O(V + E)，初始化队列需要 O(V)，每门课程最多入队和出队一次，每条依赖边最多处理一次。
- 额外空间复杂度：O(V + E)。邻接表占用 O(V + E)，入度表、队列和结果数组各占用 O(V)。

循环虽然嵌套，但不同课程的邻接表合计只有 E 条边，分批也不会重复处理已完成的课程，因此遍历的复杂度不会相乘。
