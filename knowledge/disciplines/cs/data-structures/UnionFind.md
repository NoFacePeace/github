# 并查集

并查集（Union-Find，也称 Disjoint Set Union，DSU）是一种维护多个互不相交集合的数据结构，用于动态合并集合，以及判断两个元素是否属于同一个集合。

## 1. 核心概念

并查集维护的是元素的分组关系：每个元素恰好属于一个集合，各集合之间没有公共元素。它通常支持两种操作：

- 查找（Find）：返回元素所在集合的代表元素。
- 合并（Union）：将两个元素所在的集合合并。

判断两个元素是否属于同一集合，只需比较它们的代表元素是否相同。

普通集合关注元素的添加、删除和成员判断；并查集关注分组关系的维护，通常不直接提供集合拆分或高效枚举组内元素的操作。

## 2. 表示方式

常见实现使用一片森林表示所有集合，每棵树对应一个集合：

- `parent[x]` 保存元素 `x` 的父节点。
- 根节点满足 `parent[root] == root`，作为所在集合的代表元素。
- 初始时，每个元素各自属于一个集合，父节点都是自己。

例如，初始状态为 `{A}、{B}、{C}、{D}`：

1. 合并 A 和 B，得到 `{A, B}、{C}、{D}`。
2. 合并 C 和 D，得到 `{A, B}、{C, D}`。
3. 合并 B 和 C，得到 `{A, B, C, D}`。

代表元素由具体实现和合并顺序决定，不一定是集合中最小的元素，也不保证在合并后保持不变。

## 3. 操作与优化

### 3.1 查找

沿父节点不断向上查找，直到到达根节点。若树退化为链，单次查找最坏需要 O(N) 时间。

### 3.2 合并

先找到两个元素各自的根节点：

- 根节点相同，说明它们已经属于同一个集合，无需再次合并。
- 根节点不同，将其中一个根节点挂到另一个根节点下。

必须合并根节点，而不是随意修改某个非根节点的父节点，否则可能破坏已有的分组关系。

### 3.3 路径压缩

查找根节点后，将查找路径上的节点直接指向根节点，降低后续查找的路径长度。路径压缩只改变树的形状，不改变元素所属的集合。

### 3.4 按大小合并

记录每个根节点对应集合的元素数量，将较小集合的根节点挂到较大集合的根节点下，避免树过高。

另一种常见方法是按秩合并，用秩指导合并方向。配合路径压缩后，秩不再等同于树的实际高度。

## 4. Go 实现

下面使用路径压缩和按大小合并。元素编号范围为 `[0, n)`，假设 `n` 非负，且所有操作的元素编号均有效。

```go
type UnionFind struct {
    parent []int
    size   []int
    count  int
}

func NewUnionFind(n int) *UnionFind {
    uf := &UnionFind{
        parent: make([]int, n),
        size:   make([]int, n),
        count:  n,
    }
    for i := 0; i < n; i++ {
        uf.parent[i] = i
        uf.size[i] = 1
    }
    return uf
}

func (uf *UnionFind) Find(x int) int {
    root := x
    for uf.parent[root] != root {
        root = uf.parent[root]
    }
    for x != root {
        next := uf.parent[x]
        uf.parent[x] = root
        x = next
    }
    return root
}

func (uf *UnionFind) Union(a, b int) bool {
    rootA, rootB := uf.Find(a), uf.Find(b)
    if rootA == rootB {
        return false
    }
    if uf.size[rootA] < uf.size[rootB] {
        rootA, rootB = rootB, rootA
    }
    uf.parent[rootB] = rootA
    uf.size[rootA] += uf.size[rootB]
    uf.count--
    return true
}

func (uf *UnionFind) Connected(a, b int) bool {
    return uf.Find(a) == uf.Find(b)
}

func (uf *UnionFind) Count() int {
    return uf.count
}
```

`Union` 返回是否实际合并了两个集合；只有合并成功时，集合数量才减一。`size` 只在根节点处有效，查询元素所在集合的大小时应读取 `uf.size[uf.Find(x)]`。

## 5. 复杂度

对于 N 个元素：

- 初始化：O(N) 时间。
- 查找与合并：同时使用路径压缩和按大小或按秩合并时，每次操作的摊还时间复杂度为 O(α(N))，其中 α 为反阿克曼函数，增长极慢，实际使用中接近常数。
- 查询集合数量：O(1) 时间。
- 存储空间：O(N)。

摊还复杂度描述一系列操作的平均成本，不代表每一次查找都严格为 O(1)。

## 6. 适用场景与限制

- 无向图的连通性判断：将每条边的两个端点合并，再判断任意两个顶点是否属于同一集合。
- 无向图的环检测：加入一条边之前，若两个端点已经连通，则这条边会形成环。
- Kruskal 最小生成树算法：判断候选边是否连接两个不同的连通分量。
- 等价关系分组：根据“属于同一组”的关系合并元素。
- 网格区域合并：将相邻且符合条件的格子合并，用于统计连通区域。

普通并查集适合不断合并的场景，不支持任意删除边或拆分集合，也不直接保存图中的实际连接路径。有向图的可达性不能仅靠并查集判断；需要撤销合并时，可以考虑可回滚并查集等扩展结构。
