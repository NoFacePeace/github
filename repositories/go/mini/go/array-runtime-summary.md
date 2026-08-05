# Go 数组问题与结论

## 1. 数组的内存结构

Go 数组是固定数量、相同类型元素组成的一段连续内存：

```go
var array [3]int
```

```text
array
┌────────┬────────┬────────┐
│ int 0  │ int 1  │ int 2  │
└────────┴────────┴────────┘
```

数组值没有切片那样的 `Data`、`Len`、`Cap` 头部。长度属于类型的一部分，因此 `[3]int` 和 `[4]int` 是不同类型。

## 2. 编译期数组类型

编译器内部使用类似下面的结构记录数组类型：

```go
type Array struct {
	Elem  *Type
	Bound int64
}
```

- `Elem` 表示元素类型；
- `Bound` 表示固定元素数量。

当前源码位置：

```text
src/cmd/compile/internal/types/type.go
```

编译器通过 `NewArray` 创建数组类型。元素类型和长度共同决定最终类型。

## 3. runtime 公共类型头

runtime 中的所有类型描述符都有公共头部 `internal/abi.Type`，其中包含：

```go
type Type struct {
	Size_       uintptr
	PtrBytes    uintptr
	Hash        uint32
	TFlag       TFlag
	Align_      uint8
	FieldAlign_ uint8
	Kind_       Kind
	Equal       func(unsafe.Pointer, unsafe.Pointer) bool
	GCData      *byte
	Str         NameOff
	PtrToThis   TypeOff
}
```

`Kind_` 用于说明具体类型种类。当 `Kind_ == abi.Array` 时，公共类型头实际属于一个完整的 `abi.ArrayType`。

## 4. runtime 数组类型元数据

数组的 runtime 类型元数据定义为：

```go
type ArrayType struct {
	Type
	Elem  *Type
	Slice *Type
	Len   uintptr
}
```

当前源码位置：

```text
src/internal/abi/type.go
```

字段含义：

- `Type` 是所有类型共有的头部；
- `Elem` 指向元素类型元数据；
- `Slice` 指向对应切片类型元数据；
- `Len` 保存数组固定长度。

`Type` 位于第一个字段，因此检查 `Kind_ == Array` 后，可以将 `*Type` 转换为 `*ArrayType`。

## 5. 类型元数据保存在哪里

编译器为源码中使用的数组类型生成类型描述符符号，链接器将它们整理到可执行文件的静态类型数据区。程序启动后，这些数据被映射到进程的只读内存，生命周期与程序一致，不属于普通栈或 GC 堆。

runtime 的 `moduledata` 保存类型区域信息：

```go
type moduledata struct {
	// ...
	types, typedesclen, etypes uintptr
	rodata                    uintptr
	// ...
}
```

- `types` 是类型区域起始地址；
- `etypes` 是类型区域结束地址；
- `typedesclen` 是可枚举类型描述符区域长度；
- `rodata` 是模块只读数据区域基址。

相关源码：

```text
src/runtime/symtab.go
src/runtime/type.go
src/cmd/link/internal/ld/data.go
```

## 6. 数组变量指向什么

数组变量不是指针，变量本身就是数组数据：

```go
array := [3]int{1, 2, 3}
```

- 在栈上时，`array` 是栈帧中的连续内存；
- 在堆上时，编译器通过内部指针访问堆对象，但语言层面仍是数组值；
- 作为全局变量时，数据位于静态数据区。

只有显式取地址才会得到数组指针：

```go
pointer := &array
```

`pointer` 的类型是 `*[3]int`，它指向数组数据的起始位置。

数组变量不会直接指向 `ArrayType`。数组值和类型元数据是分开的。

## 7. 普通执行时如何知道数组长度

对于普通数组操作，runtime 通常不需要动态读取长度。编译器已经知道数组类型，会把 `len(array)` 替换为常量。

```go
array := [3]int{}
length := len(array)
```

概念上编译为：

```go
length := 3
```

访问元素时，编译器使用类型中的长度和元素大小：

```text
元素地址 = 数组起始地址 + 下标 × 元素大小
```

编译器能够证明下标安全时可以消除边界检查，否则生成运行时越界检查。

相关源码：

```text
src/cmd/compile/internal/walk/builtin.go
src/cmd/compile/internal/walk/expr.go
```

## 8. 接口如何保存数组

数组装入 `any` 时，编译器生成接口装箱代码：

```go
array := [3]int{1, 2, 3}
var value any = array
```

空接口的内部结构为：

```go
type EmptyInterface struct {
	Type *Type
	Data unsafe.Pointer
}
```

```text
EmptyInterface
├── Type ──→ [3]int 对应的 ArrayType
└── Data ──→ 装箱后的数组值副本
```

数组是值类型，装入接口时保存的是数组值副本。修改原数组不会改变接口中的数组副本。

`Type` 负责说明数据是什么类型，`Data` 才负责定位具体数组数据。`ArrayType` 不保存某个数组实例的地址。

