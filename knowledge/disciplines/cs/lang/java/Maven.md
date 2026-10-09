---
color: "#5B6CDA"
---

# Maven

Apache Maven 是 Java 项目常用的构建与依赖管理工具。`mvn` 是它的命令行程序，通过项目中的 `pom.xml` 读取依赖、插件和构建配置。

## 1. 常用命令

| 命令 | 用途 |
| --- | --- |
| `mvn -version` | 查看 Maven 版本及使用的 Java 环境 |
| `mvn clean` | 清理旧构建产物，通常删除 `target/` |
| `mvn compile` | 编译主代码 |
| `mvn test` | 运行单元测试 |
| `mvn package` | 完成构建并打包，通常生成 JAR 或 WAR |
| `mvn install` | 完成构建，并将产物安装到本地 Maven 仓库 |

`compile`、`test`、`package`、`install` 是默认构建生命周期中的阶段，执行后面的阶段会先执行其前面的阶段。`clean` 属于独立的清理生命周期，可以与构建阶段组合使用。

## 2. 示例

### 2.1 清理、构建并跳过测试运行

在包含 `pom.xml` 的项目目录执行：

```sh
mvn clean install -DskipTests
```

| 部分 | 含义 |
| --- | --- |
| `mvn` | 执行 Maven |
| `clean` | 清理上次构建的产物 |
| `install` | 编译、打包，并将 JAR、POM 等产物安装到本地仓库，默认是 `~/.m2/repository/` |
| `-DskipTests` | 跳过测试运行，通常仍会编译测试代码 |

安装后，本机其他项目可以引用该构建版本。`install` 不会将产物发布到远程仓库。

如果连测试代码编译也要跳过，可以执行：

```sh
mvn clean install -Dmaven.test.skip=true
```

上述测试跳过行为适用于 Maven 常用的测试插件；自定义插件或项目配置可能影响实际行为。

[返回 Java 目录](Java.md)
