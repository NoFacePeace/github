# Go String 实现流程

## 一、核心结构

- String 的值结构
  - 保存底层字节数据指针
  - 保存字节长度
  - 不保存容量
  - 字节数据本身不属于 String 结构
- 当前仓库中的模拟结构
  - 文件：string.go
  - String.str 模拟底层数据指针
  - String.len 模拟字节长度
  - 该结构用于理解布局，不是 Go 语言内建 string 的替代类型

## 二、字符串字面量

- 编译期间
  - 双引号字符串由词法分析器解析
  - 反引号字符串由原始字符串解析逻辑处理
  - 编译器将字符串字面量放入只读数据区
  - 汇编中通常可以看到 SRODATA 标记
  - 编译器构造数据指针和字节长度
- 运行期间
  - 字符串变量只需要保存描述符
  - 字面量的字节数据通常不需要重新分配
  - 字面量内容不可修改
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/syntax/scanner.go
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/reflectdata/reflect.go

## 三、字符串拼接

- 编译期间
  - 加号表达式会被识别为字符串拼接
  - 编译器根据操作数数量选择 concatstring2 到 concatstring5
  - 操作数超过五个时选择 concatstrings
  - 如果结果不逃逸且足够小，编译器可能准备调用方栈上的临时缓冲区
  - 普通 Go 函数无法完整复刻这种调用方栈缓冲约定
- 运行期间
  - runtime.concatstrings 遍历所有输入字符串
  - 跳过空字符串
  - 累计拼接后的总字节长度
  - 没有非空字符串时返回空字符串
  - 只有一个可直接复用的非空字符串时直接返回
  - 其他情况调用 rawstringtmp
  - 临时缓冲区足够时写入调用方提供的缓冲区
  - 缓冲区不足时调用 rawstring
  - rawstring 通过 mallocgc 分配新的字节存储
  - 使用 copy 或底层内存复制将所有输入依次写入目标区域
- 相关源码
  - /Users/haotao.chen/Desktop/repositories/go/src/cmd/compile/internal/walk/expr.go
  - /Users/haotao.chen/Desktop/repositories/go/src/runtime/string.go
  - /Users/haotao.chen/Desktop/repositories/go/src/runtime/malloc.go

## 四、字节切片转换为字符串

- 编译期间
  - 编译器将 string(bytes) 转换为 runtime.slicebytetostring
  - 某些短生命周期比较或拼接场景可以使用 slicebytetostringtmp
  - 这种零拷贝路径由编译器内建优化控制
- 运行期间
  - 长度为零时返回空字符串
  - 长度为一时可以使用 runtime 的静态单字节数据
  - 结果不逃逸且临时缓冲区足够时使用临时缓冲区
  - 其他情况通过 mallocgc 分配字符串存储
  - 使用 memmove 复制字节
  - 普通转换需要复制，因为原字节切片可修改
- 当前仓库模拟
  - BytesToString 位于 string.go
  - 使用临时缓冲区和动态缓冲区模拟主要分支
  - 使用 unsafe.String 构造结果
  - 无法复刻 runtime 的静态单字节表和编译器内建零拷贝优化

## 五、字符串转换为字节切片

- 编译期间
  - 编译器将 []byte(stringValue) 转换为 runtime.stringtoslicebyte
  - 不逃逸且缓冲区足够小时可以传入临时缓冲区
- 运行期间
  - 使用调用方临时缓冲区时先清理缓冲区状态
  - 缓冲区不足时调用 rawbyteslice
  - rawbyteslice 通过 mallocgc 创建可写字节数组
  - 使用 copy 复制字符串字节
  - 返回切片描述符
- 当前仓库模拟
  - StringToBytes 位于 string.go
  - 小结果使用固定临时缓冲区
  - 大结果使用动态字节切片
  - 复制保证修改结果不会修改原字符串

## 六、字符串创建模拟

- NewString
  - 位于 string.go
  - 读取 source 的底层数据指针和长度
  - 不分配新存储
  - 用于模拟字符串描述符创建
- 适用范围
  - 适合解释字符串字面量的指针和长度布局
  - 不适合模拟从可变字节切片创建独立字符串
  - 字节切片转换应使用 BytesToString

## 七、栈与堆

- 栈候选
  - 拼接结果只在当前调用链中使用
  - 编译器确认结果不会逃逸
  - 长度满足临时缓冲区限制
- 堆分配
  - 拼接结果返回到函数外
  - 结果被长期保存或装入会逃逸的接口
  - 结果长度超过临时缓冲区
  - runtime 通过 mallocgc 创建新存储
- 重要限制
  - 普通 Go 函数无法强制字符串数据位于栈上
  - 真实栈缓冲优化依赖编译器生成的调用约定
  - 逃逸结果应以 go build -gcflags=-m 为准

## 八、结论

- string 是只读字节序列
- 字符串字面量通常引用只读数据区
- 拼接通常需要计算长度、分配存储和复制数据
- string 与 []byte 的普通转换通常需要复制
- runtime 中的临时缓冲、静态单字节表和零拷贝路径依赖编译器协作
- 当前仓库的实现用于解释主要流程，不替代编译器和 runtime 内建实现
