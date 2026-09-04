# interface

## 1. 方法集与接口实现

### 1.1 值接收者与指针接收者

```go
type User struct {
	Name string
}

func (u User) ValueMethod() {}

func (u *User) PointerMethod() {}
```

值接收者会复制接收者值，不能通过该副本直接修改原值；指针接收者接收原值的地址，可以修改原值。两者还会形成不同的方法集，进而决定类型是否实现某个接口。

### 1.2 方法集

对于已定义类型 `User`：

| 类型 | 方法集 |
| --- | --- |
| `User` | 接收者为 `User` 的方法 |
| `*User` | 接收者为 `User` 或 `*User` 的方法 |

因此，值接收者声明的方法同时属于 `User` 和 `*User` 的方法集；指针接收者声明的方法只属于 `*User` 的方法集。

### 1.3 接口实现关系

```go
type ValueInterface interface {
	ValueMethod()
}

type PointerInterface interface {
	PointerMethod()
}
```

Go 根据类型的方法集判断它是否实现接口：

| | 结构体通过值接收者实现接口 | 结构体指针通过指针接收者实现接口 |
| --- | --- | --- |
| 结构体初始化变量 `User{}` | 通过 | 不通过 |
| 结构体指针初始化变量 `&User{}` | 通过 | 通过 |

值接收者方法同时属于 `User` 和 `*User` 的方法集；指针接收者方法只属于 `*User` 的方法集。

```go
var valueInterface ValueInterface
valueInterface = User{}  // 可以
valueInterface = &User{} // 可以

var pointerInterface PointerInterface
pointerInterface = User{}  // 编译错误
pointerInterface = &User{} // 可以
```

可以通过编译期断言显式记录这种约束：

```go
var _ ValueInterface = User{}
var _ ValueInterface = (*User)(nil)
var _ PointerInterface = (*User)(nil)
```

### 1.4 方法调用的自动取址

如果变量可寻址，即使指针接收者方法不属于值类型的方法集，仍然可以通过该变量调用：

```go
u := User{}
u.PointerMethod()
```

编译器会将调用等价地改写为：

```go
(&u).PointerMethod()
```

但接口实现判断只检查方法集，不应用自动取址规则：

```go
var pointerInterface PointerInterface = User{} // 编译错误
```

因此，“值可以调用某个方法”不代表“该方法属于值类型的方法集”。

### 1.5 nil 指针与接口值

类型为 `*User` 的 nil 指针仍然具有 `*User` 的方法集，因此可以赋给接口：

```go
var user *User
var pointerInterface PointerInterface = user
```

此时接口保存了动态类型 `*User` 和动态值 `nil`，所以接口本身不等于 `nil`：

```go
fmt.Println(pointerInterface == nil) // false
```

调用接口方法是否 panic 取决于方法实现是否解引用了这个 nil 接收者。
