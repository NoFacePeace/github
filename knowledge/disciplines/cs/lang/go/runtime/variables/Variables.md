# Runtime 变量

本目录整理 Go runtime 的内部变量，说明类型、保存的状态、组织方式和使用场景。具体实现以所用 Go 版本为准。

## 1. GC 标记 worker

- [gcBgMarkWorkerPool](gcBgMarkWorkerPool.md)：以 `lfstack` 组织可用后台标记 worker 的全局池。
