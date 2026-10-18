# 接口编译期检查

Go 编译器根据静态类型和方法集，检查赋值、传参、返回值及类型断言等操作是否合法。普通接口的实现是隐式的，不需要声明 implements。

## 1. 判断是否实现接口

类型的方法集必须包含接口要求的全部方法，且方法签名一致。参数名可以不同，多出其他方法也没有影响。

```go
type Reader interface {
    Read(p []byte) (int, error)
}

type File struct{}

func (f *File) Read(p []byte) (int, error) {
    return 0, nil
}

var r Reader = &File{} // 可以：*File 实现 Reader
// var bad Reader = 123 // 编译错误：int 没有 Read 方法
```

编译器已经知道赋值来源的具体类型，因此可以在编译期拒绝不满足接口的赋值。函数参数和返回值也需要遵守相应的可赋值规则。

## 2. 值类型与指针类型分别判断

如果 Read 只定义在 `*File` 上，File 的方法集不包含这个指针接收者方法，因此不能把 File 值直接赋给 Reader。

```go
var file File
file.Read(nil) // 可以：file 可寻址，调用相当于 (&file).Read(nil)

// var r Reader = file // 编译错误，不会为接口赋值自动取地址
var r Reader = &file   // 可以
```

普通方法调用的自动取地址是语法便利，不会改变 File 是否实现 Reader 的判断。

## 3. 显式声明编译期校验

可以用空白标识符声明一条实现检查：

```go
var _ Reader = (*File)(nil)
```

它要求 `*File` 可以赋给 Reader，不满足就编译失败。它不会调用 Read，也不需要构造一个非 nil 的 File 实例。

## 4. 为什么动态断言可以编译

```go
var value any = 123
r, ok := value.(Reader)
```

这里不是要求所有 any 值都实现 Reader，而是要求在执行时检查动态类型。Reader 是合法的普通接口类型，这个断言允许存在；实际类型 int 不满足时，由执行结果返回 false。

编译器仍检查断言表达式是否合法。例如，非接口值不能直接使用普通类型断言语法；断言到具体类型时，该类型必须能够实现源接口；两个接口对同名方法要求不兼容时，也可能被判定为不可能的断言。

## 5. 编译期与执行结果的边界

| 写法 | 编译期结论 | 执行时行为 |
| --- | --- | --- |
| `var r Reader = &File{}` | 方法集满足，允许赋值 | 构造相应接口值 |
| `var r Reader = 123` | int 不满足，编译失败 | 不会执行 |
| `r, ok := value.(Reader)`，value 为 any | 允许动态检查 | 不满足时返回 nil、false |
| `r := value.(Reader)`，value 为 any | 允许动态检查 | 不满足时 panic |

编译通过表示操作符合语言规则，不保证每一次执行时断言都会成功。具体检查也可能被编译器优化，不能仅凭源码写法断定每次都执行完整的运行时查找。
