# Node exporter

## 1. 常用指标和 PromQL

node_exporter 采集主机操作系统指标并暴露 `/metrics`；Prometheus 定时抓取并保存样本，使用 PromQL 查询和计算这些样本。本文以 Linux 主机为例。

### 1.1 CPU

#### 1.1.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_cpu_seconds_total` | Counter / 秒 | 各逻辑 CPU 在不同模式下的累计时间，按 `cpu`、`mode` 区分 |

#### 1.1.2 常用标签和值

`node_cpu_seconds_total` 带有以下标签：

| label | 常见值 | 含义 |
|---|---|---|
| `cpu` | `0`、`1`、`2`、… | 逻辑 CPU 编号，标签值是字符串，不是物理插槽编号 |
| `mode` | `user` | 用户态执行时间 |
| `mode` | `nice` | 调整过 nice 优先级的用户态执行时间 |
| `mode` | `system` | 内核态执行时间 |
| `mode` | `idle` | 空闲时间 |
| `mode` | `iowait` | I/O 等待时间 |
| `mode` | `irq`、`softirq` | 硬中断、软中断处理时间 |
| `mode` | `steal` | 虚拟化环境中被宿主机用于其他工作的 CPU 时间 |

例如 `node_cpu_seconds_total{cpu="0", mode="idle"}` 表示逻辑 CPU 0 的累计空闲时间；`idle` 是标签值，累计秒数才是指标样本值。

#### 1.1.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| 整机 CPU 非空闲比例（%） | `100 * (1 - avg by (instance) ( rate(node_cpu_seconds_total{mode="idle"}[5m]) ))` | 先求每个核心的空闲比例，再取平均，用 1 减去结果并乘以 100。例如空闲比例为 0.8，结果就是 20%。这里的“使用率”采用非 idle 口径，包含 iowait、steal，不能全部解释为 CPU 正在执行计算。 |
| I/O 等待比例（%） | `100 * avg by (instance) ( rate(node_cpu_seconds_total{mode="iowait"}[5m]) )` | 用于辅助分析 I/O 等待，需要结合磁盘指标判断，不能单凭它确定磁盘瓶颈。 |

### 1.2 系统负载

负载反映正在运行、等待 CPU 或处于不可中断睡眠状态的任务压力，不是 CPU 使用率。负载高时，既可能是 CPU 竞争激烈，也可能是任务在等待 I/O。

#### 1.2.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_load1`、`node_load5`、`node_load15` | Gauge / 无单位 | 1、5、15 分钟平均负载；Linux 包含可运行任务和不可中断睡眠任务 |

#### 1.2.2 常用标签和值

这三个负载指标没有额外的 exporter 标签，通常通过 `job="node"`、`instance="10.0.0.10:9100"` 区分主机。1、5、15 分钟体现在指标名中，没有 `cpu`、`mode` 或 `period` 标签。

#### 1.2.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| 直接查询平均负载 | `node_load1` | 查询 5、15 分钟平均负载时，分别替换为 `node_load5`、`node_load15`。 |
| 每个逻辑 CPU 对应的平均负载 | `node_load1 / count by (instance) ( node_cpu_seconds_total{mode="idle"} )` | 这是负载除以逻辑 CPU 数量，不是 CPU 使用率。结果持续大于 1 时，应结合 CPU、I/O 和任务状态进一步排查。 |

### 1.3 内存与 Swap

#### 1.3.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_memory_MemTotal_bytes` | Gauge / 字节 | 系统可用的物理内存总量 |
| `node_memory_MemAvailable_bytes` | Gauge / 字节 | 无需交换即可供新应用使用的内存估算值，包含可回收部分 |
| `node_memory_MemFree_bytes` | Gauge / 字节 | 完全空闲的内存，不含可回收缓存 |
| `node_memory_SwapTotal_bytes`、`node_memory_SwapFree_bytes` | Gauge / 字节 | Swap 总量和剩余量 |

#### 1.3.2 常用标签和值

本节列出的内存与 Swap 指标没有额外的 exporter 标签，通常通过 `job`、`instance` 区分主机。总量、可用量和空闲量由不同指标名表达，没有 `type="free"` 或 `type="used"` 标签。

#### 1.3.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| 内存使用率（%） | `100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)` | 通常使用 `MemAvailable` 衡量内存余量；只用 `MemFree` 会把可回收缓存也算成不可用内存。 |
| 已用 Swap（GiB） | `(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes) / 1024^3` | Swap 占用不等于当前正在频繁换页，需要结合换页速率判断。未启用 Swap 时总量为 0，不宜直接计算百分比。 |

### 1.4 文件系统容量与 inode

