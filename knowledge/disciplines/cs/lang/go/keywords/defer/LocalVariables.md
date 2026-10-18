# defer 与局部变量

普通返回值场景中，return 先计算表达式并设置返回结果，再执行 defer。defer 修改原来的局部变量本身，不会重新计算已经设置的返回结果。

## 1. 修改局部变量不改变已设置的结果

```go
func Example() int {
    value := 1
    defer func() {
        value = 2
    }()
    return value
}
// 返回 1。
```

执行过程为：

```text
计算 return value，得到 1
    ↓
将 1 赋给返回结果
    ↓
defer 将局部变量 value 改为 2
    ↓
调用方收到已设置的结果 1
```

局部变量 value 与函数的返回结果是不同的变量。即使编译器优化了具体实现，语言层面的行为仍遵守这个顺序。

## 2. 参数求值与闭包读取

```go
func Print() {
    value := 1
    defer fmt.Println(value)
    defer func() {
        fmt.Println(value)
    }()
    value = 2
}
// 导入 fmt 后，依次输出 2、1。
```

`defer fmt.Println(value)` 在登记时保存参数 1；闭包在退出时读取外部变量，此时 value 已是 2。两个 defer 仍按照后进先出执行。

## 3. 不能据此推断深拷贝

上述规则针对变量值的赋值。返回切片、map 或指针时，复制后的值仍可能关联同一份数据；defer 对共享内容的修改可能被调用方观察到。
