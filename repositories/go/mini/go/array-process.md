# Go Array 实现流程

## 一、数组值结构

- 数组值由固定数量、相同类型的元素组成
- 元素在内存中连续排列
- 数组值本身没有 Data、Len、Cap 头部
- 数组长度属于数组类型
- 当前仓库中的模拟结构
  - 文件：array.go
  - Array 是一个结构体
  - elements 保存固定数量的连续元素
  - 该结构用于表达数组值布局
  - 真实 Go 数组值本身不是带头部的 runtime 结构体

## 二、编译期数组类型

- 语法解析
  - 数组类型由长度表达式和元素类型组成
  - [3]int 中的长度是三
  - [...]int 的长度由字面量元素数量推导
- 编译器类型
  - CompileArray 位于 array.go
  - Elem 保存元素类型
  - Bound 保存固定长度
  - CompileArray 只描述类型，不保存数组数据
- Go 编译器源码
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/syntax/nodes.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/types/type.go
  - types.NewArray 创建 TARRAY 类型
  - Array 元数据保存 Elem 和 Bound

## 三、数组字面量初始化

- 编译期间
  - [3]int{1, 2, 3} 会形成 OARRAYLIT 节点
  - 编译器先准备数组变量或临时数组
  - 缺少的元素使用零值初始化
  - fixedlit 生成逐元素写入
  - 小型字面量通常展开为直接赋值
- 大型字面量
  - 编译器可以创建只读静态初始化模板
  - 运行期间将静态模板复制到目标数组
  - 运行时仍会处理需要动态计算的元素
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/ir/node.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/complit.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/staticinit/sched.go

## 四、数组的栈和堆分配

- 编译器判断
  - 逃逸分析检查数组地址是否流向函数外
  - 地址返回、保存到全局或长期闭包通常会逃逸
  - 不逃逸时数组是栈分配候选
  - 编译器也可能将数组完全优化掉
- 大小限制
  - 显式局部变量有 MaxStackVarSize 限制
  - 当前默认值是 128 KiB
  - 超过限制时即使不逃逸也不能作为普通栈变量
  - smallframes 编译选项会降低该限制
- new 数组
  - new([3]int) 不逃逸时可以使用栈临时地址
  - 逃逸时保留堆分配路径
  - runtime.newobject 最终调用 mallocgc
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/escape/solve.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/escape/utils.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/builtin.go
  - /Users/haotao.chen/Desktop/repositories/go/src/runtime/malloc.go
- 普通 Go 代码限制
  - 逃逸分析和真实栈分配由编译器完成
  - 普通函数不能强制指定数组一定在栈上或堆上
  - go build -gcflags=-m=2 可查看编译器诊断

## 五、runtime 类型元数据

- 普通数组操作
  - 编译器已经知道数组长度和元素大小
  - len(array) 可以直接替换为编译期常量
  - array[index] 直接计算元素地址
  - 必要时插入边界检查
- ArrayType
  - 真实源码位于 /Users/haotao.chen/Desktop/repositories/go/src/internal/abi/type.go
  - Type 是公共类型头
  - Elem 指向元素类型
  - Slice 指向对应切片类型
  - Len 保存固定长度
  - ArrayType 是类型元数据，不是数组实例
- 当前仓库中的模拟
  - ArrayType 的 Type、Elem、Slice 使用 any 简化
  - Len 使用 uintptr 表示固定长度
  - 只用于解释 metadata 布局

## 六、接口和反射

- EmptyInterface
  - 文件：array.go
  - Type 保存动态类型元数据
  - Data 保存数组数据地址
  - 数组放入 any 时，接口同时携带类型引用和数据引用
- 编译期间
  - 编译器确认源类型是 [3]int
  - 生成对数组类型描述符的引用
  - 生成数组装箱和数据复制逻辑
  - 生成填充 Type 和 Data 的机器码
- 运行期间
  - 执行机器码后形成接口值
  - reflect.ValueOf 读取接口中的 Type 和 Data
  - reflect.Value 保存类型指针、数据指针和标志
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/internal/abi/iface.go
  - /Users/haotao.chen/Desktop/repositories/go/src/reflect/value.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/convert.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/reflectdata/helpers.go
- 关键区别
  - ArrayType 只保存类型信息
  - EmptyInterface 同时保存类型信息和数据地址
  - 普通数组变量本身不会额外携带类型和长度字段

## 七、反射动态数组

- ArrayOf
  - reflect.ArrayOf 根据运行时长度和元素类型创建数组类型
  - 优先查询类型缓存
  - 没有缓存时构造新的 ArrayType
  - 计算数组总大小、长度、哈希和 GC 数据
- New
  - reflect.New 根据动态数组类型分配零值数组
  - runtime.newobject 使用 mallocgc 取得存储
  - Elem 返回可操作的数组 Value
- 数组访问
  - Value.Index 读取 ArrayType.Len 做边界检查
  - 根据 Elem.Size 计算元素偏移
  - 通过 Value.ptr 加上偏移定位元素
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/reflect/type.go
  - /Users/haotao.chen/Desktop/repositories/go/src/reflect/value.go

## 八、运行时识别数组

- 普通数组
  - 编译器直接使用静态类型信息
  - runtime 不需要从数组数据中猜测长度
- 接口或反射
  - Type 指向 ArrayType
  - ArrayType.Kind 表示 Array
  - ArrayType.Len 提供固定长度
  - Data 或 ptr 指向实际数组数据
- 地址计算
  - 元素地址等于数组起始地址加下标乘元素大小
  - 类型元数据负责解释内存
  - 数据指针负责定位内存

## 九、结论

- 编译期通过 types.NewArray 创建数组类型描述
- 编译器通过 OARRAYLIT、anylit 和 fixedlit 生成数组初始化
- 数组值本身是连续元素，不携带 runtime 类型头
- 逃逸分析决定数组候选存储在栈还是堆
- ArrayType 是共享类型元数据
- EmptyInterface 负责把类型引用和数据地址组合起来
- reflect.ValueOf 读取接口中的类型和数据
- 只有反射 ArrayOf 才能根据运行时长度动态构造数组类型