#### 1.4.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_filesystem_size_bytes` | Gauge / 字节 | 文件系统总容量 |
| `node_filesystem_avail_bytes` | Gauge / 字节 | 非特权用户可用容量 |
| `node_filesystem_free_bytes` | Gauge / 字节 | 总空闲容量，可能包含保留空间 |
| `node_filesystem_files`、`node_filesystem_files_free` | Gauge / 个 | inode 总数和空闲数 |
| `node_filesystem_readonly` | Gauge / 0 或 1 | 是否以只读方式挂载 |

#### 1.4.2 常用标签和值

本节列出的文件系统指标通常带有以下标签：

| label | 常见值示例 | 含义 |
|---|---|---|
| `device` | `/dev/sda1`、`/dev/nvme0n1p1`、`/dev/mapper/vg-data` | 文件系统对应的设备或挂载来源 |
| `mountpoint` | `/`、`/data`、`/boot` | 挂载路径 |
| `fstype` | `ext4`、`xfs`、`tmpfs`、`overlay` | 文件系统类型 |

例如 `{device="/dev/sda1", mountpoint="/", fstype="ext4"}` 表示挂载在根目录的 ext4 文件系统。部分版本还会带有 `device_error` 标签，正常时通常为空字符串，异常时可能为错误说明；以实际 `/metrics` 为准。

以下查看根文件系统，其他挂载点替换 `mountpoint` 即可。

#### 1.4.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| 从普通用户可用空间衡量的容量占用（%） | `100 * (1 - node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})` | 该口径把保留空间也计入不可用部分，不一定与 `df` 的百分比完全一致。查询所有挂载点时，应按实际环境过滤伪文件系统和重复挂载。 |
| inode 使用率（%） | `100 * (1 - node_filesystem_files_free{mountpoint="/"} / node_filesystem_files{mountpoint="/"})` | 适用于提供有效 inode 总量且总量大于 0 的文件系统。小文件很多时，即使磁盘还有容量，也可能因为 inode 耗尽而无法创建文件。 |

### 1.5 磁盘 I/O

磁盘性能主要看吞吐量、IOPS、平均延迟、未完成 I/O 数量和活跃时间占比。`node_disk_*` 按 `device` 区分块设备，与文件系统挂载点不是同一层次；CPU iowait 用作主机侧的辅助指标。

#### 1.5.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_disk_read_bytes_total`、`node_disk_written_bytes_total` | Counter / 字节 | 累计读、写数据量 |
| `node_disk_reads_completed_total`、`node_disk_writes_completed_total` | Counter / 次 | 累计完成的读、写操作数 |
| `node_disk_read_time_seconds_total`、`node_disk_write_time_seconds_total` | Counter / 秒 | 读、写请求累计耗时 |
| `node_disk_io_time_seconds_total` | Counter / 秒 | 设备有 I/O 活动的累计时间 |
| `node_disk_io_time_weighted_seconds_total` | Counter / 秒（按未完成请求数加权） | 累计未完成 I/O 数量与经过时间的乘积，对其求速率可估算平均未完成 I/O 数量 |
| `node_disk_io_now` | Gauge / 个 | 当前尚未完成的 I/O 请求数，包含正在处理及排队的请求 |
| `node_disk_reads_merged_total`、`node_disk_writes_merged_total` | Counter / 次 | 累计被合并的读、写请求数 |

#### 1.5.2 常用标签和值

本节 `node_disk_*` 指标使用 `device` 区分块设备：

| label | 常见值示例 | 含义 |
|---|---|---|
| `device` | `sda`、`vda`、`nvme0n1`、`dm-0` | 内核块设备名称，通常不带 `/dev/` 前缀 |

例如 `node_disk_read_bytes_total{device="nvme0n1"}` 表示该块设备累计读取的字节数。读、写方向体现在指标名中，没有 `direction="read"` 标签；也不带文件系统的 `mountpoint`、`fstype` 标签。

#### 1.5.3 PromQL

以下磁盘查询保留每个设备的标签，`[5m]` 表示最近 5 分钟的计算窗口。查看指定设备时，可为表达式中的每个磁盘指标添加 `{instance="host-a:9100", device="nvme0n1"}`。

