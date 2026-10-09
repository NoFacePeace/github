---
color: "#5B6CDA"
---

# Java 环境配置

记录 Java 开发环境中的 JDK 路径与命令行配置。

## 1. 配置

### 1.1 JAVA_HOME

`JAVA_HOME` 用于指定 JDK 安装目录，Maven、Gradle 等工具可以通过它定位 JDK。

在 macOS 上，先安装 JDK 17，再执行：

```sh
export JAVA_HOME=$(/usr/libexec/java_home -v 17)
```

| 内容 | 含义 |
| --- | --- |
| `/usr/libexec/java_home -v 17` | 查找系统识别到的 JDK 17，并输出安装路径 |
| `$(...)` | 执行括号内的命令，将输出作为变量的值 |
| `export JAVA_HOME=...` | 设置环境变量，供当前 Shell 及其子进程读取 |

可以通过 `/usr/libexec/java_home -V` 查看系统识别到的已安装 JDK。

### 1.2 PATH

终端执行 `java` 时，通过 `PATH` 查找可执行程序。设置 `JAVA_HOME` 后，还可以将该 JDK 的 `bin` 目录放到 `PATH` 最前面：

```sh
export PATH="$JAVA_HOME/bin:$PATH"
```

以上设置只对当前 Shell 会话及其子进程有效。需要在每次打开终端时自动设置，可以将两条 `export` 命令加入对应的 Shell 配置文件，例如 zsh 的 `~/.zshrc`。

## 2. 示例

### 2.1 切换到 JDK 17 并验证

```sh
export JAVA_HOME=$(/usr/libexec/java_home -v 17)
export PATH="$JAVA_HOME/bin:$PATH"

# 查看 JDK 安装目录
echo "$JAVA_HOME"

# 查看 Java 运行时和编译器版本
java -version
javac -version
```

[返回 Java 目录](Java.md)
