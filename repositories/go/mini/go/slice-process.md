# Go Slice 实现流程

## 一、切片的核心结构

- 切片是底层数组的一段连续视图
- runtime 中的切片描述符包含三个字段
  - array
    - 指向底层数组首元素
  - len
    - 表示当前元素数量
  - cap
    - 表示从首元素开始到底层数组末尾的容量
- 切片描述符本身不保存全部元素
- 多个切片可以共享同一个底层数组
- 修改共享区域中的元素会影响其他切片
- 当前仓库中的模拟
  - 文件：slice.go
  - Slice 模拟 runtime.slice
  - array 使用 unsafe.Pointer 表示底层数组地址
  - len 和 cap 使用 int 表示长度和容量
  - Slice 只用于解释布局，不能替代 Go 原生切片

## 二、切片创建的编译器入口

- 用户代码
  - make 创建切片
  - 切片字面量创建切片
  - 从数组或切片截取区间
- 编译器内部节点
  - make 切片会形成 OMAKESLICE 节点
  - make 加 copy 可能形成 OMAKESLICECOPY 节点
  - 切片表达式会形成 OSLICE 或 OSLICE3 节点
- 编译器处理函数
  - cmd/compile/internal/walk/builtin.go
  - walkMakeSlice 处理 OMAKESLICE
  - walkMakeSliceCopy 处理 make 加 copy
- 运行时结构
  - runtime/slice.go
  - runtime.slice 保存 array、len 和 cap

## 三、逃逸分析

- 编译器先判断切片底层数组是否逃逸
- 不逃逸的情况
  - 切片只在当前函数中使用
  - 底层数组地址不返回给函数外部
  - 不保存到长期存活的堆对象
  - 可以作为栈分配候选
- 逃逸的情况
  - 返回切片
  - 保存到全局变量
  - 保存到会长期存活的接口或堆对象
  - 被闭包捕获并在外部继续使用
  - 需要使用 runtime 的堆分配路径
- 相关源码
  - cmd/compile/internal/escape/expr.go
  - cmd/compile/internal/escape/utils.go
  - OMAKESLICE 会进入逃逸分析
  - escape/utils.go 会检查元素大小、容量和栈空间限制

## 四、make 创建切片的编译期展开

### 容量是常量且切片不逃逸

- 编译器可能直接创建固定长度数组
- 概念上的展开形式
  - 声明一个 [cap]E 数组
  - 使用 arr[:len] 创建切片
  - 不调用 runtime.makeslice
- 编译器源码
  - walkMakeSlice 处理常量容量
  - types.NewArray 创建底层数组类型
  - 编译器生成数组临时变量
  - 编译器再生成切片表达式

### 容量是变量且切片不逃逸

- 编译器使用混合策略
- 先计算固定栈缓冲区最多容纳的元素数量
- K 等于变量 make 栈阈值除以元素大小
- 概念上的展开形式
  - 如果 cap 小于等于 K
    - 准备 [K]E 栈数组
    - 创建 backing[:len:cap]
  - 否则
    - 调用 runtime.makeslice
- Go 1.26.3 默认变量 make 栈阈值是 32 字节
- 当前 64 位环境中 int 通常占 8 字节
- []int 的 K 通常为 4
- 这个 4 是编译器根据字节阈值和元素大小计算出来的，不是 runtime 固定规则

### 切片逃逸或不满足栈条件

- 编译器调用 runtime.makeslice
- len 和 cap 可以转换为 int 时调用 makeslice
- 参数需要使用 int64 时调用 makeslice64
- runtime 返回底层数组指针
- 编译器使用指针、len 和 cap 组装切片描述符
- 编译器源码通过 NewSliceHeaderExpr 生成切片头部

## 五、runtime.makeslice 创建底层数组

- runtime 入口
  - runtime/slice.go 中的 makeslice
  - runtime/slice.go 中的 makeslice64
- 参数检查
  - 计算元素大小乘以 cap
  - 检查乘法是否溢出
  - 检查总大小是否超过最大分配限制
  - 检查 len 是否小于零
  - 检查 len 是否大于 cap
- 内存分配
  - 调用 runtime.mallocgc
  - et 保存元素类型信息
  - 元素包含指针时需要按类型进行 GC 扫描
  - 元素不包含指针时可以走不扫描路径
  - 新分配内存通常需要清零
- 返回结果
  - makeslice 只返回底层数组指针
  - len 和 cap 由编译器写入切片描述符

## 六、runtime.mallocgc 分配流程

- runtime/malloc.go 中的 mallocgc 是通用分配入口
- 零字节分配
  - 返回 zerobase
- 小对象分配
  - 根据大小选择对应的 size class
  - 优先从当前处理器的 mcache 获取空闲对象
- 大对象分配
  - 从堆中申请更大的 span
- GC 协作
  - 根据元素类型设置 GC 类型信息
  - 必要时执行清零
  - 必要时触发 GC
  - 返回可被 GC 追踪的内存指针

## 七、当前仓库中的 NewSlice 模拟

- 文件：slice.go
- NewSlice 使用 []int 表达切片底层数组
- 小容量路径
  - 使用 []int 初始化固定小数组
  - 模拟编译器生成栈数组并截取切片
- 中等容量路径
  - 使用固定大小的 int 数组作为栈数组示例
  - 模拟编译期确定容量且不逃逸的情况
- 大容量路径
  - 使用 make 分配底层数组
  - 模拟 runtime.makeslice 的堆分配路径
- 重要差异
  - NewSlice 是普通 Go 函数，不能真正控制栈或堆
  - 真实逃逸分析由编译器根据完整调用关系完成
  - NewSlice 的分支和中文注释只用于表达源码流程

