---
color: "#5B6CDA"
---

# kafka-consumer-groups.sh

`bin/kafka-consumer-groups.sh` 是 Apache Kafka 官方发行包自带的消费组管理工具。查询时使用 `--bootstrap-server` 指定 Kafka 集群，用 `--group` 指定消费组 ID。

## 1. 查看分区进度

```bash
bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group group-a
```

默认的 `--describe` 按 Topic 和分区显示消费进度以及当前负责该分区的消费者。

| 字段 | 含义 |
| --- | --- |
| `TOPIC`、`PARTITION` | Topic 名称和分区编号。 |
| `CURRENT-OFFSET` | 消费组在该分区的已提交 offset。 |
| `LOG-END-OFFSET` | 工具查询到的分区末尾 offset。 |
| `LAG` | `LOG-END-OFFSET - CURRENT-OFFSET`。分区尚无已提交进度时显示 `-`。 |
| `CONSUMER-ID`、`HOST`、`CLIENT-ID` | 当前负责该分区的成员 ID、客户端地址和客户端 ID。 |

例如，`group-a` 在 `orders-0` 的 `CURRENT-OFFSET` 是 `43`，`LOG-END-OFFSET` 是 `50`，则 `LAG` 为 `7`。

## 2. 查看成员分配

```bash
bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group group-a --members --verbose
```

输出按成员列出 `CONSUMER-ID`、`HOST`、`CLIENT-ID` 和分区数量（`#PARTITIONS`）。`--verbose` 还会显示每个成员具体负责的 Topic 分区（`ASSIGNMENT`）。

## 3. 默认输出的数据来源

默认的 `--describe` 用 `describeConsumerGroups()` 查询分区分配，用 `listConsumerGroupOffsets()` 查询已提交 offset，再用 `listOffsets()` 查询分区末尾 offset。工具按 Topic 和分区汇总查询结果，并计算 `LAG`。
