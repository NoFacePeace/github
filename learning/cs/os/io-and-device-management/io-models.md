# I/O 模型

以 Socket 接收数据为例，可以将一次读取理解为两个阶段：等待数据就绪，以及将数据从内核缓冲区复制到应用程序缓冲区。不同 I/O 模型的主要区别在于：应用程序如何等待数据，以及由谁发起并完成数据读取。

## 1. 基本概念

### 1.1 同步与异步

以读取数据为例，同步与异步的区别在于如何提交读取，以及何时取得完成结果。两者都可以由系统将数据写入应用程序提供的缓冲区（buffer）。

- **同步 I/O**：应用程序调用 `read` / `recv`，从本次调用取得读取结果。成功返回时，本次读到的数据已经在 buffer 中；非阻塞读取在没有数据时会返回 `EAGAIN` 等错误，需要应用程序之后重试。
- **异步 I/O**：应用程序先提交读取请求并指定 buffer，再通过完成通知或完成队列取得结果。请求成功完成时，数据已经在 buffer 中，无需再次调用 `read` / `recv`。

例如，读取 `hello` 的过程可以表示为：

```text
同步读取：调用 read(buffer) → 返回读取字节数 5，buffer 中已有 hello
异步读取：提交读取请求(buffer) → 取得完成结果 → 检查读到 5 字节，使用 buffer 中的 hello
```

异步请求提交成功并不表示读取已经完成；使用 buffer 中的数据前，需要取得完成结果，并检查错误和实际读取字节数。异步请求也可能很快完成，关键是提交请求与取得完成结果在接口上分开。

在这里介绍的五种经典模型中，阻塞 I/O、非阻塞 I/O、I/O 多路复用和信号驱动 I/O 都属于同步 I/O，只有异步 I/O 属于异步模型。

### 1.2 阻塞与非阻塞

阻塞与非阻塞描述的是：调用无法立即取得进展时，是否等待。

- **阻塞读取**：暂时没有数据时，调用线程等待，直到可以返回数据、遇到错误或其他结束条件。
- **非阻塞读取**：暂时没有数据时，立即返回 `EAGAIN` 或 `EWOULDBLOCK`，应用程序可以先执行其他工作，再尝试读取。

同步不等于阻塞，非阻塞也不等于异步。非阻塞读取返回 `EAGAIN` 后，不会留下一个等待系统继续完成的读取请求；异步请求成功提交后，系统负责执行该请求，应用程序之后取得完成结果。

### 1.3 就绪通知与完成通知

- **就绪通知**：表示可以尝试读写。例如，`epoll` 返回 Socket 可读事件后，应用程序仍需调用 `read` / `recv` 将数据读入 buffer。
- **完成通知**：表示此前提交的 I/O 请求已经结束。异步读取成功完成后，应用程序根据实际读取字节数使用 buffer 中的数据，无需再次发起读取来获取这批数据。

就绪事件也可能对应文件结束或错误，完成通知也可能报告失败，因此都需要检查实际操作结果。

非阻塞 I/O 与多路复用通常配合使用：多路复用帮助应用程序等待多个连接的就绪事件，非阻塞模式保证后续读取在无数据时立即返回。

## 2. 阻塞 I/O

应用程序调用读取接口后，如果没有数据可读，调用线程会阻塞。数据就绪并复制到应用程序缓冲区后，调用才返回。其流程简单，但等待期间该线程无法执行其他任务。

## 3. 非阻塞 I/O

应用程序将 Socket 设置为非阻塞模式后，如果没有数据可读，读取调用会立即返回错误，例如 `EAGAIN` 或 `EWOULDBLOCK`。应用程序可以稍后重试；持续循环重试会消耗 CPU。数据就绪时，应用程序仍需调用读取接口获取数据。

## 4. I/O 多路复用

应用程序通过 `select`、`poll` 或 `epoll` 等接口等待多个文件描述符的就绪事件，再对就绪的描述符执行读写。等待接口本身可以阻塞，但一个线程能够同时等待多个连接，无需为每个连接分别阻塞一个线程。

就绪通知表示当前可以尝试 I/O，并不表示数据已经读入应用程序缓冲区。实际使用时通常配合非阻塞 Socket，避免就绪状态变化导致后续读写意外阻塞。

### 4.1 select

#### 4.1.1 数据结构

