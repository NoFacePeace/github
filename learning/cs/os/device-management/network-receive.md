# 网络设备收包

Linux 网卡收包通常涉及 DMA、RX Ring、Packet Descriptor、硬中断、NAPI 和 softirq。

## 1. RX Ring 与 Packet Descriptor

RX Ring 是网卡和驱动循环复用的接收队列。队列中的 Packet Descriptor 是数据包描述符，通常记录数据缓冲区地址、长度、状态和所有权等信息，不保存完整数据包。

```text
RX Ring
├── Descriptor 0 → 数据 buffer 0
├── Descriptor 1 → 数据 buffer 1
└── Descriptor 2 → 数据 buffer 2
```

驱动预先准备描述符和数据缓冲区，并把缓冲区地址交给网卡。网卡收到数据后，通过 DMA 将数据写入缓冲区，再更新描述符状态。

## 2. NAPI

NAPI（New API）是 Linux 网络子系统中的收包机制，采用“中断加轮询”的混合模式。

```text
网卡收到数据
    ↓
DMA 写入接收 buffer
    ↓
网卡更新 RX Ring，并触发一次硬件中断
    ↓
驱动关闭或减少接收中断，调用 napi_schedule()
    ↓
NET_RX_SOFTIRQ 执行 NAPI poll
    ↓
批量处理数据包
    ↓
队列为空后恢复接收中断
```

NAPI 的 `poll` 函数通常接收 `budget` 参数，表示本轮最多处理的数据包数量：

- RX Ring 为空：调用完成逻辑，重新开启接收中断。
- 达到 `budget` 但队列仍有数据：暂不恢复中断，继续下一轮 poll。

低流量时，中断可以快速通知数据到达；高流量时，NAPI 批量处理数据，避免每个数据包都触发一次中断。

## 3. 与 softirq 的关系

NAPI poll 通常运行在网络接收 softirq 上下文中。softirq 负载较高时，相关工作可能由对应 CPU 的 `ksoftirqd` 内核线程继续处理。
