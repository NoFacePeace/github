# defer

`defer` 是 Go 关键字，用于登记一个函数调用，延迟到当前函数退出过程中执行。常见用途包括解锁、关闭资源以及配合 recover 处理 panic。

## 1. 基本用法

```go
func Work(mu *sync.Mutex) {
    mu.Lock()
    defer mu.Unlock()

    // 在锁保护下执行业务逻辑。
}
```

示例需要导入 sync。只有执行到 defer 语句，该调用才会被登记。defer 归属于当前函数，不是在 if、for 或其他普通代码块结束时执行。

## 2. 参数求值与调用时机

执行到 defer 时，就会求值并保存被调用的函数值和参数；函数调用本身延迟执行。

```go
func Example() {
    x := 1
    defer fmt.Println(x)
    x = 2
}
// 输出 1。
```

如果使用闭包读取外部变量，则变量读取发生在闭包实际执行时：

```go
func ExampleClosure() {
    x := 1
    defer func() {
        fmt.Println(x)
    }()
    x = 2
}
// 输出 2。
```

以上输出示例需要导入 fmt。也可以通过闭包的参数显式捕获登记时的值。

## 3. 多个 defer 的顺序

同一函数调用中，后登记的 defer 先执行，即后进先出：

```go
func Order() {
    defer fmt.Println("第一个")
    defer fmt.Println("第二个")
    defer fmt.Println("第三个")
}
// 依次输出：第三个、第二个、第一个。
```

顺序取决于实际执行时的登记顺序，未执行到的 defer 不参与退出处理。

## 4. 与 return 的执行顺序

```text
计算 return 后的表达式
    ↓
赋值给返回结果
    ↓
执行已登记的 defer，后进先出
    ↓
正常完成后，将最终结果返回调用方
```

如果 defer 自身引发了未恢复的 panic，函数不会沿正常返回流程把结果交给调用方。

### 4.1 普通局部变量

局部变量修改、登记时求值与闭包读取参见[defer 与局部变量](LocalVariables.md)。

```go
func Local() int {
    x := 1
    defer func() {
        x = 2
    }()
    return x
}
// 返回 1。
```

`return x` 已将当时的值 1 赋给返回结果。之后修改局部变量 x 本身，不会修改已设置的结果。

### 4.2 具名返回值

结果变量的修改及多个 defer 的影响参见[defer 与具名返回值](NamedReturns.md)。

```go
func Named() (result int) {
    defer func() {
        result = 2
    }()
    return 1
}
// 返回 2。
```

return 先将 result 设置为 1，defer 再将同一个结果变量改成 2，因此调用方收到 2。多个 defer 都修改它时，仍按照后进先出的顺序依次生效。

### 4.3 共享底层数据

切片、map 和指针的共享行为参见[defer 与共享数据](SharedData.md)。

返回时复制的是值，不会自动深拷贝引用的数据。即使使用普通局部变量，defer 修改共享内容也可能被调用方观察到：

```go
func Slice() []int {
    values := []int{1}
    defer func() {
        values[0] = 2
    }()
    return values
}
// 调用方取得的切片中，第一个元素是 2。
```

这里修改的是返回切片共享的底层数组，而不是重新设置返回结果。指针和 map 等也需要区分变量本身与其关联的数据。

## 5. panic 与退出

函数因 panic 展开调用栈时，会执行相应函数中已登记的 defer。recover 需要在延迟调用的函数中直接调用，才能在符合条件时恢复当前 G 的 panic；它不能捕获其他 G 的 panic。

`os.Exit` 直接结束进程，不执行 defer。`main.main` 返回也不会等待其他 G 退出，更不会自动执行它们尚未触发的 defer。

如果 defer 时保存的函数值是 nil，通常在退出时实际调用该函数才 panic；但登记时求值函数和参数的过程本身也可能立即 panic。

## 6. 循环中的资源释放

循环体里的 defer 会累积到当前函数退出，并不是每轮结束就执行。批量处理文件等资源时，可以把单次处理提取为函数，在该函数中 defer 关闭资源，使每次调用结束时及时释放。

## 7. 实现与源码入口

defer 由编译器和运行时配合实现，具体路径受版本及使用场景影响。常见路径包括开放编码、栈上记录及运行时分配记录，不能一概理解为每次都向某个堆链表插入节点。

- `src/cmd/compile/internal`：defer 的检查、转换与开放编码等代码生成。
- `src/runtime/panic.go`：defer 记录、延迟调用及 panic 展开等辅助逻辑。
- `src/runtime/runtime2.go`：相关内部记录结构，具体组织随版本变化。
