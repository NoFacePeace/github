# 并发与同步

并发程序中，多个执行流可能同时访问共享资源。同步机制用于约束执行顺序、保护共享数据，或协调生产者与消费者之间的节奏。

## 1. 信号量

### 1.1 基本概念

信号量（Semaphore）是一个由内核或线程库维护的计数器，用于表示可用资源数量，或者表示某个事件是否已经发生。

信号量通常包含一个非负整数值，并提供两个原子操作：

- **等待（wait、`P`、`down`）**：信号量值大于 0 时将其减一并继续执行；值为 0 时，调用线程阻塞等待。
- **发布（post、`V`、`up`）**：信号量值加一，并唤醒一个或多个等待线程。

等待和发布操作必须是原子的，否则多个线程可能同时检查到同一个可用资源，导致计数错误或资源超额使用。

### 1.2 信号量的类型

#### 1.2.1 计数信号量

计数信号量的初始值可以大于 1，适合表示多个相同资源。例如，系统中有 8 个连接池槽位，可以将信号量初始化为 8：

```text
初始化 semaphore = 8

线程申请连接：
    wait(semaphore)
    使用一个连接

线程释放连接：
    释放连接
    post(semaphore)
```

只要信号量值大于 0，新的线程就可以获得一个资源；当所有资源都被占用时，后续线程会阻塞。

#### 1.2.2 二值信号量

二值信号量的值通常只有 0 和 1，可以用于表示一个资源是否可用，或者表示一个事件是否完成。

二值信号量在部分场景下可以实现互斥效果，但它与互斥锁并不完全等价：

- 互斥锁通常要求由加锁线程解锁。
- 信号量允许一个线程 `wait`，另一个线程 `post`。
- 互斥锁通常包含线程所有权、递归加锁、优先级继承等语义；信号量主要关心计数和阻塞唤醒。

如果目标是保护临界区，优先使用互斥锁；如果目标是资源计数或线程间事件通知，可以考虑信号量。

### 1.3 POSIX 信号量

POSIX 线程信号量使用 `sem_t` 表示，常用接口如下：

| 接口 | 作用 |
| --- | --- |
| `sem_init` | 初始化进程内信号量 |
| `sem_wait` | 等待并消耗一个信号量计数 |
| `sem_trywait` | 非阻塞地尝试等待 |
| `sem_timedwait` | 在指定超时时间内等待 |
| `sem_post` | 发布一个信号量计数 |
| `sem_destroy` | 销毁进程内信号量 |

下面的示例使用初始值为 `3` 的信号量限制同时访问资源的线程数量：

```c
#include <errno.h>
#include <pthread.h>
#include <semaphore.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

sem_t slots;

void *worker(void *arg) {
    (void)arg;

    if (sem_wait(&slots) == -1) {
        perror("sem_wait");
        return NULL;
    }

    printf("thread %lu acquired a slot\n",
           (unsigned long)pthread_self());
    sleep(1);
    printf("thread %lu released a slot\n",
           (unsigned long)pthread_self());

    if (sem_post(&slots) == -1) {
        perror("sem_post");
    }
    return NULL;
}

int main(void) {
    enum { THREAD_COUNT = 6 };
    pthread_t threads[THREAD_COUNT];

    if (sem_init(&slots, 0, 3) == -1) {
        perror("sem_init");
        return EXIT_FAILURE;
    }

    for (int i = 0; i < THREAD_COUNT; i++) {
        if (pthread_create(&threads[i], NULL, worker, NULL) != 0) {
            perror("pthread_create");
            return EXIT_FAILURE;
        }
    }

    for (int i = 0; i < THREAD_COUNT; i++) {
        pthread_join(threads[i], NULL);
    }

    sem_destroy(&slots);
    return EXIT_SUCCESS;
}
```

`sem_init(&slots, 0, 3)` 中，第二个参数为 `0`，表示该信号量在线程之间共享；第三个参数表示初始计数为 `3`。因此最多有 3 个线程同时持有资源槽位。

### 1.4 信号量实现生产者与消费者

有界缓冲区通常需要两个计数信号量：

- `empty`：记录剩余空槽数量，初始值为缓冲区容量。
- `full`：记录已经放入的元素数量，初始值为 `0`。

同时还需要一个互斥锁保护缓冲区下标、元素内容等共享数据：

```text
生产者：
    wait(empty)
    lock(mutex)
    向缓冲区放入元素
    unlock(mutex)
    post(full)

消费者：
    wait(full)
    lock(mutex)
    从缓冲区取出元素
    unlock(mutex)
    post(empty)
```

这里的顺序很重要：

- 生产者先等待空槽，再进入临界区；缓冲区满时不会继续写入。
- 消费者先等待已有元素，再进入临界区；缓冲区空时不会继续读取。
- 互斥锁只保护实际访问缓冲区的短临界区，不负责表示空槽或已有元素数量。

### 1.5 进程间共享

POSIX 信号量可以用于线程间同步，也可以用于进程间同步：

- `pshared` 为 `0`：信号量在同一进程的线程之间共享。
- `pshared` 为非零值：信号量需要放在多个进程都能访问的共享内存中。

进程间使用时，信号量对象本身不能只放在某个进程的普通私有地址空间中，否则其他进程无法访问同一个对象。可以将其放入 `mmap` 创建的共享内存或其他共享内存区域。

### 1.6 常见错误

- **忘记 `post`**：线程获得资源后异常返回或提前退出，导致其他线程永久阻塞。
- **重复 `post`**：信号量计数超过实际资源数量，可能导致并发访问超出限制。
- **把信号量当作完整的互斥保护**：信号量只能表示计数或事件，复杂共享数据仍需要互斥锁保护。
- **等待顺序不一致**：多个同步对象之间形成循环等待，可能产生死锁。
- **销毁过早**：仍有线程使用信号量时调用 `sem_destroy`，行为未定义。
- **忽略中断错误**：阻塞等待可能因信号中断而失败，实际程序需要根据 `errno == EINTR` 决定是否重试。

### 1.7 与其他同步机制的选择

| 需求 | 更适合的机制 |
| --- | --- |
| 保护共享数据的临界区 | 互斥锁 |
| 表示可用资源数量 | 计数信号量 |
| 通知另一个线程某个事件已发生 | 信号量或条件变量 |
| 等待一组线程到达同一个阶段 | 屏障 |
| 对单个整数执行无锁更新 | 原子操作 |

选择同步机制时，先明确要解决的是资源数量、执行顺序、事件通知还是临界区互斥，再决定使用哪一种原语。
