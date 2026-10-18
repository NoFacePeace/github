# 方法约束

方法约束要求类型提供指定方法。已有的普通接口，例如 `fmt.Stringer`、`io.Reader`，也可以直接用作泛型约束。

## 1. 声明方法要求

```go
type Stringer interface {
    String() string
}

func Text[T Stringer](value T) string {
    return value.String()
}

type User struct {
    Name string
}

func (u User) String() string {
    return u.Name
}

var text = Text(User{Name: "Alice"})
```

类型不需要显式声明实现接口；它的方法集满足要求即可，方法名、参数和返回值必须匹配。

## 2. 值接收者与指针接收者

如果 `String` 只定义在 `*User` 上，通常应传入 `&User{...}`，此时 T 是 `*User`。普通调用中可寻址变量的方法调用简写，不会使 User 的方法集自动包含全部指针接收者方法。

## 3. 与普通接口参数的选择

只需统一调用方法时，普通接口参数通常已经足够。需要保留具体类型、返回同类型结果或关联多个参数的类型时，泛型约束更有意义。

只包含方法要求的接口也可以用于普通接口变量，不仅限于泛型约束。
