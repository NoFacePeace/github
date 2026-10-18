# 有序类型约束

标准库 `cmp.Ordered` 从 Go 1.21 开始提供，允许支持 `<`、`<=`、`>`、`>=` 的整数、浮点数和字符串类型，也包含底层类型为这些类型的自定义类型。

## 1. 基本用法

```go
import "cmp"

func Less[T cmp.Ordered](a, b T) bool {
    return a < b
}

var numberLess = Less(1, 2)
var textLess = Less("apple", "banana")
```

bool、复数、指针、切片和结构体等不满足该约束。自定义类型拥有 Compare 方法，也不会自动满足 Ordered；它仍必须符合类型集要求。

## 2. 与 comparable 的区别

comparable 保证可以进行相等比较，不保证支持大小比较。例如指针可比较是否相等，但不能用 `<` 排序。

Ordered 中的类型也可进行相等比较，但这不表示相等或排序行为一定符合全部业务需求。例如字符串按字节顺序比较，不是语言学排序。

## 3. 浮点数的注意事项

浮点 NaN 与任何值进行 `<`、`>`、`==` 比较都为 false，因此不能把普通浮点运算符直接当作数学上的全序。

需要统一处理 NaN 的排序语义时，可以考虑 `cmp.Compare` 或 `cmp.Less`：它们把 NaN 视为小于非 NaN，且 `cmp.Compare` 将两个 NaN 视为相等；这与直接使用运算符的结果不同。
