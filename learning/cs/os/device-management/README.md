# 设备管理

设备管理关注操作系统如何与硬件设备协作，包括设备驱动、中断、DMA 和网络设备收包机制。

## 1. 专题文档

- [设备驱动](device-driver.md)：操作系统如何初始化和控制硬件设备。
- [中断管理](interrupt.md)：硬中断、softirq、`ksoftirqd` 与 workqueue。
- [网络设备收包](network-receive.md)：RX Ring、Packet Descriptor、DMA 与 NAPI。
