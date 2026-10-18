# 底层类型约束

`~T` 表示所有底层类型为 T 的类型，常用于让泛型算法同时支持基础类型和基于它定义的新类型。

## 1. 与精确类型项的区别

```go
type Count int

type Integer interface {
    ~int
}

func Double[T Integer](value T) T {
    return value + value
}

var result = Double(Count(3)) // result 的类型为 Count
```

`int` 类型项只允许 int；`~int` 还允许 Count 等底层类型为 int 的类型。运算后返回 T，因此保留自定义类型。

## 2. 复合类型

```go
func Clone[S ~[]E, E any](source S) S {
    if source == nil {
        return nil
    }
    result := make(S, len(source))
    copy(result, source)
    return result
}

type Names []string

var names = Clone(Names{"Alice", "Bob"}) // 返回 Names
```

`S ~[]E` 允许底层类型为 `[]E` 的切片类型，同时保留 S 的具体类型。示例只复制元素，元素引用的对象不会递归复制。

## 3. 语法限制

`~T` 中 T 的底层类型必须是它自身，且 T 不能是接口。因此，可以写 `~int`、`~[]byte`，不能对 `type Count int` 再写 `~Count`。

`~` 描述类型集合，不是数值近似，也不会改变值或自动执行类型转换。
