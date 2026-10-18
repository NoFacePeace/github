# new

`new` 是预声明内建函数，不是关键字。传统的 `new(T)` 创建一个 T 类型变量，将其初始化为零值，并返回 `*T`。Go 1.26 起还支持表达式参数形式，用表达式的值初始化新变量。

## 1. 类型参数形式

```go
p := new(int)
// p 的类型为 *int，*p == 0。

type User struct {
    Name string
}

u := new(User)
// u 的类型为 *User，u.Name == ""。
```

可以把语义理解为声明一个零值变量后取得它的地址。返回指针不代表一定发生堆分配，存储位置由逃逸分析等决定。

## 2. 指向零值与 nil 指针

```go
var absent *int // absent == nil
present := new(int) // present != nil，指向一个值为 0 的 int
```

nil 指针没有指向可直接解引用的对象，而 new(int) 返回可使用的 int 指针。

如果 T 本身是指针，new(T) 会多一层指针：`new(*int)` 的类型是 `**int`，其指向的 `*int` 变量初始为 nil。

## 3. map、切片和 channel 的零值

```go
pointer := new(map[string]int)
// pointer 不为 nil，但 *pointer 是 nil map。

// (*pointer)["go"] = 1 // panic：向 nil map 写入。
*pointer = make(map[string]int)
(*pointer)["go"] = 1
```

new([]int) 指向 nil 切片，new(chan int) 指向 nil channel。它们不会自动完成 make 的初始化工作。通常直接用 make 创建这些类型更清楚。

## 4. 表达式参数形式

Go 1.26 起，可以直接为表达式结果创建变量并取得指针：

```go
p := new(42)
// p 的类型为 *int，*p == 42。

text := new("hello")
// text 的类型为 *string，*text == "hello"。
```

这不是对已有变量直接取地址，而是创建新变量并用表达式结果初始化。表达式求值和初始化仍遵守类型规则；旧版本 Go 不支持该语法，项目语言版本也需要允许使用它。

## 5. 使用选择

| 需求 | 常见写法 |
| --- | --- |
| 创建零值对象并取得指针 | `new(T)` |
| 创建带字段初值的结构体指针 | `&User{Name: "Alice"}` |
| 对已有可寻址变量取地址 | `&value` |
| 创建可写 map 或可用 channel | `make(...)` |

new 不负责并发同步，也不保证内部引用的对象被深拷贝。表达式初始化复制的是表达式的值，其中的指针、切片、map 等仍可能关联共享数据。