`select` 使用 `fd_set` 表示关注的文件描述符集合，分别传入读、写和异常条件集合。可以将集合理解为位图：某一位对应一个描述符，置位表示关注它。

Linux 常用的 glibc `fd_set` 固定为 1024 位，`FD_SETSIZE` 为 1024。传给辅助宏的 fd 必须有效，且满足 `0 <= fd < FD_SETSIZE`；超出范围会导致未定义行为。这是描述符编号限制，并不意味着只要连接数少于 1024 就一定可用。

#### 4.1.2 辅助宏

声明 `fd_set readfds;` 后，可以使用以下辅助宏操作集合：

| 辅助宏 | 作用 | 使用示例 |
| --- | --- | --- |
| `FD_ZERO(set)` | 清空集合，首次使用前用于初始化 | `FD_ZERO(&readfds);` |
| `FD_SET(fd, set)` | 将指定 fd 加入集合 | `FD_SET(3, &readfds);` |
| `FD_CLR(fd, set)` | 将指定 fd 从集合中移除 | `FD_CLR(3, &readfds);` |
| `FD_ISSET(fd, set)` | 判断指定 fd 是否在集合中，在集合中返回非零，否则返回 0 | `if (FD_ISSET(3, &readfds)) { /* 处理 fd 3 */ }` |

这些宏操作的是用户态的集合，不会创建或关闭 fd，也不会向内核长期注册监听。

#### 4.1.3 接口

`select()` 在 `<sys/select.h>` 中声明：

```c
int select(int nfds, fd_set *readfds, fd_set *writefds,
           fd_set *exceptfds, struct timeval *timeout);
```

| 参数 | 含义 |
| --- | --- |
| `nfds` | 所有关注集合中最大 fd 编号加一；内核检查编号小于 `nfds` 的描述符 |
| `readfds` | 关注可读条件的集合；成功返回后只保留满足可读条件的 fd |
| `writefds` | 关注可写条件的集合；成功返回后只保留满足可写条件的 fd |
| `exceptfds` | 关注异常条件的集合，例如 Socket 带外数据；并非所有 I/O 错误的统一通知集合 |
| `timeout` | 最长等待时间，由 `tv_sec`（秒）和 `tv_usec`（微秒）组成；全为 0 时立即检查并返回，传 `NULL` 时不设置超时 |

不关注的集合可以传 `NULL`。三个集合都是输入输出参数，就绪结果会覆盖传入的关注集合。

| 返回值 | 含义 |
| --- | --- |
| 大于 0 | 三个返回集合中置位的总数；同一个 fd 在不同集合中就绪时会分别计数 |
| 等于 0 | 等待超时，没有就绪描述符 |
| 等于 -1 | 调用失败，通过 `errno` 获取原因，例如信号中断 `EINTR` 或无效描述符 `EBADF` |

#### 4.1.4 调用流程

一次调用的流程如下：

1. 应用程序使用 `FD_ZERO`、`FD_SET` 构建集合，设置 `nfds` 和超时时间，调用 `select`。
2. 内核读取集合，检查所关注文件的就绪状态；在需要等待时，将当前线程关联到相应等待队列。
3. 如果没有就绪事件且允许等待，线程睡眠；被事件唤醒后，内核重新检查就绪状态，也可能因超时或信号而返回。
4. 返回的集合只保留就绪描述符。应用程序使用 `FD_ISSET` 检查集合，再执行实际读写。

```text
输入读集合：{A, B, C}
        ↓ select
输出读集合：{B, C}
        ↓
应用程序检查集合，调用 read(B)、read(C)
```

`select()` 会修改集合，在 Linux 上还会修改 `timeout`。循环等待时，每轮调用前需要重新设置或从备份恢复集合，并按本轮等待需求设置超时时间。

内核需要按描述符范围检查集合，应用程序也需要检查哪些描述符被置位；连接很多时，这些检查会增加开销。

