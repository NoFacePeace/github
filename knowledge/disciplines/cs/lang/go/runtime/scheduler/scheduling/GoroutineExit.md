# G 执行结束

普通 goroutine 的入口函数返回，或调用 `runtime.Goexit()` 结束当前 goroutine 时，会经过运行时退出路径清理，再进入调度。具体实现以所用 Go 版本为准。

## 1. 触发与清理

goroutine 入口返回会进入运行时安排的退出入口。`runtime.Goexit()` 则先执行当前 G 的 defer，再结束该 G，不是退出整个进程。

## 2. 调用路径

```text
goexit → goexit1
    ↓ mcall，切到 g0 栈
goexit0 → 清理原 G → schedule
```

运行时将结束的 G 置为 `_Gdead`，清理执行关系及必要状态，并按规则回收可复用的 G 资源，再安排后续工作。

## 3. 与暂停的区别

已结束的 goroutine 不会像挂起的 G 一样从退出位置恢复。`g` 结构可能被复用来承载新的 goroutine，这不代表原业务执行继续。

`main` 返回及程序致命错误等进程退出情况不能简单等同于普通 G 结束后的调度路径。