## 八、append 的编译器展开

- 用户代码
  - append(source, value)
  - append(source, value1, value2)
  - append(source, other...)
- 普通追加参数
  - 编译器统计追加参数数量
  - num 等于追加参数个数
- 追加另一个切片
  - 编译器使用 len(other) 作为 num
- 新长度
  - newLen 等于 oldLen 加 num
- 容量判断
  - 如果 newLen 小于等于 oldCap
    - 使用 source[:newLen]
    - 复用原底层数组
  - 如果 newLen 大于 oldCap
    - 调用 runtime.growslice
- 编译器源码
  - cmd/compile/internal/walk/builtin.go 中的 walkAppend
  - cmd/compile/internal/walk/assign.go 中的 appendSlice

## 九、runtime.growslice 扩容

- runtime 入口
  - runtime/slice.go 中的 growslice
- 参数
  - oldPtr 是旧底层数组地址
  - newLen 是扩容后的长度
  - oldCap 是旧容量
  - num 是新增元素数量
  - et 是元素类型
- 旧长度
  - oldLen 等于 newLen 减 num
- 零大小元素
  - 使用 zerobase
  - 返回 newLen 和 newLen 作为容量
- 新容量策略
  - 如果 newLen 大于 oldCap 的两倍
    - 新容量直接使用 newLen
  - 如果 oldCap 小于 256
    - 容量通常翻倍
  - 如果 oldCap 大于等于 256
    - 使用平滑增长公式
    - 增长速度逐步接近 1.25 倍
- 内存大小计算
  - 根据元素大小计算 lenmem
  - 根据元素大小计算 newlenmem
  - 根据新容量计算 capmem
  - 按分配器的 size class 向上对齐
  - 实际容量可能大于策略计算出的容量
- 分配和复制
  - 调用 mallocgc 分配新的底层数组
  - 根据元素是否包含指针选择 GC 路径
  - 清零未被 append 覆盖的尾部区域
  - 使用 memmove 复制旧元素
  - 返回新的 array、newLen 和 newCap

## 十、当前仓库中的 AppendSlice 模拟

- 文件：slice.go
- AppendSlice 接收 source 和 values
- num 等于 len(values)
- 容量足够时
  - 直接使用原 array
  - 将 values 写入可用容量区域
  - 只更新 len
- 容量不足时
  - 计算 doubleCap
  - 小容量使用翻倍策略
  - 大容量使用渐进增长策略
  - 使用 make 创建新的 int 数组
  - copy 旧元素
  - copy 新增元素
  - 返回新的 Slice 描述符
- 模拟限制
  - 真实 growslice 不负责写入普通 append 的新增值
  - 编译器会在 growslice 返回后写入新增元素
  - 当前 AppendSlice 为了便于观察，将新增值复制逻辑合并到函数中
  - 当前实现只模拟 int 元素，不能表达任意元素类型的 GC 指针信息

## 十一、copy 的编译器和运行时流程

- 拷贝数量
  - 取源切片和目标切片长度中的较小值
- 编译器优化
  - 无指针元素可能直接生成 memmove
  - 有指针元素需要使用带写屏障的 typed slice copy
  - 某些场景调用 runtime.slicecopy
- make 加 copy
  - 编译器可能识别 make 加 copy 模式
  - 无指针元素可能直接组合 mallocgc 和 memmove
  - 其他元素类型可能调用 runtime.makeslicecopy
- runtime 行为
  - 分配目标底层数组
  - 按元素类型处理 GC 扫描
  - 使用 memmove 复制数据
  - 返回底层数组指针

## 十二、无法完整复刻的逻辑

- 逃逸分析
  - 依赖完整函数调用关系
  - 普通函数无法接收真实编译器逃逸结果
- 真实栈分配
  - 依赖编译器生成栈帧和机器码
  - 普通 Go 函数无法强制底层数组留在栈上
- runtime 类型信息
  - 当前模拟使用 int
  - 真实 runtime 使用元素类型元数据
  - GC 根据类型元数据识别指针字段
- 内存分配器
  - size class 对齐
  - mcache 和 span 复用
  - 清零和 GC 触发条件
  - 无法用普通 Go 代码完全复刻
- 写屏障和竞态检测
  - 指针元素复制需要写屏障
  - race、msan、asan 由 runtime 和编译器协作处理
- 模拟代码要求
  - 使用中文注释说明源码对应关系
  - 使用中文注释说明无法完全一致的部分

## 十三、完整调用链总结

- make 创建
  - make
  - OMAKESLICE
  - 逃逸分析
  - walkMakeSlice
  - 栈数组或 makeslice
  - mallocgc
  - 组装切片描述符
- append 扩容
  - append
  - 编译器计算 num
  - 计算 newLen
  - 判断 newLen 是否超过 cap
  - growslice
  - nextslicecap
  - mallocgc
  - memmove
  - 返回新的切片描述符
- copy
  - copy
  - 编译器选择 memmove、slicecopy 或 typed slice copy
  - runtime 执行内存复制

## 十四、结论

- 切片是底层数组加 array、len、cap 三字段描述符
- make 的编译器展开决定走栈数组还是 runtime.makeslice
- 逃逸分析在编译期间决定是否允许栈分配
- makeslice 在运行期间计算内存大小并调用 mallocgc
- append 的编译器展开先计算 num 和 newLen
- growslice 在运行期间计算新容量、分配新数组并复制旧元素
- copy 根据元素类型选择不同复制路径
- 当前仓库的 NewSlice 和 AppendSlice 用普通 Go 代码表达主要流程
- 无法复刻的编译器和 runtime 机制通过中文注释和本文档说明