## 9. 编译器如何构造接口

将 `[3]int` 转换成 `any` 时，编译器会：

1. 确认源类型是 `[3]int`；
2. 生成或引用 `[3]int` 的类型描述符符号；
3. 生成数组复制或装箱代码；
4. 生成构造 `EmptyInterface` 的指令。

概念代码为：

```go
temporary := array

value := EmptyInterface{
	Type: arrayTypePointer,
	Data: unsafe.Pointer(&temporary),
}
```

编译期间只生成上述操作对应的机器指令和类型符号引用。程序运行时执行这些指令，才真正得到包含有效数据地址的接口值。

相关源码：

```text
src/cmd/compile/internal/walk/convert.go
src/cmd/compile/internal/reflectdata/helpers.go
src/cmd/compile/internal/reflectdata/reflect.go
```

## 10. reflect.ValueOf 如何取得数组信息

`reflect.ValueOf` 的参数类型是 `any`：

```go
func ValueOf(value any) Value
```

因此：

```go
reflect.ValueOf(array)
```

概念上等价于：

```go
reflect.ValueOf(any(array))
```

运行时流程：

```text
数组值
  ↓ 接口装箱
EmptyInterface
  ├── Type
  └── Data
  ↓ unpackEface
reflect.Value
  ├── typ
  ├── ptr
  └── flag
```

`reflect.ValueOf` 不会根据数组内容猜测类型，也不会按名称搜索类型。它直接读取接口中由编译器装入的 `Type` 和 `Data`。

相关源码：

```text
src/reflect/value.go
```

## 11. 为什么需要反射

runtime 类型元数据是底层内部数据，普通 Go 代码不能安全、稳定地直接访问。反射在这些元数据上提供公共 API，用于：

- 查询运行时类型和种类；
- 获取数组长度和元素类型；
- 按下标读取或修改未知数组；
- 动态创建类型和值；
- 支持 JSON、ORM、依赖注入等通用框架。

类型断言要求编译期写出目标类型，而反射可以处理编译期未知的具体类型。

## 12. 运行时能否创建数组

普通 Go 数组可以在程序运行时分配，但它的元素类型和长度必须在编译期确定：

```go
pointer := new([3]int)
```

- 不逃逸时，编译器可以在栈上创建数组；
- 逃逸时，通过 `runtime.newobject` 和 `mallocgc` 在堆上分配。

不存在普通语法形式的动态长度数组：

```go
length := 3
var array [length]int // 编译错误
```

动态长度集合应该使用切片：

```go
slice := make([]int, length)
```

反射是例外，可以在运行时创建数组类型：

```go
arrayType := reflect.ArrayOf(length, reflect.TypeOf(int(0)))
arrayValue := reflect.New(arrayType).Elem()
```

`reflect.ArrayOf` 会查询缓存或动态构造 `ArrayType`，计算 `Len`、`Size`、`Hash`、GC 数据和比较函数。通过反射创建的数组仍是真正的固定长度数组，只是类型在运行时获得。

相关源码：

```text
src/reflect/type.go
src/reflect/value.go
src/runtime/malloc.go
```

## 13. 数组在栈还是堆

数组地址没有流向生命周期更长的位置，并且数组大小不超过栈变量限制时，可以位于栈上或被编译器优化掉。

以下情况通常导致数组逃逸到堆：

- 返回局部数组的地址；
- 将数组地址保存到全局变量；
- 数组地址被长期存储；
- 被异步闭包或 goroutine 持有；
- 编译器无法证明引用不会超出当前栈帧。

当前编译器中，显式局部变量默认的最大栈对象大小是 128 KiB：

```go
MaxStackVarSize = 128 * 1024
```

超过该限制的数组即使没有普通意义上的地址逃逸，也会因为 `too large for stack` 被移到堆上。开启 `smallframes` 时限制会降低。

相关源码：

```text
src/cmd/compile/internal/ir/cfg.go
src/cmd/compile/internal/escape/utils.go
src/cmd/compile/internal/escape/solve.go
```

## 14. 最终结论

1. 数组值就是连续元素数据，不携带长度、容量或类型指针。
2. 数组的元素类型和固定长度在编译期共同构成数组类型。
3. 普通数组操作直接使用编译期信息，不需要 runtime 动态判断类型。
4. `ArrayType` 是共享的运行时类型元数据，不保存数组实例地址。
5. 接口通过 `Type` 引用数组类型元数据，通过 `Data` 引用装箱后的数组数据。
6. `reflect.ValueOf` 先接收编译器构造的接口，再取出其中的类型指针与数据指针。
7. 编译期生成“如何装箱”的代码，运行时执行代码后才真正形成 `EmptyInterface`。
8. 普通数组长度必须编译期确定；反射可以通过 `ArrayOf` 在运行时构造数组类型。
9. 栈或堆由逃逸分析、对象大小和编译器优化共同决定。

仓库中的教学实现位于 [array.go](./array.go)，思维导图位于 [array-mindmap.md](./array-mindmap.md)。
