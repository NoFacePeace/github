# defer 与共享数据

return 赋值会复制返回值，但不会自动深拷贝其关联的数据。因此，即使没有具名返回值，defer 修改返回值共享的底层数据，也可能被调用方观察到。

## 1. 切片共享底层数组

```go
func Slice() []int {
    values := []int{1}
    defer func() {
        values[0] = 2
    }()
    return values
}
// 返回切片的第一个元素为 2。
```

返回结果复制切片描述信息，仍引用同一个底层数组。修改元素会影响返回结果，但只重设局部切片变量不会重设返回结果：

```go
func Reassign() []int {
    values := []int{1}
    defer func() {
        values = []int{2}
    }()
    return values
}
// 返回的切片仍包含 1。
```

append 还可能复用或更换底层数组，不应把局部 append 的结果误认为一定更新了返回切片的长度和容量。

## 2. map 关联同一份映射数据

```go
func Map() map[string]int {
    values := map[string]int{"count": 1}
    defer func() {
        values["count"] = 2
    }()
    return values
}
// 调用方看到 count 为 2。
```

修改键值会影响共享映射。把局部变量 values 重新赋为另一个 map，则不会改变已经设置的普通返回结果。

## 3. 指针关联同一个对象

```go
func Pointer() *int {
    value := 1
    defer func() {
        value = 2
    }()
    return &value
}
// 调用方解引用返回指针，得到 2。
```

返回的是指向 value 的指针，defer 修改的仍是该对象。返回局部变量的地址在 Go 中可以合法使用，存储位置和生命周期由编译器及运行时处理。

## 4. 按值复制的边界

数组和结构体按值返回时会复制元素或字段。若字段是指针、切片或 map，复制字段后仍可能共享关联数据。

判断结果时，应先区分 defer 修改的是哪个变量，再判断它与返回结果是否关联同一个可变对象，不能简单用“局部变量”或“具名返回值”概括全部情形。

## 5. 与并发的区别

这里的共享是指多个值关联同一份数据，不意味着示例自动启动了其他 G。defer 执行完毕、函数正常返回后，调用方才继续取得结果。

如果另有 G 同时访问这些数据，仍需要锁、channel 或其他正确同步；defer 本身不提供并发保护。
