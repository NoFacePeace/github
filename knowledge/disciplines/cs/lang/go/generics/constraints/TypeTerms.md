# 类型项

约束接口可以包含具体类型项。例如 `interface{ int }` 的类型集只包含 int，用于将类型参数限制为指定类型。

## 1. 精确类型限制

```go
type OnlyInt interface {
    int
}

func Double[T OnlyInt](value T) T {
    return value + value
}

var result = Double(3)
```

这里 T 只能是 int。实际只处理一种固定类型时，普通 int 参数通常更直接；该写法主要用于理解类型项和组合约束。

## 2. 类型定义与类型别名

```go
type Count int       // 新定义类型，与 int 不同
type IntAlias = int  // 别名，与 int 是同一类型
```

Count 不满足 `interface{ int }`，IntAlias 则满足。需要同时支持 Count 等底层类型为 int 的类型时，应使用 `~int`。

## 3. 组合与限制

可以通过 `int | string` 允许多个精确类型，也可以用 `[]byte` 等复合类型作为类型项。

包含非接口类型项的接口只能作为约束，不能直接声明普通接口变量。类型项本身也不能直接是类型参数，例如不能用 `interface{ T }` 表达任意另一个类型参数的精确集合。