| 用途 | PromQL | 说明 |
|---|---|---|
| 读取吞吐量（MiB/s） | `rate(node_disk_read_bytes_total[5m]) / 1024^2` | 每秒读取的数据量。 |
| 写入吞吐量（MiB/s） | `rate(node_disk_written_bytes_total[5m]) / 1024^2` | 每秒写入的数据量。 |
| 读 IOPS（次/s） | `rate(node_disk_reads_completed_total[5m])` | 每秒完成的读请求数。 |
| 写 IOPS（次/s） | `rate(node_disk_writes_completed_total[5m])` | 每秒完成的写请求数。 |
| 读写总 IOPS（次/s） | `rate(node_disk_reads_completed_total[5m]) + rate(node_disk_writes_completed_total[5m])` | 同一设备的读写操作速率之和。 |
| 平均读取延迟（ms/次） | `1000 * rate(node_disk_read_time_seconds_total[5m]) / rate(node_disk_reads_completed_total[5m])` | 块设备统计层面的平均读取耗时。 |
| 平均写入延迟（ms/次） | `1000 * rate(node_disk_write_time_seconds_total[5m]) / rate(node_disk_writes_completed_total[5m])` | 块设备统计层面的平均写入耗时。 |
| 平均未完成 I/O 数量（个） | `rate(node_disk_io_time_weighted_seconds_total[5m])` | 包含正在处理和排队的请求，不是纯等待队列长度。 |
| 当前未完成 I/O 数量（个） | `node_disk_io_now` | 抓取时刻的瞬时值，可能漏掉两次抓取之间的短暂峰值。 |
| 设备 I/O 活跃时间占比（%） | `100 * rate(node_disk_io_time_seconds_total[5m])` | 近似反映设备持续有 I/O 的程度，不是磁盘性能上限的百分比。 |
| 平均读取大小（KiB/次） | `rate(node_disk_read_bytes_total[5m]) / rate(node_disk_reads_completed_total[5m]) / 1024` | 每次完成的读取平均涉及多少数据。 |
| 平均写入大小（KiB/次） | `rate(node_disk_written_bytes_total[5m]) / rate(node_disk_writes_completed_total[5m]) / 1024` | 每次完成的写入平均涉及多少数据。 |
| 读请求合并速率（次/s） | `rate(node_disk_reads_merged_total[5m])` | 观察块层读请求合并情况，无通用的好坏阈值。 |
| 写请求合并速率（次/s） | `rate(node_disk_writes_merged_total[5m])` | 观察块层写请求合并情况，无通用的好坏阈值。 |
| CPU I/O 等待时间占比（%） | `100 * avg by (instance) (rate(node_cpu_seconds_total{mode="iowait"}[5m]))` | CPU 辅助指标，非单盘指标；对各逻辑 CPU 的 iowait 比例取平均，不能定位到某块磁盘。 |

#### 1.5.4 如何理解 iowait

