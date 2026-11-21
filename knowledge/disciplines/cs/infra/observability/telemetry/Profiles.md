# 性能剖析

性能剖析（Profiles）记录程序在代码层面的资源使用情况，例如哪些函数占用了较多 CPU，哪些调用路径分配了较多内存。

## 1. 采样与调用栈

以 CPU 剖析为例，剖析器可以定期采集正在运行的线程的调用栈，再汇总各调用路径出现的次数。火焰图可以用来查看采样集中在哪些函数。

## 2. 排查示例

当指标显示订单服务的 CPU 使用率升高时，可以查看同一时间段的 CPU Profile。如果 JSON 编码函数占了较多采样，就可以进一步检查哪些调用路径频繁执行编码。[OpenTelemetry 性能剖析](https://opentelemetry.io/docs/concepts/signals/profiles/)
