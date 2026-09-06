# Go 底层原理学习

## 1. 构建与运行

Go 程序从源代码到实际执行，可以分为编译期、链接期和运行期：

```text
Go 源代码 → 编译期 → 目标文件 → 链接期 → 可执行文件 → 运行期
```

- 编译期：检查并优化源代码，生成目标文件。
- 链接期：组合目标文件和依赖，生成可执行文件。
- 运行期：操作系统加载可执行文件，初始化 Go Runtime 并执行程序。

`go run .` 虽然表现为一个命令，但内部仍然会先完成编译和链接，生成临时可执行文件，然后再运行该文件。

## 2. 数据结构

1. [array](./array.md)
2. [slice](./slice.md)
3. [map](./map.md)
4. [string](./string.md)

## 3. 语言机制

1. [interface](./interface.md)

## 4. 运行时

1. [scheduler](./scheduler.md)

## 5. 并发编程

### Context：取消传播与生命周期管理

`context` 属于标准库中的并发控制与请求生命周期管理，用于沿调用链传播取消信号、超时与截止时间，以及请求范围的元数据。

- 取消传播：通过 `WithCancel` 和 `Done` 通知关联任务停止工作。
- 超时控制：通过 `WithTimeout` 和 `WithDeadline` 限制任务的执行时间。
- 数据传递：通过 `WithValue` 和 `Value` 传递 trace ID 等请求元数据。

学习重点：父子 context 的取消传播机制，以及 channel、锁和定时器如何协作。取消是协作式的，需要任务主动响应；`context` 不会强制终止 goroutine，也不负责等待其退出。
