# Kahn 算法

Kahn 算法以 Arthur B. Kahn 的姓氏命名，是一种基于入度的拓扑排序算法。它反复选取没有未完成前置依赖的节点，加入结果，并移除该节点对后继节点的依赖约束。

## 1. 核心概念与步骤

对于有向边 `u → v`，约定先完成 `u`，才能执行 `v`。

- 邻接表 `graph[u]`：保存所有直接依赖 `u` 的后继节点。
- 入度 `indegree[v]`：初始化时为指向 `v` 的边数；处理过程中表示 `v` 还剩多少条未满足的前置依赖。
- 队列 `queue`：保存入度为 0、可以执行的节点。

算法步骤：

1. 遍历所有边，构建邻接表和入度表。
2. 遍历所有节点，将入度为 0 的节点加入队列。
3. 从队列取出节点 `u`，加入结果。
4. 遍历 `graph[u]`，将每个后继节点 `v` 的入度减 1；如果变成 0，则将 `v` 加入队列。
5. 重复处理，直到队列为空。
6. 如果处理节点数小于节点总数，说明存在环，无法得到覆盖全部节点的拓扑序。

这里只需更新入度，不必真的从邻接表中删除边。只有在入度变成 0 时才入队，因此每个节点最多入队一次。

## 2. 执行示例

```text
A → C → D
B → C
```

| 阶段 | 可执行节点 | 依赖变化 |
| --- | --- | --- |
| 初始化 | A、B | C 的入度为 2，D 的入度为 1 |
| 完成 A、B | C | C 的入度从 2 降为 0 |
| 完成 C | D | D 的入度从 1 降为 0 |
| 完成 D | 无 | 全部节点处理完成 |

每次取出的节点都没有未满足的前置依赖，因此追加到结果中不会违反依赖关系。如果队列为空但仍有未处理节点，剩余节点都有来自剩余图的入边；沿前驱不断追溯必然遇到重复节点，因此剩余图中存在有向环。未处理节点也可能只是依赖环的下游节点，并非每个节点都在环内。

## 3. Go 实现：输出拓扑顺序

以 LeetCode 210 课程表 II 为例：`prerequisites[i] = [a, b]` 表示先学习 `b`，再学习 `a`，对应边 `b → a`。注意，数组中两个元素的顺序与边的方向相反。

以下实现假设课程编号有效，范围为 `[0, numCourses)`。

```go
func findOrder(numCourses int, prerequisites [][]int) []int {
    graph := make([][]int, numCourses)
    indegree := make([]int, numCourses)

    for _, prerequisite := range prerequisites {
        a, b := prerequisite[0], prerequisite[1]
        graph[b] = append(graph[b], a)
        indegree[a]++
    }

    queue := make([]int, 0, numCourses)
    for course := 0; course < numCourses; course++ {
        if indegree[course] == 0 {
            queue = append(queue, course)
        }
    }

    order := make([]int, 0, numCourses)
    for head := 0; head < len(queue); head++ {
        course := queue[head]
        order = append(order, course)

        for _, next := range graph[course] {
            indegree[next]--
            if indegree[next] == 0 {
                queue = append(queue, next)
            }
        }
    }

    if len(order) != numCourses {
        return []int{}
    }
    return order
}
```

连续编号适合用切片保存入度。Go 的 nil 切片可以直接 `append`，因此不必先为 `graph[b]` 创建空切片。判环只需检查处理数量，不需要删除入度表中的元素。

## 4. 按批次输出可并行任务

每轮开始时固定当前队列的结束位置，只处理这一批节点；本轮新解锁的节点留到下一批。将上面初始化队列之后的遍历与返回部分替换为以下代码，并将函数返回类型改为 `[][]int`，即可输出批次；有环时返回 `nil`。

```go
batches := make([][]int, 0)
processed := 0

for head := 0; head < len(queue); {
    end := len(queue)
    batch := make([]int, 0, end-head)

    for head < end {
        course := queue[head]
        head++
        batch = append(batch, course)
        processed++

        for _, next := range graph[course] {
            indegree[next]--
            if indegree[next] == 0 {
                queue = append(queue, next)
            }
        }
    }
    batches = append(batches, batch)
}

if processed != numCourses {
    return nil
}
return batches
```

第 2 节的图得到 `[[A, B], [C], [D]]`，同批节点可以并行执行。同一批中的节点顺序不影响依赖约束。

当每个任务占用一个批次、并行数量不限、依赖必须在之前的批次完成时，这种分层方式将每个节点安排在最早可执行的批次，批次数最少，等于 DAG 中最长路径的节点数。它与普通 BFS 的最短距离分层含义不同：节点必须等所有前驱都完成后才能入队。

任务耗时不同时，等待整批完成可能使部分后继启动过晚；若要实现异步调度，应在每个任务实际完成时更新后继入度。并行数量受限时，仅按当前队列分批也不保证总批次数最少。

## 5. 复杂度

设 V 为节点数，E 为边数，使用邻接表表示图。

| 操作 | 时间复杂度 |
| --- | --- |
| 构建邻接表和入度表 | O(V + E) |
| 初始化零入度队列 | O(V) |
| 处理节点和依赖边 | O(V + E) |
| 合计 | O(V + E) |

虽然遍历中存在嵌套循环，但每个节点最多入队、出队一次，每条边最多访问一次，因此复杂度不会相乘。分批版本同样如此。

- 包含新建邻接表的额外空间：O(V + E)。
- 如果邻接表由调用方提供，算法自身的入度表、队列和结果占用 O(V) 空间。
- 使用 map 保存入度时，上述时间分析依赖哈希表操作的平均 O(1) 开销。
- `tmp := queue; queue = nil` 也可以隔离当前批次与下一批次；切片赋值只复制切片描述符，不复制底层数组。

## 6. 对应练习

- [LeetCode 210：课程表 II](https://leetcode.cn/problems/course-schedule-ii/)：输出任意一个合法拓扑序。
- [LeetCode 1136：并行课程](https://leetcode.cn/problems/parallel-courses/)：求无限并行条件下的最少学期数，可用按批次处理的方式求解。