参考：[select 接口文档](https://man7.org/linux/man-pages/man2/select.2.html)、[Linux select/poll 实现](https://github.com/torvalds/linux/blob/master/fs/select.c)。

#### 4.1.5 使用示例

下面通过 `select()` 等待标准输入，最多等待 5 秒。在终端中运行后，输入一行文字并按回车，即可触发可读事件。

```c
#include <stdio.h>
#include <unistd.h>
#include <sys/select.h>

int main(void) {
    int fd = STDIN_FILENO;  // 标准输入，fd 为 0

    fd_set readfds;
    FD_ZERO(&readfds);
    FD_SET(fd, &readfds);

    struct timeval timeout = {
        .tv_sec = 5,
        .tv_usec = 0
    };

    printf("请输入内容，等待 5 秒：\n");
    fflush(stdout);

    // 最大 fd 为 0，因此 nfds 为 1；只关注可读事件。
    int ready = select(fd + 1, &readfds, NULL, NULL, &timeout);
    if (ready == -1) {
        perror("select");
        return 1;
    }
    if (ready == 0) {
        printf("等待超时\n");
        return 0;
    }

    if (FD_ISSET(fd, &readfds)) {
        char buffer[1024];
        ssize_t n = read(fd, buffer, sizeof(buffer) - 1);
        if (n > 0) {
            buffer[n] = '\0';
            printf("读取到：%s", buffer);
        } else if (n == 0) {
            printf("输入已结束（EOF）\n");
        } else {
            perror("read");
            return 1;
        }
    }
    return 0;
}
```

保存为 `select_example.c`，在 Linux 或 macOS 上编译运行：

```bash
cc -Wall -Wextra select_example.c -o select_example
./select_example
```

输入结束也会使标准输入可读，此时 `read()` 返回 `0`，因此可读不一定代表有新的数据。

这个示例只等待一次，并在系统调用报错时退出。

### 4.2 poll

#### 4.2.1 数据结构

`poll` 使用 `pollfd` 数组描述关注的对象，每一项分别保存描述符、关注事件和返回事件：

```c
struct pollfd {
    int fd;         // 文件描述符
    short events;   // 应用程序关注的事件，例如 POLLIN
    short revents;  // 内核返回的事件
};
```

数组将 fd 保存为整数，不使用 fd 编号作为位图下标，因此没有 `fd_set` 的固定 1024 位限制。例如，一个元素就可以保存 `fd = 1500`。数组规模仍受内存及进程资源上限约束。

`events` 与 `revents` 是事件位掩码，可以组合多个标志：

| 标志 | 含义 |
| --- | --- |
| `POLLIN` | 可读 |
| `POLLOUT` | 可写；不保证任意大小的写入都不会阻塞 |
| `POLLPRI` | 异常条件，例如 TCP 带外数据 |
| `POLLERR` | 错误 |
| `POLLHUP` | 挂断；管道或流式 Socket 中仍可能有未读数据 |
| `POLLNVAL` | fd 未打开 |

前三项可放入 `events`，通过按位或组合，例如 `POLLIN | POLLOUT`；通过按位与检查 `revents`。后三项只用于返回结果，无需请求。将某项 `fd` 设为负数，可以让本轮调用忽略该项。

#### 4.2.2 接口

`poll()` 在 `<poll.h>` 中声明：

```c
int poll(struct pollfd *fds, nfds_t nfds, int timeout);
```

| 参数 | 含义 |
| --- | --- |
| `fds` | 关注数组；应用填写 `fd`、`events`，内核填写 `revents` |
| `nfds` | 数组元素个数，不是最大 fd 加一 |
| `timeout` | 等待时间，单位为毫秒；0 表示立即检查，负数表示不设置超时 |

| 返回值 | 含义 |
| --- | --- |
| 大于 0 | `revents` 非零的数组元素个数；同一项有多个事件只计一次 |
| 等于 0 | 等待超时 |
| 等于 -1 | 调用失败，通过 `errno` 获取原因，例如信号中断 `EINTR` |

#### 4.2.3 调用流程

一次调用的流程如下：

1. 应用程序构建数组，填写各项的 `fd` 和 `events`，调用 `poll`。
2. 内核读取数组，逐项检查文件的就绪状态，并在需要时注册等待队列。
3. 没有事件时可进入睡眠；被唤醒后重新检查，或者因超时、信号而返回。
4. 内核通过各项的 `revents` 返回事件。应用程序遍历数组，对有事件的描述符执行读写。

```text
输入数组：[{A, POLLIN}, {B, POLLIN}, {C, POLLIN}]
        ↓ poll
返回事件：[A: 无事件, B: POLLIN, C: POLLIN]
        ↓
应用程序遍历 revents，调用 read(B)、read(C)
```

`events` 不会被返回结果覆盖，循环等待时可以复用关注数组，内核会重新填写 `revents`。

`poll` 每轮仍需传入数组，内核和应用程序仍需遍历相关条目。它改善了 `select` 的接口和编号限制，没有消除大量连接下的扫描开销。另外，`poll` 等待时可以睡眠，并不是一直占用 CPU 忙轮询。

参考：[poll 接口文档](https://man7.org/linux/man-pages/man2/poll.2.html)、[Linux select/poll 实现](https://github.com/torvalds/linux/blob/master/fs/select.c)。

#### 4.2.4 使用示例

下面与 `select` 示例一样，等待标准输入，最多等待 5 秒。终端中输入文字并按回车后读取一次，然后退出。

```c
#include <stdio.h>
#include <unistd.h>
#include <poll.h>

int main(void) {
    struct pollfd fds[1] = {
        {.fd = STDIN_FILENO, .events = POLLIN, .revents = 0}
    };

    printf("请输入内容，等待 5 秒：\n");
    fflush(stdout);

    int ready = poll(fds, 1, 5000);
    if (ready == -1) {
        perror("poll");
        return 1;
    }
    if (ready == 0) {
        printf("等待超时\n");
        return 0;
    }
    if (fds[0].revents & POLLNVAL) {
        fprintf(stderr, "标准输入 fd 无效\n");
        return 1;
    }
    if (fds[0].revents & POLLERR) {
        fprintf(stderr, "标准输入发生错误\n");
        return 1;
    }
    // 挂断时仍可能有剩余数据，通过 read 的结果判断是否 EOF。
    if (fds[0].revents & (POLLIN | POLLHUP)) {
        char buffer[1024];
        ssize_t n = read(fds[0].fd, buffer, sizeof(buffer) - 1);
        if (n > 0) {
            buffer[n] = '\0';
            printf("读取到：%s", buffer);
        } else if (n == 0) {
            printf("输入已结束（EOF）\n");
        } else {
            perror("read");
            return 1;
        }
    }
    return 0;
}
```

保存为 `poll_example.c`，在 Linux 或 macOS 上编译运行：

```bash
cc -Wall -Wextra poll_example.c -o poll_example
./poll_example
```

这个示例只读取一次。持续读取管道或流式 Socket 时，不能仅因出现 `POLLHUP` 就丢弃剩余数据，应继续处理读取结果，直到 EOF 或错误。

### 4.3 epoll

#### 4.3.1 数据结构

`epoll` 是 Linux 提供的 I/O 就绪通知机制。每个 epoll 实例在内核中维护关注集合与就绪集合，应用程序通过一个 epoll 文件描述符访问实例。

Linux 实现中，关注项通过[红黑树](../../data-structures/trees/red-black-tree.md)组织，另外维护就绪链表。它没有 `fd_set` 的固定 1024 位限制，但仍受文件描述符、内核内存和监听项数量等资源限制。

应用程序通过 `struct epoll_event` 配置关注事件、接收返回事件：

```c
struct epoll_event {
    uint32_t events;    // 事件位掩码
    epoll_data_t data;  // 应用程序设置的关联数据
};
```

`epoll_data_t` 是联合体，可以保存 `fd`、指针或整数标识。内核在返回事件时带回注册时设置的 `data`；`data.fd` 需要应用程序主动填写，不会自动设置为被监听的 fd。

| 事件标志 | 含义 |
| --- | --- |
| `EPOLLIN` | 可读 |
| `EPOLLOUT` | 可写；不保证任意大小的写入都不会阻塞 |
| `EPOLLPRI` | 异常条件，例如 TCP 带外数据 |
| `EPOLLERR` | 错误，无需显式请求也会报告 |
| `EPOLLHUP` | 挂断，无需显式请求也会报告；仍可能存在未读数据 |
| `EPOLLRDHUP` | 流式 Socket 对端关闭连接或关闭写方向，需显式请求 |

使用按位或组合关注标志，通过按位与检查返回事件。触发方式由后文的 `EPOLLET` 等选项控制。

#### 4.3.2 接口

核心接口在 `<sys/epoll.h>` 中声明：

```c
int epoll_create1(int flags);
int epoll_ctl(int epfd, int op, int fd, struct epoll_event *event);
int epoll_wait(int epfd, struct epoll_event *events,
               int maxevents, int timeout);
```

| 接口 | 参数说明 | 返回值 |
| --- | --- | --- |
| `epoll_create1` | `flags` 为 0 或 `EPOLL_CLOEXEC`，后者使返回的 fd 在成功执行 exec 时自动关闭 | 成功返回非负的实例 fd，失败返回 -1 |
| `epoll_ctl` | `epfd` 为实例 fd，`op` 为操作，`fd` 为目标 fd，`event` 为关注配置 | 成功返回 0，失败返回 -1 |
| `epoll_wait` | `events` 为输出数组，`maxevents` 为其容量且必须大于 0；`timeout` 单位为毫秒，0 表示立即检查，-1 表示无限等待 | 成功返回本次事件条目数，超时返回 0，失败返回 -1 |

`epoll_ctl` 支持以下操作：

| 操作 | 作用 |
| --- | --- |
| `EPOLL_CTL_ADD` | 添加目标 fd 及其关注配置 |
| `EPOLL_CTL_MOD` | 更新已注册目标的关注事件和关联数据 |
| `EPOLL_CTL_DEL` | 移除目标的关注项，`event` 可传 `NULL`；不会关闭目标 fd |

接口返回 -1 时，通过 `errno` 检查原因，例如 `epoll_wait` 被信号中断时的 `EINTR`。`maxevents` 只限制单次返回的条目数，不是整个实例的监听数量上限。不再使用实例时调用 `close(epfd)`，被监听的 fd 由应用程序另行管理。

普通磁盘文件通常不支持注册到 epoll，`epoll_ctl` 会以 `EPERM` 失败；Socket、管道等是典型的监听对象。

参考：[epoll_create1](https://man7.org/linux/man-pages/man2/epoll_create.2.html)、[epoll_ctl](https://man7.org/linux/man-pages/man2/epoll_ctl.2.html)、[epoll_wait](https://man7.org/linux/man-pages/man2/epoll_wait.2.html)。

#### 4.3.3 调用流程

1. 应用程序调用 `epoll_create1` 创建实例。
2. 使用 `epoll_ctl` 注册目标 fd。内核保存关注项，将回调关联到目标的等待队列，并检查已有的就绪状态。
3. 调用 `epoll_wait` 等待事件。目标发生事件时，回调将对应项标记为就绪，并在需要时唤醒等待线程。
4. 内核检查就绪项并返回一批事件；应用程序遍历结果，执行实际读写。
5. 处理完事件后继续调用 `epoll_wait()`；关注对象或事件变化时才调用 `epoll_ctl()` 更新注册。

```text
epoll_ctl 注册 A、B、C → 内核保存关注集合
                              ↓
B 有数据 → 等待队列回调 → B 进入就绪链表
                              ↓
epoll_wait → 检查就绪项，向应用程序返回事件
                              ↓
                          read(B)
```

`epoll_wait` 从就绪项中检查并返回事件，无需每次由应用程序重新提交完整关注集合。它仍然需要把事件结果复制到用户空间，也不会自动把 Socket 数据读入应用 buffer。

**按需唤醒**：符合通知条件的事件到达后，回调记录就绪状态，并根据等待线程的状态触发唤醒。

| 线程状态 | 是否有尚未返回的 `epoll_wait()` | 事件到达后的处理 |
| --- | --- | --- |
| 正在 `epoll_wait()` 中睡眠 | 是 | 记录就绪状态，唤醒线程，使其变为可运行 |
| 正在处理业务，尚未再次调用 `epoll_wait()` | 否 | 记录就绪状态，供下一次调用检查，无需唤醒该线程 |
| 已被先前事件唤醒，但尚未获得 CPU | 是 | 继续记录就绪状态；线程恢复后可能在本次调用中取得这些事件 |

以上按单个线程说明；唤醒只使线程变为可运行，不保证立即获得 CPU 执行。

#### 4.3.4 红黑树操作

红黑树用于管理关注集合，主要在注册变更和资源清理时访问：

| 场景 | 红黑树操作 |
| --- | --- |
| `EPOLL_CTL_ADD` | 查找是否已注册；未注册时插入新节点，必要时调整树的平衡 |
| `EPOLL_CTL_MOD` | 查找已有节点，更新关注事件和关联数据，通常不改变树结构 |
| `EPOLL_CTL_DEL` | 查找并删除节点，必要时调整树的平衡 |
| 被监听的底层文件对象释放 | 通过反向关联找到监听项，直接删除对应节点 |
| epoll 实例释放 | 遍历并清理树中的监听项 |

Socket 收到数据时，等待队列回调已经关联到对应的 `epitem`，可以直接将其标记为就绪，无需按 fd 搜索红黑树。`epoll_wait()` 的常规事件返回流程检查的是就绪链表，也不需要遍历整棵红黑树。

```text
epoll_ctl → 红黑树：管理关注项
数据到达 → 等待队列回调 → epitem → 就绪链表 → epoll_wait
```

LT 下重复通知、ET 下等待新事件，主要影响就绪事件的处理，不会因每次通知就重新插入或删除红黑树节点。红黑树的查找、插入和删除具有 O(log N) 的复杂度，其中 N 为关注项数量；这不代表整个 `epoll_ctl()` 的全部工作都只有红黑树操作。

参考：[Linux epoll 实现](https://github.com/torvalds/linux/blob/master/fs/eventpoll.c)。

#### 4.3.5 触发模式

`epoll` 支持两种主要触发方式：

| 模式 | 通知行为 | 读取时的处理方式 |
| --- | --- | --- |
| 水平触发（LT，默认） | 只要仍满足就绪条件，后续等待仍可返回该事件 | 可以分批读取，未读完时仍可继续获得通知 |
| 边缘触发（ET，设置 `EPOLLET`） | 按新的就绪事件通知，不保证仅因仍有未读数据而再次通知 | 通常使用非阻塞读取，读到 `EAGAIN` 后再等待 |

例如，收到可读事件时有 100 字节，应用程序只读了 40 字节：LT 下剩余数据会使连接继续保持可读；ET 下如果直接回去等待，剩余的 60 字节可能无法触发下一次通知。

实际事件循环还应处理读取返回 `0`、错误和部分读写，并控制单个连接的处理量，避免一个连接长期占用线程。ET 下如果主动暂停读取，需要自行安排后续处理，不能只依赖新的边缘通知。

参考：[epoll 接口与触发模式](https://man7.org/linux/man-pages/man7/epoll.7.html)、[Linux epoll 实现](https://github.com/torvalds/linux/blob/master/fs/eventpoll.c)。

#### 4.3.6 关闭与自动清理

当 `close(fd)` 关闭了对底层打开文件对象的最后一个引用，内核会在释放该对象时自动清理对应的 epoll 监听项，无需先调用 `EPOLL_CTL_DEL`。

fd 是进程文件描述符表的索引，底层对象是内核中的 `struct file`。`dup()` 或 `fork()` 可以让多个 fd 引用同一个对象，因此关闭一个 fd 不一定触发对象释放：

```c
int fd2 = dup(fd);  // fd 和 fd2 引用同一个底层打开文件对象
close(fd);         // fd2 仍持有引用，不能认为监听项已自动移除
```

如果需要明确停止监听，应在关闭注册时使用的 fd 前删除关注项：

```c
epoll_ctl(epfd, EPOLL_CTL_DEL, fd, NULL);  // 实际代码需要检查返回值
close(fd);
```

内核能够直接找到待清理的监听项，是因为注册时建立了双向关联：

- epoll 实例的红黑树保存 `epitem` 监听项，每项关联被监听的文件对象。
- 文件对象通过反向关联记录监听它的 `epitem`，每个监听项又记录所属的 epoll 实例。

```text
close(fd)
    ↓ 通过当前进程的文件描述符表查找
底层文件对象 struct file
    ↓ 最后一个引用释放，进入对象清理
沿文件对象的反向关联找到 epitem
    ↓ 找到所属 epoll 实例
注销等待队列回调，移除红黑树节点及就绪链表关联
```

因此，清理时不需要扫描所有进程的 fd 或系统中的所有 epoll 实例。`epitem` 本身包含红黑树节点，内核可以直接删除该节点，无需再按 fd 从树根查找；删除仍可能涉及红黑树的平衡调整。

对端关闭连接只会改变本地 Socket 的状态，可能产生 EOF、`EPOLLRDHUP` 或 `EPOLLHUP`，不等于本地文件对象已释放；应用程序仍需处理事件并关闭自己的 fd。

参考：[epoll 的关闭与重复描述符说明](https://man7.org/linux/man-pages/man7/epoll.7.html)、[Linux epoll 清理实现](https://github.com/torvalds/linux/blob/master/fs/eventpoll.c)。

#### 4.3.7 使用示例

下面使用默认的 LT 模式等待标准输入，最多等待 5 秒。示例用于 Linux，需在终端中输入文字并按回车，或通过管道传入数据；不要将普通磁盘文件直接重定向为标准输入。

```c
#include <stdio.h>
#include <unistd.h>
#include <sys/epoll.h>

int main(void) {
    int epfd = epoll_create1(EPOLL_CLOEXEC);
    if (epfd == -1) {
        perror("epoll_create1");
        return 1;
    }

    struct epoll_event interest = {
        .events = EPOLLIN,
        .data.fd = STDIN_FILENO
    };
    if (epoll_ctl(epfd, EPOLL_CTL_ADD, STDIN_FILENO, &interest) == -1) {
        perror("epoll_ctl");
        close(epfd);
        return 1;
    }

    printf("请输入内容，等待 5 秒：\n");
    fflush(stdout);

    struct epoll_event events[1];
    int ready = epoll_wait(epfd, events, 1, 5000);
    int status = 0;
    if (ready == -1) {
        perror("epoll_wait");
        status = 1;
    } else if (ready == 0) {
        printf("等待超时\n");
    } else if (events[0].events & EPOLLERR) {
        fprintf(stderr, "标准输入发生错误\n");
        status = 1;
    } else if (events[0].events & (EPOLLIN | EPOLLHUP)) {
        char buffer[1024];
        ssize_t n = read(events[0].data.fd, buffer, sizeof(buffer) - 1);
        if (n > 0) {
            buffer[n] = '\0';
            printf("读取到：%s", buffer);
        } else if (n == 0) {
            printf("输入已结束（EOF）\n");
        } else {
            perror("read");
            status = 1;
        }
    }
    close(epfd);
    return status;
}
```

保存为 `epoll_example.c`，在 Linux 上编译运行：

```bash
cc -Wall -Wextra epoll_example.c -o epoll_example
./epoll_example
printf 'hello\n' | ./epoll_example
```

这个示例只有一个输入源，没有其他线程竞争读取，只等待并读取一次。实际服务器通常使用非阻塞 Socket，并处理 `EAGAIN`、EOF、错误和部分读写；不能仅添加 `EPOLLET` 就把此示例当作完整的 ET 事件循环。

### 4.4 三种机制对比

`select`、`poll` 和 `epoll` 都提供就绪通知：可以让等待线程睡眠，在事件发生后返回就绪信息，实际读写仍由应用程序执行。主要区别是关注集合如何保存、等待关系保留多久，以及唤醒后如何找到就绪对象。

#### 4.4.1 接口与数据结构

| 对比项 | select | poll | epoll |
| --- | --- | --- | --- |
| 核心接口 | `select()` | `poll()` | `epoll_create1()`、`epoll_ctl()`、`epoll_wait()` |
| 关注集合 | 用户态 `fd_set` 位图，每次调用传入 | 用户态 `pollfd` 数组，每次调用传入 | 内核保存关注集合，通过 `epoll_ctl` 更新 |
| 返回结果 | 覆盖输入集合，只保留就绪 fd | 填写数组各项的 `revents` | 将一批就绪事件写入输出数组 |
| 数量参数 | `nfds` 为最大关注 fd 加一 | `nfds` 为数组元素个数 | `maxevents` 为单次输出数组容量 |
| 固定 1024 限制 | Linux 常用 glibc 的 `fd_set` 要求 fd 小于 1024 | 无此位图限制 | 无此位图限制 |
| 后续等待 | 恢复或重建集合后再次传入 | 可复用数组，关注项变化时修改数组 | 复用内核中的注册，只在关注项变化时更新 |

`poll` 和 `epoll` 仍受文件描述符、内存等资源限制。`epoll` 是 Linux 特有接口，`select` 和 `poll` 则在多种 Unix 系统上提供。

#### 4.4.2 等待关系与回调

等待关系是内核记录的“对象发生事件时通知谁”。对象的等待队列保存等待项及其回调，网络数据则保存在 Socket 的接收缓冲区中。

| 对比项 | select / poll | epoll |
| --- | --- | --- |
| 目标对象上的等待项 | 本次调用需要等待时建立，调用返回时清理 | 注册监听时建立，随监听项保留 |
| 事件回调 | 标记本次等待被触发，按需唤醒等待线程 | 标记对应监听项就绪，按需唤醒等待 epoll 实例的线程 |
| 回调由谁指定 | 内核的 select/poll 实现 | 内核的 epoll 实现 |
| 返回后保留什么 | 不保留本次调用的临时等待关系 | 保留目标对象到 epoll 的注册关系；本次线程等待已结束 |

Socket 收到数据时触发自身等待队列，执行等待项中注册的回调。Socket 不需要判断应用使用了哪种 I/O 多路复用接口，回调的不同实现决定后续行为。

`epoll` 中需要区分两层关系：

```text
目标 Socket 的等待队列
    → 持续注册的回调：将对应 epitem 标记为就绪
    → epoll 实例的就绪链表
    → epoll 实例的等待队列：按需唤醒等待线程
```

`epoll_wait()` 返回后，线程无需继续等待；下一次调用发现没有就绪事件且允许等待时，才重新加入实例的等待队列并睡眠。目标对象上的注册回调仍然保留，因此线程不在等待时到达的事件也可以被记录。

#### 4.4.3 唤醒与就绪检查

以关注 A、B、C，只有 B 收到数据为例：

```text
select / poll：
B 收到数据 → 回调唤醒线程 → 重新扫描 A、B、C → 返回 B 就绪

epoll：
B 收到数据 → 回调将 B 标记为就绪 → 按需唤醒线程
                                      ↓
                              检查就绪项 → 返回 B
```

唤醒只是让睡眠线程具备被调度执行的条件，不代表线程立即运行，也不代表数据已经复制到应用程序缓冲区。三种机制返回后都需要检查实际读写结果。

如果事件先于 `epoll_wait()` 到达，回调可以先记录就绪项；后续调用检查到可用事件时直接返回，无需先睡眠再唤醒。内核通过同步与再次检查就绪状态，避免事件在检查和进入睡眠之间到达而丢失唤醒。

`select` 和 `poll` 同样会在睡眠前检查目标的就绪状态，因此数据提前到达也不会仅因当时没有等待线程而被忽略。区别在于它们需要重新扫描本次关注集合，而 epoll 的常规返回路径检查就绪链表中的候选项。

#### 4.4.4 开销与适用场景

| 机制 | 主要开销 | 适用场景 |
| --- | --- | --- |
| select | 按 fd 范围扫描集合、传递集合、建立和清理临时等待项 | fd 编号较小、关注对象较少的简单程序 |
| poll | 扫描数组、传递数组、建立和清理临时等待项 | 需要跨 Unix 系统使用，或需要处理较大 fd 编号的程序 |
| epoll | 维护内核关注集合、处理回调、检查并复制就绪事件 | Linux 上关注对象多、每轮只有部分对象活跃的事件循环 |

epoll 的收益来自复用监听关系和减少全量扫描，不是独有的“唤醒能力”。当多数连接都活跃或关注集合频繁变化时，优势需要结合实际负载评估；不能把整个 epoll 的开销概括为 O(1)。

参考：[Linux select/poll 实现](https://github.com/torvalds/linux/blob/master/fs/select.c)、[Linux epoll 实现](https://github.com/torvalds/linux/blob/master/fs/eventpoll.c)。

## 5. 信号驱动 I/O

应用程序启用信号通知后，可以继续执行其他任务。当发生相关 I/O 事件时，内核通过信号通知应用程序，应用程序再发起读取。信号通知与实际数据读取是两个独立步骤。

## 6. 异步 I/O

应用程序提交 I/O 请求后可以继续执行其他任务，由系统完成请求，并通过完成通知或完成队列告知结果。对于读取请求，完成结果说明本次读取已结束，可根据返回的字节数或错误信息处理缓冲区。

## 7. 模型对比

| 模型 | 等待或通知方式 | 数据读取方式 |
| --- | --- | --- |
| 阻塞 I/O | 读取调用等待数据 | 本次读取调用完成数据复制后返回 |
| 非阻塞 I/O | 无数据时立即返回，稍后重试 | 应用程序再次调用读取接口 |
| I/O 多路复用 | 等待多个描述符的就绪事件 | 应用程序收到就绪事件后调用读取接口 |
| 信号驱动 I/O | 通过信号通知 I/O 事件 | 应用程序收到通知后调用读取接口 |
| 异步 I/O | 提交请求后获取完成结果 | 系统执行已提交的读取请求 |

阻塞与非阻塞描述调用在无法立即取得进展时是否等待；就绪通知与完成通知则区分通知发生在哪个阶段。非阻塞 I/O 和 I/O 多路复用都不等同于异步 I/O。

## 8. 实践方向

- 对比阻塞与非阻塞读取的行为。
- 使用 select、poll 和 epoll 编写 I/O 多路复用示例。

[返回 I/O 学习目录](README.md)
