# 网络 I/O

网络 I/O 关注应用程序如何通过 Socket 收发数据，以及数据如何从网卡经过内核网络协议栈到达应用程序。

## 1. 数据接收路径

```text
网卡
  ↓ DMA
主机内存中的接收 buffer
  ↓
RX Ring / Packet Descriptor
  ↓
驱动与 NAPI
  ↓
网络协议栈
  ↓
skb
  ↓
Socket 接收缓冲区
  ↓ recv()
应用程序 buffer
```

## 2. `skb`

Linux 中的 `skb` 通常指 `struct sk_buff`，是网络协议栈用来表示和管理数据包的内核对象。它保存数据地址、长度、协议层位置、网络设备、校验和等信息，通常不等于完整的数据包内存本身。

## 3. Socket 缓冲区

Socket 接收缓冲区保存应用程序尚未读取的数据；Socket 发送缓冲区保存尚未发送完成或尚未由协议栈处理完的数据。

```text
接收：网络协议栈 → Socket 接收缓冲区 → recv()
发送：send() → Socket 发送缓冲区 → 网络协议栈 → 网卡
```

因此，`skb` 是内核中的数据包对象，Socket 缓冲区则是面向某个 Socket 的数据排队空间，两者不是同一个概念。