`node_cpu_seconds_total` 的定义和标签见 [CPU 小节](#11-cpu)。这里保留 iowait 查询，用于辅助排查磁盘性能。

`iowait` 大致表示 CPU 空闲、同时存在相关未完成 I/O 时被记账的等待时间。它不是任务等待时间占比，也不是等待磁盘的任务数量占比。例如一个 CPU 在 10 秒内有 2 秒被记录为 iowait，对应比例为 20%。

如果任务等待 I/O 期间，CPU 持续执行其他可运行任务，时间会记入 `user`、`system` 等模式，iowait 就可能很低。但仅仅“有其他任务”还不够，要看这些任务是否让 CPU 忙起来；多核主机的其他核心仍可能记录 iowait。

因此，**iowait 低不代表磁盘快，也不代表没有任务等待 I/O**。Linux 内核也指出 iowait 在多核环境中难以准确归属，某些情况下计数还可能下降，应把它作为辅助线索。参见 [Linux /proc/stat 说明](https://docs.kernel.org/filesystems/proc.html)。

#### 1.5.5 结合指标判断性能

| 现象 | 分析方向 |
|---|---|
| 平均延迟升高，未完成 I/O 数量持续增加 | 请求可能在积压，需要结合工作负载和存储能力排查。 |
| 吞吐量高，延迟稳定 | 可能是正常的大量读写，不能仅凭流量判定故障。 |
| IOPS 高，吞吐量不高 | 通常说明平均 I/O 较小，可用平均读写大小验证。 |
| 活跃时间接近 100% | 表示持续有 I/O，不意味着 SSD、NVMe 或 RAID 已达到最大性能。 |
| iowait 低，但磁盘延迟高 | CPU 可能正在执行其他任务，不能据此排除 I/O 瓶颈。 |

没有读写时，平均延迟和平均 I/O 大小的分母可能为 0，结果可能为 `NaN`，展示时应处理为空值。平均延迟不是应用请求延迟，也无法直接给出 P95、P99。请求大小不能单独用来判断随机或顺序访问。

汇总设备前应区分物理盘、分区和 device mapper 等层次，避免重复计算。磁盘统计字段含义及内核版本相关的计时限制见 [Linux I/O 统计说明](https://docs.kernel.org/admin-guide/iostats.html)。

### 1.6 网络

#### 1.6.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `node_network_receive_bytes_total`、`node_network_transmit_bytes_total` | Counter / 字节 | 网卡累计接收、发送数据量 |
| `node_network_receive_packets_total`、`node_network_transmit_packets_total` | Counter / 个 | 累计接收、发送包数 |
| `node_network_receive_drop_total`、`node_network_transmit_drop_total` | Counter / 个 | 累计接收、发送丢包数 |
| `node_network_receive_errs_total`、`node_network_transmit_errs_total` | Counter / 个 | 累计接收、发送错误数 |

#### 1.6.2 常用标签和值

本节列出的网络指标均使用 `device` 区分网络接口：

| label | 常见值示例 | 含义 |
|---|---|---|
| `device` | `eth0`、`ens192`、`enp0s3` | 网络接口名，具体命名取决于系统 |
| `device` | `lo` | 回环接口 |
| `device` | `bond0`、`br0`、`vethabc123` | 聚合接口、网桥、虚拟以太网接口等 |

例如 `node_network_receive_bytes_total{device="eth0"}` 表示 eth0 累计接收的字节数。接收和发送体现在指标名中，没有 `direction` 标签，也不按远端 IP、端口或进程拆分。

#### 1.6.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| 每张网卡接收速率（Mbit/s） | `8 * rate(node_network_receive_bytes_total{device!="lo"}[5m]) / 1e6` | 字节乘以 8 转为 bit；这里 Mbit 使用十进制。发送速率替换为 `node_network_transmit_bytes_total`。 |
| 每张网卡接收丢包速率（个/s） | `rate(node_network_receive_drop_total{device!="lo"}[5m])` | 这里是每秒丢包数量，不是丢包百分比。示例只排除回环网卡；汇总主机流量时还需按实际网络结构选择接口，避免 bridge、veth、bond 及其成员接口重复计数。 |

### 1.7 抓取状态与运行时间

#### 1.7.1 指标

| 指标 | 类型 / 单位 | 含义 |
|---|---|---|
| `up` | Gauge / 0 或 1 | Prometheus 最近一次抓取目标是否成功 |
| `node_scrape_collector_success` | Gauge / 0 或 1 | 各 collector 是否采集成功 |
| `node_boot_time_seconds` | Gauge / 秒 | 主机启动时刻的 Unix 时间戳 |

#### 1.7.2 常用标签和值

| 指标 | label 与值示例 | 含义 |
|---|---|---|
| `up` | `job="node"`、`instance="10.0.0.10:9100"` | Prometheus 生成的目标抓取状态，带有目标标签 |
| `node_scrape_collector_success` | `collector="cpu"`、`"loadavg"`、`"meminfo"`、`"filesystem"`、`"diskstats"`、`"netdev"`、`"stat"` | 每个值分别标识一个 collector；Prometheus 抓取后还会附加目标标签 |
| `node_boot_time_seconds` | 无额外 exporter 标签 | Prometheus 中通常通过 `job`、`instance` 区分主机 |

例如 `node_scrape_collector_success{collector="diskstats"} 1` 中，`diskstats` 是标签值，`1` 是表示成功的指标样本值。

#### 1.7.3 PromQL

| 用途 | PromQL | 说明 |
|---|---|---|
| Prometheus 是否成功抓取目标 | `up` | `1` 表示最近一次抓取成功，`0` 表示失败。`up` 是 Prometheus 生成的指标，不是 node_exporter 暴露的主机指标。失败也可能是网络、权限或 exporter 异常，不能直接认定主机宕机；目标被移出服务发现后，还可能没有该序列。 |
| 各 collector 是否采集成功 | `node_scrape_collector_success` | 按 `collector` 标签查看各采集器结果：`1` 成功、`0` 失败。HTTP 抓取成功不保证每个 collector 都采集成功。 |
| 主机运行时间（小时） | `(time() - node_boot_time_seconds) / 3600` | 启动时间是 Unix 时间戳。该表达式依赖主机和 Prometheus 时钟基本同步。 |

### 1.8 使用注意与参考

指标是否存在取决于操作系统、内核、node_exporter 版本、权限以及启用的 collector。缺失指标不代表指标值为 0，应先检查目标 `/metrics` 中的 `HELP`、`TYPE` 和标签，以及 collector 状态。本文示例未在实际监控环境执行，需要按抓取配置和设备布局调整。

- [node_exporter 官方说明与 collector 列表](https://github.com/prometheus/node_exporter/blob/master/README.md)
- [官方测试输出：指标名称、类型和 HELP 说明](https://raw.githubusercontent.com/prometheus/node_exporter/master/collector/fixtures/e2e-64k-page-output.txt)
- [Prometheus 查询基础](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Prometheus 查询函数](https://prometheus.io/docs/prometheus/latest/querying/functions/)
