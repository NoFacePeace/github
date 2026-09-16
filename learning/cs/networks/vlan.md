# VLAN

VLAN（Virtual Local Area Network，虚拟局域网）是一种在同一个物理交换网络中划分多个逻辑局域网的二层技术，主要由 IEEE 802.1Q 定义。

例如，可以在同一组交换机上划分两个 VLAN：

```text
VLAN 10：研发部门
VLAN 20：财务部门
```

属于同一 VLAN 的接口处于同一个二层广播域；属于不同 VLAN 的设备即使连接到同一台交换机，默认也不能直接进行二层通信。不同 VLAN 之间需要通过路由器或三层交换机进行三层转发。

VLAN 可以用于：

- 隔离不同部门、业务或安全级别的网络。
- 缩小广播域，减少广播流量的影响范围。
- 在不改变物理布线的情况下灵活划分逻辑网络。
- 让多个 VLAN 共享交换机之间的同一条物理链路。

VLAN 与以太网关系密切，但不是 IEEE 802.3 基础以太网标准本身。基础以太网主要由 IEEE 802.3 定义，VLAN 标签和相关交换规则则由 IEEE 802.1Q 定义，两者都工作在数据链路层。

## 1. 802.1Q 帧格式

IEEE 802.1Q 会在源 MAC 和原 EtherType 字段之间插入一个 4 字节的 VLAN 标签：

```text
普通 Ethernet II 帧：
| 目标 MAC | 源 MAC | EtherType |     数据      | FCS    |
| 6 字节   | 6 字节 | 2 字节    | 46～1500 字节 | 4 字节 |

带 802.1Q 标签的帧：
| 目标 MAC | 源 MAC | TPID   | TCI    | EtherType |     数据      | FCS    |
| 6 字节   | 6 字节 | 2 字节 | 2 字节 | 2 字节    | 46～1500 字节 | 4 字节 |
```

因此，普通 Ethernet II 帧头为 14 字节，带一个 802.1Q 标签的帧头为 18 字节：

```text
普通帧头：
目标 MAC 6 + 源 MAC 6 + EtherType 2 = 14 字节

带 VLAN 标签的帧头：
目标 MAC 6 + 源 MAC 6 + VLAN 标签 4 + EtherType 2 = 18 字节
```

在普通 1500 字节 MTU 不变的情况下，带 VLAN 标签的以太网帧最大长度通常从 1518 字节增加到 1522 字节。VLAN 标签增加的是二层头部开销，不会占用 IP 包的 1500 字节 MTU。

## 2. VLAN 标签字段

4 字节 VLAN 标签由 TPID 和 TCI 两部分组成：

```text
| TPID：16 bit | PCP：3 bit | DEI：1 bit | VID：12 bit |
|<-- 2 字节 -->|<------------- TCI：2 字节 ----------->|
```

- TPID（Tag Protocol Identifier）：标签协议标识，常见值为 `0x8100`，表示这是一个 IEEE 802.1Q VLAN 标签。
- PCP（Priority Code Point）：3 bit 的流量优先级，取值为 `0～7`。
- DEI（Drop Eligible Indicator）：表示发生拥塞时，该帧是否可以被优先丢弃。
- VID（VLAN Identifier）：12 bit 的 VLAN 编号。取值 `0` 和 `4095` 有特殊用途，通常可配置的 VLAN ID 为 `1～4094`。

接收带标签帧的交换机可以读取 VID，并只在属于相应 VLAN 的端口之间转发该帧。

## 3. Access 与 Trunk

交换机端口常见的 VLAN 工作模式包括 Access 和 Trunk：

- Access 端口：通常连接电脑、打印机等终端设备，只属于一个 VLAN。终端收发的普通以太网帧通常不携带 VLAN 标签，由交换机在内部将其关联到该端口所属的 VLAN。
- Trunk 端口：通常用于连接交换机、路由器或虚拟化主机，可以承载多个 VLAN 的流量。交换机通过 802.1Q 标签区分帧所属的 VLAN。

一个典型的转发过程如下：

```text
终端 A
  |
Access 端口（VLAN 10）
  |
交换机添加或关联 VLAN 10 信息
  |
Trunk 链路（帧携带 VLAN 10 标签）
  |
另一台交换机
  |
Access 端口（VLAN 10）
  |
终端 B
```

Access 端口和 Trunk 端口的具体标签处理方式取决于交换机配置。Trunk 链路还可能配置 Native VLAN，其帧是否携带标签需要根据设备实现和配置判断。
