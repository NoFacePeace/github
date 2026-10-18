# make

`make` 是预声明内建函数，不是关键字，用于创建并初始化切片、map 和 channel。它返回对应类型的值，不是统一返回指针。

## 1. 创建切片

```go
values := make([]int, 3, 5)
// len(values) == 3，cap(values) == 5，三个元素均为 0。
```

第二个参数是长度，第三个参数是可选容量，省略容量时与长度相同。容量不能小于长度，长度和容量不能为负；编译期可确定的非法参数会被拒绝，执行时才确定的非法值会导致 panic。

`make([]int, 0)` 得到非 nil 的空切片；`var values []int` 的零值则为 nil。两者都可以 append，但 nil 判断不同。

## 2. 创建 map

```go
counts := make(map[string]int)
counts["go"] = 1

reserved := make(map[string]int, 100)
_ = reserved
```

可选大小提示用于估计初始空间，不是 map 长度，也不是容量上限。新 map 的长度为零，之后可以继续添加条目。

nil map 可以读取、查询长度和删除条目，但直接写入会 panic；make 创建的 map 则可正常写入。

## 3. 创建 channel

```go
unbuffered := make(chan int)
buffered := make(chan int, 3)
```

省略容量或容量为零时是无缓冲 channel，发送和接收需要配对。容量大于零时带缓冲，但缓冲已满的发送、缓冲为空的接收仍可能等待。

make 创建 channel 不会自动创建用于发送、接收或监听它的 goroutine。

## 4. 与 new 的区别

| 写法 | 结果 |
| --- | --- |
| `make(map[string]int)` | 已初始化、可以写入的 map |
| `new(map[string]int)` | 指向 map 零值的指针，解引用结果为 nil map |
| `make([]int, 3)` | 长度为 3 的切片 |
| `new([]int)` | 指向 nil 切片的指针 |

make 只支持切片、map、channel；不能写 `make(int)` 或 `make(User)`。

## 5. 分配与并发

make 描述语言层面的创建和初始化，不意味着底层数据一定分配在堆上。实际存储位置和是否消除分配，由编译器及运行时决定。

make 不会自动保护共享数据。多个 G 并发读写普通 map、切片元素等，仍需要遵守正确的同步规则。
