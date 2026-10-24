# 锁函数

MySQL 的 Locking Functions 操作用户级命名锁，用于应用间协作互斥。锁由服务端管理，不需要创建业务锁表；以下按 MySQL 8.4 说明。

## 1. 函数与返回值

| 函数 | 用途与主要返回值 |
| --- | --- |
| `GET_LOCK(name, timeout)` | 申请锁；1 成功，0 等待超时，NULL 出错；timeout 单位为秒 |
| `RELEASE_LOCK(name)` | 释放一次持有；1 成功，0 当前会话未持有，NULL 锁不存在 |
| `RELEASE_ALL_LOCKS()` | 释放当前会话全部命名锁，返回释放次数 |
| `IS_FREE_LOCK(name)` | 1 空闲，0 已占用，NULL 出错 |
| `IS_USED_LOCK(name)` | 返回持有者连接 ID，未持有时返回 NULL |

## 2. 基本使用

```sql
-- 会话 A：不等待，必须检查返回值为 1 后再执行工作。
SELECT GET_LOCK('demo.report', 0);

-- 会话 B：A 持有期间返回 0。
SELECT GET_LOCK('demo.report', 0);

-- 会话 A：完成后在同一连接释放。
SELECT RELEASE_LOCK('demo.report');
```

## 3. 生命周期

锁绑定会话，COMMIT 和 ROLLBACK 不会释放；显式释放或服务端确认会话结束后释放。无需通过定时 SQL 给锁续期，网络中断的检测可能存在延迟。

同一会话可重复取得同名锁，需要释放相应次数。使用连接池时保留持锁连接，先释放锁再归还；归还连接池不等于断开连接。

## 4. 与数据锁的区别

命名锁依赖应用主动遵守名称约定，不自动阻止其他连接修改表或记录。它作用于单个 MySQL 服务端实例，不自动跨多个数据库实例协调。

多个应用共享该实例时，锁名应带应用前缀；名称最长 64 字符。多锁等待仍可能死锁，锁也不会自动停止业务代码。

## 5. 参考资料

- [MySQL 8.4：Locking Functions](https://dev.mysql.com/doc/refman/8.4/en/locking-functions.html)。
